package exercise

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type ExerciseRepository interface {
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error)
	GetOrCreateExercise(ctx context.Context, name, userID string) (db.Exercise, error)
	GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name, userID string) (db.Exercise, error)
	GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error)
	GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error)
	GetExerciseMetricsHistory(ctx context.Context, req GetExerciseMetricsHistoryRequest, userID string) ([]ExerciseMetricsHistoryPoint, MetricsHistoryBucket, error)
	UpdateExerciseName(ctx context.Context, id int32, name, userID string) error
	GetExerciseHistorical1RM(ctx context.Context, id int32, userID string) (db.GetExerciseHistorical1RMRow, error)
	GetExerciseBestE1rmWithWorkout(ctx context.Context, exerciseID int32, userID string) (db.GetExerciseBestE1rmWithWorkoutRow, error)
	UpdateExerciseHistorical1RMManual(ctx context.Context, id int32, historical1rm *float64, userID string) error
	SetExerciseHistorical1RM(ctx context.Context, id int32, historical1rm *float64, sourceWorkoutID *int32, userID string) error
	DeleteExercise(ctx context.Context, id int32, userID string) error
}

type ExerciseService struct {
	logger *slog.Logger
	repo   ExerciseRepository
}

func NewService(logger *slog.Logger, repo ExerciseRepository) *ExerciseService {
	return &ExerciseService{
		logger: logger,
		repo:   repo,
	}
}

func (es *ExerciseService) ListExercises(ctx context.Context) ([]db.Exercise, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	exercises, err := es.repo.ListExercises(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	return exercises, nil
}

func (es *ExerciseService) GetExerciseWithSets(ctx context.Context, id int32) ([]ExerciseWithSetsResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		return nil, &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", id)}
	}

	sets, err := es.repo.GetExerciseWithSets(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise with sets: %w", err)
	}

	// Convert database rows to response type
	response, err := es.convertExerciseWithSetsRows(sets)
	if err != nil {
		return nil, fmt.Errorf("failed to convert exercise with sets rows: %w", err)
	}

	return response, nil
}

func (es *ExerciseService) GetOrCreateExercise(ctx context.Context, name string) (*db.Exercise, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	exercise, err := es.repo.GetOrCreateExercise(ctx, name, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create exercise: %w", err)
	}
	return &exercise, nil
}

func (es *ExerciseService) GetRecentSetsForExercise(ctx context.Context, id int32) ([]db.GetRecentSetsForExerciseRow, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	sets, err := es.repo.GetRecentSetsForExercise(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent sets for exercise: %w", err)
	}
	return sets, nil
}

// MARK: UpdateExerciseName
func (es *ExerciseService) UpdateExerciseName(ctx context.Context, id int32, name string) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		return &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", id)}
	}

	if err := es.repo.UpdateExerciseName(ctx, id, name, userID); err != nil {
		return fmt.Errorf("failed to update exercise name: %w", err)
	}

	return nil
}

// MARK: DeleteExercise
func (es *ExerciseService) DeleteExercise(ctx context.Context, id int32) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, id, userID)
	if err != nil {
		return &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", id)}
	}

	if err := es.repo.DeleteExercise(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete exercise: %w", err)
	}

	return nil
}

// convertExerciseWithSetsRows converts database rows to response type, handling pgtype.Numeric to float64 conversion
func (es *ExerciseService) convertExerciseWithSetsRows(rows []db.GetExerciseWithSetsRow) ([]ExerciseWithSetsResponse, error) {
	response := make([]ExerciseWithSetsResponse, len(rows))

	for i, row := range rows {
		// Convert weight from pgtype.Numeric to *float64
		var weight *float64
		if row.Weight.Valid {
			f64, err := row.Weight.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("failed to convert weight: %w", err)
			}
			weight = &f64.Float64
		}

		// Convert volume from pgtype.Numeric to float64
		var volume float64
		if row.Volume.Valid {
			f64, err := row.Volume.Float64Value()
			if err != nil {
				return nil, fmt.Errorf("failed to convert volume: %w", err)
			}
			volume = f64.Float64
		}

		// Convert exercise_order and set_order from int32 to *int32
		var exerciseOrder *int32
		if row.ExerciseOrder != 0 {
			exerciseOrder = &row.ExerciseOrder
		}

		var setOrder *int32
		if row.SetOrder != 0 {
			setOrder = &row.SetOrder
		}

		// Convert workout_notes from pgtype.Text to *string
		var workoutNotes *string
		if row.WorkoutNotes.Valid {
			workoutNotes = &row.WorkoutNotes.String
		}

		historical1rm, err := floatPtrFromNumeric(row.Historical1rm)
		if err != nil {
			return nil, err
		}

		var historical1rmUpdatedAt *time.Time
		if row.Historical1rmUpdatedAt.Valid {
			t := row.Historical1rmUpdatedAt.Time
			historical1rmUpdatedAt = &t
		}

		var historical1rmSourceWorkoutID *int32
		if row.Historical1rmSourceWorkoutID.Valid {
			id := row.Historical1rmSourceWorkoutID.Int32
			historical1rmSourceWorkoutID = &id
		}

		response[i] = ExerciseWithSetsResponse{
			WorkoutID:                    row.WorkoutID,
			WorkoutDate:                  row.WorkoutDate.Time,
			WorkoutNotes:                 workoutNotes,
			SetID:                        row.SetID,
			Weight:                       weight,
			Reps:                         row.Reps,
			SetType:                      row.SetType,
			ExerciseID:                   row.ExerciseID,
			ExerciseName:                 row.ExerciseName,
			Historical1RM:                historical1rm,
			Historical1RMUpdatedAt:       historical1rmUpdatedAt,
			Historical1RMSourceWorkoutID: historical1rmSourceWorkoutID,
			ExerciseOrder:                exerciseOrder,
			SetOrder:                     setOrder,
			Volume:                       volume,
		}
	}

	return response, nil
}
