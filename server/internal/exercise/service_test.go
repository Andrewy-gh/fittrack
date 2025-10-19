package exercise

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
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

func TestExerciseService_UpdateExerciseName(t *testing.T) {
	tests := []struct {
		name        string
		exerciseID  int32
		newName     string
		userID      string
		setupMock   func(*MockExerciseRepository)
		expectError bool
		errorType   string
	}{
		{
			name:       "successful update",
			exerciseID: 1,
			newName:    "Updated Exercise",
			userID:     "user-123",
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetExercise", mock.Anything, int32(1), "user-123").Return(db.Exercise{ID: 1, Name: "Old Exercise"}, nil)
				m.On("UpdateExerciseName", mock.Anything, int32(1), "Updated Exercise", "user-123").Return(nil)
			},
			expectError: false,
		},
		{
			name:       "exercise not found",
			exerciseID: 999,
			newName:    "Updated Exercise",
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
			newName:    "Updated Exercise",
			userID:     "",
			setupMock: func(m *MockExerciseRepository) {
				// No mock setup needed as authentication check happens first
			},
			expectError: true,
			errorType:   "ErrUnauthorized",
		},
		{
			name:       "update operation fails",
			exerciseID: 1,
			newName:    "Updated Exercise",
			userID:     "user-123",
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetExercise", mock.Anything, int32(1), "user-123").Return(db.Exercise{ID: 1, Name: "Old Exercise"}, nil)
				m.On("UpdateExerciseName", mock.Anything, int32(1), "Updated Exercise", "user-123").Return(assert.AnError)
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

			err := service.UpdateExerciseName(ctx, tt.exerciseID, tt.newName)

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

func TestGroupSetsByWorkout(t *testing.T) {
	tests := []struct {
		name     string
		sets     []db.GetExerciseWithSetsRow
		expected []WorkoutGroup
	}{
		{
			name: "empty sets",
			sets: []db.GetExerciseWithSetsRow{},
			expected: []WorkoutGroup{},
		},
		{
			name: "single workout with one set",
			sets: []db.GetExerciseWithSetsRow{
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        100,
					Weight:       mustParseInt4(225),
					Reps:         5,
					SetType:      "normal",
					Volume:       1125,
				},
			},
			expected: []WorkoutGroup{
				{
					WorkoutID:   1,
					Date:        "2025-01-15T10:00:00Z",
					Notes:       "Heavy day",
					TotalReps:   5,
					TotalVolume: 1125,
					Sets: []SetInfo{
						{SetID: 100, Weight: 225, Reps: 5, SetType: "normal"},
					},
				},
			},
		},
		{
			name: "single workout with multiple sets",
			sets: []db.GetExerciseWithSetsRow{
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        100,
					Weight:       mustParseInt4(225),
					Reps:         5,
					SetType:      "normal",
					Volume:       1125,
				},
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        101,
					Weight:       mustParseInt4(225),
					Reps:         5,
					SetType:      "normal",
					Volume:       1125,
				},
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        102,
					Weight:       mustParseInt4(225),
					Reps:         4,
					SetType:      "normal",
					Volume:       900,
				},
			},
			expected: []WorkoutGroup{
				{
					WorkoutID:   1,
					Date:        "2025-01-15T10:00:00Z",
					Notes:       "Heavy day",
					TotalReps:   14,
					TotalVolume: 3150,
					Sets: []SetInfo{
						{SetID: 100, Weight: 225, Reps: 5, SetType: "normal"},
						{SetID: 101, Weight: 225, Reps: 5, SetType: "normal"},
						{SetID: 102, Weight: 225, Reps: 4, SetType: "normal"},
					},
				},
			},
		},
		{
			name: "multiple workouts",
			sets: []db.GetExerciseWithSetsRow{
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        100,
					Weight:       mustParseInt4(225),
					Reps:         5,
					SetType:      "normal",
					Volume:       1125,
				},
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Heavy day"),
					SetID:        101,
					Weight:       mustParseInt4(225),
					Reps:         5,
					SetType:      "normal",
					Volume:       1125,
				},
				{
					WorkoutID:    2,
					WorkoutDate:  mustParseTimestamp("2025-01-17T14:00:00Z"),
					WorkoutNotes: mustParseText("Light day"),
					SetID:        200,
					Weight:       mustParseInt4(135),
					Reps:         10,
					SetType:      "normal",
					Volume:       1350,
				},
				{
					WorkoutID:    2,
					WorkoutDate:  mustParseTimestamp("2025-01-17T14:00:00Z"),
					WorkoutNotes: mustParseText("Light day"),
					SetID:        201,
					Weight:       mustParseInt4(135),
					Reps:         10,
					SetType:      "normal",
					Volume:       1350,
				},
			},
			expected: []WorkoutGroup{
				{
					WorkoutID:   1,
					Date:        "2025-01-15T10:00:00Z",
					Notes:       "Heavy day",
					TotalReps:   10,
					TotalVolume: 2250,
					Sets: []SetInfo{
						{SetID: 100, Weight: 225, Reps: 5, SetType: "normal"},
						{SetID: 101, Weight: 225, Reps: 5, SetType: "normal"},
					},
				},
				{
					WorkoutID:   2,
					Date:        "2025-01-17T14:00:00Z",
					Notes:       "Light day",
					TotalReps:   20,
					TotalVolume: 2700,
					Sets: []SetInfo{
						{SetID: 200, Weight: 135, Reps: 10, SetType: "normal"},
						{SetID: 201, Weight: 135, Reps: 10, SetType: "normal"},
					},
				},
			},
		},
		{
			name: "workout with null weight",
			sets: []db.GetExerciseWithSetsRow{
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseText("Bodyweight"),
					SetID:        100,
					Weight:       mustParseInt4Null(),
					Reps:         10,
					SetType:      "normal",
					Volume:       0,
				},
			},
			expected: []WorkoutGroup{
				{
					WorkoutID:   1,
					Date:        "2025-01-15T10:00:00Z",
					Notes:       "Bodyweight",
					TotalReps:   10,
					TotalVolume: 0,
					Sets: []SetInfo{
						{SetID: 100, Weight: 0, Reps: 10, SetType: "normal"},
					},
				},
			},
		},
		{
			name: "workout with null notes",
			sets: []db.GetExerciseWithSetsRow{
				{
					WorkoutID:    1,
					WorkoutDate:  mustParseTimestamp("2025-01-15T10:00:00Z"),
					WorkoutNotes: mustParseTextNull(),
					SetID:        100,
					Weight:       mustParseInt4(135),
					Reps:         10,
					SetType:      "normal",
					Volume:       1350,
				},
			},
			expected: []WorkoutGroup{
				{
					WorkoutID:   1,
					Date:        "2025-01-15T10:00:00Z",
					Notes:       "",
					TotalReps:   10,
					TotalVolume: 1350,
					Sets: []SetInfo{
						{SetID: 100, Weight: 135, Reps: 10, SetType: "normal"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := groupSetsByWorkout(tt.sets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for test data
func mustParseTimestamp(s string) pgtype.Timestamptz {
	var t pgtype.Timestamptz
	err := t.Scan(s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustParseText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func mustParseTextNull() pgtype.Text {
	return pgtype.Text{String: "", Valid: false}
}

func mustParseInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func mustParseInt4Null() pgtype.Int4 {
	return pgtype.Int4{Int32: 0, Valid: false}
}

