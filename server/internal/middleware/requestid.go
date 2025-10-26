package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// RequestID creates a middleware that adds a unique request ID to each request.
// It generates a UUID v4 for each request or uses a client-provided X-Request-ID header.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client provided a request ID
			requestID := r.Header.Get("X-Request-ID")

			// Generate a new UUID v4 if no request ID was provided
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)

			// Store request ID in context for access by handlers
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
