package workout

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type WorkoutRepository interface {
	ListWorkouts(ctx context.Context, userID string) ([]db.Workout, error)
	GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error)
	GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error)
	ListWorkoutFocusValues(ctx context.Context, userID string) ([]string, error)
	GetContributionData(ctx context.Context, userID string) ([]db.GetContributionDataRow, error)
	SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error
	UpdateWorkout(ctx context.Context, id int32, reformatted *ReformattedRequest, userID string) error
	DeleteWorkout(ctx context.Context, id int32, userID string) error
}

type WorkoutService struct {
	logger *slog.Logger
	repo   WorkoutRepository
}

func NewService(logger *slog.Logger, repo WorkoutRepository) *WorkoutService {
	return &WorkoutService{
		logger: logger,
		repo:   repo,
	}
}

func (ws *WorkoutService) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}
	workouts, err := ws.repo.ListWorkouts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}

func (ws *WorkoutService) GetWorkoutWithSets(ctx context.Context, id int32) ([]WorkoutWithSetsResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}
	workoutWithSets, err := ws.repo.GetWorkoutWithSets(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout with sets: %w", err)
	}

	// Convert database rows to response type
	response, err := ws.convertWorkoutWithSetsRows(workoutWithSets)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workout with sets rows: %w", err)
	}

	return response, nil
}

func (ws *WorkoutService) CreateWorkout(ctx context.Context, requestBody CreateWorkoutRequest) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}
	// Transform the request to our internal format
	reformatted, err := ws.transformRequest(requestBody)
	if err != nil {
		return fmt.Errorf("failed to transform request: %w", err)
	}

	// Use repository to save the workout
	if err := ws.repo.SaveWorkout(ctx, reformatted, userID); err != nil {
		return fmt.Errorf("failed to save workout: %w", err)
	}

	return nil
}

// UpdateWorkout updates an existing workout (PUT endpoint)
// Returns 204 No Content on success
func (ws *WorkoutService) UpdateWorkout(ctx context.Context, id int32, req UpdateWorkoutRequest) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}

	// First, validate that the workout exists and belongs to the user
	// This helps provide better error messages (404 vs generic error)
	_, err := ws.repo.GetWorkout(ctx, id, userID)
	if err != nil {
		return &apperrors.NotFound{Resource: "workout", ID: fmt.Sprintf("%d", id)}
	}

	// Transform the request to our internal format (same as CreateWorkout)
	reformatted, err := ws.transformUpdateRequest(req)
	if err != nil {
		return fmt.Errorf("failed to transform update request: %w", err)
	}

	// Perform the update
	if err := ws.repo.UpdateWorkout(ctx, id, reformatted, userID); err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}

	return nil
}

// DeleteWorkout deletes an existing workout (DELETE endpoint)
// Returns 204 No Content on success
func (ws *WorkoutService) DeleteWorkout(ctx context.Context, id int32) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}

	// First, validate that the workout exists and belongs to the user
	// This helps provide better error messages (404 vs generic error)
	_, err := ws.repo.GetWorkout(ctx, id, userID)
	if err != nil {
		return &apperrors.NotFound{Resource: "workout", ID: fmt.Sprintf("%d", id)}
	}

	// Call repository to delete the workout
	if err := ws.repo.DeleteWorkout(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	return nil
}

// ListWorkoutFocusValues retrieves all distinct workout focus values for the authenticated user
func (ws *WorkoutService) ListWorkoutFocusValues(ctx context.Context) ([]string, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}
	focusValues, err := ws.repo.ListWorkoutFocusValues(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workout focus values: %w", err)
	}

	// Ensure we always return an empty slice, not nil
	if focusValues == nil {
		focusValues = []string{}
	}

	return focusValues, nil
}

// GetContributionData retrieves contribution graph data for the past 52 weeks
func (ws *WorkoutService) GetContributionData(ctx context.Context) (*ContributionDataResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "workout", UserID: ""}
	}

	rows, err := ws.repo.GetContributionData(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contribution data: %w", err)
	}

	// Convert rows to ContributionDay slice and calculate levels
	days := ws.convertContributionRows(rows)

	return &ContributionDataResponse{Days: days}, nil
}

// convertContributionRows converts database rows to ContributionDay slice with calculated levels
func (ws *WorkoutService) convertContributionRows(rows []db.GetContributionDataRow) []ContributionDay {
	if len(rows) == 0 {
		return []ContributionDay{}
	}

	counts := make([]int, len(rows))
	for i, row := range rows {
		counts[i] = int(row.Count)
	}

	thresholds := ws.calculateLevelThresholds(counts)

	days := make([]ContributionDay, len(rows))
	for i, row := range rows {
		workouts := ws.parseWorkouts(row.Workouts)

		days[i] = ContributionDay{
			Date:     row.Date.Time.Format("2006-01-02"),
			Count:    int(row.Count),
			Level:    ws.calculateLevel(int(row.Count), thresholds),
			Workouts: workouts,
		}
	}

	return days
}

// calculateLevelThresholds returns thresholds for levels 1-4
// Uses percentiles (25th, 50th, 75th) when 10+ workout days, otherwise static thresholds
func (ws *WorkoutService) calculateLevelThresholds(counts []int) [4]int {
	// Filter out zero counts for threshold calculation
	nonZeroCounts := make([]int, 0, len(counts))
	for _, c := range counts {
		if c > 0 {
			nonZeroCounts = append(nonZeroCounts, c)
		}
	}

	// Use static thresholds if fewer than 30 workout days
	if len(nonZeroCounts) < 30 {
		return [4]int{1, 6, 11, 16} // 0, 1-5, 6-10, 11-15, 16+
	}

	// Sort for percentile calculation
	sorted := make([]int, len(nonZeroCounts))
	copy(sorted, nonZeroCounts)
	sort.Ints(sorted)

	// Calculate percentiles (25th, 50th, 75th)
	p25 := percentile(sorted, 25)
	p50 := percentile(sorted, 50)
	p75 := percentile(sorted, 75)

	// Ensure thresholds are strictly increasing and at least 1
	t1 := max(1, p25)
	t2 := max(t1+1, p50)
	t3 := max(t2+1, p75)
	t4 := t3 + 1

	return [4]int{t1, t2, t3, t4}
}

// percentile calculates the p-th percentile of a sorted slice
func percentile(sorted []int, p int) int {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * (len(sorted) - 1)) / 100
	return sorted[idx]
}

// calculateLevel returns the level (0-4) based on count and thresholds
func (ws *WorkoutService) calculateLevel(count int, thresholds [4]int) int {
	if count == 0 {
		return 0
	}
	if count < thresholds[1] {
		return 1
	}
	if count < thresholds[2] {
		return 2
	}
	if count < thresholds[3] {
		return 3
	}
	return 4
}

// parseWorkouts parses JSON workout data into []WorkoutSummary
func (ws *WorkoutService) parseWorkouts(workoutsJSON []byte) []WorkoutSummary {
	if workoutsJSON == nil || len(workoutsJSON) == 0 {
		return []WorkoutSummary{}
	}

	var workouts []WorkoutSummary
	if err := json.Unmarshal(workoutsJSON, &workouts); err != nil {
		ws.logger.Warn("failed to parse workouts JSON", "error", err)
		return []WorkoutSummary{}
	}

	return workouts
}

// Generic transform function that works with both request types
func transformWorkoutRequest[T WorkoutRequestTransformable](logger *slog.Logger, request T, requireDate bool) (*ReformattedRequest, error) {
	// Parse date
	datePtr := request.GetDate()
	if requireDate && datePtr == nil {
		return nil, fmt.Errorf("date is required")
	}

	var parsedDate time.Time
	var err error
	if datePtr != nil {
		parsedDate, err = time.Parse("2006-01-02T15:04:05Z07:00", *datePtr)
		if err != nil {
			logger.Error("failed to parse date", "error", err)
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
	}

	// Create workout data
	workout := WorkoutData{
		Date:         parsedDate,
		Notes:        request.GetNotes(),
		WorkoutFocus: request.GetWorkoutFocus(),
	}

	// Process exercises and sets
	exerciseMap := make(map[string]bool)
	var exercises []ExerciseData
	var sets []SetData

	for _, exercise := range request.GetExercises() {
		if !exerciseMap[exercise.GetName()] {
			exerciseMap[exercise.GetName()] = true
			exercises = append(exercises, ExerciseData{
				Name: exercise.GetName(),
			})
		}

		for _, set := range exercise.GetSets() {
			sets = append(sets, SetData{
				ExerciseName: exercise.GetName(),
				Weight:       set.GetWeight(),
				Reps:         set.GetReps(),
				SetType:      set.GetSetType(),
			})
		}
	}

	return &ReformattedRequest{
		Workout:   workout,
		Exercises: exercises,
		Sets:      sets,
	}, nil
}

// Convenience wrappers
func (ws *WorkoutService) transformRequest(request CreateWorkoutRequest) (*ReformattedRequest, error) {
	return transformWorkoutRequest(ws.logger, request, false) // Date is required in struct, not optional
}

func (ws *WorkoutService) transformUpdateRequest(request UpdateWorkoutRequest) (*ReformattedRequest, error) {
	return transformWorkoutRequest(ws.logger, request, false) // Date is optional for updates (partial updates allowed)
}

// convertWorkoutWithSetsRows converts database rows to response type, handling pgtype.Numeric to float64 conversion
func (ws *WorkoutService) convertWorkoutWithSetsRows(rows []db.GetWorkoutWithSetsRow) ([]WorkoutWithSetsResponse, error) {
	response := make([]WorkoutWithSetsResponse, len(rows))

	for i, row := range rows {
		var weight *float64
		if row.Weight.Valid {
			f64, err := row.Weight.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("failed to convert weight: %w", err)
			}
			weight = &f64.Float64
		}

		var volume float64
		if row.Volume.Valid {
			f64, err := row.Volume.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("failed to convert volume: %w", err)
			}
			volume = f64.Float64
		}

		var exerciseOrder *int32
		if row.ExerciseOrder != 0 {
			exerciseOrder = &row.ExerciseOrder
		}

		var setOrder *int32
		if row.SetOrder != 0 {
			setOrder = &row.SetOrder
		}

		var workoutNotes *string
		if row.WorkoutNotes.Valid {
			workoutNotes = &row.WorkoutNotes.String
		}

		var workoutFocus *string
		if row.WorkoutFocus.Valid {
			workoutFocus = &row.WorkoutFocus.String
		}

		response[i] = WorkoutWithSetsResponse{
			WorkoutID:     row.WorkoutID,
			WorkoutDate:   row.WorkoutDate.Time,
			WorkoutNotes:  workoutNotes,
			WorkoutFocus:  workoutFocus,
			SetID:         row.SetID,
			Weight:        weight,
			Reps:          row.Reps,
			SetType:       row.SetType,
			ExerciseID:    row.ExerciseID,
			ExerciseName:  row.ExerciseName,
			ExerciseOrder: exerciseOrder,
			SetOrder:      setOrder,
			Volume:        volume,
		}
	}

	return response, nil
}
