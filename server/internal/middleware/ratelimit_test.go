package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit_NoUserID(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create handler with rate limit
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := RateLimit(logger, 5)(nextHandler)

	// Create request without user ID in context
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should pass through without rate limiting
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())

	// Should not have rate limit headers
	assert.Empty(t, rr.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimit_WithinLimit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := "test-user-123"

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := RateLimit(logger, 5)(nextHandler)

	// Make 3 requests (within limit of 5)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := user.WithContext(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// Should succeed
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
		assert.Equal(t, "success", rr.Body.String())

		// Check rate limit headers
		assert.Equal(t, "5", rr.Header().Get("X-RateLimit-Limit"))
		remaining := rr.Header().Get("X-RateLimit-Remaining")
		assert.NotEmpty(t, remaining)

		remainingInt, err := strconv.Atoi(remaining)
		assert.NoError(t, err)
		assert.Equal(t, 5-(i+1), remainingInt, "remaining should decrease")
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := "test-user-456"
	limit := int64(3)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := RateLimit(logger, limit)(nextHandler)

	// Make requests up to the limit
	for i := 0; i < int(limit); i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := user.WithContext(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := user.WithContext(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should return 429
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Check Retry-After header exists
	retryAfter := rr.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter)

	retryAfterInt, err := strconv.Atoi(retryAfter)
	assert.NoError(t, err)
	assert.Greater(t, retryAfterInt, 0)

	// Check response body
	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Contains(t, response["message"], "rate limit exceeded")
	assert.Contains(t, response["message"], "retry after")
}

func TestRateLimit_DifferentUsers(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	user1 := "user-1"
	user2 := "user-2"
	limit := int64(2)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := RateLimit(logger, limit)(nextHandler)

	// User 1 makes 2 requests (hits limit)
	for i := 0; i < int(limit); i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := user.WithContext(req.Context(), user1)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// User 1's next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := user.WithContext(req.Context(), user1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// User 2 should still be able to make requests
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx = user.WithContext(req.Context(), user2)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "user 2 should not be rate limited")
}

func TestRateLimit_ConfigurableLimit(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := "test-user-789"

	tests := []struct {
		name  string
		limit int64
	}{
		{
			name:  "limit of 1",
			limit: 1,
		},
		{
			name:  "limit of 10",
			limit: 10,
		},
		{
			name:  "limit of 100",
			limit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			handler := RateLimit(logger, tt.limit)(nextHandler)

			// Make one request
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			ctx := user.WithContext(req.Context(), userID+tt.name) // Unique user per test
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			// Check that the limit header matches configured limit
			limitHeader := rr.Header().Get("X-RateLimit-Limit")
			assert.Equal(t, strconv.FormatInt(tt.limit, 10), limitHeader)
		})
	}
}

func TestRateLimit_Headers(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := "test-user-headers"
	limit := int64(10)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimit(logger, limit)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := user.WithContext(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Verify all rate limit headers are present
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Limit"), "should have X-RateLimit-Limit header")
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Remaining"), "should have X-RateLimit-Remaining header")
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"), "should have X-RateLimit-Reset header")

	// Verify header values are valid
	limitHeader, err := strconv.ParseInt(rr.Header().Get("X-RateLimit-Limit"), 10, 64)
	assert.NoError(t, err)
	assert.Equal(t, limit, limitHeader)

	remaining, err := strconv.ParseInt(rr.Header().Get("X-RateLimit-Remaining"), 10, 64)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, remaining, int64(0))
	assert.LessOrEqual(t, remaining, limit)

	reset, err := strconv.ParseInt(rr.Header().Get("X-RateLimit-Reset"), 10, 64)
	assert.NoError(t, err)
	assert.Greater(t, reset, time.Now().Unix())
}

func TestRateLimit_RetryAfterFormat(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	userID := "test-user-retry"
	limit := int64(1)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RateLimit(logger, limit)(nextHandler)

	// First request succeeds
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx := user.WithContext(req.Context(), userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Second request is rate limited
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	ctx = user.WithContext(req.Context(), userID)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Verify Retry-After is a valid integer in seconds
	retryAfter := rr.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter)

	retryAfterInt, err := strconv.Atoi(retryAfter)
	assert.NoError(t, err, "Retry-After should be a valid integer")
	assert.GreaterOrEqual(t, retryAfterInt, 0, "Retry-After should be non-negative")
	assert.LessOrEqual(t, retryAfterInt, 60, "Retry-After should be reasonable (within 60 seconds for 1-minute window)")
}
