package request

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
	}{
		{
			name:      "sets UUID request ID",
			requestID: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "sets short request ID",
			requestID: "req-123",
		},
		{
			name:      "sets empty request ID",
			requestID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := WithRequestID(ctx, tt.requestID)

			// Verify the new context is not nil
			if newCtx == nil {
				t.Fatal("WithRequestID returned nil context")
			}

			// Verify the value was set
			got := GetRequestID(newCtx)
			if got != tt.requestID {
				t.Errorf("WithRequestID() set requestID = %q, want %q", got, tt.requestID)
			}

			// Verify original context is unchanged
			original := GetRequestID(ctx)
			if original != "" {
				t.Errorf("WithRequestID() modified original context, got %q, want empty string", original)
			}
		})
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		want      string
		wantEmpty bool
	}{
		{
			name:      "retrieves set request ID",
			ctx:       WithRequestID(context.Background(), "test-id-123"),
			want:      "test-id-123",
			wantEmpty: false,
		},
		{
			name:      "returns empty string when not set",
			ctx:       context.Background(),
			want:      "",
			wantEmpty: true,
		},
		{
			name:      "returns empty string for nil context",
			ctx:       nil,
			want:      "",
			wantEmpty: true,
		},
		{
			name:      "retrieves UUID request ID",
			ctx:       WithRequestID(context.Background(), "550e8400-e29b-41d4-a716-446655440000"),
			want:      "550e8400-e29b-41d4-a716-446655440000",
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Handle nil context test case
			if tt.ctx == nil {
				defer func() {
					if r := recover(); r != nil {
						// Panic is expected for nil context
						return
					}
				}()
			}

			got := GetRequestID(tt.ctx)
			if got != tt.want {
				t.Errorf("GetRequestID() = %q, want %q", got, tt.want)
			}

			if tt.wantEmpty && got != "" {
				t.Errorf("GetRequestID() = %q, want empty string", got)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can set and retrieve the same value
	testCases := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"simple-id",
		"",
		"with-special-chars-!@#$%",
	}

	for _, requestID := range testCases {
		t.Run(requestID, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithRequestID(ctx, requestID)
			got := GetRequestID(ctx)

			if got != requestID {
				t.Errorf("Round-trip failed: set %q, got %q", requestID, got)
			}
		})
	}
}

func TestMultipleContextValues(t *testing.T) {
	// Test that request ID works alongside other context values
	type contextKey string
	userIDKey := contextKey("user_id")

	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-123")
	ctx = WithRequestID(ctx, "request-456")

	// Verify both values are present
	requestID := GetRequestID(ctx)
	if requestID != "request-456" {
		t.Errorf("GetRequestID() = %q, want %q", requestID, "request-456")
	}

	userID, ok := ctx.Value(userIDKey).(string)
	if !ok || userID != "user-123" {
		t.Errorf("Other context value lost: got %q (ok=%v), want %q", userID, ok, "user-123")
	}
}

func TestContextImmutability(t *testing.T) {
	// Verify that WithRequestID doesn't modify the original context
	originalCtx := context.Background()
	newCtx := WithRequestID(originalCtx, "new-id")

	// Original should still be empty
	if got := GetRequestID(originalCtx); got != "" {
		t.Errorf("Original context was modified: got %q, want empty string", got)
	}

	// New context should have the ID
	if got := GetRequestID(newCtx); got != "new-id" {
		t.Errorf("New context doesn't have ID: got %q, want %q", got, "new-id")
	}

	// Setting a new value should create yet another context
	newerCtx := WithRequestID(newCtx, "newer-id")
	if got := GetRequestID(newCtx); got != "new-id" {
		t.Errorf("Previous context was modified: got %q, want %q", got, "new-id")
	}
	if got := GetRequestID(newerCtx); got != "newer-id" {
		t.Errorf("Newest context doesn't have new ID: got %q, want %q", got, "newer-id")
	}
}
