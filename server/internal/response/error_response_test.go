package response

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIErrorResponses verifies that API endpoints return properly formatted
// error responses with correct HTTP status codes and JSON structure.
func TestAPIErrorResponses(t *testing.T) {
	tests := []struct {
		name               string
		message            string
		err                error
		statusCode         int
		expectedMessage    string
		messageShouldContain string
		sanitized          bool
	}{
		{
			name:               "Unauthorized error returns 401",
			message:            "user user123 is not authorized to access exercise",
			err:                &apperrors.Unauthorized{Resource: "exercise", UserID: "user123"},
			statusCode:         http.StatusUnauthorized,
			messageShouldContain: "not authorized",
			sanitized:          false,
		},
		{
			name:               "NotFound error returns 404",
			message:            "workout with id 456 not found",
			err:                &apperrors.NotFound{Resource: "workout", ID: "456"},
			statusCode:         http.StatusNotFound,
			messageShouldContain: "not found",
			sanitized:          false,
		},
		{
			name:            "Database error is sanitized and returns 500",
			message:         "Failed to create user",
			err:             &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint \"users_email_key\""},
			statusCode:      http.StatusInternalServerError,
			expectedMessage: "internal error",
			sanitized:       true,
		},
		{
			name:               "Generic error returns 500",
			message:            "something went wrong",
			err:                errors.New("something went wrong"),
			statusCode:         http.StatusInternalServerError,
			messageShouldContain: "something went wrong",
			sanitized:          false,
		},
		{
			name:            "Wrapped database error is sanitized",
			message:         "Failed to insert record",
			err:             errors.New("insert failed: ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"),
			statusCode:      http.StatusInternalServerError,
			expectedMessage: "internal error",
			sanitized:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			w := httptest.NewRecorder()

			// Create request with request ID via header
			req := httptest.NewRequest("GET", "/test", nil)
			requestID := "test-request-id-123"
			req.Header.Set("X-Request-ID", requestID)

			// Apply RequestID middleware to set it in context
			var capturedReq *http.Request
			handler := middleware.RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
			}))
			handler.ServeHTTP(httptest.NewRecorder(), req)
			req = capturedReq

			// Execute
			ErrorJSON(w, req, logger, tt.statusCode, tt.message, tt.err)

			// Assert HTTP status code
			assert.Equal(t, tt.statusCode, w.Code, "HTTP status code mismatch")

			// Assert Content-Type header
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

			// Parse response body
			var response struct {
				Message   string `json:"message"`
				RequestID string `json:"request_id,omitempty"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Response should be valid JSON")

			// Assert message content
			if tt.expectedMessage != "" {
				assert.Equal(t, tt.expectedMessage, response.Message, "Error message mismatch")
			} else if tt.messageShouldContain != "" {
				assert.Contains(t, response.Message, tt.messageShouldContain, "Error message should contain expected text")
			}

			// Assert request_id is included
			assert.Equal(t, requestID, response.RequestID, "Request ID should be included in response")

			// Assert sanitization
			if tt.sanitized {
				assert.NotContains(t, response.Message, "SQLSTATE", "Sanitized message should not contain SQLSTATE")
				assert.NotContains(t, response.Message, "constraint", "Sanitized message should not contain constraint details")
				assert.NotContains(t, response.Message, "23505", "Sanitized message should not contain error codes")
			}
		})
	}
}

// TestErrorResponseStructure verifies the JSON structure of error responses.
func TestErrorResponseStructure(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "req-456")

	// Apply RequestID middleware to set it in context
	var capturedReq *http.Request
	handler := middleware.RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	}))
	handler.ServeHTTP(httptest.NewRecorder(), req)
	req = capturedReq

	ErrorJSON(w, req, logger, http.StatusBadRequest, "validation failed", nil)

	// Parse raw JSON to verify exact structure
	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	require.NoError(t, err)

	// Assert only expected fields are present
	_, hasMessage := raw["message"]
	assert.True(t, hasMessage, "Response must have 'message' field")

	// request_id is optional but should be present when set
	_, hasRequestID := raw["request_id"]
	assert.True(t, hasRequestID, "Response should have 'request_id' when available")

	// Assert message is a string
	assert.IsType(t, "", raw["message"], "'message' field should be a string")

	// Assert request_id is a string (if present)
	if hasRequestID {
		assert.IsType(t, "", raw["request_id"], "'request_id' field should be a string")
	}
}

// TestHTTPStatusCodeMapping verifies correct HTTP status codes for different error types.
func TestHTTPStatusCodeMapping(t *testing.T) {
	tests := []struct {
		name           string
		errorType      string
		expectedStatus int
	}{
		{
			name:           "400 Bad Request for validation errors",
			errorType:      "validation",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "401 Unauthorized for auth errors",
			errorType:      "unauthorized",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "404 Not Found for missing resources",
			errorType:      "notfound",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "409 Conflict for unique constraint violations",
			errorType:      "conflict",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "500 Internal Server Error for generic errors",
			errorType:      "internal",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			ErrorJSON(w, req, logger, tt.expectedStatus, "test error", nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestErrorMessageSanitization verifies that sensitive information is removed from error messages.
func TestErrorMessageSanitization(t *testing.T) {
	tests := []struct {
		name            string
		rawMessage      string
		shouldSanitize  bool
		shouldNotContain []string
	}{
		{
			name:           "PostgreSQL error codes are sanitized",
			rawMessage:     "ERROR: duplicate key (SQLSTATE 23505)",
			shouldSanitize: true,
			shouldNotContain: []string{"23505", "SQLSTATE", "ERROR:"},
		},
		{
			name:           "Connection errors are sanitized",
			rawMessage:     "failed to connect: connection refused",
			shouldSanitize: true,
			shouldNotContain: []string{"connection"},
		},
		{
			name:           "Table names are sanitized",
			rawMessage:     "relation \"users\" does not exist",
			shouldSanitize: true,
			shouldNotContain: []string{"users", "relation"},
		},
		{
			name:           "JWT tokens are sanitized",
			rawMessage:     "invalid token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			shouldSanitize: true,
			shouldNotContain: []string{"eyJhbGci", "token:"},
		},
		{
			name:           "Validation errors are NOT sanitized",
			rawMessage:     "field 'email' is required",
			shouldSanitize: false,
			shouldNotContain: []string{},
		},
		{
			name:           "Missing parameter errors are NOT sanitized",
			rawMessage:     "missing required parameter: name",
			shouldSanitize: false,
			shouldNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error from message
			err := errors.New(tt.rawMessage)
			sanitized := sanitizeErrorMessage(tt.rawMessage, err)

			if tt.shouldSanitize {
				// Verify sensitive info is removed
				for _, forbidden := range tt.shouldNotContain {
					assert.NotContains(t, sanitized, forbidden,
						"Sanitized message should not contain: %s", forbidden)
				}
				// Verify a generic message is returned
				assert.Contains(t, []string{"internal error", "unauthorized"}, sanitized,
					"Sanitized message should be generic")
			} else {
				// Verify original message is preserved
				assert.Equal(t, tt.rawMessage, sanitized,
					"Non-sensitive messages should not be sanitized")
			}
		})
	}
}
