package workout

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/jackc/pgx/v5/pgtype"
)

// MARK: insertWorkout
func insertWorkout(ctx context.Context, qtx *db.Queries, workout PGWorkoutData, userID string) (db.Workout, error) {
	workoutID, err := qtx.CreateWorkout(ctx, db.CreateWorkoutParams{
		Date:         workout.Date,
		Notes:        workout.Notes,
		WorkoutFocus: workout.WorkoutFocus,
		UserID:       userID,
	})
	if err != nil {
		return db.Workout{}, fmt.Errorf("failed to create workout: %w", err)
	}

	// Create Workout from returned ID
	return db.Workout{
		ID:           workoutID,
		Date:         workout.Date,
		Notes:        workout.Notes,
		WorkoutFocus: workout.WorkoutFocus,
		UserID:       userID,
	}, nil
}

// MARK: getOrCreateExercises
func getOrCreateExercises(ctx context.Context, logger *slog.Logger, exerciseRepo exercise.ExerciseRepository, qtx *db.Queries, exercises []ExerciseData, userID string) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		logger.Debug("attempting to get or create exercise", "exercise_name", exercise.Name, "user_id", userID)
		var dbExercise db.Exercise
		var err error
		if exercise.CatalogID != nil {
			dbExercise.ID, err = qtx.GetOrCreateCatalogExercise(ctx, db.GetOrCreateCatalogExerciseParams{
				Name: exercise.Name, UserID: userID, CatalogID: pgtype.Text{String: *exercise.CatalogID, Valid: true},
			})
		} else {
			dbExercise, err = exerciseRepo.GetOrCreateExerciseTx(ctx, qtx, exercise.Name, userID)
		}
		if err != nil {
			logger.Error("failed to get/create exercise", "exercise_name", exercise.Name, "error", err)
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		logger.Debug("successfully got/created exercise", "exercise_name", exercise.Name, "exercise_id", dbExercise.ID)
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// MARK: insertSets
func insertSets(ctx context.Context, logger *slog.Logger, qtx *db.Queries, sets []PGSetData, workoutID int32, exerciseMap map[string]int32, userID string) error {
	if len(sets) == 0 {
		return nil
	}
	var ids, reps, exerciseOrder, setOrder []int32
	var weights []pgtype.Numeric
	var types []string
	for _, set := range sets {
		exerciseID, exists := exerciseMap[set.ExerciseName]
		if !exists {
			errMsg := fmt.Sprintf("exercise not found in exercise map: %s", set.ExerciseName)
			logger.Error(errMsg, "exercise_name", set.ExerciseName, "available_exercises", exerciseMap)
			return fmt.Errorf("exercise not found: %s", set.ExerciseName)
		}
		ids = append(ids, exerciseID)
		weights = append(weights, set.Weight)
		reps = append(reps, set.Reps)
		types = append(types, set.SetType)
		exerciseOrder = append(exerciseOrder, set.ExerciseOrder)
		setOrder = append(setOrder, set.SetOrder)
	}
	if err := qtx.BulkCreateSets(ctx, workoutID, userID, ids, weights, reps, types, exerciseOrder, setOrder); err != nil {
		logger.Error("failed to create workout sets", "error", err, "workout_id", workoutID, "set_count", len(sets))
		return fmt.Errorf("failed to create workout sets: %w", err)
	}
	return nil
}

// ErrInvalidSetNumeric indicates a set cannot be represented by the workout storage schema.
var ErrInvalidSetNumeric = errors.New("set numeric value is outside storage limits")

// MARK: convertToPGTypes
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
		WorkoutFocus: pgtype.Text{
			String: "",
			Valid:  false,
		},
	}

	if reformatted.Workout.Notes != nil && *reformatted.Workout.Notes != "" {
		pgWorkout.Notes = pgtype.Text{
			String: *reformatted.Workout.Notes,
			Valid:  true,
		}
	}

	if reformatted.Workout.WorkoutFocus != nil && *reformatted.Workout.WorkoutFocus != "" {
		pgWorkout.WorkoutFocus = pgtype.Text{
			String: *reformatted.Workout.WorkoutFocus,
			Valid:  true,
		}
	}

	// Convert sets with ordering information
	var pgSets []PGSetData

	// Create exercise name to order mapping
	exerciseOrderMap := make(map[string]int32)
	for i, exercise := range reformatted.Exercises {
		exerciseOrderMap[exercise.Name] = int32(i + 1) // 1-based ordering
	}

	// Create set order counters per exercise
	setOrderCounters := make(map[string]int32)

	for _, set := range reformatted.Sets {
		if set.Reps < 0 || int64(set.Reps) > math.MaxInt32 {
			return nil, ErrInvalidSetNumeric
		}
		var roundedWeight string
		if set.Weight != nil {
			if math.IsNaN(*set.Weight) || math.IsInf(*set.Weight, 0) || *set.Weight < 0 {
				return nil, ErrInvalidSetNumeric
			}
			// Match the existing float-to-numeric conversion before applying SQL bounds.
			roundedWeight = fmt.Sprintf("%.1f", *set.Weight)
			tenths, err := strconv.ParseInt(strings.ReplaceAll(roundedWeight, ".", ""), 10, 64)
			if err != nil || tenths > 9999999999 {
				return nil, ErrInvalidSetNumeric
			}
			// NUMERIC(8,2) rounds to 1000000.00 at 999999.995. With a
			// weight in tenths, Epley is tenths*(30+reps)/300. Compare
			// using integer division to avoid overflow and floating-point ties.
			if set.SetType == "working" && tenths > 299999998/(30+int64(set.Reps)) {
				return nil, ErrInvalidSetNumeric
			}
		}
		// Increment set counter for this exercise
		setOrderCounters[set.ExerciseName]++

		pgSet := PGSetData{
			ExerciseName: set.ExerciseName,
			Weight: pgtype.Numeric{
				Valid: false,
			},
			Reps:          int32(set.Reps),
			SetType:       set.SetType,
			ExerciseOrder: exerciseOrderMap[set.ExerciseName],
			SetOrder:      setOrderCounters[set.ExerciseName],
		}

		if set.Weight != nil {
			// Convert float64 to pgtype.Numeric with proper precision
			if err := pgSet.Weight.Scan(roundedWeight); err != nil {
				return nil, fmt.Errorf("failed to convert weight to numeric: %w", err)
			}
		}

		pgSets = append(pgSets, pgSet)
	}

	return &PGReformattedRequest{
		Workout:   pgWorkout,
		Exercises: reformatted.Exercises,
		Sets:      pgSets,
	}, nil
}
