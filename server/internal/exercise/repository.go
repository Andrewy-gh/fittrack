package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type exerciseRepository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewRepository creates a new instance of ExerciseRepository
func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) ExerciseRepository {
	return &exerciseRepository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (er *exerciseRepository) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	exercises, err := er.queries.ListExercises(ctx)
	if err != nil {
		er.logger.Error("failed to list exercises", "error", err)
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (er *exerciseRepository) GetExercise(ctx context.Context, id int32) (db.Exercise, error) {
	exercise, err := er.queries.GetExercise(ctx, id)
	if err != nil {
		er.logger.Error("failed to get exercise", "error", err)
		return db.Exercise{}, fmt.Errorf("failed to get exercise: %w", err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetOrCreateExercise(ctx context.Context, name string) (db.Exercise, error) {
	exercise, err := er.queries.GetOrCreateExercise(ctx, name)
	if err != nil {
		er.logger.Error("failed to get or create exercise", "error", err)
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetExerciseWithSets(ctx context.Context, id int32) (db.GetExerciseWithSetsRow, error) {
	exerciseWithSets, err := er.queries.GetExerciseWithSets(ctx, id)
	if err != nil {
		er.logger.Error("failed to get exercise with sets", "exercise_id", id, "error", err)
		return db.GetExerciseWithSetsRow{}, fmt.Errorf("failed to get exercise with sets: %w", err)
	}
	return exerciseWithSets, nil
}

// func (er *exerciseRepository) CreateExercise(ctx context.Context, name string) (db.Exercise, error) {
// 	exercise, err := er.queries.CreateExercise(ctx, name)
// 	if err != nil {
// 		return db.Exercise{}, fmt.Errorf("failed to create exercise: %w", err)
// 	}
// 	return exercise, nil
// }

// func (er *exerciseRepository) UpdateExercise(ctx context.Context, id int32, name string) (db.Exercise, error) {
// 	exercise, err := er.queries.UpdateExercise(ctx, db.UpdateExerciseParams{
// 		ID:   id,
// 		Name: name,
// 	})
// 	if err != nil {
// 		return db.Exercise{}, fmt.Errorf("failed to update exercise: %w", err)
// 	}
// 	return exercise, nil
// }

// func (er *exerciseRepository) DeleteExercise(ctx context.Context, id int32) error {
// 	err := er.queries.DeleteExercise(ctx, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete exercise: %w", err)
// 	}
// 	return nil
// }
