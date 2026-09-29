package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

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
