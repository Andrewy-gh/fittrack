package workout

import (
	"bytes"
	"context"
	"encoding/json"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

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
				m.On("SaveWorkoutWithID", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(int32(1), nil)
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
				m.On("SaveWorkoutWithID", mock.Anything, mock.AnythingOfType("*workout.ReformattedRequest"), userID).Return(int32(0), assert.AnError)
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
				m.On("GetWorkout", mock.Anything, id, userID).Return(db.Workout{}, pgx.ErrNoRows)
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
