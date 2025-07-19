package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type workoutRepository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) WorkoutRepository {
	return &workoutRepository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

// MARK: ListWorkouts
func (wr *workoutRepository) ListWorkouts(ctx context.Context, userId string) ([]db.Workout, error) {
	pgUserId := pgtype.Text{String: userId, Valid: true}

	workouts, err := wr.queries.ListWorkouts(ctx, pgUserId)
	if err != nil {
		wr.logger.Error("list workouts query failed", "error", err)
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}

	if workouts == nil {
		workouts = []db.Workout{}
	}

	return workouts, nil
}

// MARK: GetWorkoutWithSets
func (wr *workoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
	params := db.GetWorkoutWithSetsParams{
		ID:     id,
		UserID: pgtype.Text{String: userID, Valid: true},
	}
	workoutWithSets, err := wr.queries.GetWorkoutWithSets(ctx, params)
	if err != nil {
		wr.logger.Error("get workout with sets query failed", "workout_id", id, "error", err)
		return nil, fmt.Errorf("failed to get workout with sets (id: %d): %w", id, err)
	}

	if workoutWithSets == nil {
		workoutWithSets = []db.GetWorkoutWithSetsRow{}
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
	if err := wr.insertSets(ctx, qtx, pgData.Sets, workout.ID, exerciseMap); err != nil {
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

// MARK: Utilities

// MARK: insertWorkout
func (wr *workoutRepository) insertWorkout(ctx context.Context, qtx *db.Queries, workout PGWorkoutData, userID string) (db.Workout, error) {
	return qtx.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:   workout.Date,
		Notes:  workout.Notes,
		UserID: pgtype.Text{String: userID, Valid: true},
	})
}

// MARK: getOrCreateExercises
func (wr *workoutRepository) getOrCreateExercises(ctx context.Context, qtx *db.Queries, exercises []PGExerciseData, userID string) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		dbExercise, err := qtx.GetOrCreateExercise(ctx, db.GetOrCreateExerciseParams{
			Name:   exercise.Name,
			UserID: pgtype.Text{String: userID, Valid: true},
		})
		if err != nil {
			wr.logger.Error("failed to get/create exercise", "exercise_name", exercise.Name, "error", err)
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// MARK: insertSets
func (wr *workoutRepository) insertSets(ctx context.Context, qtx *db.Queries, sets []PGSetData, workoutID int32, exerciseMap map[string]int32) error {
	for _, set := range sets {
		exerciseID, exists := exerciseMap[set.ExerciseName]
		if !exists {
			wr.logger.Error("exercise not found", "exercise_name", set.ExerciseName)
			return fmt.Errorf("exercise not found: %s", set.ExerciseName)
		}

		_, err := qtx.CreateSet(ctx, db.CreateSetParams{
			ExerciseID: exerciseID,
			WorkoutID:  workoutID,
			Weight:     set.Weight,
			Reps:       set.Reps,
			SetType:    set.SetType,
		})
		if err != nil {
			wr.logger.Error("failed to create set", "exercise_name", set.ExerciseName, "error", err)
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
