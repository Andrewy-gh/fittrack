package workout

import (
	"context"
	"fmt"
	"log/slog"

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
func getOrCreateExercises(ctx context.Context, logger *slog.Logger, exerciseRepo exercise.ExerciseRepository, qtx *db.Queries, exercises []PGExerciseData, userID string) (map[string]int32, error) {
	exerciseMap := make(map[string]int32)

	for _, exercise := range exercises {
		logger.Info("attempting to get or create exercise", "exercise_name", exercise.Name, "user_id", userID)
		dbExercise, err := exerciseRepo.GetOrCreateExerciseTx(ctx, qtx, exercise.Name, userID)
		if err != nil {
			logger.Error("failed to get/create exercise", "exercise_name", exercise.Name, "error", err)
			return nil, fmt.Errorf("failed to get/create exercise %s: %w", exercise.Name, err)
		}
		logger.Info("successfully got/created exercise", "exercise_name", exercise.Name, "exercise_id", dbExercise.ID)
		exerciseMap[exercise.Name] = dbExercise.ID
	}

	return exerciseMap, nil
}

// MARK: insertSets
func insertSets(ctx context.Context, logger *slog.Logger, qtx *db.Queries, sets []PGSetData, workoutID int32, exerciseMap map[string]int32, userID string) error {
	for _, set := range sets {
		exerciseID, exists := exerciseMap[set.ExerciseName]
		if !exists {
			errMsg := fmt.Sprintf("exercise not found in exercise map: %s", set.ExerciseName)
			logger.Error(errMsg, "exercise_name", set.ExerciseName, "available_exercises", exerciseMap)
			return fmt.Errorf("exercise not found: %s", set.ExerciseName)
		}

		logger.Info("attempting to create set",
			"exercise_name", set.ExerciseName,
			"exercise_id", exerciseID,
			"workout_id", workoutID,
			"reps", set.Reps,
			"weight", set.Weight,
			"set_type", set.SetType,
			"exercise_order", set.ExerciseOrder,
			"set_order", set.SetOrder,
			"user_id", userID)

		_, err := qtx.CreateSet(ctx, db.CreateSetParams{
			ExerciseID:    exerciseID,
			WorkoutID:     workoutID,
			Weight:        set.Weight,
			Reps:          set.Reps,
			SetType:       set.SetType,
			UserID:        userID,
			ExerciseOrder: set.ExerciseOrder,
			SetOrder:      set.SetOrder,
		})
		if err != nil {
			errMsg := fmt.Sprintf("failed to create set for exercise %s (ID: %d)", set.ExerciseName, exerciseID)
			logger.Error(errMsg, "error", err, "set_details", set, "user_id", userID)
			return fmt.Errorf("failed to create set for exercise %s: %w", set.ExerciseName, err)
		}
	}

	return nil
}

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

	// Convert exercises
	var pgExercises []PGExerciseData
	for _, exercise := range reformatted.Exercises {
		pgExercises = append(pgExercises, PGExerciseData(exercise))
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
			if err := pgSet.Weight.Scan(fmt.Sprintf("%.1f", *set.Weight)); err != nil {
				return nil, fmt.Errorf("failed to convert weight to numeric: %w", err)
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
