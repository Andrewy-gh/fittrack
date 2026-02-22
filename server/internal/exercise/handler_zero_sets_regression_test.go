package exercise

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExerciseHandler_GetExerciseWithSets_ZeroSetsResponseShapeRegression(t *testing.T) {
	const (
		userID     = "test-user-id"
		exerciseID = int32(99)
	)

	t.Run("empty slice stays json array", func(t *testing.T) {
		mockRepo := &MockExerciseRepository{}
		mockRepo.On("GetExerciseDetail", mock.Anything, exerciseID, userID).Return(db.GetExerciseDetailRow{
			ID:        exerciseID,
			Name:      "Zero Set Exercise",
			UserID:    userID,
			CreatedAt: pgtype.Timestamptz{Time: time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC), Valid: true},
		}, nil)
		mockRepo.On("GetExerciseWithSets", mock.Anything, exerciseID, userID).Return([]db.GetExerciseWithSetsRow{}, nil)

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := NewHandler(logger, validator.New(), NewService(logger, mockRepo))

		req := httptest.NewRequest(http.MethodGet, "/api/exercises/99", nil)
		req = req.WithContext(context.WithValue(context.Background(), user.UserIDKey, userID))
		req.SetPathValue("id", "99")
		w := httptest.NewRecorder()

		handler.GetExerciseWithSets(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var raw map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
		require.Contains(t, raw, "exercise")
		require.Contains(t, raw, "sets")
		sets, ok := raw["sets"].([]any)
		require.True(t, ok, "sets should be a JSON array")
		assert.Len(t, sets, 0)

		var typedResp ExerciseDetailResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &typedResp))
		assert.Equal(t, exerciseID, typedResp.Exercise.ID)
		assert.Nil(t, typedResp.Exercise.BestE1RM)
		require.NotNil(t, typedResp.Sets, "sets should be present as [] instead of null")
		assert.Len(t, typedResp.Sets, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("nil slice stays json array", func(t *testing.T) {
		mockRepo := &MockExerciseRepository{}
		mockRepo.On("GetExerciseDetail", mock.Anything, exerciseID, userID).Return(db.GetExerciseDetailRow{
			ID:        exerciseID,
			Name:      "Zero Set Exercise",
			UserID:    userID,
			CreatedAt: pgtype.Timestamptz{Time: time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC), Valid: true},
		}, nil)
		mockRepo.On("GetExerciseWithSets", mock.Anything, exerciseID, userID).Return(([]db.GetExerciseWithSetsRow)(nil), nil)

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := NewHandler(logger, validator.New(), NewService(logger, mockRepo))

		req := httptest.NewRequest(http.MethodGet, "/api/exercises/99", nil)
		req = req.WithContext(context.WithValue(context.Background(), user.UserIDKey, userID))
		req.SetPathValue("id", "99")
		w := httptest.NewRecorder()

		handler.GetExerciseWithSets(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var raw map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
		sets, ok := raw["sets"].([]any)
		require.True(t, ok, "sets should be a JSON array")
		assert.Len(t, sets, 0)

		var typedResp ExerciseDetailResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &typedResp))
		assert.Nil(t, typedResp.Exercise.BestE1RM)
		require.NotNil(t, typedResp.Sets, "sets should be present as [] instead of null")
		assert.Len(t, typedResp.Sets, 0)

		mockRepo.AssertExpectations(t)
	})
}
