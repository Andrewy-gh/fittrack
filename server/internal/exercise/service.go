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
	GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error)
	UpdateExerciseName(ctx context.Context, id int32, name, userID string) error
	DeleteExercise(ctx context.Context, id int32, userID string) error
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

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		es.logger.Debug("exercise not found", "exercise_id", id, "user_id", userID, "error", err)
		return nil, &ErrNotFound{Message: "exercise not found"}
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

func (es *ExerciseService) GetRecentSetsForExercise(ctx context.Context, id int32) ([]db.GetRecentSetsForExerciseRow, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &ErrUnauthorized{Message: "user not authenticated"}
	}

	sets, err := es.repo.GetRecentSetsForExercise(ctx, id, userID)
	if err != nil {
		es.logger.Error("failed to get recent sets for exercise", "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "exercise_id", id, "user_id", userID)
		return nil, fmt.Errorf("failed to get recent sets for exercise: %w", err)
	}
	return sets, nil
}

// MARK: UpdateExerciseName
func (es *ExerciseService) UpdateExerciseName(ctx context.Context, id int32, name string) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &ErrUnauthorized{Message: "user not authenticated"}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		es.logger.Debug("exercise not found for update", "exercise_id", id, "user_id", userID, "error", err)
		return &ErrNotFound{Message: "exercise not found"}
	}

	if err := es.repo.UpdateExerciseName(ctx, id, name, userID); err != nil {
		es.logger.Error("failed to update exercise name", "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "exercise_id", id, "user_id", userID)
		return fmt.Errorf("failed to update exercise name: %w", err)
	}

	es.logger.Info("exercise name updated successfully", "exercise_id", id, "user_id", userID)
	return nil
}

// MARK: DeleteExercise
func (es *ExerciseService) DeleteExercise(ctx context.Context, id int32) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &ErrUnauthorized{Message: "user not authenticated"}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		es.logger.Debug("exercise not found for update", "exercise_id", id, "user_id", userID, "error", err)
		return &ErrNotFound{Message: "exercise not found"}
	}

	if err := es.repo.DeleteExercise(ctx, id, userID); err != nil {
		es.logger.Error("failed to delete exercise", "error", err)
		es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "exercise_id", id, "user_id", userID)
		return fmt.Errorf("failed to delete exercise: %w", err)
	}

	es.logger.Info("exercise deleted successfully", "exercise_id", id, "user_id", userID)
	return nil
}
