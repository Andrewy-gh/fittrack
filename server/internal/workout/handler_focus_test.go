package workout

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
