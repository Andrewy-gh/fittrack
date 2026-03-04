package exercise

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mustNumeric(t *testing.T, s string) pgtype.Numeric {
	t.Helper()
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		t.Fatalf("failed to scan numeric %q: %v", s, err)
	}
	return n
}

func TestExerciseHandler_UpdateExerciseHistorical1RM(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()

	t.Run("manual_set", func(t *testing.T) {
		repo := new(MockExerciseRepository)
		service := NewService(logger, repo)
		handler := NewHandler(logger, validator, service)

		userID := "user-123"
		exerciseID := int32(7)
		val := 315.5

		repo.On("GetExercise", mock.Anything, exerciseID, userID).
			Return(db.Exercise{ID: exerciseID, Name: "Bench"}, nil)
		repo.On("UpdateExerciseHistorical1RMManual", mock.Anything, exerciseID, mock.AnythingOfType("*float64"), userID).
			Return(nil).
			Run(func(args mock.Arguments) {
				p := args.Get(2).(*float64)
				assert.InDelta(t, val, *p, 0.001)
			})

		body, _ := json.Marshal(UpdateExerciseHistorical1RMRequest{Mode: "manual", Historical1RM: &val})
		ctx := context.WithValue(context.Background(), user.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodPatch, "/api/exercises/7/historical-1rm", bytes.NewBuffer(body)).WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(exerciseID)))
		w := httptest.NewRecorder()

		handler.UpdateExerciseHistorical1RM(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		repo.AssertExpectations(t)
	})

	t.Run("recompute_sets_source_workout", func(t *testing.T) {
		repo := new(MockExerciseRepository)
		service := NewService(logger, repo)
		handler := NewHandler(logger, validator, service)

		userID := "user-123"
		exerciseID := int32(7)
		workoutID := int32(55)

		repo.On("GetExercise", mock.Anything, exerciseID, userID).
			Return(db.Exercise{ID: exerciseID, Name: "Bench"}, nil)
		repo.On("GetExerciseBestE1rmWithWorkout", mock.Anything, exerciseID, userID).
			Return(db.GetExerciseBestE1rmWithWorkoutRow{WorkoutID: workoutID, E1rm: mustNumeric(t, "300.00")}, nil)
		repo.On("SetExerciseHistorical1RM", mock.Anything, exerciseID, mock.AnythingOfType("*float64"), &workoutID, userID).
			Return(nil).
			Run(func(args mock.Arguments) {
				p := args.Get(2).(*float64)
				assert.InDelta(t, 300.0, *p, 0.001)
			})

		body, _ := json.Marshal(UpdateExerciseHistorical1RMRequest{Mode: "recompute"})
		ctx := context.WithValue(context.Background(), user.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodPatch, "/api/exercises/7/historical-1rm", bytes.NewBuffer(body)).WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(exerciseID)))
		w := httptest.NewRecorder()

		handler.UpdateExerciseHistorical1RM(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		repo.AssertExpectations(t)
	})

	t.Run("validation_error_bad_mode", func(t *testing.T) {
		repo := new(MockExerciseRepository)
		service := NewService(logger, repo)
		handler := NewHandler(logger, validator, service)

		userID := "user-123"
		exerciseID := int32(7)

		body := []byte(`{"mode":"bad"}`)
		ctx := context.WithValue(context.Background(), user.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodPatch, "/api/exercises/7/historical-1rm", bytes.NewBuffer(body)).WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(exerciseID)))
		w := httptest.NewRecorder()

		handler.UpdateExerciseHistorical1RM(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
