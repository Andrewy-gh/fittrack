package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/helpers"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
)

// ExerciseRepository is an interface for exercise data access
type ExerciseRepository interface {
	ListExercises(ctx context.Context, userID pgtype.Text) ([]db.Exercise, error)
	GetExercise(ctx context.Context, id int32) (db.Exercise, error)
	GetOrCreateExercise(ctx context.Context, name string) (db.Exercise, error)
	GetExerciseWithSets(ctx context.Context, id int32) ([]db.GetExerciseWithSetsRow, error)
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

// ListExercises retrieves all exercises for the authenticated user
func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, fmt.Errorf("user not authenticated")
	}

	pgUserID := helpers.ToPgText(userID)

	exercises, err := es.repo.ListExercises(ctx, pgUserID)
	if err != nil {
		es.logger.Error("service failed to list exercises", "error", err)
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (es *ExerciseService) GetExerciseWithSets(ctx context.Context, id int32) ([]db.GetExerciseWithSetsRow, error) {
	sets, err := es.repo.GetExerciseWithSets(ctx, id)
	if err != nil {
		es.logger.Error("service failed to get exercise with sets", "exercise_id", id, "error", err)
		return nil, fmt.Errorf("failed to get exercise with sets: %w", err)
	}
	return sets, nil
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
