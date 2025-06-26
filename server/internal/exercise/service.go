package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// ExerciseRepository is an interface for exercise data access
type ExerciseRepository interface {
	ListExercises(ctx context.Context) ([]db.Exercise, error)
	GetExercise(ctx context.Context, id int32) (db.Exercise, error)
	GetOrCreateExercise(ctx context.Context, name string) (db.Exercise, error)
}

// ExerciseService handles exercise business logic
type ExerciseService struct {
	logger *slog.Logger
	repo   ExerciseRepository
}

// NewService creates a new ExerciseService
func NewService(logger *slog.Logger, repo ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		logger: logger,
		repo:   repo,
	}
}

// ListExercises retrieves all exercises
func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	exercises, err := es.repo.ListExercises(ctx)
	if err != nil {
		es.logger.Error("failed to list exercises", "error", err)
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

// GetExercise retrieves a single exercise by ID
func (es *ExerciseService) GetExercise(ctx context.Context, id int32) (db.Exercise, error) {
	exercise, err := es.repo.GetExercise(ctx, id)
	if err != nil {
		es.logger.Error("failed to get exercise", "id", id, "error", err)
		return db.Exercise{}, fmt.Errorf("failed to get exercise: %w", err)
	}
	return exercise, nil
}

// GetOrCreateExercise gets an existing exercise by name or creates a new one if it doesn't exist
func (es *ExerciseService) GetOrCreateExercise(ctx context.Context, name string) (db.Exercise, error) {
	exercise, err := es.repo.GetOrCreateExercise(ctx, name)
	if err != nil {
		es.logger.Error("repository failed to get or create exercise", "exercise_name", name, "error", err)
		return exercise, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return exercise, nil
}
