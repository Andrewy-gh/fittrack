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
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorkoutFocus_CreateWorkout(t *testing.T) {
	userID := "test-user-id"
	
	// Helper function for string pointers
	stringPtr := func(s string) *string { return &s }

	tests := []struct {
		name          string
		requestBody   interface{}
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name: "successful creation with workout focus",
			requestBody: CreateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtr("Upper Body"),
				Exercises: []ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []SetInput{
							{
								Reps:    10,
								SetType: "working",
							},
						},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("SaveWorkout", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name: "successful creation with empty workout focus",
			requestBody: CreateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtr(""),
				Exercises: []ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []SetInput{
							{
								Reps:    10,
								SetType: "working",
							},
						},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("SaveWorkout", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name: "successful creation with nil workout focus",
			requestBody: CreateWorkoutRequest{
				Date: "2023-01-15T10:00:00Z",
				Exercises: []ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []SetInput{
							{
								Reps:    10,
								SetType: "working",
							},
						},
					},
				},
			},
			setupMock: func(m *MockWorkoutRepository) {
				m.On("SaveWorkout", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name: "validation error - workout focus too long",
			requestBody: CreateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtr(string(make([]byte, 300))), // Exceeds 256 char limit
				Exercises: []ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []SetInput{
							{
								Reps:    10,
								SetType: "working",
							},
						},
					},
				},
			},
			setupMock:    func(m *MockWorkoutRepository) {},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusBadRequest,
			expectJSON:   true,
			expectedError: "validation error occurred",
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
				require.NoError(t, err)
				req = httptest.NewRequest("POST", "/api/workouts", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("POST", "/api/workouts", nil)
			}

			w := httptest.NewRecorder()

			// Execute
			handler.CreateWorkout(w, req.WithContext(tt.ctx))

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				if tt.expectedError != "" {
					var resp errorResponse
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					assert.NoError(t, err)
					assert.Contains(t, resp.Message, tt.expectedError)
				}
			} else if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutFocus_UpdateWorkout(t *testing.T) {
	userID := "test-user-id"
	workoutID := int32(1)

	// Helper function for string pointers
	stringPtr := func(s string) *string { return &s }

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
			name:      "successful update with workout focus",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtrHelper("Push Day"),
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
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
			name:      "successful update with empty workout focus",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtrHelper(""),
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
			name:      "successful partial update - focus only",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z", // Date is required for validation
				WorkoutFocus: stringPtrHelper("Leg Day"),
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
			name:      "validation error - workout focus too long",
			workoutID: "1",
			requestBody: UpdateWorkoutRequest{
				Date:         "2023-01-15T10:00:00Z",
				WorkoutFocus: stringPtr(string(make([]byte, 300))), // Exceeds 256 char limit
				Exercises: []UpdateExercise{
					{
						Name: "placeholder",
						Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
					},
				},
			},
			setupMock:    func(m *MockWorkoutRepository) {},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusBadRequest,
			expectedError: "validation error occurred",
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
				require.NoError(t, err)
				req = httptest.NewRequest("PUT", "/api/workouts/"+tt.workoutID, bytes.NewBuffer(body))
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

func TestWorkoutFocus_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool, cleanup := setupTestDatabaseForWorkoutFocus(t)
	defer cleanup()

	// Setup test data
	userID := "focus-test-user"

	// Initialize components with real database
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	queries := db.New(pool)

	// Initialize repositories
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := NewRepository(logger, queries, pool, exerciseRepo)
	workoutService := NewService(logger, workoutRepo)
	handler := NewHandler(logger, validator, workoutService)

	t.Run("CreateWorkout_WithFocus", func(t *testing.T) {
		ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
		ctx = user.WithContext(ctx, userID)

		createReq := CreateWorkoutRequest{
			Date:         "2023-01-15T10:00:00Z",
			WorkoutFocus: stringPtrHelper("Upper Body Focus"),
			Exercises: []ExerciseInput{
				{
					Name: "Bench Press",
					Sets: []SetInput{
						{Reps: 10, SetType: "working"},
					},
				},
			},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/workouts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.CreateWorkout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		// Verify that the workout was created successfully by checking the response
		var result map[string]bool
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.True(t, result["success"])
		
		// Verify the workout was created with the focus by querying the database directly
		var actualFocus string
		var actualFocusValid bool
		err = pool.QueryRow(ctx, "SELECT workout_focus, workout_focus IS NOT NULL FROM workout WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1",
			userID).Scan(&actualFocus, &actualFocusValid)
		assert.NoError(t, err)
		assert.Equal(t, "Upper Body Focus", actualFocus)
		assert.True(t, actualFocusValid)
	})

	t.Run("UpdateWorkout_Focus", func(t *testing.T) {
		// Create a workout first
		ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
		ctx = user.WithContext(ctx, userID)

		// Create a workout
		createReq := CreateWorkoutRequest{
			Date: "2023-01-15T10:00:00Z",
			Exercises: []ExerciseInput{
				{
					Name: "Squats",
					Sets: []SetInput{
						{Reps: 10, SetType: "working"},
					},
				},
			},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/workouts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.CreateWorkout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Get the created workout ID
		var result []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		workoutID := result[0].ID

		// Update the workout with focus
		updateReq := UpdateWorkoutRequest{
			WorkoutFocus: stringPtrHelper("Leg Day Focus"),
			Exercises: []UpdateExercise{
				{
					Name: "placeholder",
					Sets: []UpdateSet{{Reps: 1, SetType: "working"}},
				},
			},
		}

		body, _ = json.Marshal(updateReq)
		req = httptest.NewRequest("PUT", fmt.Sprintf("/api/workouts/%d", workoutID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		req.SetPathValue("id", fmt.Sprintf("%d", workoutID))

		w = httptest.NewRecorder()
		handler.UpdateWorkout(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify the update by checking the workout directly from database
		var actualFocus string
		var actualFocusValid bool
		err = pool.QueryRow(ctx, "SELECT workout_focus FROM workout WHERE id = $1 AND user_id = $2",
			workoutID, userID).Scan(&actualFocus, &actualFocusValid)
		assert.NoError(t, err)
		assert.Equal(t, "Leg Day Focus", actualFocus)
		assert.True(t, actualFocusValid)
	})

	t.Run("ListWorkouts_WithFocus", func(t *testing.T) {
		ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
		ctx = user.WithContext(ctx, userID)

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ListWorkouts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var workouts []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &workouts)
		assert.NoError(t, err)
		assert.NotEmpty(t, workouts)

		// Find the workout with focus
		var found bool
		for _, workout := range workouts {
			if workout.WorkoutFocus.Valid && workout.WorkoutFocus.String == "Leg Day Focus" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find workout with focus in list")
	})

	t.Run("GetWorkoutWithSets_WithFocus", func(t *testing.T) {
		ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
		ctx = user.WithContext(ctx, userID)

		// Get the workout with focus
		var workoutID int32
		err := pool.QueryRow(ctx, "SELECT id FROM workout WHERE workout_focus = $1 AND user_id = $2",
			"Leg Day Focus", userID).Scan(&workoutID)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/workouts/%d", workoutID), nil)
		req = req.WithContext(ctx)
		req.SetPathValue("id", fmt.Sprintf("%d", workoutID))

		w := httptest.NewRecorder()
		handler.GetWorkoutWithSets(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var workout []db.GetWorkoutWithSetsRow
		err = json.Unmarshal(w.Body.Bytes(), &workout)
		assert.NoError(t, err)
		assert.NotEmpty(t, workout)
		assert.Equal(t, "Leg Day Focus", workout[0].WorkoutFocus.String)
		assert.True(t, workout[0].WorkoutFocus.Valid)
	})
}

func stringPtrHelper(s string) *string {
	return &s
}

// Helper functions for integration tests

func setupTestDatabaseForWorkoutFocus(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	
	// Use the same setup as other integration tests
	pool, cleanup := setupTestDatabase(t)
	
	// Setup specific test users if needed
	setupSpecificTestUsersForFocus(t, pool)
	
	return pool, func() {
		cleanupSpecificTestUsersForFocus(t, pool)
		cleanup()
	}
}

func setupSpecificTestUsersForFocus(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	userIDs := []string{"focus-test-user"}
	setupSpecificTestUsers(t, pool, userIDs)
}

func cleanupSpecificTestUsersForFocus(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	userIDs := []string{"focus-test-user"}
	cleanupSpecificTestUsers(t, pool, userIDs)
}