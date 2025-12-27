package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimit creates a rate limiting middleware that limits requests per user
func RateLimit(logger *slog.Logger, requestsPerMinute int64) func(http.Handler) http.Handler {
	// Create in-memory store
	store := memory.NewStore()

	// Create rate limiter with configured rate
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  requestsPerMinute,
	}

	instance := limiter.New(store, rate)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context
			userID, ok := user.Current(r.Context())
			if !ok {
				// If no user ID in context, skip rate limiting
				// This allows unauthenticated endpoints to bypass rate limiting
				next.ServeHTTP(w, r)
				return
			}

			// Get rate limit context for this user
			context, err := instance.Get(r.Context(), userID)
			if err != nil {
				response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "internal error", err)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

			// Check if rate limit exceeded
			if context.Reached {
				// Calculate retry after duration in seconds
				retryAfter := time.Until(time.Unix(context.Reset, 0))
				retryAfterSeconds := int(retryAfter.Seconds())
				if retryAfterSeconds < 0 {
					retryAfterSeconds = 0
				}

				// Set Retry-After header before sending error response
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))

				// Return standardized error response
				response.ErrorJSON(w, r, logger, http.StatusTooManyRequests,
					fmt.Sprintf("rate limit exceeded, retry after %d seconds", retryAfterSeconds), nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
