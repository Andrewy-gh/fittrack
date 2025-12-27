// Package request provides utilities for managing request-scoped context values.
package request

import "context"

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// RequestIDKey is the context key for storing request IDs.
const RequestIDKey contextKey = "request_id"

// WithRequestID adds a request ID to the context.
// Returns a new context with the request ID value set.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present in the context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
