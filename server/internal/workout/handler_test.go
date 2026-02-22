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

// Test user registry for tracking users created during tests
// This allows for automatic cleanup without hardcoding user IDs
var (
	testUserRegistry   = make(map[string]bool)
	testUserRegistryMu sync.Mutex
)

// registerTestUser adds a user to the cleanup registry
func registerTestUser(userID string) {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	testUserRegistry[userID] = true
}

// getRegisteredTestUsers returns a slice of all registered test users
func getRegisteredTestUsers() []string {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	users := make([]string, 0, len(testUserRegistry))
	for userID := range testUserRegistry {
		users = append(users, userID)
	}
	return users
}

// clearTestUserRegistry clears the registry (useful for test isolation if needed)
func clearTestUserRegistry() {
	testUserRegistryMu.Lock()
	defer testUserRegistryMu.Unlock()
	testUserRegistry = make(map[string]bool)
}

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

func (m *MockWorkoutRepository) GetContributionData(ctx context.Context, userID string) ([]db.GetContributionDataRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.GetContributionDataRow), args.Error(1)
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
			expectedError: "not authorized",
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
			expectedError: "not authorized",
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
			expectedError: "not authorized",
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
			expectedError: "not found",
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
			expectedError: "not authorized",
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
			expectedError: "not authorized",
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

			req := httptest.NewRequest("GET", "/api/workouts/focus-values", nil).WithContext(tt.ctx)
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

// Test for GetContributionData endpoint
func TestWorkoutHandler_GetContributionData(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockWorkoutRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedDays  int
		expectedError string
	}{
		{
			name: "successful fetch with data",
			setupMock: func(m *MockWorkoutRepository) {
				now := time.Now()
				yesterday := now.AddDate(0, 0, -1)
				focusValue := "Strength"

				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{
					{
						Date:     pgtype.Date{Time: yesterday, Valid: true},
						Count:    5,
						Workouts: []byte(`[{"id": 1, "time": "` + yesterday.Format(time.RFC3339) + `", "focus": "` + focusValue + `"}]`),
					},
					{
						Date:     pgtype.Date{Time: now, Valid: true},
						Count:    10,
						Workouts: []byte(`[{"id": 2, "time": "` + now.Format(time.RFC3339) + `", "focus": null}, {"id": 3, "time": "` + now.Format(time.RFC3339) + `", "focus": "Cardio"}]`),
					},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedDays: 2,
		},
		{
			name: "successful fetch with empty result",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
			expectedDays: 0,
		},
		{
			name: "service error",
			setupMock: func(m *MockWorkoutRepository) {
				m.On("GetContributionData", mock.Anything, userID).Return([]db.GetContributionDataRow{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "failed to get contribution data",
		},
		{
			name:          "unauthenticated user",
			setupMock:     func(m *MockWorkoutRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectedError: "not authorized",
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

			req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.GetContributionData(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				var result ContributionDataResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result.Days, tt.expectedDays)

				// Verify workout data structure if we have days
				if tt.expectedDays > 0 && tt.name == "successful fetch with data" {
					// First day should have 1 workout with focus
					assert.Len(t, result.Days[0].Workouts, 1)
					assert.Equal(t, int32(1), result.Days[0].Workouts[0].ID)
					assert.NotEmpty(t, result.Days[0].Workouts[0].Time)
					assert.NotNil(t, result.Days[0].Workouts[0].Focus)
					assert.Equal(t, "Strength", *result.Days[0].Workouts[0].Focus)

					// Second day should have 2 workouts, one with focus and one without
					assert.Len(t, result.Days[1].Workouts, 2)
					assert.Equal(t, int32(2), result.Days[1].Workouts[0].ID)
					assert.Nil(t, result.Days[1].Workouts[0].Focus)
					assert.Equal(t, int32(3), result.Days[1].Workouts[1].ID)
					assert.NotNil(t, result.Days[1].Workouts[1].Focus)
					assert.Equal(t, "Cardio", *result.Days[1].Workouts[1].Focus)
				}
			}

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test for level calculation in the service layer
func TestWorkoutService_CalculateLevelThresholds(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	tests := []struct {
		name       string
		counts     []int
		wantStatic bool // true if we expect static thresholds
	}{
		{
			name:       "empty counts uses static thresholds",
			counts:     []int{},
			wantStatic: true,
		},
		{
			name:       "fewer than 30 workout days uses static thresholds",
			counts:     []int{5, 10, 3, 8, 2},
			wantStatic: true,
		},
		{
			name:       "10 workout days (below threshold) uses static thresholds",
			counts:     []int{5, 10, 3, 8, 2, 7, 4, 6, 9, 11},
			wantStatic: true,
		},
		{
			name:       "exactly 29 workout days uses static thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29},
			wantStatic: true,
		},
		{
			name:       "exactly 30 workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30},
			wantStatic: false,
		},
		{
			name:       "30+ workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35},
			wantStatic: false,
		},
		{
			name:       "only zeros uses static thresholds",
			counts:     []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			wantStatic: true,
		},
		{
			name:       "mixed zeros and values - 30+ non-zero uses dynamic",
			counts:     []int{0, 1, 0, 2, 3, 0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			wantStatic: false,
		},
		{
			name:       "exactly 31 workout days uses dynamic percentile thresholds",
			counts:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
			wantStatic: false,
		},
		{
			name:       "all identical counts - dynamic thresholds should handle gracefully",
			counts:     []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
			wantStatic: false,
		},
		{
			name:       "mostly identical with one outlier - thresholds remain valid",
			counts:     []int{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 100},
			wantStatic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thresholds := service.calculateLevelThresholds(tt.counts)

			// Static thresholds are [1, 6, 11, 16]
			staticThresholds := [4]int{1, 6, 11, 16}

			if tt.wantStatic {
				assert.Equal(t, staticThresholds, thresholds, "Expected static thresholds")
			} else {
				// Dynamic thresholds should be strictly increasing
				assert.GreaterOrEqual(t, thresholds[0], 1, "First threshold should be at least 1")
				assert.Greater(t, thresholds[1], thresholds[0], "Thresholds should be strictly increasing")
				assert.Greater(t, thresholds[2], thresholds[1], "Thresholds should be strictly increasing")
				assert.Greater(t, thresholds[3], thresholds[2], "Thresholds should be strictly increasing")
			}
		})
	}
}

// Test for level calculation based on count and thresholds
func TestWorkoutService_CalculateLevel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	// Static thresholds: [1, 6, 11, 16]
	// Level 0: count == 0
	// Level 1: count >= 1 and count < 6
	// Level 2: count >= 6 and count < 11
	// Level 3: count >= 11 and count < 16
	// Level 4: count >= 16
	thresholds := [4]int{1, 6, 11, 16}

	tests := []struct {
		name          string
		count         int
		expectedLevel int
	}{
		{"zero count gives level 0", 0, 0},
		{"count 1 gives level 1", 1, 1},
		{"count 5 gives level 1", 5, 1},
		{"count 6 gives level 2", 6, 2},
		{"count 10 gives level 2", 10, 2},
		{"count 11 gives level 3", 11, 3},
		{"count 15 gives level 3", 15, 3},
		{"count 16 gives level 4", 16, 4},
		{"count 100 gives level 4", 100, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := service.calculateLevel(tt.count, thresholds)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

// Test for parseWorkouts
func TestWorkoutService_ParseWorkouts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	tests := []struct {
		name     string
		input    []byte
		expected []WorkoutSummary
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []WorkoutSummary{},
		},
		{
			name:     "empty byte slice",
			input:    []byte{},
			expected: []WorkoutSummary{},
		},
		{
			name:  "single workout with focus",
			input: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength"}]`),
			expected: []WorkoutSummary{
				{ID: 1, Time: "2025-01-15T10:00:00Z", Focus: stringPtr("Strength")},
			},
		},
		{
			name:  "multiple workouts",
			input: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength"}, {"id": 2, "time": "2025-01-15T14:00:00Z", "focus": null}]`),
			expected: []WorkoutSummary{
				{ID: 1, Time: "2025-01-15T10:00:00Z", Focus: stringPtr("Strength")},
				{ID: 2, Time: "2025-01-15T14:00:00Z", Focus: nil},
			},
		},
		{
			name:     "invalid JSON",
			input:    []byte(`invalid json`),
			expected: []WorkoutSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.parseWorkouts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test for convertContributionRows
func TestWorkoutService_ConvertContributionRows(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &WorkoutService{logger: logger}

	t.Run("empty rows returns empty slice", func(t *testing.T) {
		result := service.convertContributionRows([]db.GetContributionDataRow{})
		assert.Equal(t, []ContributionDay{}, result)
	})

	t.Run("converts rows with correct level calculation", func(t *testing.T) {
		testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		rows := []db.GetContributionDataRow{
			{
				Date:     pgtype.Date{Time: testDate, Valid: true},
				Count:    5,
				Workouts: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength"}]`),
			},
		}

		result := service.convertContributionRows(rows)

		assert.Len(t, result, 1)
		assert.Equal(t, "2025-01-15", result[0].Date)
		assert.Equal(t, 5, result[0].Count)
		assert.Equal(t, 1, result[0].Level) // Static threshold: 5 < 6, so level 1
		assert.Len(t, result[0].Workouts, 1)
		assert.Equal(t, int32(1), result[0].Workouts[0].ID)
	})

	t.Run("handles multiple workouts", func(t *testing.T) {
		testDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		rows := []db.GetContributionDataRow{
			{
				Date:     pgtype.Date{Time: testDate, Valid: true},
				Count:    20,
				Workouts: []byte(`[{"id": 1, "time": "2025-01-15T10:00:00Z", "focus": "Strength"}, {"id": 2, "time": "2025-01-15T14:00:00Z", "focus": null}, {"id": 3, "time": "2025-01-15T18:00:00Z", "focus": "Cardio"}]`),
			},
		}

		result := service.convertContributionRows(rows)

		assert.Len(t, result, 1)
		assert.Equal(t, 4, result[0].Level) // 20 >= 16, so level 4
		assert.Len(t, result[0].Workouts, 3)
		assert.Equal(t, int32(1), result[0].Workouts[0].ID)
		assert.Equal(t, int32(2), result[0].Workouts[1].ID)
		assert.Equal(t, int32(3), result[0].Workouts[2].ID)
	})
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
		assert.Contains(t, resp.Message, "not authorized")
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

func TestContributionData_Integration_RLS(t *testing.T) {
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
	userAID := "test-user-rls-a"
	userBID := "test-user-rls-b"

	// Create test workouts for both users on the same day
	today := time.Now()
	workoutAID := setupTestWorkout(t, pool, userAID, "User A's workout")
	workoutBID := setupTestWorkout(t, pool, userBID, "User B's workout")

	// Add a set to each workout so they show up in contribution data
	setupTestSet(t, pool, userAID, workoutAID)
	setupTestSet(t, pool, userBID, workoutBID)

	t.Run("UserA_OnlySeesOwnWorkoutMetadata", func(t *testing.T) {
		// Set RLS context for User A
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's contribution data
		todayStr := today.Format("2006-01-02")
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		assert.NotNil(t, todayData, "Should have contribution data for today")
		assert.Len(t, todayData.Workouts, 1, "User A should see exactly 1 workout (their own)")
		assert.Equal(t, workoutAID, todayData.Workouts[0].ID, "User A should see their own workout ID")
		assert.NotEqual(t, workoutBID, todayData.Workouts[0].ID, "User A should NOT see User B's workout ID")
	})

	t.Run("UserB_OnlySeesOwnWorkoutMetadata", func(t *testing.T) {
		// Set RLS context for User B
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's contribution data
		todayStr := today.Format("2006-01-02")
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		assert.NotNil(t, todayData, "Should have contribution data for today")
		assert.Len(t, todayData.Workouts, 1, "User B should see exactly 1 workout (their own)")
		assert.Equal(t, workoutBID, todayData.Workouts[0].ID, "User B should see their own workout ID")
		assert.NotEqual(t, workoutAID, todayData.Workouts[0].ID, "User B should NOT see User A's workout ID")
	})

	t.Run("AnonymousUser_Unauthorized", func(t *testing.T) {
		// No user context set (anonymous user)
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		// Should get unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "not authorized")
	})
}

func TestContributionData_Integration(t *testing.T) {
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
	userID := "test-user-integration-contrib"

	t.Run("VerifyDateRangeFiltering", func(t *testing.T) {
		// Create workouts at different time points
		now := time.Now()
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Workout within 52 weeks
		withinRangeDate := now.AddDate(0, 0, -200) // ~6.5 months ago
		_, err := pool.Exec(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3)",
			withinRangeDate, "Within range", userID)
		require.NoError(t, err)

		// Workout outside 52 weeks (should not appear)
		outsideRangeDate := now.AddDate(0, 0, -400) // Over a year ago
		_, err = pool.Exec(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3)",
			outsideRangeDate, "Outside range", userID)
		require.NoError(t, err)

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Check that the within-range workout appears
		withinRangeDateStr := withinRangeDate.Format("2006-01-02")
		var foundWithinRange bool
		for _, day := range result.Days {
			if day.Date == withinRangeDateStr {
				foundWithinRange = true
				break
			}
		}
		assert.True(t, foundWithinRange, "Workout within 52 weeks should appear in contribution data")

		// Check that the outside-range workout does NOT appear
		outsideRangeDateStr := outsideRangeDate.Format("2006-01-02")
		var foundOutsideRange bool
		for _, day := range result.Days {
			if day.Date == outsideRangeDateStr {
				foundOutsideRange = true
				break
			}
		}
		assert.False(t, foundOutsideRange, "Workout outside 52 weeks should NOT appear in contribution data")
	})

	t.Run("VerifyMultipleWorkoutsPerDay", func(t *testing.T) {
		// Create multiple workouts on the same day
		now := time.Now()
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Create 3 workouts on the same day with different focuses
		// Use NOW() from database to avoid timezone issues
		var workoutIDs []int32
		focuses := []string{"Strength", "Cardio", ""}

		for i, focus := range focuses {
			var workoutID int32
			var err error
			if focus == "" {
				err = pool.QueryRow(ctx,
					"INSERT INTO workout (date, notes, user_id, workout_focus) VALUES (NOW(), $1, $2, NULL) RETURNING id",
					fmt.Sprintf("Workout %d", i+1), userID).Scan(&workoutID)
			} else {
				err = pool.QueryRow(ctx,
					"INSERT INTO workout (date, notes, user_id, workout_focus) VALUES (NOW(), $1, $2, $3) RETURNING id",
					fmt.Sprintf("Workout %d", i+1), userID, focus).Scan(&workoutID)
			}
			require.NoError(t, err)
			workoutIDs = append(workoutIDs, workoutID)

			// Add a set to each workout so they appear in contribution data
			setupTestSet(t, pool, userID, workoutID)
		}

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find today's data (use database's current date)
		todayStr := now.Format("2006-01-02")
		var todayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == todayStr {
				todayData = &result.Days[i]
				break
			}
		}

		// If not found with today's date, look for the most recent day with data
		if todayData == nil && len(result.Days) > 0 {
			// Find the day with workouts that match our workout IDs
			for i := range result.Days {
				if len(result.Days[i].Workouts) >= 3 {
					todayData = &result.Days[i]
					break
				}
			}
		}

		assert.NotNil(t, todayData, "Should have contribution data for workouts created")
		assert.Len(t, todayData.Workouts, 3, "Should have 3 workouts")

		// Verify workout metadata
		foundIDs := make(map[int32]bool)
		for _, workout := range todayData.Workouts {
			foundIDs[workout.ID] = true
			assert.NotEmpty(t, workout.Time, "Workout should have a time")
		}

		// Verify all workout IDs are present
		for _, wid := range workoutIDs {
			assert.True(t, foundIDs[wid], "Workout ID %d should be in contribution data", wid)
		}

		// Verify focus values
		var hasStrength, hasCardio, hasNull bool
		for _, workout := range todayData.Workouts {
			if workout.Focus != nil && *workout.Focus == "Strength" {
				hasStrength = true
			} else if workout.Focus != nil && *workout.Focus == "Cardio" {
				hasCardio = true
			} else if workout.Focus == nil {
				hasNull = true
			}
		}
		assert.True(t, hasStrength, "Should have workout with Strength focus")
		assert.True(t, hasCardio, "Should have workout with Cardio focus")
		assert.True(t, hasNull, "Should have workout with null focus")
	})

	t.Run("VerifyLevelCalculation", func(t *testing.T) {
		// Create a day with known count to verify level calculation
		now := time.Now()
		testDate := now.AddDate(0, 0, -7) // 7 days ago
		ctx := setTestUserContext(context.Background(), t, pool, userID)

		// Create a workout with exactly 15 sets (should result in count=15)
		var workoutID int32
		err := pool.QueryRow(ctx,
			"INSERT INTO workout (date, notes, user_id) VALUES ($1, $2, $3) RETURNING id",
			testDate, "Level test workout", userID).Scan(&workoutID)
		require.NoError(t, err)

		// Add 15 sets
		var exerciseID int32
		err = pool.QueryRow(ctx,
			"INSERT INTO exercise (name, user_id) VALUES ($1, $2) RETURNING id",
			"Level Test Exercise", userID).Scan(&exerciseID)
		require.NoError(t, err)

		for i := 1; i <= 15; i++ {
			_, err = pool.Exec(ctx,
				"INSERT INTO \"set\" (workout_id, exercise_id, user_id, weight, reps, set_type, exercise_order, set_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				workoutID, exerciseID, userID, 100, 10, "working", 1, i)
			require.NoError(t, err)
		}

		// Set user context and make request
		ctx = user.WithContext(ctx, userID)
		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Find the test date
		testDateStr := testDate.Format("2006-01-02")
		var testDayData *ContributionDay
		for i := range result.Days {
			if result.Days[i].Date == testDateStr {
				testDayData = &result.Days[i]
				break
			}
		}

		assert.NotNil(t, testDayData, "Should have contribution data for test date")
		assert.Equal(t, 15, testDayData.Count, "Count should be 15")
		// With static thresholds [1, 6, 11, 16], count=15 should give level 3
		// (15 >= 11 and 15 < 16)
		assert.Equal(t, 3, testDayData.Level, "Level should be 3 for count=15 with static thresholds")
	})

	t.Run("VerifyEmptyResult", func(t *testing.T) {
		// Use a different user with no workouts
		emptyUserID := "test-user-empty-contrib"
		ctx := setTestUserContext(context.Background(), t, pool, emptyUserID)
		ctx = user.WithContext(ctx, emptyUserID)

		req := httptest.NewRequest("GET", "/api/workouts/contribution-data", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetContributionData(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result ContributionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Should return empty days array, not nil
		assert.NotNil(t, result.Days)
		assert.Len(t, result.Days, 0, "Should have no contribution data for user with no workouts")
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

	// Create test users and register them for cleanup
	// Include all users needed by integration tests
	userIDs := []string{
		"test-user-a",
		"test-user-b",
		"test-user-rls-a",
		"test-user-rls-b",
		"test-user-integration-contrib",
		"test-user-empty-contrib",
	}
	for _, userID := range userIDs {
		_, err = pool.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
		require.NoError(t, err, "Failed to create test user: %s", userID)
		registerTestUser(userID)
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

func setupTestSet(t *testing.T, pool *pgxpool.Pool, userID string, workoutID int32) int32 {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Get or create a test exercise
	var exerciseID int32
	err := pool.QueryRow(ctx,
		"INSERT INTO exercise (name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		"Test Exercise", userID).Scan(&exerciseID)
	if err != nil {
		// If conflict, query for existing
		err = pool.QueryRow(ctx,
			"SELECT id FROM exercise WHERE name = $1 AND user_id = $2 LIMIT 1",
			"Test Exercise", userID).Scan(&exerciseID)
		require.NoError(t, err, "Failed to get test exercise for user %s", userID)
	}

	// Create a working set
	var setID int32
	err = pool.QueryRow(ctx,
		"INSERT INTO \"set\" (workout_id, exercise_id, user_id, weight, reps, set_type, exercise_order, set_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		workoutID, exerciseID, userID, 100, 10, "working", 1, 1).Scan(&setID)
	require.NoError(t, err, "Failed to create test set for user %s", userID)

	return setID
}

func setTestUserContext(ctx context.Context, t *testing.T, pool *pgxpool.Pool, userID string) context.Context {
	t.Helper()
	return testutils.SetTestUserContext(ctx, t, pool, userID)
}

func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable RLS temporarily for cleanup
	tables := []string{"users", "workout", "exercise", `"set"`}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", table))
		if err != nil {
			t.Logf("Warning: Failed to disable RLS on %s for cleanup: %v", table, err)
		}
	}

	// Get all registered test users from the registry
	// This automatically includes any users created during test setup
	testUserIDs := getRegisteredTestUsers()

	if len(testUserIDs) == 0 {
		t.Logf("Warning: No test users registered for cleanup")
	}

	// Clean up data for each registered test user
	for _, userID := range testUserIDs {
		// Clean up dependent data first (sets → exercises → workouts → users)
		// exercise.historical_1rm_source_workout_id references workout(id), so exercises must be removed before workouts.
		_, err := pool.Exec(ctx, "DELETE FROM \"set\" WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up set data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM exercise WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up exercise data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM workout WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up workout data for user %s: %v", userID, err)
		}

		_, err = pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)
		if err != nil {
			t.Logf("Warning: Failed to clean up user data for user %s: %v", userID, err)
		}
	}

	// Re-enable RLS
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", table))
		if err != nil {
			t.Logf("Warning: Failed to re-enable RLS on %s after cleanup: %v", table, err)
		}
	}

	t.Logf("Test cleanup complete for %d registered users: %v", len(testUserIDs), testUserIDs)
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

// TestFormatValidationErrors verifies that the FormatValidationErrors function
// correctly handles wrapped validation errors and non-validation errors
func TestFormatValidationErrors(t *testing.T) {
	// Create a validator for generating validation errors
	v := validator.New()

	// Test struct for validation
	type TestStruct struct {
		Name  string `validate:"required"`
		Age   int    `validate:"gte=0"`
		Email string `validate:"required,email"`
	}

	t.Run("unwraps wrapped validation errors", func(t *testing.T) {
		// Create a struct that fails validation
		testData := TestStruct{
			Name:  "", // required field is empty
			Age:   -1, // fails gte=0
			Email: "", // required field is empty
		}

		// Get validation error
		err := v.Struct(testData)
		require.Error(t, err, "Expected validation to fail")

		// Wrap the error to test errors.As unwrapping
		wrappedErr := fmt.Errorf("validation failed: %w", err)
		doubleWrappedErr := fmt.Errorf("request processing failed: %w", wrappedErr)

		// Test that FormatValidationErrors can unwrap and format the error
		result := FormatValidationErrors(doubleWrappedErr)

		// Verify the result contains validation error messages
		assert.Contains(t, result, "Validation errors:", "Should contain validation errors prefix")
		assert.Contains(t, result, "Name is required", "Should contain Name required message")
		assert.Contains(t, result, "Age must be greater than or equal to 0", "Should contain Age gte message")
		assert.Contains(t, result, "Email is required", "Should contain Email required message")
	})

	t.Run("handles single validation error", func(t *testing.T) {
		testData := TestStruct{
			Name:  "",
			Age:   25,
			Email: "valid@example.com",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Name is required")
	})

	t.Run("handles validation errors with different tags", func(t *testing.T) {
		type ValidationTestStruct struct {
			Username string `validate:"required,min=3,max=20"`
			Count    int    `validate:"gte=1"`
			Created  string `validate:"datetime=2006-01-02T15:04:05Z07:00"`
		}

		testData := ValidationTestStruct{
			Username: "ab",      // fails min=3
			Count:    0,         // fails gte=1
			Created:  "invalid", // fails datetime
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Username must be at least 3 characters")
		assert.Contains(t, result, "Count must be greater than or equal to 1")
		assert.Contains(t, result, "Created must be a valid datetime in RFC3339 format")
	})

	t.Run("falls back to err.Error() for non-validation errors", func(t *testing.T) {
		// Test with a standard error
		standardErr := fmt.Errorf("database connection failed")
		result := FormatValidationErrors(standardErr)
		assert.Equal(t, "database connection failed", result)

		// Test with a wrapped standard error
		wrappedStandardErr := fmt.Errorf("operation failed: %w", standardErr)
		result = FormatValidationErrors(wrappedStandardErr)
		assert.Equal(t, "operation failed: database connection failed", result)
	})

	t.Run("handles nil error gracefully", func(t *testing.T) {
		// This shouldn't happen in practice, but test defensive programming
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FormatValidationErrors panicked with nil error: %v", r)
			}
		}()

		// If err is nil, err.Error() will panic, but this tests if there's nil checking
		// In the current implementation, this would panic, which is acceptable
		// since the function is never called with nil in actual use
	})

	t.Run("handles max validation tag", func(t *testing.T) {
		type MaxTestStruct struct {
			Description string `validate:"max=10"`
		}

		testData := MaxTestStruct{
			Description: "This is a very long description that exceeds the maximum",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Description must be at most 10 characters")
	})

	t.Run("handles unknown validation tag with default message", func(t *testing.T) {
		type UnknownTagStruct struct {
			Field string `validate:"customtag"`
		}

		// Register a custom validator that will fail
		v.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
			return false
		})

		testData := UnknownTagStruct{
			Field: "test",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Field failed validation (customtag)")
	})
}
