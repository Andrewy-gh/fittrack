package aichat

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultChatWorkoutLimit = 5
	maxChatWorkoutLimit     = 20
)

type ChatDataReader interface {
	ListWorkoutsWithSets(ctx context.Context, userID string, filter WorkoutHistoryFilter) ([]ChatWorkoutView, error)
	ResolveExerciseNames(ctx context.Context, userID string, query string) ([]string, error)
	TrainingSnapshot(ctx context.Context, userID string) (*TrainingSnapshot, error)
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

func normalizeWorkoutHistoryFilter(filter WorkoutHistoryFilter) WorkoutHistoryFilter {
	if filter.LastN <= 0 {
		filter.LastN = defaultChatWorkoutLimit
	}
	if filter.LastN > maxChatWorkoutLimit {
		filter.LastN = maxChatWorkoutLimit
	}
	return filter
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

func formatSetWeight(weight float64) string {
	text := fmt.Sprintf("%.1f", weight)
	return strings.TrimSuffix(strings.TrimSuffix(text, "0"), ".")
}

var _ ChatDataReader = (*repository)(nil)
