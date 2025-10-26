package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name                   string
		clientProvidedID       string
		expectClientID         bool
		expectGeneratedID      bool
	}{
		{
			name:              "no client-provided request ID - should generate",
			clientProvidedID:  "",
			expectClientID:    false,
			expectGeneratedID: true,
		},
		{
			name:              "client-provided request ID - should use it",
			clientProvidedID:  "client-request-123",
			expectClientID:    true,
			expectGeneratedID: false,
		},
		{
			name:              "client-provided UUID - should use it",
			clientProvidedID:  "550e8400-e29b-41d4-a716-446655440000",
			expectClientID:    true,
			expectGeneratedID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedRequestID string

			// Create a test handler that captures the request ID from context
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequestID = GetRequestID(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with RequestID middleware
			handler := RequestID()(nextHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if tt.clientProvidedID != "" {
				req.Header.Set("X-Request-ID", tt.clientProvidedID)
			}

			// Record response
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check that X-Request-ID header is set in response
			responseID := rr.Header().Get("X-Request-ID")
			if responseID == "" {
				t.Error("Expected X-Request-ID header in response, got none")
			}

			// Check if client-provided ID was used
			if tt.expectClientID {
				if responseID != tt.clientProvidedID {
					t.Errorf("X-Request-ID = %v, want %v", responseID, tt.clientProvidedID)
				}
				if capturedRequestID != tt.clientProvidedID {
					t.Errorf("context request ID = %v, want %v", capturedRequestID, tt.clientProvidedID)
				}
			}

			// Check if a new UUID was generated
			if tt.expectGeneratedID {
				if _, err := uuid.Parse(responseID); err != nil {
					t.Errorf("Expected valid UUID, got %v (error: %v)", responseID, err)
				}
				if capturedRequestID != responseID {
					t.Errorf("context request ID = %v, want %v", capturedRequestID, responseID)
				}
			}

			// Verify request ID is stored in context
			if capturedRequestID == "" {
				t.Error("Expected request ID in context, got empty string")
			}

			// Verify response header matches context value
			if responseID != capturedRequestID {
				t.Errorf("Response header ID (%v) does not match context ID (%v)", responseID, capturedRequestID)
			}
		})
	}
}

func TestGetRequestID_NoRequestID(t *testing.T) {
	// Test GetRequestID when no request ID is in context
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	requestID := GetRequestID(req.Context())

	if requestID != "" {
		t.Errorf("Expected empty string when no request ID in context, got %v", requestID)
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	// Test that each request gets a unique ID
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID()(nextHandler)

	// Make multiple requests
	requestIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		requestID := rr.Header().Get("X-Request-ID")
		if requestIDs[requestID] {
			t.Errorf("Duplicate request ID generated: %v", requestID)
		}
		requestIDs[requestID] = true
	}

	if len(requestIDs) != 10 {
		t.Errorf("Expected 10 unique request IDs, got %d", len(requestIDs))
	}
}
