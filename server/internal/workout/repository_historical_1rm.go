package workout

import (
	"context"
	"fmt"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (wr *workoutRepository) updateHistorical1rmFromWorkout(ctx context.Context, qtx *db.Queries, workoutID int32, userID string) error {
	rows, err := qtx.GetWorkoutBestE1rmByExercise(ctx, db.GetWorkoutBestE1rmByExerciseParams{
		WorkoutID: workoutID,
		UserID:    userID,
	})
	if err != nil {
		return fmt.Errorf("get workout best e1rm by exercise failed: %w", err)
	}

	for _, row := range rows {
		if err := qtx.UpdateExerciseHistorical1RMFromWorkoutIfBetter(ctx, db.UpdateExerciseHistorical1RMFromWorkoutIfBetterParams{
			ID:            row.ExerciseID,
			Historical1rm: row.BestE1rm,
			Historical1rmSourceWorkoutID: pgtype.Int4{
				Int32: workoutID,
				Valid: true,
			},
			UserID: userID,
		}); err != nil {
			return fmt.Errorf("update exercise historical 1rm from workout if better failed (exercise_id: %d): %w", row.ExerciseID, err)
		}
	}

	return nil
}

func (wr *workoutRepository) recomputeHistorical1rmForExercisesSourcedFromWorkout(ctx context.Context, qtx *db.Queries, workoutID int32, userID string) error {
	exerciseIDs, err := qtx.ListExercisesWithHistorical1RMSourceWorkout(ctx, db.ListExercisesWithHistorical1RMSourceWorkoutParams{
		UserID: userID,
		Historical1rmSourceWorkoutID: pgtype.Int4{
			Int32: workoutID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("list exercises with historical 1rm sourced from workout failed: %w", err)
	}

	for _, exerciseID := range exerciseIDs {
		if err := wr.recomputeHistorical1rmForExercise(ctx, qtx, exerciseID, userID); err != nil {
			return err
		}
	}

	return nil
}

func (wr *workoutRepository) recomputeHistorical1rmForExercise(ctx context.Context, qtx *db.Queries, exerciseID int32, userID string) error {
	best, err := qtx.GetExerciseBestE1rmWithWorkout(ctx, db.GetExerciseBestE1rmWithWorkoutParams{
		UserID:     userID,
		ExerciseID: exerciseID,
	})
	switch {
	case err == nil:
		return qtx.SetExerciseHistorical1RM(ctx, db.SetExerciseHistorical1RMParams{
			ID:            exerciseID,
			Historical1rm: best.E1rm,
			Historical1rmSourceWorkoutID: pgtype.Int4{
				Int32: best.WorkoutID,
				Valid: true,
			},
			UserID: userID,
		})
	case err == pgx.ErrNoRows:
		return qtx.SetExerciseHistorical1RM(ctx, db.SetExerciseHistorical1RMParams{
			ID:                           exerciseID,
			Historical1rm:                pgtype.Numeric{Valid: false},
			Historical1rmSourceWorkoutID: pgtype.Int4{Valid: false},
			UserID:                       userID,
		})
	default:
		return fmt.Errorf("get exercise best e1rm with workout failed (exercise_id: %d): %w", exerciseID, err)
	}
}

func (wr *workoutRepository) recomputeHistorical1rmForExerciseExcludingWorkout(ctx context.Context, qtx *db.Queries, exerciseID int32, excludedWorkoutID int32, userID string) error {
	best, err := qtx.GetExerciseBestE1rmWithWorkoutExcludingWorkout(ctx, db.GetExerciseBestE1rmWithWorkoutExcludingWorkoutParams{
		UserID:     userID,
		ExerciseID: exerciseID,
		WorkoutID:  excludedWorkoutID,
	})
	switch {
	case err == nil:
		return qtx.SetExerciseHistorical1RM(ctx, db.SetExerciseHistorical1RMParams{
			ID:            exerciseID,
			Historical1rm: best.E1rm,
			Historical1rmSourceWorkoutID: pgtype.Int4{
				Int32: best.WorkoutID,
				Valid: true,
			},
			UserID: userID,
		})
	case err == pgx.ErrNoRows:
		return qtx.SetExerciseHistorical1RM(ctx, db.SetExerciseHistorical1RMParams{
			ID:                           exerciseID,
			Historical1rm:                pgtype.Numeric{Valid: false},
			Historical1rmSourceWorkoutID: pgtype.Int4{Valid: false},
			UserID:                       userID,
		})
	default:
		return fmt.Errorf("get exercise best e1rm excluding workout failed (exercise_id: %d): %w", exerciseID, err)
	}
}
