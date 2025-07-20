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
	GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error)
	SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error
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

	// Convert to PG types

	// Use repository to save the workout
	if err := ws.repo.SaveWorkout(ctx, reformatted, userID); err != nil {
		ws.logger.Error("failed to save workout", "error", err)
		return fmt.Errorf("failed to save workout: %w", err)
	}

	return nil
}

func (ws *WorkoutService) transformRequest(request CreateWorkoutRequest) (*ReformattedRequest, error) {
	// Parse date
	parsedDate, err := time.Parse("2006-01-02T15:04:05Z07:00", request.Date)
	if err != nil {
		ws.logger.Error("failed to parse date", "error", err)
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create workout data
	workout := WorkoutData{
		Date:  parsedDate,
		Notes: request.Notes,
	}

	// Process exercises and sets (your existing logic)
	exerciseMap := make(map[string]bool)
	var exercises []ExerciseData
	var sets []SetData

	for _, exercise := range request.Exercises {
		if !exerciseMap[exercise.Name] {
			exerciseMap[exercise.Name] = true
			exercises = append(exercises, ExerciseData{
				Name: exercise.Name,
			})
		}

		for _, set := range exercise.Sets {
			sets = append(sets, SetData{
				ExerciseName: exercise.Name,
				Weight:       set.Weight,
				Reps:         set.Reps,
				SetType:      set.SetType,
			})
		}
	}

	return &ReformattedRequest{
		Workout:   workout,
		Exercises: exercises,
		Sets:      sets,
	}, nil
}
