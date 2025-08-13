package workout

import (
	"context"
	"fmt"
	"log/slog"

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
	workouts, err := wr.queries.ListWorkouts(ctx, userId)
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

	if workouts == nil {
		workouts = []db.Workout{}
	}

	// Log empty results that might indicate RLS filtering
	if len(workouts) == 0 {
		wr.logger.Debug("list workouts returned empty results",
			"user_id", userId,
			"potential_rls_filtering", true)
	}

	return workouts, nil
}

// MARK: GetWorkoutWithSets
func (wr *workoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
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
	updatedWorkout, err := qtx.UpdateWorkout(ctx, db.UpdateWorkoutParams{
		ID:     id,
		Date:   pgData.Workout.Date,
		Notes:  pgData.Workout.Notes,
		UserID: userID,
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
		"workout_id", updatedWorkout.ID,
		"updated_at", updatedWorkout.UpdatedAt,
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

// MARK: Utilities

// MARK: insertWorkout
func (wr *workoutRepository) insertWorkout(ctx context.Context, qtx *db.Queries, workout PGWorkoutData, userID string) (db.Workout, error) {
	return qtx.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:   workout.Date,
		Notes:  workout.Notes,
		UserID: userID,
	})
}

// MARK: getOrCreateExercises
func (wr *workoutRepository) getOrCreateExercises(ctx context.Context, qtx *db.Queries, exercises []PGExerciseData, userID string) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		wr.logger.Info("attempting to get or create exercise", "exercise_name", exercise.Name, "user_id", userID)
		dbExercise, err := wr.exerciseRepo.GetOrCreateExerciseTx(ctx, qtx, exercise.Name, userID)
		if err != nil {
			wr.logger.Error("failed to get/create exercise", "exercise_name", exercise.Name, "error", err)
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		wr.logger.Info("successfully got/created exercise", "exercise_name", exercise.Name, "exercise_id", dbExercise.ID)
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// MARK: insertSets
func (wr *workoutRepository) insertSets(ctx context.Context, qtx *db.Queries, sets []PGSetData, workoutID int32, exerciseMap map[string]int32, userID string) error {
	for _, set := range sets {
		exerciseID, exists := exerciseMap[set.ExerciseName]
		if !exists {
			errMsg := fmt.Sprintf("exercise not found in exercise map: %s", set.ExerciseName)
			wr.logger.Error(errMsg, "exercise_name", set.ExerciseName, "available_exercises", exerciseMap)
			return fmt.Errorf("exercise not found: %s", set.ExerciseName)
		}

		wr.logger.Info("attempting to create set",
			"exercise_name", set.ExerciseName,
			"exercise_id", exerciseID,
			"workout_id", workoutID,
			"reps", set.Reps,
			"weight", set.Weight,
			"set_type", set.SetType,
			"user_id", userID)

		_, err := qtx.CreateSet(ctx, db.CreateSetParams{
			ExerciseID: exerciseID,
			WorkoutID:  workoutID,
			Weight:     set.Weight,
			Reps:       set.Reps,
			SetType:    set.SetType,
			UserID:     userID,
		})
		if err != nil {
			errMsg := fmt.Sprintf("failed to create set for exercise %s (ID: %d)", set.ExerciseName, exerciseID)
			wr.logger.Error(errMsg, "error", err, "set_details", set, "user_id", userID)
			return fmt.Errorf("failed to create set for exercise %s: %w", set.ExerciseName, err)
		}
	}

	return nil
}


// MARK: convertToPGTypes
func (wr *workoutRepository) convertToPGTypes(reformatted *ReformattedRequest) (*PGReformattedRequest, error) {
	// Convert workout
	pgWorkout := PGWorkoutData{
		Date: pgtype.Timestamptz{
			Time:  reformatted.Workout.Date,
			Valid: true,
		},
		Notes: pgtype.Text{
			String: "",
			Valid:  false,
		},
	}

	if reformatted.Workout.Notes != nil {
		pgWorkout.Notes = pgtype.Text{
			String: *reformatted.Workout.Notes,
			Valid:  true,
		}
	}

	// Convert exercises
	var pgExercises []PGExerciseData
	for _, exercise := range reformatted.Exercises {
		pgExercises = append(pgExercises, PGExerciseData(exercise))
	}

	// Convert sets
	var pgSets []PGSetData
	for _, set := range reformatted.Sets {
		pgSet := PGSetData{
			ExerciseName: set.ExerciseName,
			Weight: pgtype.Int4{
				Int32: 0,
				Valid: false,
			},
			Reps:    int32(set.Reps),
			SetType: set.SetType,
		}

		if set.Weight != nil {
			pgSet.Weight = pgtype.Int4{
				Int32: int32(*set.Weight),
				Valid: true,
			}
		}

		pgSets = append(pgSets, pgSet)
	}

	return &PGReformattedRequest{
		Workout:   pgWorkout,
		Exercises: pgExercises,
		Sets:      pgSets,
	}, nil
}
