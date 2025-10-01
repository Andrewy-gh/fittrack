package exercise

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

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *MockExerciseRepository) GetRecentSetsForExercise(ctx context.Context, id int32, userID string) ([]db.GetRecentSetsForExerciseRow, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).([]db.GetRecentSetsForExerciseRow), args.Error(1)
}

func (m *MockExerciseRepository) DeleteExercise(ctx context.Context, id int32, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
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
			name:       "empty results - returns 200 with empty array",
			exerciseID: "999",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExerciseWithSets", mock.Anything, id, userID).Return([]db.GetExerciseWithSetsRow{}, nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusOK,
			expectJSON:   true,
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

func TestExerciseHandler_GetRecentSetsForExercise(t *testing.T) {
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
				m.On("GetRecentSetsForExercise", mock.Anything, id, userID).Return([]db.GetRecentSetsForExerciseRow{
					{SetID: 1, Reps: 10},
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
				m.On("GetRecentSetsForExercise", mock.Anything, id, userID).Return([]db.GetRecentSetsForExerciseRow{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Failed to get recent sets for exercise",
		},
		{
			name:          "unauthenticated user",
			exerciseID:    "1",
			setupMock:     func(m *MockExerciseRepository, id int32) {},
			ctx:           context.Background(),
			expectedCode:  http.StatusUnauthorized,
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

			req := httptest.NewRequest("GET", "/api/exercises/"+tt.exerciseID+"/recent-sets", nil).WithContext(tt.ctx)
			req.SetPathValue("id", tt.exerciseID)
			w := httptest.NewRecorder()

			// Execute
			handler.GetRecentSetsForExercise(w, req)

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

func TestExerciseHandler_DeleteExercise(t *testing.T) {
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
			name:       "successful deletion",
			exerciseID: "1",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExercise", mock.Anything, id, userID).Return(db.Exercise{ID: id, Name: "Bench Press"}, nil)
				m.On("DeleteExercise", mock.Anything, id, userID).Return(nil)
			},
			ctx:          context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode: http.StatusNoContent,
			expectJSON:   false,
		},
		{
			name:          "missing exercise ID",
			exerciseID:    "",
			setupMock:     func(m *MockExerciseRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "Missing exercise ID",
		},
		{
			name:          "invalid exercise ID",
			exerciseID:    "invalid",
			setupMock:     func(m *MockExerciseRepository, id int32) {},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusBadRequest,
			expectJSON:    true,
			expectedError: "Invalid exercise ID",
		},
		{
			name:       "exercise not found",
			exerciseID: "999",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExercise", mock.Anything, id, userID).Return(db.Exercise{}, assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusNotFound,
			expectJSON:    true,
			expectedError: "exercise not found",
		},
		{
			name:       "delete operation fails",
			exerciseID: "1",
			setupMock: func(m *MockExerciseRepository, id int32) {
				m.On("GetExercise", mock.Anything, id, userID).Return(db.Exercise{ID: id, Name: "Bench Press"}, nil)
				m.On("DeleteExercise", mock.Anything, id, userID).Return(assert.AnError)
			},
			ctx:           context.WithValue(context.Background(), user.UserIDKey, userID),
			expectedCode:  http.StatusInternalServerError,
			expectJSON:    true,
			expectedError: "Failed to delete exercise",
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

			req := httptest.NewRequest("DELETE", "/api/exercises/"+tt.exerciseID, nil).WithContext(tt.ctx)
			req.SetPathValue("id", tt.exerciseID)
			w := httptest.NewRecorder()

			// Execute
			handler.DeleteExercise(w, req)

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

// === INTEGRATION TESTS (RLS Testing) ===
// These tests use a real database connection to test Row Level Security policies

func TestExerciseHandlerRLSIntegration(t *testing.T) {
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

	// Initialize repository and service
	exerciseRepo := NewRepository(logger, queries, pool)
	exerciseService := NewService(logger, exerciseRepo)
	handler := NewHandler(logger, validator, exerciseService)

	// Test data
	userAID := "test-user-a"
	userBID := "test-user-b"

	// Create test data with proper RLS context
	exerciseAID := setupTestExercise(t, pool, userAID, "User A's Bench Press")
	exerciseBID := setupTestExercise(t, pool, userBID, "User B's Squat")

	t.Run("Scenario1_UserA_CanRetrieveOwnExercises", func(t *testing.T) {
		// Set RLS context for User A
		ctx := setTestUserContext(context.Background(), t, pool, userAID)
		ctx = user.WithContext(ctx, userAID)

		req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListExercises(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var exercises []db.Exercise
		err := json.Unmarshal(w.Body.Bytes(), &exercises)
		assert.NoError(t, err)

		// User A should see only their own exercise
		assert.Len(t, exercises, 1)
		assert.Equal(t, exerciseAID, exercises[0].ID)
		assert.Equal(t, userAID, exercises[0].UserID)
		assert.Equal(t, "User A's Bench Press", exercises[0].Name)
	})

	t.Run("Scenario2_UserB_CannotRetrieveUserAExercises", func(t *testing.T) {
		// Set RLS context for User B
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListExercises(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var exercises []db.Exercise
		err := json.Unmarshal(w.Body.Bytes(), &exercises)
		assert.NoError(t, err)

		// User B should see only their own exercise, not User A's
		assert.Len(t, exercises, 1)
		assert.Equal(t, exerciseBID, exercises[0].ID)
		assert.Equal(t, userBID, exercises[0].UserID)
		assert.Equal(t, "User B's Squat", exercises[0].Name)
		assert.NotEqual(t, exerciseAID, exercises[0].ID, "User B should not see User A's exercise")
	})

	t.Run("Scenario3_AnonymousUser_CannotAccessExercises", func(t *testing.T) {
		// No user context set (anonymous user)
		ctx := context.Background()

		req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ListExercises(w, req)

		// Should get unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "user not authenticated")
	})

	t.Run("Scenario4_GetSpecificExercise_UserB_CannotAccessUserAExercise", func(t *testing.T) {
		// User B tries to access User A's specific exercise
		ctx := setTestUserContext(context.Background(), t, pool, userBID)
		ctx = user.WithContext(ctx, userBID)

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/exercises/%d", exerciseAID), nil).WithContext(ctx)
		req.SetPathValue("id", fmt.Sprintf("%d", exerciseAID))
		w := httptest.NewRecorder()

		handler.GetExerciseWithSets(w, req)

		// Should return 200 with empty array due to RLS filtering (no sets visible)
		assert.Equal(t, http.StatusOK, w.Code)
		var exerciseWithSets []db.GetExerciseWithSetsRow
		err := json.Unmarshal(w.Body.Bytes(), &exerciseWithSets)
		assert.NoError(t, err)
		assert.Empty(t, exerciseWithSets, "User B should not see User A's exercise sets due to RLS")
	})

	t.Run("Scenario5_CreateExercise_UserIsolation", func(t *testing.T) {
		// User A creates an exercise
		ctxA := setTestUserContext(context.Background(), t, pool, userAID)
		ctxA = user.WithContext(ctxA, userAID)

		createReq := CreateExerciseRequest{Name: "User A's Deadlift"}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/exercises", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxA)
		w := httptest.NewRecorder()

		handler.GetOrCreateExercise(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var createdExercise db.Exercise
		err := json.Unmarshal(w.Body.Bytes(), &createdExercise)
		assert.NoError(t, err)
		assert.Equal(t, "User A's Deadlift", createdExercise.Name)
		assert.Equal(t, userAID, createdExercise.UserID)

		// User B should not see User A's newly created exercise
		ctxB := setTestUserContext(context.Background(), t, pool, userBID)
		ctxB = user.WithContext(ctxB, userBID)

		listReq := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctxB)
		listW := httptest.NewRecorder()

		handler.ListExercises(listW, listReq)

		assert.Equal(t, http.StatusOK, listW.Code)
		var exercises []db.Exercise
		err = json.Unmarshal(listW.Body.Bytes(), &exercises)
		assert.NoError(t, err)

		// User B should still see only their own exercise (not the new one from User A)
		assert.Len(t, exercises, 1)
		assert.Equal(t, "User B's Squat", exercises[0].Name)
		assert.NotEqual(t, "User A's Deadlift", exercises[0].Name)
	})

	t.Run("Scenario6_ConcurrentRequests_ProperIsolation", func(t *testing.T) {
		// Test concurrent requests from different users
		var wg sync.WaitGroup
		results := make(map[string][]db.Exercise)
		mu := sync.Mutex{}

		// User A request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userAID)
			ctx = user.WithContext(ctx, userAID)

			req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListExercises(w, req)

			if w.Code == http.StatusOK {
				var exercises []db.Exercise
				json.Unmarshal(w.Body.Bytes(), &exercises)
				mu.Lock()
				results[userAID] = exercises
				mu.Unlock()
			}
		}()

		// User B request
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := setTestUserContext(context.Background(), t, pool, userBID)
			ctx = user.WithContext(ctx, userBID)

			req := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ListExercises(w, req)

			if w.Code == http.StatusOK {
				var exercises []db.Exercise
				json.Unmarshal(w.Body.Bytes(), &exercises)
				mu.Lock()
				results[userBID] = exercises
				mu.Unlock()
			}
		}()

		wg.Wait()

		// Verify isolation - each user should only see their own exercises
		assert.True(t, len(results[userAID]) >= 1, "User A should have at least one exercise")
		assert.True(t, len(results[userBID]) >= 1, "User B should have at least one exercise")

		// Check that User A's results contain their exercise
		found := false
		for _, ex := range results[userAID] {
			if ex.ID == exerciseAID {
				found = true
				assert.Equal(t, userAID, ex.UserID)
				break
			}
		}
		assert.True(t, found, "User A should see their own exercise")

		// Check that User B's results contain their exercise and not User A's
		found = false
		for _, ex := range results[userBID] {
			assert.NotEqual(t, exerciseAID, ex.ID, "User B should not see User A's exercise")
			assert.Equal(t, userBID, ex.UserID, "User B should only see their own exercises")
			if ex.ID == exerciseBID {
				found = true
			}
		}
		assert.True(t, found, "User B should see their own exercise")
	})

	t.Run("Scenario7_DeleteExercise_UserA_CanDeleteOwnExercise", func(t *testing.T) {
		// User A creates an exercise to delete
		ctxA := setTestUserContext(context.Background(), t, pool, userAID)
		ctxA = user.WithContext(ctxA, userAID)

		createReq := CreateExerciseRequest{Name: "User A's Exercise To Delete"}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/exercises", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxA)
		w := httptest.NewRecorder()

		handler.GetOrCreateExercise(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var createdExercise db.Exercise
		err := json.Unmarshal(w.Body.Bytes(), &createdExercise)
		assert.NoError(t, err)

		// User A deletes their own exercise
		deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/exercises/%d", createdExercise.ID), nil).WithContext(ctxA)
		deleteReq.SetPathValue("id", fmt.Sprintf("%d", createdExercise.ID))
		deleteW := httptest.NewRecorder()

		handler.DeleteExercise(deleteW, deleteReq)
		assert.Equal(t, http.StatusNoContent, deleteW.Code)

		// Verify exercise is deleted - listing should not include it
		listReq := httptest.NewRequest("GET", "/api/exercises", nil).WithContext(ctxA)
		listW := httptest.NewRecorder()

		handler.ListExercises(listW, listReq)
		assert.Equal(t, http.StatusOK, listW.Code)

		var exercises []db.Exercise
		err = json.Unmarshal(listW.Body.Bytes(), &exercises)
		assert.NoError(t, err)

		for _, ex := range exercises {
			assert.NotEqual(t, createdExercise.ID, ex.ID, "Deleted exercise should not appear in list")
		}
	})

	t.Run("Scenario8_DeleteExercise_UserB_CannotDeleteUserAExercise", func(t *testing.T) {
		// User B tries to delete User A's exercise
		ctxB := setTestUserContext(context.Background(), t, pool, userBID)
		ctxB = user.WithContext(ctxB, userBID)

		deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/exercises/%d", exerciseAID), nil).WithContext(ctxB)
		deleteReq.SetPathValue("id", fmt.Sprintf("%d", exerciseAID))
		w := httptest.NewRecorder()

		handler.DeleteExercise(w, deleteReq)

		// Should return 404 due to RLS - User B cannot see or delete User A's exercise
		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "exercise not found")
	})

	t.Run("Scenario9_DeleteExercise_CascadeDeletesSets", func(t *testing.T) {
		// User A creates an exercise and adds sets to it
		ctxA := setTestUserContext(context.Background(), t, pool, userAID)
		ctxA = user.WithContext(ctxA, userAID)

		// Create exercise
		createReq := CreateExerciseRequest{Name: "User A's Exercise With Sets"}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/exercises", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxA)
		w := httptest.NewRecorder()

		handler.GetOrCreateExercise(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var createdExercise db.Exercise
		err := json.Unmarshal(w.Body.Bytes(), &createdExercise)
		assert.NoError(t, err)

		// Add sets to the exercise
		setupTestSetsForExercise(t, pool, userAID, createdExercise.ID, 3)

		// Verify sets exist
		getReq := httptest.NewRequest("GET", fmt.Sprintf("/api/exercises/%d", createdExercise.ID), nil).WithContext(ctxA)
		getReq.SetPathValue("id", fmt.Sprintf("%d", createdExercise.ID))
		getW := httptest.NewRecorder()

		handler.GetExerciseWithSets(getW, getReq)
		assert.Equal(t, http.StatusOK, getW.Code)

		var exerciseWithSets []db.GetExerciseWithSetsRow
		err = json.Unmarshal(getW.Body.Bytes(), &exerciseWithSets)
		assert.NoError(t, err)
		assert.Len(t, exerciseWithSets, 3, "Exercise should have 3 sets")

		// Delete the exercise
		deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/exercises/%d", createdExercise.ID), nil).WithContext(ctxA)
		deleteReq.SetPathValue("id", fmt.Sprintf("%d", createdExercise.ID))
		deleteW := httptest.NewRecorder()

		handler.DeleteExercise(deleteW, deleteReq)
		assert.Equal(t, http.StatusNoContent, deleteW.Code)

		// Verify sets were cascade deleted - trying to get exercise with sets should return empty
		getReq2 := httptest.NewRequest("GET", fmt.Sprintf("/api/exercises/%d", createdExercise.ID), nil).WithContext(ctxA)
		getReq2.SetPathValue("id", fmt.Sprintf("%d", createdExercise.ID))
		getW2 := httptest.NewRecorder()

		handler.GetExerciseWithSets(getW2, getReq2)
		assert.Equal(t, http.StatusOK, getW2.Code)

		var exerciseWithSets2 []db.GetExerciseWithSetsRow
		err = json.Unmarshal(getW2.Body.Bytes(), &exerciseWithSets2)
		assert.NoError(t, err)
		assert.Empty(t, exerciseWithSets2, "Exercise and its sets should be deleted")
	})

	t.Run("Scenario10_DeleteExercise_AnonymousUser_CannotDelete", func(t *testing.T) {
		// Anonymous user (no context) tries to delete an exercise
		ctx := context.Background()

		deleteReq := httptest.NewRequest("DELETE", fmt.Sprintf("/api/exercises/%d", exerciseAID), nil).WithContext(ctx)
		deleteReq.SetPathValue("id", fmt.Sprintf("%d", exerciseAID))
		w := httptest.NewRecorder()

		handler.DeleteExercise(w, deleteReq)

		// Should return 401 Unauthorized
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp errorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "user not authenticated")
	})
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

	_, err = pool.Exec(ctx, "ALTER TABLE exercise ENABLE ROW LEVEL SECURITY")
	require.NoError(t, err, "Failed to enable RLS on exercise table")

	// RLS policies should already exist from migration - no need to recreate them
	// Just verify they exist
	var policyCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM pg_policies WHERE tablename IN ('users', 'exercise')").Scan(&policyCount)
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

func setupTestExercise(t *testing.T, pool *pgxpool.Pool, userID, name string) int32 {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Create exercise
	var exerciseID int32
	err := pool.QueryRow(ctx,
		"INSERT INTO exercise (name, user_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id",
		name, userID).Scan(&exerciseID)
	require.NoError(t, err, "Failed to create test exercise for user %s", userID)

	return exerciseID
}

func setupTestSetsForExercise(t *testing.T, pool *pgxpool.Pool, userID string, exerciseID int32, numSets int) {
	t.Helper()
	ctx := context.Background()

	// Set user context for RLS
	ctx = testutils.SetTestUserContext(ctx, t, pool, userID)

	// Create a workout for the sets
	var workoutID int32
	err := pool.QueryRow(ctx,
		"INSERT INTO workout (date, user_id, created_at, updated_at) VALUES (NOW(), $1, NOW(), NOW()) RETURNING id",
		userID).Scan(&workoutID)
	require.NoError(t, err, "Failed to create test workout for sets")

	// Create sets for the exercise
	for i := 0; i < numSets; i++ {
		_, err := pool.Exec(ctx,
			`INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
			exerciseID, workoutID, 100+i*10, 10-i, "normal", userID, 1, i+1)
		require.NoError(t, err, "Failed to create test set %d", i+1)
	}
}

func setTestUserContext(ctx context.Context, t *testing.T, pool *pgxpool.Pool, userID string) context.Context {
	t.Helper()
	return testutils.SetTestUserContext(ctx, t, pool, userID)
}

func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable RLS temporarily for cleanup
	_, err := pool.Exec(ctx, "ALTER TABLE users DISABLE ROW LEVEL SECURITY; ALTER TABLE exercise DISABLE ROW LEVEL SECURITY;")
	if err != nil {
		t.Logf("Warning: Failed to disable RLS for cleanup: %v", err)
	}

	// Clean up test data
	_, err = pool.Exec(ctx, "DELETE FROM exercise WHERE user_id IN ('test-user-a', 'test-user-b')")
	if err != nil {
		t.Logf("Warning: Failed to clean up exercise data: %v", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM users WHERE user_id IN ('test-user-a', 'test-user-b')")
	if err != nil {
		t.Logf("Warning: Failed to clean up user data: %v", err)
	}

	// Re-enable RLS
	_, err = pool.Exec(ctx, "ALTER TABLE users ENABLE ROW LEVEL SECURITY; ALTER TABLE exercise ENABLE ROW LEVEL SECURITY;")
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
