package exercise

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
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

func (er *exerciseRepository) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	pgUserID := pgtype.Text{String: userID, Valid: true}
	exercises, err := er.queries.ListExercises(ctx, pgUserID)
	if err != nil {
		er.logger.Error("failed to list exercises", "error", err)
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (er *exerciseRepository) GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error) {
	params := db.GetExerciseParams{
		ID:     id,
		UserID: pgtype.Text{String: userID, Valid: true},
	}
	exercise, err := er.queries.GetExercise(ctx, params)
	if err != nil {
		er.logger.Error("failed to get exercise", "id", id, "error", err)
		return db.Exercise{}, fmt.Errorf("failed to get exercise: %w", err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetOrCreateExercise(ctx context.Context, name string, userID string) (db.Exercise, error) {
	params := db.GetOrCreateExerciseParams{
		Name:   name,
		UserID: pgtype.Text{String: userID, Valid: true},
	}
	exercise, err := er.queries.GetOrCreateExercise(ctx, params)
	if err != nil {
		er.logger.Error("failed to get or create exercise", "error", err)
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error) {
	params := db.GetExerciseWithSetsParams{
		ExerciseID: id,
		UserID:     pgtype.Text{String: userID, Valid: true},
	}
	sets, err := er.queries.GetExerciseWithSets(ctx, params)
	if err != nil {
		er.logger.Error("failed to get exercise with sets", "exercise_id", id, "error", err)
		return nil, fmt.Errorf("failed to get exercise with sets: %w", err)
	}
	return sets, nil
}
