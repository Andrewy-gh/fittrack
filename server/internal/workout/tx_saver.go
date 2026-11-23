package workout

import (
	"context"
	"fmt"
	"log/slog"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
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
	reformatted, err := transformWorkoutRequest(s.logger, newCreateWorkoutDraft(requestBody))
	if err != nil {
		return 0, fmt.Errorf("failed to transform request: %w", err)
	}

	reformatted.Recommendations = requestBody.Recommendations
	return saveReformattedWorkoutTx(ctx, qtx, s.logger, s.exerciseRepo, reformatted, userID)
}

// saveReformattedWorkoutTx writes a normalized workout in the caller's transaction.
func saveReformattedWorkoutTx(ctx context.Context, qtx *db.Queries, logger *slog.Logger, exerciseRepo exercise.ExerciseRepository, reformatted *ReformattedRequest, userID string) (int32, error) {
	pgData, err := convertToPGTypes(reformatted)
	if err != nil {
		logger.Error("failed to convert to PG types", "error", err)
		return 0, fmt.Errorf("failed to convert to PG types: %w", err)
	}

	workoutRow, err := insertWorkout(ctx, qtx, pgData.Workout, userID)
	if err != nil {
		logger.Error("failed to insert workout", "error", err)
		return 0, fmt.Errorf("failed to insert workout: %w", err)
	}

	exerciseMap, err := getOrCreateExercises(ctx, logger, exerciseRepo, qtx, pgData.Exercises, userID)
	if err != nil {
		logger.Error("failed to get/create exercises", "error", err)
		return 0, fmt.Errorf("failed to get/create exercises: %w", err)
	}

	if err := insertSets(ctx, logger, qtx, pgData.Sets, workoutRow.ID, exerciseMap, userID); err != nil {
		logger.Error("failed to insert sets", "error", err)
		return 0, fmt.Errorf("failed to insert sets: %w", err)
	}

	if err := recommendation.SaveContext(ctx, qtx, userID, workoutRow.ID, exerciseMap, reformatted.Recommendations); err != nil {
		return 0, fmt.Errorf("save recommendation context: %w", err)
	}
	if err := updateHistorical1rmFromWorkout(ctx, qtx, workoutRow.ID, userID); err != nil {
		logger.Error("failed to update historical 1RM from workout", "error", err, "workout_id", workoutRow.ID)
		return 0, fmt.Errorf("failed to update historical 1RM from workout: %w", err)
	}

	return workoutRow.ID, nil
}

var _ TxSaver = (*txSaver)(nil)
