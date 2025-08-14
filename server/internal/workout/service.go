package workout

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type WorkoutRepository interface {
	ListWorkouts(ctx context.Context, userID string) ([]db.Workout, error)
	GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error)
	GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error)
	SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error
	UpdateWorkout(ctx context.Context, id int32, reformatted *ReformattedRequest, userID string) error
}

type WorkoutService struct {
	logger *slog.Logger
	repo   WorkoutRepository
}

type ErrUnauthorized struct {
	Message string
}

func (e *ErrUnauthorized) Error() string {
	return e.Message
}

type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return e.Message
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
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}
	workouts, err := ws.repo.ListWorkouts(ctx, userID)
	if err != nil {
		ws.logger.Error("failed to list workouts", "error", err)
		ws.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}

func (ws *WorkoutService) GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}
	workoutWithSets, err := ws.repo.GetWorkoutWithSets(ctx, id, userID)
	if err != nil {
		ws.logger.Error("failed to get workout with sets", "error", err)
		ws.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "workout_id", id, "user_id", userID)
		return nil, fmt.Errorf("failed to get workout with sets: %w", err)
	}
	return workoutWithSets, nil
}

func (ws *WorkoutService) CreateWorkout(ctx context.Context, requestBody CreateWorkoutRequest) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &ErrUnauthorized{Message: "user not authenticated"}
	}
	// Transform the request to our internal format
	reformatted, err := ws.transformRequest(requestBody)
	if err != nil {
		ws.logger.Error("failed to transform request", "error", err)
		return fmt.Errorf("failed to transform request: %w", err)
	}

	// Use repository to save the workout
	if err := ws.repo.SaveWorkout(ctx, reformatted, userID); err != nil {
		ws.logger.Error("failed to save workout", "error", err)
		ws.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
		return fmt.Errorf("failed to save workout: %w", err)
	}

	return nil
}

// UpdateWorkout updates an existing workout (PUT endpoint)
// Returns 204 No Content on success
func (ws *WorkoutService) UpdateWorkout(ctx context.Context, id int32, req UpdateWorkoutRequest) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &ErrUnauthorized{Message: "user not authenticated"}
	}

	// First, validate that the workout exists and belongs to the user
	// This helps provide better error messages (404 vs generic error)
	_, err := ws.repo.GetWorkout(ctx, id, userID)
	if err != nil {
		ws.logger.Debug("workout not found for update", "workout_id", id, "user_id", userID, "error", err)
		return &ErrNotFound{Message: "workout not found"}
	}

	// Transform the request to our internal format (same as CreateWorkout)
	reformatted, err := ws.transformUpdateRequest(req)
	if err != nil {
		ws.logger.Error("failed to transform update request", "error", err, "workout_id", id, "user_id", userID)
		return fmt.Errorf("failed to transform update request: %w", err)
	}

	// Perform the update
	if err := ws.repo.UpdateWorkout(ctx, id, reformatted, userID); err != nil {
		ws.logger.Error("failed to update workout", "error", err, "workout_id", id, "user_id", userID)
		ws.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err))
		return fmt.Errorf("failed to update workout: %w", err)
	}

	ws.logger.Info("workout updated successfully", "workout_id", id, "user_id", userID)
	return nil
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
		Date:  parsedDate,
		Notes: request.GetNotes(),
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
