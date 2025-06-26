package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutRepository defines the interface for workout data operations
type WorkoutRepository interface {
	// SaveWorkout saves a complete workout with all its related data
	SaveWorkout(ctx context.Context, pgData *PGReformattedRequest) error
	ListWorkouts(ctx context.Context) ([]db.Workout, error)
	GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error)
}

type workoutRepository struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

// NewRepository creates a new instance of WorkoutRepository
func NewRepository(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) WorkoutRepository {
	return &workoutRepository{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (wr *workoutRepository) SaveWorkout(ctx context.Context, pgData *PGReformattedRequest) error {
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
	workout, err := wr.insertWorkout(ctx, qtx, pgData.Workout)
	if err != nil {
		wr.logger.Error("failed to insert workout", "error", err)
		return fmt.Errorf("failed to insert workout: %w", err)
	}

	// Step 2: Get or create exercises and build exercise name->ID mapping
	exerciseMap, err := wr.getOrCreateExercises(ctx, qtx, pgData.Exercises)
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

// insertWorkout creates a single workout using SQLC
func (wr *workoutRepository) insertWorkout(ctx context.Context, qtx *db.Queries, workout PGWorkoutData) (db.Workout, error) {
	return qtx.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:  workout.Date,
		Notes: workout.Notes,
	})
}

// getOrCreateExercises efficiently handles all exercises using SQLC's GetOrCreateExercise
func (wr *workoutRepository) getOrCreateExercises(ctx context.Context, qtx *db.Queries, exercises []PGExerciseData) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		dbExercise, err := qtx.GetOrCreateExercise(ctx, exercise.Name)
		if err != nil {
			wr.logger.Error("failed to get/create exercise", "exercise_name", exercise.Name, "error", err)
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// insertSets creates all sets using SQLC's CreateSet
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

func (wr *workoutRepository) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	workouts, err := wr.queries.ListWorkouts(ctx)
	if err != nil {
		wr.logger.Error("list workouts query failed", "error", err)
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}

func (wr *workoutRepository) GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error) {
	workoutWithSets, err := wr.queries.GetWorkoutWithSets(ctx, id)
	if err != nil {
		wr.logger.Error("get workout with sets query failed", "workout_id", id, "error", err)
		return nil, fmt.Errorf("failed to get workout with sets (id: %d): %w", id, err)
	}
	return workoutWithSets, nil
}
