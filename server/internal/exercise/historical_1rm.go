package exercise

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ExerciseHistorical1RMResponse struct {
	Historical1RM                *float64   `json:"historical_1rm,omitempty"`
	Historical1RMUpdatedAt       *time.Time `json:"historical_1rm_updated_at,omitempty"`
	Historical1RMSourceWorkoutID *int32     `json:"historical_1rm_source_workout_id,omitempty"`
	ComputedBestE1RM             *float64   `json:"computed_best_e1rm,omitempty"`
	ComputedBestWorkoutID        *int32     `json:"computed_best_workout_id,omitempty"`
}

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

// MARK: GetExerciseHistorical1RM
func (es *ExerciseService) GetExerciseHistorical1RM(ctx context.Context, id int32) (*ExerciseHistorical1RMResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		return nil, &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", id)}
	}

	row, err := es.repo.GetExerciseHistorical1RM(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise historical 1rm: %w", err)
	}

	var resp ExerciseHistorical1RMResponse

	resp.Historical1RM, err = floatPtrFromNumeric(row.Historical1rm)
	if err != nil {
		return nil, err
	}

	if row.Historical1rmUpdatedAt.Valid {
		t := row.Historical1rmUpdatedAt.Time
		resp.Historical1RMUpdatedAt = &t
	}

	if row.Historical1rmSourceWorkoutID.Valid {
		id := row.Historical1rmSourceWorkoutID.Int32
		resp.Historical1RMSourceWorkoutID = &id
	}

	best, err := es.repo.GetExerciseBestE1rmWithWorkout(ctx, id, userID)
	switch {
	case err == nil:
		resp.ComputedBestE1RM, err = floatPtrFromNumeric(best.E1rm)
		if err != nil {
			return nil, err
		}
		workoutID := best.WorkoutID
		resp.ComputedBestWorkoutID = &workoutID
	case err == pgx.ErrNoRows:
		// No working sets yet.
	default:
		return nil, fmt.Errorf("failed to get computed best e1rm: %w", err)
	}

	return &resp, nil
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
