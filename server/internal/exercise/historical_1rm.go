package exercise

import (
	"context"
	"fmt"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func floatPtrFromNumeric(n pgtype.Numeric) (*float64, error) {
	if !n.Valid {
		return nil, nil
	}
	f64, err := n.Float64Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert numeric to float64: %w", err)
	}
	return &f64.Float64, nil
}

// MARK: UpdateExerciseHistorical1RM
func (es *ExerciseService) UpdateExerciseHistorical1RM(ctx context.Context, id int32, req UpdateExerciseHistorical1RMRequest) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		return &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", id)}
	}

	mode := req.Mode
	if mode == "" {
		mode = "manual"
	}

	switch mode {
	case "recompute":
		best, err := es.repo.GetExerciseBestE1rmWithWorkout(ctx, id, userID)
		switch {
		case err == nil:
			bestVal, err := floatPtrFromNumeric(best.E1rm)
			if err != nil {
				return err
			}
			workoutID := best.WorkoutID
			return es.repo.SetExerciseHistorical1RM(ctx, id, bestVal, &workoutID, userID)
		case err == pgx.ErrNoRows:
			return es.repo.SetExerciseHistorical1RM(ctx, id, nil, nil, userID)
		default:
			return fmt.Errorf("failed to recompute exercise historical 1rm: %w", err)
		}
	case "manual":
		return es.repo.UpdateExerciseHistorical1RMManual(ctx, id, req.Historical1RM, userID)
	default:
		return fmt.Errorf("invalid mode: %q", mode)
	}
}
