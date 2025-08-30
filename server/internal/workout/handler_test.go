package workout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWorkoutRepository implements the WorkoutRepository interface for testing
type MockWorkoutRepository struct {
	mock.Mock
}

func (m *MockWorkoutRepository) SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error {
	args := m.Called(ctx, reformatted, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(db.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) ListWorkouts(ctx context.Context, userID string) ([]db.Workout, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetWorkoutWithSetsRow), args.Error(1)
}

func (m *MockWorkoutRepository) UpdateWorkout(ctx context.Context, id int32, reformatted *ReformattedRequest, userID string) error {
	args := m.Called(ctx, id, reformatted, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) DeleteWorkout(ctx context.Context, id int32, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) ListWorkoutFocusValues(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

type errorResponse struct {
	Message string `json:"message"`
}

func TestWorkoutHandler_ListWorkouts(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name: "successful fetch",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkouts", mock.Anything, userID).Return([]db.Workout{
					{
						ID: 1,
						Date: pgtype.Timestamptz{
							Time:  time.Now(),
							Valid: true,
						},
						Notes: pgtype.Text{String: "Morning workout", Valid: true},
					},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name: "internal server error",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkouts", mock.Anything, userID).Return([]db.Workout{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectJSON:    true,
			expectedError: "failed to list workouts",
		},
		{
			name:          "unauthenticated user",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectJSON:    true,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockWorkoutRepository{}
			tt.setupMock(mockRepo)

			// Create service with mock repo
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.ListWorkouts(w, req)

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

func TestWorkoutHandler_GetWorkoutWithSets(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		workoutID     string
		setupMock     func(*MockWorkoutRepository, int32)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name:      "successful fetch",
			workoutID: "1",
			setupMock: func(m *MockWorkoutRepository, id int32) {
				m.On("GetWorkoutWithSets", mock.Anything, id, userID).Return([]db.GetWorkoutWithSetsRow{
					{
						WorkoutID: id,
						WorkoutDate: pgtype.Timestamptz{
							Time:  time.Now(),
							Valid: true,
						},
					},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:          "missing workout ID",
			workoutID:     "",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Missing workout ID",
		},
		{
			name:          "invalid workout ID",
			workoutID:     "invalid",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid workout ID",
		},
		{
			name:      "service error",
			workoutID: "999",
			setupMock: func(m *MockWorkoutRepository, id int32) {
				m.On("GetWorkoutWithSets", mock.Anything, id, userID).Return([]db.GetWorkoutWithSetsRow{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get workout with sets",
		},
		{
			name:          "unauthenticated user",
			workoutID:     "1",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockWorkoutRepository)
			var id int32
			if tt.workoutID != "" {
				parsedID, err := strconv.Atoi(tt.workoutID)
				if err == nil {
					id = int32(parsedID)
				}
			}
			tt.setupMock(mockRepo, id)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/workouts/"+tt.workoutID, nil).WithContext(tt.ctx)
			if tt.workoutID != "" {
				req.SetPathValue("id", tt.workoutID)
			}
			w := httptest.NewRecorder()

			// Execute
			handler.GetWorkoutWithSets(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var result []db.GetWorkoutWithSetsRow
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutHandler_CreateWorkout(t *testing.T) {
	userID := "test-user-id"

	validRequest := CreateWorkoutRequest{
		Date: time.Now().Format(time.RFC3339),
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
	}

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
			name:        "successful creation",
			requestBody: validRequest,
			setupMock: func(m *MockWorkoutRepository) {
				m.On("SaveWorkout", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:          "invalid JSON",
			requestBody:   "invalid json string",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "failed to decode request body",
		},
		{
			name: "validation error - missing date",
			requestBody: CreateWorkoutRequest{
				Exercises: []ExerciseInput{
					{
						Name: "Bench Press",
						Sets: []SetInput{
							{Reps: 10, SetType: "working"},
						},
					},
				},
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "validation error occurred",
		},
		{
			name: "validation error - empty exercises",
			requestBody: CreateWorkoutRequest{
				Date: time.Now().Format(time.RFC3339),
			},
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "validation error occurred",
		},
		{
			name:        "service error",
			requestBody: validRequest,
			setupMock: func(m *MockWorkoutRepository) {
				m.On("SaveWorkout", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectJSON:    true,
			expectedError: "failed to create workout",
		},
		{
			name:          "unauthenticated user",
			requestBody:   validRequest,
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectJSON:    true,
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
					req = httptest.NewRequest("POST", "/api/workouts", bytes.NewBufferString(tt.requestBody.(string)))
				} else {
					req = httptest.NewRequest("POST", "/api/workouts", bytes.NewBuffer(body))
				}
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

func TestWorkoutHandler_DeleteWorkout(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		workoutID     string
		setupMock     func(*MockWorkoutRepository, int32)
		ctx           context.Context
		expectedCode  int
		expectedError string
	}{
		{
			name:      "successful deletion",
			workoutID: "1",
			setupMock: func(m *MockWorkoutRepository, id int32) {
				// First call for existence check
				m.On("GetWorkout", mock.Anything, id, userID).Return(db.Workout{
					ID:     id,
					UserID: userID,
				}, nil)
				// Second call for actual deletion
				m.On("DeleteWorkout", mock.Anything, id, userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusNoContent,
		},
		{
			name:          "missing workout ID",
			workoutID:     "",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Missing workout ID",
		},
		{
			name:          "invalid workout ID",
			workoutID:     "invalid",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid workout ID",
		},
		{
			name:      "workout not found",
			workoutID: "999",
			setupMock: func(m *MockWorkoutRepository, id int32) {
				m.On("GetWorkout", mock.Anything, id, userID).Return(db.Workout{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusNotFound,
			expectedError: "workout not found",
		},
		{
			name:      "service error during deletion",
			workoutID: "2",
			setupMock: func(m *MockWorkoutRepository, id int32) {
				// First call succeeds (workout exists)
				m.On("GetWorkout", mock.Anything, id, userID).Return(db.Workout{
					ID:     id,
					UserID: userID,
				}, nil)
				// Second call fails (deletion error)
				m.On("DeleteWorkout", mock.Anything, id, userID).Return(assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to delete workout",
		},
		{
			name:          "unauthenticated user",
			workoutID:     "1",
			setupMock:     func(m *MockWorkoutRepository, id int32) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockWorkoutRepository)
			var id int32
			if tt.workoutID != "" {
				parsedID, err := strconv.Atoi(tt.workoutID)
				if err == nil {
					id = int32(parsedID)
				}
			}
			tt.setupMock(mockRepo, id)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("DELETE", "/api/workouts/"+tt.workoutID, nil).WithContext(tt.ctx)
			if tt.workoutID != "" {
				req.SetPathValue("id", tt.workoutID)
			}
			w := httptest.NewRecorder()

			// Execute
			handler.DeleteWorkout(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			// Check for expected errors (if any)
			if tt.expectedError != "" {
				var resp errorResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Message, tt.expectedError)
			}

			// Ensure no content for successful deletion
			if tt.expectedCode == http.StatusNoContent {
				assert.Empty(t, w.Body.String())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test for ListWorkoutFocusValues endpoint
func TestWorkoutHandler_ListWorkoutFocusValues(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedBody  []string
		expectedError string
	}{
		{
			name: "successful fetch with values",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{"Chest", "Back", "Legs"}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedBody: []string{"Chest", "Back", "Legs"},
		},
		{
			name: "successful fetch with empty result",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedBody: []string{},
		},
		{
			name: "service error",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("ListWorkoutFocusValues", mock.Anything, userID).Return([]string{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to list workout focus values",
		},
		{
			name:          "unauthenticated user",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockWorkoutRepository)
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := &WorkoutService{
				repo:   mockRepo,
				logger: logger,
			}
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/workout-focus-values", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.ListWorkoutFocusValues(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var result []string
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				// Ensure we always return an empty slice, not nil
				if tt.expectedBody == nil {
					tt.expectedBody = []string{}
				}
				assert.Equal(t, tt.expectedBody, result)
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutHandler_Integration_ListWorkouts_RLS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup real database connection
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Initialize components with real database
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	queries := db.New(pool)
	
	// Initialize repositories
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := NewRepository(logger, queries, pool, exerciseRepo)
	workoutService := NewService(logger, workoutRepo)
	handler := NewHandler(logger, validator, workoutService)

	// Test data
	userAID := "test-user-a"
	userBID := "test-user-b"

	// Create test data with proper RLS context
	workoutAID := setupTestWorkout(t, pool, userAID, "User A's workout")
	workoutBID := setupTestWorkout(t, pool, userBID, "User B's workout")

	t.Run("Scenario1_UserA_CanRetrieveOwnWorkout", func(t *testing.T) {
		// Set RLS context for User A
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var workouts []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &workouts)
		assert.NoError(t, err)

		// User A should see only their own workout
		assert.Len(t, workouts, 1)
		assert.Equal(t, workoutAID, workouts[0].ID)
		assert.Equal(t, userAID, workouts[0].UserID)
	})

	t.Run("Scenario2_UserB_CannotRetrieveUserAWorkout", func(t *testing.T) {
		// Set RLS context for User B
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var workouts []db.Workout
		err := json.Unmarshal(w.Body.Bytes(), &workouts)
		assert.NoError(t, err)

		// User B should see only their own workout, not User A's
		assert.Len(t, workouts, 1)
		assert.Equal(t, workoutBID, workouts[0].ID)
		assert.Equal(t, userBID, workouts[0].UserID)
		assert.NotEqual(t, workoutAID, workouts[0].ID, "User B should not see User A's workout")
	})

	t.Run("Scenario3_AnonymousUser_CannotAccessWorkouts", func(t *testing.T) {
		// No user context set (anonymous user)
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListWorkouts(w, req)

		// Should get unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "user not authenticated")
	})

	t.Run("Scenario4_GetSpecificWorkout_UserB_CannotAccessUserAWorkout", func(t *testing.T) {
		// User B tries to access User A's specific workout
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/workouts/%d", workoutAID), nil).WithContext(ctx)
		req.SetPathValue("id", fmt.Sprintf("%d", workoutAID))
		w := httptest.NewRecorder()

		handler.GetWorkoutWithSets(w, req)

		// Should succeed but return empty results due to RLS
		assert.Equal(t, http.StatusOK, w.Code)
		var result []db.GetWorkoutWithSetsRow
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Empty(t, result, "User B should not see User A's workout data")
	})

	t.Run("Scenario5_ConcurrentRequests_ProperIsolation", func(t *testing.T) {
		// Test concurrent requests from different users
		var wg sync.WaitGroup
		results := make(map[string][]db.Workout)
		mu := sync.Mutex{}

		// User A request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userAID)
			ctx = user.WithContext(ctx, userAID)

			req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListWorkouts(w, req)

			if w.Code == http.StatusOK {
				var workouts []db.Workout
				json.Unmarshal(w.Body.Bytes(), &workouts)
				mu.Lock()
				results[userAID] = workouts
				mu.Unlock()
			}
		}()

		// User B request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userBID)
			ctx = user.WithContext(ctx, userBID)

			req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListWorkouts(w, req)

			if w.Code == http.StatusOK {
				var workouts []db.Workout
				json.Unmarshal(w.Body.Bytes(), &workouts)
				mu.Lock()
				results[userBID] = workouts
				mu.Unlock()
			}
		}()

		wg.Wait()

		// Verify isolation - each user should only see their own workouts
		assert.Len(t, results[userAID], 1)
		assert.Equal(t, workoutAID, results[userAID][0].ID)
		assert.Equal(t, userAID, results[userAID][0].UserID)

		assert.Len(t, results[userBID], 1)
		assert.Equal(t, workoutBID, results[userBID][0].ID)
		assert.Equal(t, userBID, results[userBID][0].UserID)
	})
}

func BenchmarkWorkoutHandler_ListWorkouts(b *testing.B) {
	userID := "test-user-id"
	mockRepo := &MockWorkoutRepository{}
	mockRepo.On("ListWorkouts", mock.Anything, userID).Return([]db.Workout{
		{ID: 1, Date: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
	}, nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	service := &WorkoutService{
		repo:   mockRepo,
		logger: logger,
	}
	handler := NewHandler(logger, validator, service)

	ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ListWorkouts(w, req)
	}
}

// === HELPER FUNCTIONS FOR INTEGRATION TESTS ===

func getTestDatabaseURL() string {
	// Try to get from environment first (for CI/CD)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		return dbURL
	}

	// Load environment variables from .env file for local development
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Try again after loading .env
	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Final fallback to default test database
		dbURL = "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
	}
	return dbURL
}

func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	// Create connection pool
	pool, err := pgxpool.New(ctx, getTestDatabaseURL())
	require.NoError(t, err, "Failed to create database pool")

	// Test connectivity
	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping database")

	// Setup RLS policies
	setupRLS(t, pool)

	// Setup users table entries
	setupTestUsers(t, pool)
	
	// Backfill order columns for existing test data if they exist
	// This ensures tests work with both old and new database schemas
	backfillOrderColumnsForTests(t, pool)

	return pool, func() {
		cleanupTestData(t, pool)
		pool.Close()
	}
}

func setupRLS(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Check if current_user_id function exists (should exist from migration)
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'current_user_id')").Scan(&exists)
	require.NoError(t, err, "Failed to check if current_user_id function exists")
	require.True(t, exists, "current_user_id function should exist from migration")

	// Ensure RLS is enabled on tables (should be enabled from migration)
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on users table")

	_, err = pool.Exec(ctx, "ALTER TABLE workout ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on workout table")

	_, err = pool.Exec(ctx, "ALTER TABLE exercise ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on exercise table")

	_, err = pool.Exec(ctx, "ALTER TABLE \"set\" ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on set table")

	// RLS policies should already exist from migration - no need to recreate them
	// Just verify they exist
	var policyCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM pg_policies WHERE tablename IN ('users', 'workout', 'exercise', 'set')").Scan(&policyCount)
	require.NoError(t, err, "Failed to count RLS policies")
	require.Greater(t, policyCount, 0, "RLS policies should exist from migration")
}


func setupTestUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	// Temporarily disable RLS for setup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to disable RLS for setup")

	// Create test users
	userIDs := []string{"test-user-a", "test-user-b"}
	for _, userID := range userIDs {
		_, err = pool.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
		require.NoError(t, err, "Failed to create test user: %s", userID)
	}

	// Re-enable RLS
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to re-enable RLS")
}

func setupTestWorkout(t *testing.T, pool *pgxpool.Pool, userID, notes string) int32 {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Create workout
	var workoutID int32
	err := pool.QueryRow(ctx, 
		"INSERT INTO workout (date, notes, user_id) VALUES (NOW(), $1, $2) RETURNING id",
		notes, userID).Scan(&workoutID)
	require.NoError(t, err, "Failed to create test workout for user %s", userID)

	return workoutID
}

func setTestUserContext(ctx context.Context, t *testing.T, pool *pgxpool.Pool, userID string) context.Context {
	t.Helper()
	return testutils.SetTestUserContext(ctx, t, pool, userID)
}

func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable RLS temporarily for cleanup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY; ALTER TABLE workout DISABLE ROW LEVEL SECURITY; ALTER TABLE exercise DISABLE ROW LEVEL SECURITY; ALTER TABLE \"set\" DISABLE ROW LEVEL SECURITY;")
	if err != nil {
		t.Logf("Warning: Failed to disable RLS for cleanup: %v", err)
	}

	// More thorough cleanup - remove ALL test data
	// This helps prevent test pollution between runs
	testUserPattern := "test-user-%"
	
	// Start with dependent tables first
	_, err = pool.Exec(ctx, "DELETE FROM \"set\" WHERE workout_id IN (SELECT id FROM workout WHERE user_id LIKE $1)", testUserPattern)
	if err != nil {
		t.Logf("Warning: Failed to clean up set data: %v", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM workout WHERE user_id LIKE $1", testUserPattern)
	if err != nil {
		t.Logf("Warning: Failed to clean up workout data: %v", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM exercise WHERE user_id LIKE $1", testUserPattern)
	if err != nil {
		t.Logf("Warning: Failed to clean up exercise data: %v", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM users WHERE user_id LIKE $1", testUserPattern)
	if err != nil {
		t.Logf("Warning: Failed to clean up user data: %v", err)
	}

	// Log cleanup results
	var remainingWorkouts int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM workout WHERE user_id LIKE $1", testUserPattern).Scan(&remainingWorkouts)
	if err == nil {
		t.Logf("Cleanup complete. Remaining test workouts: %d", remainingWorkouts)
	}

	// Re-enable RLS
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY; ALTER TABLE workout ENABLE ROW LEVEL SECURITY; ALTER TABLE exercise ENABLE ROW LEVEL SECURITY; ALTER TABLE \"set\" ENABLE ROW LEVEL SECURITY;")
	if err != nil {
		t.Logf("Warning: Failed to re-enable RLS after cleanup: %v", err)
	}
}

func backfillOrderColumnsForTests(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	
	// Backfill order columns for all test users
	// This ensures tests work whether or not the migration has been applied
	testUsers := []string{"test-user-a", "test-user-b"}
	
	for _, userID := range testUsers {
		// Set user context for RLS
		ctxUser := testutils.SetTestUserContext(ctx, t, pool, userID)
		
		// Backfill order columns for this user
		testutils.BackfillSetOrderColumns(ctxUser, t, pool, userID)
	}
}
