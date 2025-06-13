package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
)

// ExerciseService handles exercise business logic
type ExerciseService struct {
	logger  *slog.Logger
	queries *db.Queries
}

func NewService(logger *slog.Logger, queries *db.Queries) *ExerciseService {
	return &ExerciseService{
		logger:  logger,
		queries: queries,
	}
}

func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	exercises, err := es.queries.ListExercises(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (es *ExerciseService) GetExercise(ctx context.Context, id int32) (db.Exercise, error) {
	exercise, err := es.queries.GetExercise(ctx, id)
	if err != nil {
		return exercise, fmt.Errorf("failed to get exercise: %w", err)
	}
	return exercise, nil
}

func (es *ExerciseService) GetOrCreateExercise(ctx context.Context, name string) (db.Exercise, error) {
	exercise, err := es.queries.GetOrCreateExercise(ctx, name)
	if err != nil {
		return exercise, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return exercise, nil
}
