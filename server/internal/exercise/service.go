package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type ExerciseRepository interface {
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error)
	GetOrCreateExercise(ctx context.Context, name, userID string) (db.Exercise, error)
	GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name, userID string) (db.Exercise, error)
	GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error)
}

type ErrUnauthorized struct {
	Message string
}

func (e *ErrUnauthorized) Error() string {
	return e.Message
}

type ExerciseService struct {
	logger *slog.Logger
	repo   ExerciseRepository
}

func NewService(logger *slog.Logger, repo ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		logger: logger,
		repo:   repo,
	}
}

func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}

	exercises, err := es.repo.ListExercises(ctx, userID)
	if err != nil {
		es.logger.Error("failed to list exercises", "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (es *ExerciseService) GetExerciseWithSets(ctx context.Context, id int32) ([]db.GetExerciseWithSetsRow, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}

	sets, err := es.repo.GetExerciseWithSets(ctx, id, userID)
	if err != nil {
		es.logger.Error("failed to get exercise with sets", "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "exercise_id", id, "user_id", userID)
		return nil, fmt.Errorf("failed to get exercise with sets: %w", err)
	}
	return sets, nil
}

func (es *ExerciseService) GetOrCreateExercise(ctx context.Context, name string) (*db.Exercise, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}

	exercise, err := es.repo.GetOrCreateExercise(ctx, name, userID)
	if err != nil {
		es.logger.Error("repository failed to get or create exercise", "exercise_name", name, "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "exercise_name", name, "user_id", userID)
		return nil, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return &exercise, nil
}
