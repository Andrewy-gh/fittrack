package workout

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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWorkoutRepository implements the WorkoutRepository interface for testing
type MockWorkoutRepository struct {
	mock.Mock
}

func (m *MockWorkoutRepository) SaveWorkout(ctx context.Context, reformatted *ReformattedRequest, userID string) error {
	args := m.Called(ctx, reformatted, userID)
	return args.Error(0)
}

func (m *MockWorkoutRepository) ListWorkouts(ctx context.Context, userID string) ([]db.Workout, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Workout), args.Error(1)
}

func (m *MockWorkoutRepository) GetWorkoutWithSets(ctx context.Context, id int32, userID string) ([]db.GetWorkoutWithSetsRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetWorkoutWithSetsRow), args.Error(1)
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

func BenchmarkWorkoutHandler_ListWorkouts(b *testing.B) {
	userID := "test-user-id"
	mockRepo := &MockWorkoutRepository{}
	mockRepo.On("ListWorkouts", mock.Anything, userID).Return([]db.Workout{
		{ID: 1, Date: pgtype.Timestamptz{Time: time.Now(), Valid: true}},
	}, nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validator := validator.New()
	service := &WorkoutService{repo: mockRepo, logger: logger}
	handler := NewHandler(logger, validator, service)

	ctx := context.WithValue(context.Background(), user.UserIDKey, userID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/workouts", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		handler.ListWorkouts(w, req)
	}
}
