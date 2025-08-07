package auth

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJWKSCache struct {
	mock.Mock
}

func (m *MockJWKSCache) GetUserIDFromToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) EnsureUser(ctx context.Context, userID string) (db.Users, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Users), args.Error(1)
}

type MockDBTX struct {
	mock.Mock
}

func (m *MockDBTX) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments[0])
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *MockDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

func TestAuthenticator_Middleware(t *testing.T) {
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
			mockJWKSCache := &MockJWKSCache{}
			mockUserService := &MockUserService{}
			tt.setupMocks(mockJWKSCache, mockUserService)

			auth := &Authenticator{
				logger:      logger,
				jwkCache:    mockJWKSCache,
				userService: mockUserService,
			}

			var capturedContext context.Context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()

			auth.Middleware(nextHandler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectContext {
				userID, ok := user.Current(capturedContext)
				assert.True(t, ok, "Expected user ID in context")
				assert.Equal(t, tt.expectedUserID, userID)
			}

			mockJWKSCache.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
		})
	}
}

func TestAuthenticator_Middleware_SessionUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name           string
		setupMocks     func(*MockDBTX, *MockJWKSCache, *MockUserService)
		expectedStatus int
		expectContext  bool
		expectedUserID string
	}{
		{
			name: "successfully set session user ID",
			setupMocks: func(mockDB *MockDBTX, mockJWKS *MockJWKSCache, mockUserSvc *MockUserService) {
				mockJWKS.On("GetUserIDFromToken", "valid-token").Return("user-123", nil)
				mockUserSvc.On("EnsureUser", mock.Anything, "user-123").Return(db.Users{UserID: "user-123"}, nil)
				tag := pgconn.NewCommandTag("SET")
				mockDB.On("Exec", mock.Anything, setUserIDQuery, "user-123").Return(tag, nil)
			},
			expectedStatus: http.StatusOK,
			expectContext:  true,
			expectedUserID: "user-123",
		},
		{
			name: "error setting session user ID",
			setupMocks: func(mockDB *MockDBTX, mockJWKS *MockJWKSCache, mockUserSvc *MockUserService) {
				err := fmt.Errorf("database error")
				mockJWKS.On("GetUserIDFromToken", "valid-token").Return("user-123", nil)
				mockUserSvc.On("EnsureUser", mock.Anything, "user-123").Return(db.Users{UserID: "user-123"}, nil)
				tag := pgconn.NewCommandTag("SET")
				mockDB.On("Exec", mock.Anything, setUserIDQuery, "user-123").Return(tag, err)
			},
			expectedStatus: http.StatusInternalServerError,
			expectContext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockDBTX{}
			mockJWKSCache := &MockJWKSCache{}
			mockUserService := &MockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockDB, mockJWKSCache, mockUserService)
			}

			auth := &Authenticator{
				logger:      logger,
				jwkCache:    mockJWKSCache,
				userService: mockUserService,
				dbPool:      mockDB,
			}

			var capturedContext context.Context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("x-stack-access-token", "valid-token")

			w := httptest.NewRecorder()

			auth.Middleware(nextHandler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectContext {
				userID, ok := user.Current(capturedContext)
				assert.True(t, ok, "Expected user ID in context")
				assert.Equal(t, tt.expectedUserID, userID)
			}

			mockJWKSCache.AssertExpectations(t)
			mockUserService.AssertExpectations(t)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestAuthenticator_Middleware_NilDBPool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockJWKSCache := &MockJWKSCache{}
	mockUserService := &MockUserService{}

	mockJWKSCache.On("GetUserIDFromToken", "valid-token").Return("user-123", nil)
	mockUserService.On("EnsureUser", mock.Anything, "user-123").Return(db.Users{UserID: "user-123"}, nil)

	auth := &Authenticator{
		logger:      logger,
		jwkCache:    mockJWKSCache,
		userService: mockUserService,
		dbPool:      nil, // Explicitly nil DB pool
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("x-stack-access-token", "valid-token")

	w := httptest.NewRecorder()

	auth.Middleware(nextHandler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockJWKSCache.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestJWKSCache_GetUserIDFromToken(t *testing.T) {
	// This test verifies the interface implementation
	// More comprehensive tests would require mocking the jwk library
	var _ JWKSProvider = (*JWKSCache)(nil)
}

func TestJWKSCache_GetUserIDFromToken_ErrorCases(t *testing.T) {
	// This is a simplified test - in a real scenario, you'd want to mock the jwk library
	// or use test vectors with known good/bad tokens

	// Test with malformed token
	t.Run("malformed token", func(t *testing.T) {
		cache := &JWKSCache{}
		_, err := cache.GetUserIDFromToken("malformed.token.here")
		assert.Error(t, err, "Expected error for malformed token")
	})

	// Test with empty token
	t.Run("empty token", func(t *testing.T) {
		cache := &JWKSCache{}
		_, err := cache.GetUserIDFromToken("")
		assert.Error(t, err, "Expected error for empty token")
	})
}

func TestErrorResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	response.ErrorJSON(w, req, logger, http.StatusUnauthorized, "test error", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
