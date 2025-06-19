package workout

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkoutService handles workout business logic
type WorkoutService struct {
	logger  *slog.Logger
	queries *db.Queries
	conn    *pgxpool.Pool
}

func NewService(logger *slog.Logger, queries *db.Queries, conn *pgxpool.Pool) *WorkoutService {
	return &WorkoutService{
		logger:  logger,
		queries: queries,
		conn:    conn,
	}
}

func (ws *WorkoutService) ListWorkouts(ctx context.Context) ([]db.Workout, error) {
	workouts, err := ws.queries.ListWorkouts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	return workouts, nil
}

func (ws *WorkoutService) GetWorkoutWithSets(ctx context.Context, id int32) ([]db.GetWorkoutWithSetsRow, error) {
	workouts, err := ws.queries.GetWorkoutWithSets(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout with sets: %w", err)
	}
	return workouts, nil
}

func (ws *WorkoutService) CreateWorkout(ctx context.Context, requestBody CreateWorkoutRequest) error {
	reformatted, err := transformRequest(requestBody)
	if err != nil {
		return fmt.Errorf("failed to transform request: %w", err)
	}

	pgData, err := convertToPGTypes(reformatted)
	if err != nil {
		return fmt.Errorf("failed to convert to PG types: %w", err)
	}

	// Insert into database
	if err := ws.SaveWorkout(ctx, pgData); err != nil {
		return fmt.Errorf("failed to save workout: %w", err)
	}

	return nil
}

// SaveWorkout efficiently inserts the PG-converted workout data using SQLC
func (ws *WorkoutService) SaveWorkout(ctx context.Context, pgData *PGReformattedRequest) error {
	// Start transaction
	tx, err := ws.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Will be ignored if tx.Commit() succeeds

	// Create queries instance with transaction
	qtx := ws.queries.WithTx(tx)

	// Step 1: Insert workout and get ID
	workout, err := ws.insertWorkout(ctx, qtx, pgData.Workout)
	if err != nil {
		return fmt.Errorf("failed to insert workout: %w", err)
	}

	// Step 2: Get or create exercises and build exercise name->ID mapping
	exerciseMap, err := ws.getOrCreateExercises(ctx, qtx, pgData.Exercises)
	if err != nil {
		return fmt.Errorf("failed to get/create exercises: %w", err)
	}

	// Step 3: Insert all sets
	if err := ws.insertSets(ctx, qtx, pgData.Sets, workout.ID, exerciseMap); err != nil {
		return fmt.Errorf("failed to insert sets: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// insertWorkout creates a single workout using SQLC
func (ws *WorkoutService) insertWorkout(ctx context.Context, qtx *db.Queries, workout PGWorkoutData) (db.Workout, error) {
	return qtx.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:  workout.Date,
		Notes: workout.Notes,
	})
}

// getOrCreateExercises efficiently handles all exercises using SQLC's GetOrCreateExercise
func (ws *WorkoutService) getOrCreateExercises(ctx context.Context, qtx *db.Queries, exercises []PGExerciseData) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		dbExercise, err := qtx.GetOrCreateExercise(ctx, exercise.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// insertSets creates all sets using SQLC's CreateSet
func (ws *WorkoutService) insertSets(ctx context.Context, qtx *db.Queries, sets []PGSetData, workoutID int32, exerciseMap map[string]int32) error {
	for _, set := range sets {
		exerciseID, exists := exerciseMap[set.ExerciseName]
		if !exists {
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
			return fmt.Errorf("failed to create set for exercise %s: %w", set.ExerciseName, err)
		}
	}

	return nil
}

func transformRequest(request CreateWorkoutRequest) (*ReformattedRequest, error) {
	// Parse date
	parsedDate, err := time.Parse("2006-01-02T15:04:05Z07:00", request.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create workout data
	workout := WorkoutData{
		Date:  parsedDate,
		Notes: request.Notes,
	}

	// Process exercises and sets (your existing logic)
	exerciseMap := make(map[string]bool)
	var exercises []ExerciseData
	var sets []SetData

	for _, exercise := range request.Exercises {
		if !exerciseMap[exercise.Name] {
			exerciseMap[exercise.Name] = true
			exercises = append(exercises, ExerciseData{
				Name: exercise.Name,
			})
		}

		for _, set := range exercise.Sets {
			sets = append(sets, SetData{
				ExerciseName: exercise.Name,
				Weight:       set.Weight,
				Reps:         set.Reps,
				SetType:      set.SetType,
			})
		}
	}

	return &ReformattedRequest{
		Workout:   workout,
		Exercises: exercises,
		Sets:      sets,
	}, nil
}

func convertToPGTypes(reformatted *ReformattedRequest) (*PGReformattedRequest, error) {
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
