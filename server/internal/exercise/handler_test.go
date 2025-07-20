package exercise

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExerciseRepository implements the ExerciseRepository interface for testing
type MockExerciseRepository struct {
	mock.Mock
}

func (m *MockExerciseRepository) ListExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) GetExercise(ctx context.Context, id int32, userID string) (db.Exercise, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) GetExerciseWithSets(ctx context.Context, id int32, userID string) ([]db.GetExerciseWithSetsRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetExerciseWithSetsRow), args.Error(1)
}

func (m *MockExerciseRepository) GetOrCreateExercise(ctx context.Context, name string, userID string) (db.Exercise, error) {
	args := m.Called(ctx, name, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) GetOrCreateExerciseTx(ctx context.Context, qtx *db.Queries, name string, userID string) (db.Exercise, error) {
	args := m.Called(ctx, qtx, name, userID)
	return args.Get(0).(db.Exercise), args.Error(1)
}

type errorResponse struct {
	Message string `json:"message"`
}

func TestExerciseHandler_ListExercises(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		setupMock     func(*MockExerciseRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name: "successful fetch",
			setupMock: func(m *MockExerciseRepository) {
				m.On("ListExercises", mock.Anything, userID).Return([]db.Exercise{
					{ID: 1, Name: "Bench Press"},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name: "internal server error",
			setupMock: func(m *MockExerciseRepository) {
				m.On("ListExercises", mock.Anything, userID).Return([]db.Exercise{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectJSON:    true,
			expectedError: "Failed to list exercises",
		},
		{
			name:          "unauthenticated user",
			setupMock:     func(m *MockExerciseRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectJSON:    true,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockExerciseRepository{}
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := NewService(logger, mockRepo)
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(tt.ctx)
			w := httptest.NewRecorder()

			// Execute
			handler.ListExercises(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				if tt.expectedError != "" {
					assertJSONError(t, w, tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExerciseHandler_GetExerciseWithSets(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		exerciseID    string
		setupMock     func(*MockExerciseRepository, int32)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name:       "successful fetch",
			exerciseID: "1",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExerciseWithSets", mock.Anything, id, userID).Return([]db.GetExerciseWithSetsRow{
					{ExerciseID: id, ExerciseName: "Bench Press"},
				}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:          "invalid exercise ID",
			exerciseID:    "invalid",
			setupMock:     func(m *MockExerciseRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid exercise ID",
		},
		{
			name:       "service error",
			exerciseID: "999",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExerciseWithSets", mock.Anything, id, userID).Return([]db.GetExerciseWithSetsRow{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get exercise with sets",
		},
		{
			name:       "exercise not found",
			exerciseID: "999",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExerciseWithSets", mock.Anything, id, userID).Return([]db.GetExerciseWithSetsRow{}, nil)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusNotFound,
			expectJSON:    true,
			expectedError: "No sets found for this exercise",
		},
		{
			name:          "unauthenticated user",
			exerciseID:    "1",
			setupMock:     func(m *MockExerciseRepository, id int32) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectJSON:    true,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockExerciseRepository{}
			var id int32
			if tt.exerciseID != "" {
				idInt, err := strconv.Atoi(tt.exerciseID)
				if err == nil {
					id = int32(idInt)
				}
			}
			tt.setupMock(mockRepo, id)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := NewService(logger, mockRepo)
			handler := NewHandler(logger, validator, service)

			req := httptest.NewRequest("GET", "/api/exercises/"+tt.exerciseID, nil).WithContext(tt.ctx)
			req.SetPathValue("id", tt.exerciseID)
			w := httptest.NewRecorder()

			// Execute
			handler.GetExerciseWithSets(w, req)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				if tt.expectedError != "" {
					assertJSONError(t, w, tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestExerciseHandler_GetOrCreateExercise(t *testing.T) {
	userID := "test-user-id"

	tests := []struct {
		name          string
		requestBody   interface{}
		setupMock     func(*MockExerciseRepository)
		ctx           context.Context
		expectedCode  int
		expectJSON    bool
		expectedError string
	}{
		{
			name:        "successful creation",
			requestBody: CreateExerciseRequest{Name: "Bench Press"},
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetOrCreateExercise", mock.Anything, "Bench Press", userID).Return(db.Exercise{ID: 1, Name: "Bench Press"}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
		},
		{
			name:          "invalid JSON",
			requestBody:   "invalid json string",
			setupMock:     func(m *MockExerciseRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "Failed to decode request body",
		},
		{
			name:          "validation error",
			requestBody:   CreateExerciseRequest{Name: ""},
			setupMock:     func(m *MockExerciseRepository) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "Validation failed",
		},
		{
			name:        "service error",
			requestBody: CreateExerciseRequest{Name: "Bench Press"},
			setupMock: func(m *MockExerciseRepository) {
				m.On("GetOrCreateExercise", mock.Anything, "Bench Press", userID).Return(db.Exercise{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectJSON:    true,
			expectedError: "Failed to get or create exercise",
		},
		{
			name:          "unauthenticated user",
			requestBody:   CreateExerciseRequest{Name: "Bench Press"},
			setupMock:     func(m *MockExerciseRepository) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
			expectJSON:    true,
			expectedError: "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := &MockExerciseRepository{}
			tt.setupMock(mockRepo)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			validator := validator.New()
			service := NewService(logger, mockRepo)
			handler := NewHandler(logger, validator, service)

			// Prepare request
			var req *http.Request
			if tt.requestBody != nil {
				var body []byte
				var err error

				switch v := tt.requestBody.(type) {
				case string:
					body = []byte(v)
				default:
					body, err = json.Marshal(v)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
				}
				req = httptest.NewRequest("POST", "/api/exercises", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest("POST", "/api/exercises", nil)
			}

			w := httptest.NewRecorder()

			// Execute
			handler.GetOrCreateExercise(w, req.WithContext(tt.ctx))

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectJSON {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				if tt.expectedError != "" {
					assertJSONError(t, w, tt.expectedError)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// Add this function to help with test assertions
func assertJSONError(t *testing.T, w *httptest.ResponseRecorder, expectedError string) {
	var resp errorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Message, expectedError)
}
