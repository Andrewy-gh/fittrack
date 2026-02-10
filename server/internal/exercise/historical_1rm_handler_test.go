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
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
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

func TestExerciseHandler_GetExerciseHistorical1RM(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()

	t.Run("success_with_computed_best", func(t *testing.T) {
		repo := new(MockExerciseRepository)
		service := NewService(logger, repo)
		handler := NewHandler(logger, validator, service)

		userID := "user-123"
		exerciseID := int32(7)

		repo.On("GetExercise", mock.Anything, exerciseID, userID).
			Return(db.Exercise{ID: exerciseID, Name: "Bench"}, nil)

		ts := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
		repo.On("GetExerciseHistorical1RM", mock.Anything, exerciseID, userID).
			Return(db.GetExerciseHistorical1RMRow{
				Historical1rm: mustNumeric(t, "250.00"),
				Historical1rmUpdatedAt: pgtype.Timestamptz{
					Time:  ts,
					Valid: true,
				},
				Historical1rmSourceWorkoutID: pgtype.Int4{Int32: 99, Valid: true},
			}, nil)

		repo.On("GetExerciseBestE1rmWithWorkout", mock.Anything, exerciseID, userID).
			Return(db.GetExerciseBestE1rmWithWorkoutRow{
				WorkoutID: 123,
				E1rm:      mustNumeric(t, "255.50"),
			}, nil)

		ctx := context.WithValue(context.Background(), user.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodGet, "/api/exercises/7/historical-1rm", nil).WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(exerciseID)))
		w := httptest.NewRecorder()

		handler.GetExerciseHistorical1RM(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var got ExerciseHistorical1RMResponse
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))

		if assert.NotNil(t, got.Historical1RM) {
			assert.InDelta(t, 250.0, *got.Historical1RM, 0.001)
		}
		if assert.NotNil(t, got.Historical1RMUpdatedAt) {
			assert.True(t, got.Historical1RMUpdatedAt.Equal(ts))
		}
		if assert.NotNil(t, got.Historical1RMSourceWorkoutID) {
			assert.Equal(t, int32(99), *got.Historical1RMSourceWorkoutID)
		}
		if assert.NotNil(t, got.ComputedBestE1RM) {
			assert.InDelta(t, 255.5, *got.ComputedBestE1RM, 0.001)
		}
		if assert.NotNil(t, got.ComputedBestWorkoutID) {
			assert.Equal(t, int32(123), *got.ComputedBestWorkoutID)
		}

		repo.AssertExpectations(t)
	})

	t.Run("success_no_working_sets", func(t *testing.T) {
		repo := new(MockExerciseRepository)
		service := NewService(logger, repo)
		handler := NewHandler(logger, validator, service)

		userID := "user-123"
		exerciseID := int32(7)

		repo.On("GetExercise", mock.Anything, exerciseID, userID).
			Return(db.Exercise{ID: exerciseID, Name: "Bench"}, nil)
		repo.On("GetExerciseHistorical1RM", mock.Anything, exerciseID, userID).
			Return(db.GetExerciseHistorical1RMRow{
				Historical1rm:                pgtype.Numeric{Valid: false},
				Historical1rmUpdatedAt:       pgtype.Timestamptz{Valid: false},
				Historical1rmSourceWorkoutID: pgtype.Int4{Valid: false},
			}, nil)
		repo.On("GetExerciseBestE1rmWithWorkout", mock.Anything, exerciseID, userID).
			Return(db.GetExerciseBestE1rmWithWorkoutRow{}, pgx.ErrNoRows)

		ctx := context.WithValue(context.Background(), user.UserIDKey, userID)
		req := httptest.NewRequest(http.MethodGet, "/api/exercises/7/historical-1rm", nil).WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(exerciseID)))
		w := httptest.NewRecorder()

		handler.GetExerciseHistorical1RM(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var got ExerciseHistorical1RMResponse
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Nil(t, got.ComputedBestE1RM)
		assert.Nil(t, got.ComputedBestWorkoutID)

		repo.AssertExpectations(t)
	})
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
