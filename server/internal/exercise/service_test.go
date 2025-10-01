package exercise

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



func TestExerciseService_GetRecentSetsForExercise(t *testing.T) {
	mockRepo := new(MockExerciseRepository)
	service := NewService(nil, mockRepo)

	exerciseID := int32(1)
	userID := "user-123"

	expectedSets := []db.GetRecentSetsForExerciseRow{
		{SetID: 1, Reps: 10},
		{SetID: 2, Reps: 8},
		{SetID: 3, Reps: 12},
	}

	mockRepo.On("GetRecentSetsForExercise", mock.Anything, exerciseID, userID).Return(expectedSets, nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, user.UserIDKey, userID)

	sets, err := service.GetRecentSetsForExercise(ctx, exerciseID)
	if err != nil {
		t.Errorf("error was not expected while getting recent sets: %s", err)
	}

	if len(sets) != 3 {
		t.Errorf("expected 3 sets, but got %d", len(sets))
	}

	mockRepo.AssertExpectations(t)
}

func TestExerciseService_DeleteExercise(t *testing.T) {
	tests := []struct {
		name          string
		exerciseID    int32
		userID        string
		setupMock     func(*MockExerciseRepository)
		expectError   bool
		errorType     string
	}{
		{
			name:       "successful deletion",
			exerciseID: 1,
			userID:     "user-123",
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetExercise", mock.Anything, int32(1), "user-123").Return(db.Exercise{ID: 1, Name: "Bench Press"}, nil)
				m.On("DeleteExercise", mock.Anything, int32(1), "user-123").Return(nil)
			},
			expectError: false,
		},
		{
			name:       "exercise not found",
			exerciseID: 999,
			userID:     "user-123",
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetExercise", mock.Anything, int32(999), "user-123").Return(db.Exercise{}, assert.AnError)
			},
			expectError: true,
			errorType:   "ErrNotFound",
		},
		{
			name:       "unauthenticated user",
			exerciseID: 1,
			userID:     "",
			setupMock: func(m *MockExerciseRepository) {
				// No mock setup needed as authentication check happens first
			},
			expectError: true,
			errorType:   "ErrUnauthorized",
		},
		{
			name:       "delete operation fails",
			exerciseID: 1,
			userID:     "user-123",
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetExercise", mock.Anything, int32(1), "user-123").Return(db.Exercise{ID: 1, Name: "Bench Press"}, nil)
				m.On("DeleteExercise", mock.Anything, int32(1), "user-123").Return(assert.AnError)
			},
			expectError: true,
			errorType:   "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockExerciseRepository)
			tt.setupMock(mockRepo)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			service := NewService(logger, mockRepo)

			ctx := context.Background()
			if tt.userID != "" {
				ctx = context.WithValue(ctx, user.UserIDKey, tt.userID)
			}

			err := service.DeleteExercise(ctx, tt.exerciseID)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				switch tt.errorType {
				case "ErrUnauthorized":
					if _, ok := err.(*ErrUnauthorized); !ok {
						t.Errorf("expected ErrUnauthorized but got %T", err)
					}
				case "ErrNotFound":
					if _, ok := err.(*ErrNotFound); !ok {
						t.Errorf("expected ErrNotFound but got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

