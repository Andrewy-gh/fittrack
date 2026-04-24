package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
)

type TxSaver interface {
	SaveWorkoutTx(ctx context.Context, qtx *db.Queries, requestBody CreateWorkoutRequest, userID string) (int32, error)
}

type txSaver struct {
	logger       *slog.Logger
	exerciseRepo exercise.ExerciseRepository
}

func NewTxSaver(logger *slog.Logger, exerciseRepo exercise.ExerciseRepository) TxSaver {
	return &txSaver{
		logger:       logger,
		exerciseRepo: exerciseRepo,
	}
}

func (s *txSaver) SaveWorkoutTx(ctx context.Context, qtx *db.Queries, requestBody CreateWorkoutRequest, userID string) (int32, error) {
	reformatted, err := transformWorkoutRequest(s.logger, requestBody, false)
	if err != nil {
		return 0, fmt.Errorf("failed to transform request: %w", err)
	}

	pgData, err := convertToPGTypes(reformatted)
	if err != nil {
		s.logger.Error("failed to convert to PG types", "error", err)
		return 0, fmt.Errorf("failed to convert to PG types: %w", err)
	}

	workoutRow, err := insertWorkout(ctx, qtx, pgData.Workout, userID)
	if err != nil {
		s.logger.Error("failed to insert workout", "error", err)
		return 0, fmt.Errorf("failed to insert workout: %w", err)
	}

	exerciseMap, err := getOrCreateExercises(ctx, s.logger, s.exerciseRepo, qtx, pgData.Exercises, userID)
	if err != nil {
		s.logger.Error("failed to get/create exercises", "error", err)
		return 0, fmt.Errorf("failed to get/create exercises: %w", err)
	}

	if err := insertSets(ctx, s.logger, qtx, pgData.Sets, workoutRow.ID, exerciseMap, userID); err != nil {
		s.logger.Error("failed to insert sets", "error", err)
		return 0, fmt.Errorf("failed to insert sets: %w", err)
	}

	if err := updateHistorical1rmFromWorkout(ctx, qtx, workoutRow.ID, userID); err != nil {
		s.logger.Error("failed to update historical 1RM from workout", "error", err, "workout_id", workoutRow.ID)
		return 0, fmt.Errorf("failed to update historical 1RM from workout: %w", err)
	}

	return workoutRow.ID, nil
}

var _ TxSaver = (*txSaver)(nil)
