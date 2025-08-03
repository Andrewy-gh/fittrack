package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJWKSCache implements the JWKSProvider interface for testing
type MockJWKSCache struct {
	mock.Mock
}

func (m *MockJWKSCache) GetUserIDFromToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

// MockUserService implements the UserServiceProvider interface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) EnsureUser(ctx context.Context, userID string) (db.Users, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Users), args.Error(1)
}

func TestAuthenticator_Middleware(t *testing.T) {
	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		setupMocks     func(*MockJWKSCache, *MockUserService)
		expectedStatus int
		expectContext  bool
		expectedUserID string
	}{
		{
			name: "bypass non-API paths",
			path: "/health",
			headers: map[string]string{
				"x-stack-access-token": "valid-token",
			},
			setupMocks: func(jwkCache *MockJWKSCache, userService *MockUserService) {
				// No mocks needed as middleware should bypass
			},
			expectedStatus: http.StatusOK,
			expectContext:  false,
		},
		{
			name:    "missing access token",
			path:    "/api/test",
			headers: map[string]string{},
			setupMocks: func(jwkCache *MockJWKSCache, userService *MockUserService) {
				// No mocks needed as middleware should return early
			},
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name: "invalid access token",
			path: "/api/test",
			headers: map[string]string{
				"x-stack-access-token": "invalid-token",
			},
			setupMocks: func(jwkCache *MockJWKSCache, userService *MockUserService) {
				jwkCache.On("GetUserIDFromToken", "invalid-token").Return("", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name: "user service failure",
			path: "/api/test",
			headers: map[string]string{
				"x-stack-access-token": "valid-token",
			},
			setupMocks: func(jwkCache *MockJWKSCache, userService *MockUserService) {
				jwkCache.On("GetUserIDFromToken", "valid-token").Return("user-123", nil)
				userService.On("EnsureUser", mock.Anything, "user-123").Return(db.Users{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectContext:  false,
		},
		{
			name: "successful authentication with existing user",
			path: "/api/test",
			headers: map[string]string{
				"x-stack-access-token": "valid-token",
			},
			setupMocks: func(jwkCache *MockJWKSCache, userService *MockUserService) {
				jwkCache.On("GetUserIDFromToken", "valid-token").Return("user-123", nil)
				userService.On("EnsureUser", mock.Anything, "user-123").Return(db.Users{UserID: "user-123"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectContext:  true,
			expectedUserID: "user-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockJWKSCache := &MockJWKSCache{}
			mockUserService := &MockUserService{}
			tt.setupMocks(mockJWKSCache, mockUserService)

			// Create authenticator
			auth := &Authenticator{
				logger:      logger,
				jwkCache:    mockJWKSCache,
				userService: mockUserService,
			}

			// Create test handler that verifies context
			var capturedContext context.Context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			// Create test request
			req := httptest.NewRequest("GET", tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute middleware
			auth.Middleware(nextHandler).ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert context
			if tt.expectContext {
				userID, ok := user.Current(capturedContext)
				assert.True(t, ok, "Expected user ID in context")
				assert.Equal(t, tt.expectedUserID, userID)
			}

			// Assert mocks
			mockJWKSCache.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

func TestAuthenticator_Middleware_UserCreationFlow(t *testing.T) {
	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create mocks
	mockJWKSCache := &MockJWKSCache{}
	mockUserService := &MockUserService{}

	// Setup expectations for user creation flow
	mockJWKSCache.On("GetUserIDFromToken", "valid-token").Return("new-user-456", nil)
	mockUserService.On("EnsureUser", mock.Anything, "new-user-456").Return(db.Users{UserID: "new-user-456"}, nil)

	// Create authenticator
	auth := &Authenticator{
		logger:      logger,
		jwkCache:    mockJWKSCache,
		userService: mockUserService,
	}

	// Create test handler that verifies context
	var capturedContext context.Context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	// Create test request
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-stack-access-token", "valid-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute middleware
	auth.Middleware(nextHandler).ServeHTTP(w, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert context contains the new user ID
	userID, ok := user.Current(capturedContext)
	assert.True(t, ok, "Expected user ID in context")
	assert.Equal(t, "new-user-456", userID)

	// Assert mocks
	mockJWKSCache.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// TestJWKSCache_GetUserIDFromToken tests the JWKSCache implementation
func TestJWKSCache_GetUserIDFromToken(t *testing.T) {
	// This test would require more sophisticated mocking of the underlying jwk library
	// For now, we'll just verify that the method signature works as expected
	// A more complete implementation would test the actual token parsing logic

	// Verify that JWKSCache implements JWKSProvider interface
	var _ JWKSProvider = (*JWKSCache)(nil)
}

// TestErrorResponse verifies that error responses are properly formatted
func TestErrorResponse(t *testing.T) {
	// Create a test logger that discards output
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create test response recorder and request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	// Call ErrorJSON directly to verify format
	response.ErrorJSON(w, req, logger, http.StatusUnauthorized, "test error", nil)

	// Assert status code
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Assert content type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
