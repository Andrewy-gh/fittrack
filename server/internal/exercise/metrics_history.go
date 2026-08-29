package exercise

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

type MetricsHistoryBucket string

const (
	MetricsHistoryBucketWorkout MetricsHistoryBucket = "workout"
)

type ExerciseMetricsHistoryPoint struct {
	X                    string    `json:"x"`
	Date                 time.Time `json:"date"`
	WorkoutID            *int32    `json:"workout_id,omitempty"`
	SessionBestE1RM      float64   `json:"session_best_e1rm"`
	SessionAvgE1RM       float64   `json:"session_avg_e1rm"`
	SessionAvgIntensity  float64   `json:"session_avg_intensity"`
	SessionBestIntensity float64   `json:"session_best_intensity"`
	TotalVolumeWorking   float64   `json:"total_volume_working"`
}

type ExerciseMetricsHistoryResponse struct {
	Bucket MetricsHistoryBucket          `json:"bucket"`
	Points []ExerciseMetricsHistoryPoint `json:"points"`
}

func (es *ExerciseService) GetExerciseMetricsHistory(ctx context.Context, exerciseID int32) (*ExerciseMetricsHistoryResponse, error) {
	userID, ok := user.Current(ctx)
	if !ok {
		return nil, &apperrors.Unauthorized{Resource: "exercise", UserID: ""}
	}

	_, err := es.repo.GetExercise(ctx, exerciseID, userID)
	if err != nil {
		return nil, &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprintf("%d", exerciseID)}
	}

	points, bucket, err := es.repo.GetExerciseMetricsHistory(ctx, exerciseID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise metrics history: %w", err)
	}

	return &ExerciseMetricsHistoryResponse{
		Bucket: bucket,
		Points: points,
	}, nil
}
