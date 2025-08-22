package workout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkoutHandler_UpdateWorkout(t *testing.T) {
	userID := "test-user-id"
	workoutID := int32(1)

	// Helper function for string pointers
	stringPtr := func(s string) *string { return &s }
	// Helper function for int pointers
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name          string
		workoutID     string
		requestBody   interface{}
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectedError string
	}{
		{
			name:      "successful full update",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr("Updated workout notes"),
				Exercises: []UpdateExercise{
					{
						Name: "Updated Exercise",
						Sets: []UpdateSet{
							{Weight: intPtr(225), Reps: 8, SetType: "working"},
						},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				// Mock GetWorkout call (service checks if workout exists)
				m.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{ID: workoutID}, nil)
				m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.ReformattedRequest"), userID).
					Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusNoContent,
		},
		{
			name:      "successful partial update - notes only",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr("Just updating notes"),
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{ID: workoutID}, nil)
				m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.ReformattedRequest"), userID).
					Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusNoContent,
		},
		{
			name:      "successful partial update - date only",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date: "2023-01-15T15:30:00Z",
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{ID: workoutID}, nil)
				m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.ReformattedRequest"), userID).
					Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusNoContent,
		},
		{
			name:          "invalid workout ID - non-numeric",
			workoutID:     "invalid",
			requestBody:   UpdateWorkoutRequest{
				Date: "2023-01-15T10:00:00Z",
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid workout ID",
		},
		{
			name:          "missing workout ID",
			workoutID:     "",
			requestBody:   UpdateWorkoutRequest{
				Date: "2023-01-15T10:00:00Z",
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Missing workout ID",
		},
		{
			name:          "invalid JSON body",
			workoutID:     "1",
			requestBody:   "invalid json string",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "failed to decode request body",
		},
		{
			name:      "validation error - invalid date format",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date: "invalid-date-format",
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation error occurred",
		},
		{
			name:      "validation error - notes too long",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr(string(make([]byte, 500))), // Exceeds 256 char limit
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation error occurred",
		},
		{
			name:      "validation error - exercise missing name",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:      "2023-01-15T10:00:00Z",
				Exercises: []UpdateExercise{
					{
						// Missing Name field
						Sets: []UpdateSet{
							{Reps: 10, SetType: "working"},
						},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation error occurred",
		},
		{
			name:      "validation error - set missing required fields",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:      "2023-01-15T10:00:00Z",
				Exercises: []UpdateExercise{
					{
						Name: "Valid Exercise",
						Sets: []UpdateSet{
							{
								// Missing Reps and SetType
							},
						},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "validation error occurred",
		},
		{
			name:      "service error - workout not found",
			workoutID: "999",
			requestBody: UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr("Updating non-existent workout"),
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				// GetWorkout returns error for non-existent workout
				m.On("GetWorkout", mock.Anything, int32(999), userID).Return(db.Workout{}, fmt.Errorf("workout not found"))
				// UpdateWorkout should not be called since GetWorkout returns error
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusNotFound,
			expectedError: "workout not found",
		},
		{
			name:      "service error - database failure",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr("Valid update"),
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetWorkout", mock.Anything, workoutID, userID).Return(db.Workout{ID: workoutID}, nil)
				m.On("UpdateWorkout", mock.Anything, workoutID, mock.AnythingOfType("*workout.ReformattedRequest"), userID).
					Return(assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to update workout",
		},
		{
			name:          "unauthenticated user",
			workoutID:     "1",
			requestBody:   UpdateWorkoutRequest{
				Date:  "2023-01-15T10:00:00Z",
				Notes: stringPtr("Unauthorized update"),
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockWorkoutRepository{}
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			// Prepare request
			var req *http.Request
			if tt.requestBody != nil {
				body, err := json.Marshal(tt.requestBody)
				if err != nil {
					// For the "invalid json string" test case
					req = httptest.NewRequest("PUT", "/api/workouts/"+tt.workoutID, bytes.NewBufferString(tt.requestBody.(string)))
				} else {
					req = httptest.NewRequest("PUT", "/api/workouts/"+tt.workoutID, bytes.NewBuffer(body))
				}
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("PUT", "/api/workouts/"+tt.workoutID, nil)
			}

			if tt.workoutID != "" {
				req.SetPathValue("id", tt.workoutID)
			}
			w := httptest.NewRecorder()

			// Execute
			handler.UpdateWorkout(w, req.WithContext(tt.ctx))

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedError != "" {
				var resp errorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Message, tt.expectedError)
			} else if tt.expectedCode == http.StatusNoContent {
				// Success case - should have no content
				assert.Empty(t, w.Body.String())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutHandler_UpdateWorkout_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Setup test data
	userAID := "test-user-a"
	userBID := "test-user-b"
	
	// Create test workout for User A
	workoutID := setupTestWorkout(t, pool, userAID, "Original workout")
	
	// Initialize components with real database
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	queries := db.New(pool)
	
	// Initialize repositories
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := NewRepository(logger, queries, pool, exerciseRepo)
	workoutService := NewService(logger, workoutRepo)
	handler := NewHandler(logger, validator, workoutService)

	t.Run("UserA_CanUpdateOwnWorkout", func(t *testing.T) {
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		updateReq := UpdateWorkoutRequest{
			Date:  "2023-01-15T10:00:00Z",
			Notes: stringPtr("Updated notes via integration test"),
			Exercises: []UpdateExercise{
				{
					Name: "placeholder",
					Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
				},
			},
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(workoutID)))

		w := httptest.NewRecorder()
		handler.UpdateWorkout(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		
		// Verify the update by checking the workout directly from database
		// Since our test workout has no sets, GetWorkoutWithSets would return empty
		// Let's create a workout with sets in this test, or verify via database directly
		var actualNotes string
		err := pool.QueryRow(ctx, "SELECT notes FROM workout WHERE id = $1 AND user_id = $2",
			workoutID, userAID).Scan(&actualNotes)
		assert.NoError(t, err)
		assert.Equal(t, "Updated notes via integration test", actualNotes)
	})

	t.Run("UserB_CannotUpdateUserA_Workout", func(t *testing.T) {
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		updateReq := UpdateWorkoutRequest{
			Date:  "2023-01-15T10:00:00Z",
			Notes: stringPtr("Malicious update attempt from User B"),
			Exercises: []UpdateExercise{
				{
					Name: "placeholder",
					Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
				},
			},
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(workoutID)))

		w := httptest.NewRecorder()
		handler.UpdateWorkout(w, req)

		// Should fail due to RLS - workout not found for userB
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		// Double-check userA's workout was not modified
		ctxA := setTestUserContext(context.Background(), t, pool, userAID)
		ctxA = user.WithContext(ctxA, userAID)
		
		// Verify via direct database query that the workout was not modified
		var actualNotes string
		err := pool.QueryRow(ctxA, "SELECT notes FROM workout WHERE id = $1 AND user_id = $2",
			workoutID, userAID).Scan(&actualNotes)
		assert.NoError(t, err)
		assert.NotEqual(t, "Malicious update attempt from User B", actualNotes)
	})

	t.Run("Exercise_And_Set_Updates", func(t *testing.T) {
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		// Create a complex update with exercises and sets
		updateReq := UpdateWorkoutRequest{
			Date:  "2023-01-15T10:00:00Z",
			Notes: stringPtr("Updated with new exercises"),
			Exercises: []UpdateExercise{
				{
					Name: "Bench Press",
					Sets: []UpdateSet{
						{Weight: intPtr(135), Reps: 10, SetType: "warmup"},
						{Weight: intPtr(185), Reps: 8, SetType: "working"},
						{Weight: intPtr(225), Reps: 5, SetType: "working"},
					},
				},
				{
					Name: "Squats",
					Sets: []UpdateSet{
						{Weight: intPtr(95), Reps: 10, SetType: "warmup"},
						{Weight: intPtr(135), Reps: 8, SetType: "working"},
					},
				},
			},
		}

		body, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		req.SetPathValue("id", strconv.Itoa(int(workoutID)))

		w := httptest.NewRecorder()
		handler.UpdateWorkout(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		
		// Backfill order columns for the sets we just created
		// This ensures the subsequent GetWorkoutWithSets query works correctly
		testutils.BackfillSetOrderColumns(ctx, t, pool, userAID)
		
		// Verify the update by getting the workout and checking exercises and sets
		getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/workouts/%d", workoutID), nil)
		getReq = getReq.WithContext(ctx)
		getReq.SetPathValue("id", strconv.Itoa(int(workoutID)))
		
		getW := httptest.NewRecorder()
		handler.GetWorkoutWithSets(getW, getReq)
		
		var workout []db.GetWorkoutWithSetsRow
		err := json.Unmarshal(getW.Body.Bytes(), &workout)
		assert.NoError(t, err)
		
		// Group sets by exercise
		exerciseSets := make(map[string][]db.GetWorkoutWithSetsRow)
		for _, row := range workout {
		if row.ExerciseName != "" {
			exerciseSets[row.ExerciseName] = append(
				exerciseSets[row.ExerciseName], row)
			}
		}
		
		// Check for Bench Press and its 3 sets
		benchSets, hasBench := exerciseSets["Bench Press"]
		assert.True(t, hasBench, "Bench Press exercise should exist")
		assert.Len(t, benchSets, 3, "Bench Press should have 3 sets")
		
		// Check for Squats and its 2 sets
		squatSets, hasSquats := exerciseSets["Squats"]
		assert.True(t, hasSquats, "Squats exercise should exist")
		assert.Len(t, squatSets, 2, "Squats should have 2 sets")
	})
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function for int pointers
func intPtr(i int) *int {
	return &i
}
