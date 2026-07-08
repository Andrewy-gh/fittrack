package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultChatWorkoutLimit     = 5
	maxChatWorkoutLimit         = 20
	defaultExerciseStatsWindow  = "3m"
	maxExerciseStatsTrendPoints = 8
)

type ChatDataReader interface {
	ListWorkoutsWithSets(ctx context.Context, userID string, filter WorkoutHistoryFilter) ([]ChatWorkoutView, error)
	ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error)
	ExerciseStats(ctx context.Context, userID string, exerciseName string, window string) (*ExerciseStatsView, error)
	TrainingSnapshot(ctx context.Context, userID string) (*TrainingSnapshot, error)
	TrainingProfile(ctx context.Context, userID string) (*TrainingProfile, error)
	UpdateTrainingProfile(ctx context.Context, userID string, update TrainingProfileUpdate) (*TrainingProfile, error)
}

func (r *repository) ListWorkoutsWithSets(ctx context.Context, userID string, filter WorkoutHistoryFilter) ([]ChatWorkoutView, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	filter = normalizeWorkoutHistoryFilter(filter)
	rows, err := r.queries.ListWorkoutsWithSetsForChat(ctx, db.ListWorkoutsWithSetsForChatParams{
		UserID:       userID,
		StartDate:    timePtrToPg(filter.StartDate),
		EndDate:      timePtrToPg(filter.EndDate),
		ExerciseName: textToPg(strings.TrimSpace(filter.ExerciseName)),
		WorkoutFocus: textToPg(strings.TrimSpace(filter.WorkoutFocus)),
		RowLimit:     int32(filter.LastN),
	})
	if err != nil {
		return nil, fmt.Errorf("list workouts with sets for ai chat: %w", err)
	}

	workouts, err := mapChatWorkoutRows(rows)
	if err != nil {
		return nil, err
	}
	return workouts, nil
}

func (r *repository) ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := r.queries.ListExerciseNameMatches(ctx, db.ListExerciseNameMatchesParams{
		UserID:    userID,
		NameQuery: strings.TrimSpace(query),
	})
	if err != nil {
		return nil, fmt.Errorf("list exercise name matches for ai chat: %w", err)
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names, nil
}

func (r *repository) TrainingSnapshot(ctx context.Context, userID string) (*TrainingSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	stats, err := r.queries.GetChatWorkoutSnapshotStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get ai chat workout snapshot stats: %w", err)
	}
	topRows, err := r.queries.ListTopExercisesByFrequency(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list top exercises by frequency for ai chat: %w", err)
	}

	snapshot := &TrainingSnapshot{
		WorkoutsLast30D: stats.WorkoutsLast30d,
		TopExercises:    make([]string, 0, len(topRows)),
	}
	if stats.LastWorkoutDate.Valid {
		snapshot.LastWorkoutDate = stats.LastWorkoutDate.Time.Format("2006-01-02")
	}
	for _, row := range topRows {
		snapshot.TopExercises = append(snapshot.TopExercises, row.Name)
	}

	return snapshot, nil
}

func (r *repository) TrainingProfile(ctx context.Context, userID string) (*TrainingProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	row, err := r.queries.GetUserTrainingProfile(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get ai chat user training profile: %w", err)
	}

	profile, err := trainingProfileFromRow(row)
	if err != nil {
		return nil, err
	}

	if !hasTrainingProfileContent(profile) {
		return nil, nil
	}
	return profile, nil
}

func (r *repository) UpdateTrainingProfile(ctx context.Context, userID string, update TrainingProfileUpdate) (*TrainingProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.UpsertUserTrainingProfileForChatParams{
		UserID:                          userID,
		PrimaryGoal:                     optionalProfileText(update.PrimaryGoal),
		ExperienceLevel:                 optionalProfileText(update.ExperienceLevel),
		PreferredSessionDurationMinutes: optionalProfileInt(update.PreferredSessionDurationMinutes),
		UsualTrainingLocation:           optionalProfileText(update.UsualTrainingLocation),
		SourceConversationID:            optionalProfileInt(update.SourceConversationID),
		SourceMessageID:                 optionalProfileInt(update.SourceMessageID),
	}

	var err error
	params.AvailableEquipment, err = optionalProfileStringArray(update.AvailableEquipment)
	if err != nil {
		return nil, fmt.Errorf("encode profile available equipment: %w", err)
	}
	params.AvoidedExercises, err = optionalProfileStringArray(update.AvoidedExercises)
	if err != nil {
		return nil, fmt.Errorf("encode profile avoided exercises: %w", err)
	}
	params.MovementLimitations, err = optionalProfileStringArray(update.MovementLimitations)
	if err != nil {
		return nil, fmt.Errorf("encode profile movement limitations: %w", err)
	}

	row, err := r.queries.UpsertUserTrainingProfileForChat(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("upsert ai chat user training profile: %w", err)
	}
	return trainingProfileFromRow(row)
}

func (r *repository) ExerciseStats(ctx context.Context, userID string, exerciseName string, window string) (*ExerciseStatsView, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	window = normalizeExerciseStatsWindow(window)
	exerciseName = strings.TrimSpace(exerciseName)
	stats := &ExerciseStatsView{
		ExerciseName: exerciseName,
		Window:       window,
	}
	if exerciseName == "" {
		stats.Message = "No exercise name was provided."
		return stats, nil
	}

	exerciseRow, err := r.queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{
		Name:   exerciseName,
		UserID: userID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			stats.Message = fmt.Sprintf("No exercise named %q exists for this user.", exerciseName)
			return stats, nil
		}
		return nil, fmt.Errorf("get exercise for ai chat stats: %w", err)
	}
	stats.ExerciseName = exerciseRow.Name

	bestRow, err := r.queries.GetExerciseBestE1rmWithWorkout(ctx, db.GetExerciseBestE1rmWithWorkoutParams{
		UserID:     userID,
		ExerciseID: exerciseRow.ID,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("get exercise best e1rm for ai chat stats: %w", err)
	}
	if err == nil && bestRow.E1rm.Valid {
		best, err := bestRow.E1rm.Float64Value()
		if err != nil {
			return nil, fmt.Errorf("convert ai chat exercise best e1rm: %w", err)
		}
		stats.BestE1RM = &ExerciseBestE1RMView{Weight: best.Float64}
		if workoutDate, err := r.workoutDate(ctx, userID, bestRow.WorkoutID); err == nil {
			stats.BestE1RM.Date = workoutDate
		}
	}

	points, err := r.exerciseStatsTrendRows(ctx, userID, exerciseRow.ID, window)
	if err != nil {
		return nil, err
	}
	stats.SessionCount = len(points)
	stats.Trend = compactExerciseStatsTrend(points, maxExerciseStatsTrendPoints)

	recentSets, err := r.queries.GetRecentSetsForExercise(ctx, db.GetRecentSetsForExerciseParams{
		ExerciseID: exerciseRow.ID,
		UserID:     userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get recent sets for ai chat exercise stats: %w", err)
	}
	if len(recentSets) > 0 {
		latestWorkoutID := recentSets[0].WorkoutID
		if recentSets[0].WorkoutDate.Valid {
			stats.LastSessionDate = recentSets[0].WorkoutDate.Time.Format("2006-01-02")
		}
		for _, row := range recentSets {
			if row.WorkoutID != latestWorkoutID {
				continue
			}
			setText, err := formatRecentExerciseStatSet(row.Weight, row.Reps)
			if err != nil {
				return nil, err
			}
			stats.LastSessionSets = append(stats.LastSessionSets, setText)
		}
	}

	if stats.BestE1RM == nil && stats.SessionCount == 0 && len(stats.LastSessionSets) == 0 {
		stats.Message = fmt.Sprintf("No working-set stats were found for %s.", stats.ExerciseName)
	}
	return stats, nil
}

func pgTextString(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func optionalProfileText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: strings.TrimSpace(*value), Valid: true}
}

func optionalProfileInt(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func optionalProfileStringArray(value *[]string) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	cleaned := cleanProfileStringList(*value)
	raw, err := json.Marshal(cleaned)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func trainingProfileFromRow(row db.UserTrainingProfile) (*TrainingProfile, error) {
	profile := &TrainingProfile{
		PrimaryGoal:           pgTextString(row.PrimaryGoal),
		ExperienceLevel:       pgTextString(row.ExperienceLevel),
		UsualTrainingLocation: pgTextString(row.UsualTrainingLocation),
	}
	if row.PreferredSessionDurationMinutes.Valid {
		profile.PreferredSessionDurationMinutes = row.PreferredSessionDurationMinutes.Int32
	}

	var err error
	profile.AvailableEquipment, err = decodeProfileStringArray(row.AvailableEquipment)
	if err != nil {
		return nil, fmt.Errorf("decode profile available equipment: %w", err)
	}
	profile.AvoidedExercises, err = decodeProfileStringArray(row.AvoidedExercises)
	if err != nil {
		return nil, fmt.Errorf("decode profile avoided exercises: %w", err)
	}
	profile.MovementLimitations, err = decodeProfileStringArray(row.MovementLimitations)
	if err != nil {
		return nil, fmt.Errorf("decode profile movement limitations: %w", err)
	}
	profile.MovementLimitationsRecorded = row.MovementLimitations != nil
	return profile, nil
}

func decodeProfileStringArray(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	cleaned := make([]string, 0, len(values))
	return append(cleaned, cleanProfileStringList(values)...), nil
}

func cleanProfileStringList(values []string) []string {
	const maxProfileListItems = 20
	cleaned := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if len(trimmed) > 120 {
			trimmed = trimmed[:120]
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, trimmed)
		if len(cleaned) == maxProfileListItems {
			break
		}
	}
	return cleaned
}

func hasTrainingProfileContent(profile *TrainingProfile) bool {
	if profile == nil {
		return false
	}
	return strings.TrimSpace(profile.PrimaryGoal) != "" ||
		strings.TrimSpace(profile.ExperienceLevel) != "" ||
		profile.PreferredSessionDurationMinutes > 0 ||
		strings.TrimSpace(profile.UsualTrainingLocation) != "" ||
		len(profile.AvailableEquipment) > 0 ||
		len(profile.AvoidedExercises) > 0 ||
		profile.MovementLimitationsRecorded
}

func (r *repository) workoutDate(ctx context.Context, userID string, workoutID int32) (string, error) {
	row, err := r.queries.GetWorkout(ctx, db.GetWorkoutParams{ID: workoutID, UserID: userID})
	if err != nil {
		return "", err
	}
	if !row.Date.Valid {
		return "", nil
	}
	return row.Date.Time.Format("2006-01-02"), nil
}

type exerciseStatsTrendRow struct {
	WorkoutID          int32
	WorkoutDay         time.Time
	SessionBestE1RM    float64
	SessionAvgE1RM     float64
	TotalVolumeWorking float64
}

func (r *repository) exerciseStatsTrendRows(ctx context.Context, userID string, exerciseID int32, window string) ([]exerciseStatsTrendRow, error) {
	switch normalizeExerciseStatsWindow(window) {
	case "1y":
		rows, err := r.queries.GetExerciseMetricsHistoryRawYear(ctx, db.GetExerciseMetricsHistoryRawYearParams{ExerciseID: exerciseID, UserID: userID})
		if err != nil {
			return nil, fmt.Errorf("get exercise metrics history year for ai chat stats: %w", err)
		}
		return mapExerciseMetricsYearRows(rows), nil
	case "all":
		rows, err := r.queries.GetExerciseMetricsHistoryRawAll(ctx, db.GetExerciseMetricsHistoryRawAllParams{ExerciseID: exerciseID, UserID: userID})
		if err != nil {
			return nil, fmt.Errorf("get exercise metrics history all for ai chat stats: %w", err)
		}
		return mapExerciseMetricsAllRows(rows), nil
	default:
		rows, err := r.queries.GetExerciseMetricsHistoryRaw6M(ctx, db.GetExerciseMetricsHistoryRaw6MParams{ExerciseID: exerciseID, UserID: userID})
		if err != nil {
			return nil, fmt.Errorf("get exercise metrics history 6m for ai chat stats: %w", err)
		}
		return filterLastThreeMonths(mapExerciseMetrics6MRows(rows)), nil
	}
}

func normalizeWorkoutHistoryFilter(filter WorkoutHistoryFilter) WorkoutHistoryFilter {
	if filter.LastN <= 0 {
		filter.LastN = defaultChatWorkoutLimit
	}
	if filter.LastN > maxChatWorkoutLimit {
		filter.LastN = maxChatWorkoutLimit
	}
	return filter
}

func normalizeExerciseStatsWindow(window string) string {
	switch strings.ToLower(strings.TrimSpace(window)) {
	case "3m", "3mo", "3 months", "three months":
		return "3m"
	case "1y", "1yr", "1 year", "year":
		return "1y"
	case "all", "all-time", "all time", "lifetime":
		return "all"
	default:
		return defaultExerciseStatsWindow
	}
}

func mapChatWorkoutRows(rows []db.ListWorkoutsWithSetsForChatRow) ([]ChatWorkoutView, error) {
	workouts := make([]ChatWorkoutView, 0)
	workoutIndexes := make(map[int32]int)
	exerciseIndexes := make(map[int32]map[string]int)

	for _, row := range rows {
		workoutIndex, ok := workoutIndexes[row.WorkoutID]
		if !ok {
			workout := ChatWorkoutView{
				Date: row.Date.Time.Format("2006-01-02"),
			}
			if row.WorkoutFocus.Valid {
				workout.Focus = row.WorkoutFocus.String
			}
			if row.Notes.Valid {
				workout.Notes = row.Notes.String
			}
			workoutIndexes[row.WorkoutID] = len(workouts)
			exerciseIndexes[row.WorkoutID] = make(map[string]int)
			workouts = append(workouts, workout)
			workoutIndex = len(workouts) - 1
		}

		if !row.ExerciseName.Valid {
			continue
		}
		exerciseName := row.ExerciseName.String
		perWorkout := exerciseIndexes[row.WorkoutID]
		exerciseIndex, ok := perWorkout[exerciseName]
		if !ok {
			workouts[workoutIndex].Exercises = append(workouts[workoutIndex].Exercises, ChatExerciseView{Name: exerciseName})
			exerciseIndex = len(workouts[workoutIndex].Exercises) - 1
			perWorkout[exerciseName] = exerciseIndex
		}

		if row.Reps.Valid && row.SetType.Valid {
			setText, err := formatChatSet(row.Weight, row.Reps.Int32, row.SetType.String)
			if err != nil {
				return nil, err
			}
			workouts[workoutIndex].Exercises[exerciseIndex].Sets = append(workouts[workoutIndex].Exercises[exerciseIndex].Sets, setText)
		}
	}

	return workouts, nil
}

func mapExerciseMetrics6MRows(rows []db.GetExerciseMetricsHistoryRaw6MRow) []exerciseStatsTrendRow {
	points := make([]exerciseStatsTrendRow, 0, len(rows))
	for _, row := range rows {
		if !row.WorkoutDay.Valid {
			continue
		}
		points = append(points, exerciseStatsTrendRow{
			WorkoutID:          row.WorkoutID,
			WorkoutDay:         row.WorkoutDay.Time,
			SessionBestE1RM:    row.SessionBestE1rm,
			SessionAvgE1RM:     row.SessionAvgE1rm,
			TotalVolumeWorking: row.TotalVolumeWorking,
		})
	}
	return points
}

func mapExerciseMetricsYearRows(rows []db.GetExerciseMetricsHistoryRawYearRow) []exerciseStatsTrendRow {
	points := make([]exerciseStatsTrendRow, 0, len(rows))
	for _, row := range rows {
		if !row.WorkoutDay.Valid {
			continue
		}
		points = append(points, exerciseStatsTrendRow{
			WorkoutID:          row.WorkoutID,
			WorkoutDay:         row.WorkoutDay.Time,
			SessionBestE1RM:    row.SessionBestE1rm,
			SessionAvgE1RM:     row.SessionAvgE1rm,
			TotalVolumeWorking: row.TotalVolumeWorking,
		})
	}
	return points
}

func mapExerciseMetricsAllRows(rows []db.GetExerciseMetricsHistoryRawAllRow) []exerciseStatsTrendRow {
	points := make([]exerciseStatsTrendRow, 0, len(rows))
	for _, row := range rows {
		if !row.WorkoutDay.Valid {
			continue
		}
		points = append(points, exerciseStatsTrendRow{
			WorkoutID:          row.WorkoutID,
			WorkoutDay:         row.WorkoutDay.Time,
			SessionBestE1RM:    row.SessionBestE1rm,
			SessionAvgE1RM:     row.SessionAvgE1rm,
			TotalVolumeWorking: row.TotalVolumeWorking,
		})
	}
	return points
}

func filterLastThreeMonths(points []exerciseStatsTrendRow) []exerciseStatsTrendRow {
	if len(points) == 0 {
		return nil
	}
	cutoff := points[len(points)-1].WorkoutDay.AddDate(0, -3, 0)
	filtered := make([]exerciseStatsTrendRow, 0, len(points))
	for _, point := range points {
		if !point.WorkoutDay.Before(cutoff) {
			filtered = append(filtered, point)
		}
	}
	return filtered
}

func compactExerciseStatsTrend(points []exerciseStatsTrendRow, maxPoints int) []ExerciseStatsTrendPoint {
	if maxPoints <= 0 || len(points) == 0 {
		return nil
	}
	if len(points) <= maxPoints {
		return mapExerciseStatsTrendPoints(points)
	}

	compact := make([]exerciseStatsTrendRow, 0, maxPoints)
	lastIndex := len(points) - 1
	for i := 0; i < maxPoints; i++ {
		index := i * lastIndex / (maxPoints - 1)
		compact = append(compact, points[index])
	}
	return mapExerciseStatsTrendPoints(compact)
}

func mapExerciseStatsTrendPoints(points []exerciseStatsTrendRow) []ExerciseStatsTrendPoint {
	mapped := make([]ExerciseStatsTrendPoint, 0, len(points))
	for _, point := range points {
		mapped = append(mapped, ExerciseStatsTrendPoint{
			Date:      point.WorkoutDay.Format("2006-01-02"),
			BestE1RM:  point.SessionBestE1RM,
			AvgE1RM:   point.SessionAvgE1RM,
			Volume:    point.TotalVolumeWorking,
			WorkoutID: point.WorkoutID,
		})
	}
	return mapped
}

func formatChatSet(weight pgtype.Numeric, reps int32, setType string) (string, error) {
	setType = strings.TrimSpace(setType)
	if weight.Valid {
		f64, err := weight.Float64Value()
		if err != nil {
			return "", fmt.Errorf("convert ai chat workout set weight: %w", err)
		}
		return fmt.Sprintf("%sx%d %s", formatSetWeight(f64.Float64), reps, setType), nil
	}
	return fmt.Sprintf("%d reps %s", reps, setType), nil
}

func formatRecentExerciseStatSet(weight pgtype.Numeric, reps int32) (string, error) {
	if weight.Valid {
		f64, err := weight.Float64Value()
		if err != nil {
			return "", fmt.Errorf("convert ai chat exercise stat set weight: %w", err)
		}
		return fmt.Sprintf("%sx%d", formatSetWeight(f64.Float64), reps), nil
	}
	return fmt.Sprintf("%d reps", reps), nil
}

func formatSetWeight(weight float64) string {
	text := fmt.Sprintf("%.1f", weight)
	return strings.TrimSuffix(strings.TrimSuffix(text, "0"), ".")
}

var _ ChatDataReader = (*repository)(nil)
