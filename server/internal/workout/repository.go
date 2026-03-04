package workout

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type workoutRepository struct {
	logger       *slog.Logger
	queries      *db.Queries
	conn         *pgxpool.Pool
	exerciseRepo exercise.ExerciseRepository
}

func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool, exerciseRepo exercise.ExerciseRepository) WorkoutRepository {
	return &workoutRepository{
		logger:       logger,
		queries:      queries,
		conn:         conn,
		exerciseRepo: exerciseRepo,
	}
}

// MARK: ListWorkouts
func (wr *workoutRepository) ListWorkouts(ctx context.Context, userId string) ([]db.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	workoutRows, err := wr.queries.ListWorkouts(ctx, userId)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("list workouts query failed - RLS policy violation",
				"error", err,
				"user_id", userId,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("list workouts query failed", "error", err, "user_id", userId)
		}
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}

	if workoutRows == nil {
		return []db.Workout{}, nil
	}

	// Convert ListWorkoutsRow to Workout
	workouts := make([]db.Workout, len(workoutRows))
	for i, row := range workoutRows {
		workouts[i] = db.Workout{
			ID:           row.ID,
			Date:         row.Date,
			Notes:        row.Notes,
			WorkoutFocus: row.WorkoutFocus,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
			// UserID is not returned by optimized query but was used for filtering
			UserID: userId,
		}
	}

	// Log empty results that might indicate RLS filtering
	if len(workouts) == 0 {
		wr.logger.Debug("list workouts returned empty results",
			"user_id", userId,
			"potential_rls_filtering", true)
	}

	return workouts, nil
}

// MARK: GetWorkout
func (wr *workoutRepository) GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetWorkoutParams{
		ID:     id,
		UserID: userID,
	}
	workoutRow, err := wr.queries.GetWorkout(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("get workout query failed - RLS policy violation",
				"error", err,
				"workout_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("get workout query failed",
				"workout_id", id,
				"user_id", userID,
				"error", err)
		}
		return db.Workout{}, fmt.Errorf("failed to get workout (id: %d): %w", id, err)
	}

	// Convert GetWorkoutRow to Workout
	workout := db.Workout{
		ID:           workoutRow.ID,
		Date:         workoutRow.Date,
		Notes:        workoutRow.Notes,
		WorkoutFocus: workoutRow.WorkoutFocus,
		CreatedAt:    workoutRow.CreatedAt,
		UpdatedAt:    workoutRow.UpdatedAt,
		// UserID is not returned by optimized query but was used for filtering
		UserID: userID,
	}

	return workout, nil
}

// MARK: GetWorkoutWithSets
func (wr *workoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	params := db.GetWorkoutWithSetsParams{
		ID:     id,
		UserID: userID,
	}
	workoutWithSets, err := wr.queries.GetWorkoutWithSets(ctx, params)
	if err != nil {
		// Check if this might be an RLS-related error
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("get workout with sets query failed - RLS policy violation",
				"error", err,
				"workout_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("get workout with sets query failed",
				"workout_id", id,
				"user_id", userID,
				"error", err)
		}
		return nil, fmt.Errorf("failed to get workout with sets (id: %d): %w", id, err)
	}

	if workoutWithSets == nil {
		workoutWithSets = []db.GetWorkoutWithSetsRow{}
	}

	// Log empty results that might indicate RLS filtering
	if len(workoutWithSets) == 0 {
		wr.logger.Debug("get workout with sets returned empty results",
			"workout_id", id,
			"user_id", userID,
			"potential_rls_filtering", true)
	}

	return workoutWithSets, nil
}

// MARK: SaveWorkout
func (wr *workoutRepository) SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pgData, err := wr.convertToPGTypes(reformatted)
	if err != nil {
		wr.logger.Error("failed to convert to PG types", "error", err)
		return fmt.Errorf("failed to convert to PG types: %w", err)
	}

	// Start transaction
	tx, err := wr.conn.Begin(ctx)
	if err != nil {
		wr.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create queries instance with transaction
	qtx := wr.queries.WithTx(tx)

	// Step 1: Insert workout and get ID
	workout, err := wr.insertWorkout(ctx, qtx, pgData.Workout, userID)
	if err != nil {
		wr.logger.Error("failed to insert workout", "error", err)
		return fmt.Errorf("failed to insert workout: %w", err)
	}

	// Step 2: Get or create exercises and build exercise name->ID mapping
	exerciseMap, err := wr.getOrCreateExercises(ctx, qtx, pgData.Exercises, userID)
	if err != nil {
		wr.logger.Error("failed to get/create exercises", "error", err)
		return fmt.Errorf("failed to get/create exercises: %w", err)
	}

	// Step 3: Insert all sets
	if err := wr.insertSets(ctx, qtx, pgData.Sets, workout.ID, exerciseMap, userID); err != nil {
		wr.logger.Error("failed to insert sets", "error", err)
		return fmt.Errorf("failed to insert sets: %w", err)
	}

	// Step 4: Update historical 1RM from this workout (auto PR detection).
	if err := wr.updateHistorical1rmFromWorkout(ctx, qtx, workout.ID, userID); err != nil {
		wr.logger.Error("failed to update historical 1RM from workout", "error", err, "workout_id", workout.ID)
		return fmt.Errorf("failed to update historical 1RM from workout: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		wr.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MARK: UpdateWorkout (PUT endpoint)
// Updates workout metadata (date/notes) and exercises/sets
// Uses a replace strategy for exercises/sets (deletes existing and recreates)
func (wr *workoutRepository) UpdateWorkout(ctx context.Context, id int32, reformatted *ReformattedRequest, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Convert to PG types
	pgData, err := wr.convertToPGTypes(reformatted)
	if err != nil {
		wr.logger.Error("failed to convert to PG types for update", "error", err)
		return fmt.Errorf("failed to convert to PG types: %w", err)
	}

	// Start transaction
	tx, err := wr.conn.Begin(ctx)
	if err != nil {
		wr.logger.Error("failed to begin transaction for update", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create queries instance with transaction
	qtx := wr.queries.WithTx(tx)

	// Step 1: Update workout metadata
	_, err = qtx.UpdateWorkout(ctx, db.UpdateWorkoutParams{
		ID:           id,
		Date:         pgData.Workout.Date,
		Notes:        pgData.Workout.Notes,
		WorkoutFocus: pgData.Workout.WorkoutFocus,
		UserID:       userID,
	})
	if err != nil {
		// Check for RLS violations
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("update workout failed - RLS policy violation",
				"error", err,
				"workout_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("update workout failed",
				"workout_id", id,
				"user_id", userID,
				"error", err)
		}
		return fmt.Errorf("failed to update workout (id: %d): %w", id, err)
	}

	wr.logger.Info("workout metadata updated successfully",
		"workout_id", id,
		"user_id", userID)

	// Step 2: Handle exercise/set updates (replace strategy - delete all and recreate)
	if len(reformatted.Exercises) > 0 {
		wr.logger.Info("processing exercise/set updates",
			"workout_id", id,
			"exercise_count", len(reformatted.Exercises))

		// Delete all existing sets for this workout
		if err := qtx.DeleteSetsByWorkout(ctx, db.DeleteSetsByWorkoutParams{
			WorkoutID: id,
			UserID:    userID,
		}); err != nil {
			wr.logger.Error("failed to delete existing sets", "error", err, "workout_id", id)
			return fmt.Errorf("failed to delete existing sets: %w", err)
		}
		wr.logger.Info("deleted existing sets for workout", "workout_id", id)

		// Get or create exercises and build exercise name->ID mapping
		exerciseMap, err := wr.getOrCreateExercises(ctx, qtx, pgData.Exercises, userID)
		if err != nil {
			wr.logger.Error("failed to get/create exercises for update", "error", err)
			return fmt.Errorf("failed to get/create exercises: %w", err)
		}

		// Insert new sets
		if err := wr.insertSets(ctx, qtx, pgData.Sets, id, exerciseMap, userID); err != nil {
			wr.logger.Error("failed to insert new sets", "error", err)
			return fmt.Errorf("failed to insert new sets: %w", err)
		}

		// If this workout currently owns any historical 1RM, recompute after the replace update.
		if err := wr.recomputeHistorical1rmForExercisesSourcedFromWorkout(ctx, qtx, id, userID); err != nil {
			wr.logger.Error("failed to recompute historical 1RM after workout update", "error", err, "workout_id", id)
			return fmt.Errorf("failed to recompute historical 1RM after workout update: %w", err)
		}

		// Update historical 1RM from this workout (auto PR detection).
		if err := wr.updateHistorical1rmFromWorkout(ctx, qtx, id, userID); err != nil {
			wr.logger.Error("failed to update historical 1RM from workout", "error", err, "workout_id", id)
			return fmt.Errorf("failed to update historical 1RM from workout: %w", err)
		}

		wr.logger.Info("successfully updated exercises and sets",
			"workout_id", id,
			"exercises_processed", len(reformatted.Exercises),
			"sets_created", len(pgData.Sets))
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		wr.logger.Error("failed to commit update transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MARK: DeleteWorkout
// Deletes a workout and all associated sets (CASCADE delete automatically handles sets)
func (wr *workoutRepository) DeleteWorkout(ctx context.Context, id int32, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Delete can invalidate historical 1RM values sourced from this workout.
	// Keep it transactional so cascade deletes and recomputes are consistent.
	tx, err := wr.conn.Begin(ctx)
	if err != nil {
		wr.logger.Error("failed to begin transaction for delete", "error", err)
		return fmt.Errorf("failed to begin delete transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := wr.queries.WithTx(tx)

	// Identify exercises whose PR is currently attributed to this workout.
	exerciseIDs, err := qtx.ListExercisesWithHistorical1RMSourceWorkout(ctx, db.ListExercisesWithHistorical1RMSourceWorkoutParams{
		UserID: userID,
		Historical1rmSourceWorkoutID: pgtype.Int4{
			Int32: id,
			Valid: true,
		},
	})
	if err != nil {
		wr.logger.Error("failed to list exercises with historical 1RM sourced from workout", "error", err, "workout_id", id)
		return fmt.Errorf("failed to list exercises with historical 1RM sourced from workout: %w", err)
	}

	// Clear/recompute PR attribution *before* deleting to satisfy FK constraint.
	for _, exerciseID := range exerciseIDs {
		if err := wr.recomputeHistorical1rmForExerciseExcludingWorkout(ctx, qtx, exerciseID, id, userID); err != nil {
			wr.logger.Error("failed to recompute historical 1RM before workout delete", "error", err, "exercise_id", exerciseID, "workout_id", id)
			return fmt.Errorf("failed to recompute historical 1RM before workout delete: %w", err)
		}
	}

	// Delete the workout - CASCADE will automatically delete associated sets.
	if err := qtx.DeleteWorkout(ctx, db.DeleteWorkoutParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		// Check for RLS violations
		if db.IsRowLevelSecurityError(err) {
			wr.logger.Error("delete workout failed - RLS policy violation",
				"error", err,
				"workout_id", id,
				"user_id", userID,
				"error_type", "rls_violation")
		} else {
			wr.logger.Error("delete workout failed",
				"workout_id", id,
				"user_id", userID,
				"error", err)
		}
		return fmt.Errorf("failed to delete workout (id: %d): %w", id, err)
	}

	if err := tx.Commit(ctx); err != nil {
		wr.logger.Error("failed to commit delete transaction", "error", err)
		return fmt.Errorf("failed to commit delete transaction: %w", err)
	}

	wr.logger.Info("workout deleted successfully (sets cascaded)",
		"workout_id", id,
		"user_id", userID)

	return nil
}

var _ WorkoutRepository = (*workoutRepository)(nil)
