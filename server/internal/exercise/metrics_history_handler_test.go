package exercise

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExerciseHandler_GetExerciseMetricsHistory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(ioDiscard{}, nil))
	validate := validator.New()

	userID := "user-123"
	ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

	t.Run("400 invalid range", func(t *testing.T) {
		mockRepo := new(MockExerciseRepository)
		service := NewService(logger, mockRepo)
		handler := NewHandler(logger, validate, service)

		req := httptest.NewRequest(http.MethodGet, "/api/exercises/1/metrics-history?range=BAD", nil).WithContext(ctx)
		req.SetPathValue("id", "1")

		w := httptest.NewRecorder()
		handler.GetExerciseMetricsHistory(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("200 default range M", func(t *testing.T) {
		mockRepo := new(MockExerciseRepository)
		service := NewService(logger, mockRepo)
		handler := NewHandler(logger, validate, service)

		exerciseID := int32(1)
		mockRepo.On("GetExercise", mock.Anything, exerciseID, userID).Return(db.Exercise{ID: exerciseID, Name: "Bench"}, nil)

		wantReq := GetExerciseMetricsHistoryRequest{ExerciseID: exerciseID, Range: "M"}
		points := []ExerciseMetricsHistoryPoint{
			{
				X:                    "1",
				Date:                 time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				WorkoutID:            ptrInt32(1),
				SessionBestE1RM:      250,
				SessionAvgE1RM:       230,
				SessionAvgIntensity:  85,
				SessionBestIntensity: 102,
				TotalVolumeWorking:   12000,
			},
		}

		mockRepo.
			On("GetExerciseMetricsHistory", mock.Anything, wantReq, userID).
			Return(points, MetricsHistoryBucketWorkout, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/exercises/1/metrics-history", nil).WithContext(ctx)
		req.SetPathValue("id", "1")

		w := httptest.NewRecorder()
		handler.GetExerciseMetricsHistory(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("404 when exercise missing", func(t *testing.T) {
		mockRepo := new(MockExerciseRepository)
		service := NewService(logger, mockRepo)
		handler := NewHandler(logger, validate, service)

		exerciseID := int32(99)
		mockRepo.On("GetExercise", mock.Anything, exerciseID, userID).Return(db.Exercise{}, errors.New("missing"))

		req := httptest.NewRequest(http.MethodGet, "/api/exercises/99/metrics-history?range=W", nil).WithContext(ctx)
		req.SetPathValue("id", "99")

		w := httptest.NewRecorder()
		handler.GetExerciseMetricsHistory(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func ptrInt32(v int32) *int32 { return &v }
