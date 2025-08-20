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

func (er *exerciseRepository) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	exercises, err := er.queries.ListExercises(ctx, userID)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("list exercises query failed - RLS policy violation",
				"error", err,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("list exercises query failed", "error", err, "user_id", userID)
		}
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}

	if exercises == nil {
		exercises = []db.Exercise{}
	}

	// Log empty results that might indicate RLS filtering
	if len(exercises) == 0 {
		er.logger.Debug("list exercises returned empty results",
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return exercises, nil
}

func (er *exerciseRepository) GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error) {
	params := db.GetExerciseParams{
		ID:     id,
		UserID: userID,
	}
	exercise, err := er.queries.GetExercise(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get exercise query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get exercise query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return db.Exercise{}, fmt.Errorf("failed to get exercise (id: %d): %w", id, err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetOrCreateExercise(ctx context.Context, name string, userID string) (db.Exercise, error) {
	params := db.GetOrCreateExerciseParams{
		Name:   name,
		UserID: userID,
	}
	exercise, err := er.queries.GetOrCreateExercise(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get or create exercise query failed - RLS policy violation",
				"error", err,
				"exercise_name", name,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get or create exercise query failed",
				"exercise_name", name,
				"user_id", userID,
				"error", err)
		}
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return exercise, nil
}

func (er *exerciseRepository) GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error) {
	params := db.GetExerciseWithSetsParams{
		ExerciseID: id,
		UserID:     userID,
	}
	sets, err := er.queries.GetExerciseWithSets(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get exercise with sets query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get exercise with sets query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return nil, fmt.Errorf("failed to get exercise with sets (id: %d): %w", id, err)
	}

	if sets == nil {
		sets = []db.GetExerciseWithSetsRow{}
	}

	// Log empty results that might indicate RLS filtering
	if len(sets) == 0 {
		er.logger.Debug("get exercise with sets returned empty results",
			"exercise_id", id,
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return sets, nil
}

func (er *exerciseRepository) GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name, userID string) (db.Exercise, error) {
	er.logger.Info("GetOrCreateExerciseTx called", "exercise_name", name, "user_id", userID)

	params := db.GetOrCreateExerciseParams{
		Name:   name,
		UserID: userID,
	}

	er.logger.Info("calling GetOrCreateExercise with params", "params", params)
	exercise, err := qtx.GetOrCreateExercise(ctx, params)
	if err != nil {
		er.logger.Error("failed to get or create exercise",
			"error", err,
			"exercise_name", name,
			"user_id", userID)
		return db.Exercise{}, fmt.Errorf("failed to get or create exercise: %w", err)
	}

	er.logger.Info("successfully got/created exercise",
		"exercise_id", exercise.ID,
		"exercise_name", exercise.Name,
		"user_id", exercise.UserID)

	return exercise, nil
}

func (er *exerciseRepository) GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error) {
	params := db.GetRecentSetsForExerciseParams{
		ExerciseID: id,
		UserID:     userID,
	}
	sets, err := er.queries.GetRecentSetsForExercise(ctx, params)
	if err != nil {
		if db.IsRowLevelSecurityError(err) {
			er.logger.Error("get recent sets for exercise query failed - RLS policy violation",
				"error", err,
				"exercise_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			er.logger.Error("get recent sets for exercise query failed",
				"exercise_id", id,
				"user_id", userID,
				"error", err)
		}
		return nil, fmt.Errorf("failed to get recent sets for exercise (id: %d): %w", id, err)
	}

	if sets == nil {
		sets = []db.GetRecentSetsForExerciseRow{}
	}

	if len(sets) == 0 {
		er.logger.Debug("get recent sets for exercise returned empty results",
			"exercise_id", id,
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return sets, nil
}
