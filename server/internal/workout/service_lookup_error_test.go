package workout

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorkoutService_UpdateWorkout_LookupErrorClassification(t *testing.T) {
	userID := "user-123"
	workoutID := int32(1)
	request := UpdateWorkoutRequest{
		Date: time.Now().UTC().Format(time.RFC3339),
		Exercises: []UpdateExercise{{
			Name: "Bench Press",
			Sets: []UpdateSet{{Reps: 5, SetType: "working"}},
		}},
	}

	tests := []struct {
		name         string
		lookupError  error
		wantNotFound bool
	}{
		{
			name:         "no rows returns not found",
			lookupError:  pgx.ErrNoRows,
			wantNotFound: true,
		},
		{
			name:        "database failure is wrapped",
			lookupError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWorkoutRepository)
			mockRepo.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{}, tt.lookupError)

			service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockRepo)
			ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

			err := service.UpdateWorkout(ctx, workoutID, request)
			require.Error(t, err)

			var notFoundErr *apperrors.NotFound
			if tt.wantNotFound {
				require.True(t, errors.As(err, &notFoundErr), "expected NotFound but got %T", err)
			} else {
				require.False(t, errors.As(err, &notFoundErr), "unexpected NotFound for %v", err)
				require.ErrorIs(t, err, tt.lookupError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_DeleteWorkout_LookupErrorClassification(t *testing.T) {
	userID := "user-123"
	workoutID := int32(1)

	tests := []struct {
		name         string
		lookupError  error
		wantNotFound bool
	}{
		{
			name:         "no rows returns not found",
			lookupError:  pgx.ErrNoRows,
			wantNotFound: true,
		},
		{
			name:        "database failure is wrapped",
			lookupError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWorkoutRepository)
			mockRepo.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{}, tt.lookupError)

			service := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), mockRepo)
			ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

			err := service.DeleteWorkout(ctx, workoutID)
			require.Error(t, err)

			var notFoundErr *apperrors.NotFound
			if tt.wantNotFound {
				require.True(t, errors.As(err, &notFoundErr), "expected NotFound but got %T", err)
			} else {
				require.False(t, errors.As(err, &notFoundErr), "unexpected NotFound for %v", err)
				require.ErrorIs(t, err, tt.lookupError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
