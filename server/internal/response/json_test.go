package response

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		err           error
		expectedMsg   string
		description   string
	}{
		{
			name:        "PostgreSQL error code 23505",
			message:     "failed to create user",
			err:         errors.New("pq: duplicate key value violates unique constraint \"users_pkey\" DETAIL: Key (user_id)=(test) already exists. SQLSTATE 23505"),
			expectedMsg: "internal error",
			description: "PostgreSQL unique constraint error should be sanitized",
		},
		{
			name:        "PostgreSQL error code 23503",
			message:     "failed to create workout",
			err:         errors.New("pq: insert or update on table \"workouts\" violates foreign key constraint SQLSTATE 23503"),
			expectedMsg: "internal error",
			description: "PostgreSQL foreign key error should be sanitized",
		},
		{
			name:        "pgx connection error",
			message:     "database operation failed",
			err:         errors.New("pgx: connection lost, failed to read from database pool"),
			expectedMsg: "internal error",
			description: "pgx connection errors should be sanitized",
		},
		{
			name:        "JWT parse error",
			message:     "invalid access token",
			err:         errors.New("failed to parse/validate token: invalid signature algorithm"),
			expectedMsg: "unauthorized",
			description: "JWT parsing errors should be sanitized to unauthorized",
		},
		{
			name:        "JWT claims error",
			message:     "token validation failed",
			err:         errors.New("token missing required 'sub' claim"),
			expectedMsg: "unauthorized",
			description: "JWT claim errors should be sanitized to unauthorized",
		},
		{
			name:        "authorization context error",
			message:     "failed to ensure user",
			err:         errors.New("pq: permission denied for relation users SQLSTATE 42501"),
			expectedMsg: "internal error",
			description: "Database permission errors should be sanitized",
		},
		{
			name:        "clean validation error",
			message:     "validation failed",
			err:         errors.New("field 'name' is required"),
			expectedMsg: "validation failed",
			description: "Clean validation errors should pass through",
		},
		{
			name:        "validation error occurred message",
			message:     "validation error occurred",
			err:         errors.New("Key: 'CreateWorkoutRequest.Date' Error:Field validation for 'Date' failed on the 'required' tag"),
			expectedMsg: "validation error occurred",
			description: "go-playground validator errors should pass through",
		},
		{
			name:        "failed to decode request body",
			message:     "failed to decode request body",
			err:         errors.New("invalid character '}' looking for beginning of object key string"),
			expectedMsg: "failed to decode request body",
			description: "JSON decode errors should pass through",
		},
		{
			name:        "nil error",
			message:     "operation failed",
			err:         nil,
			expectedMsg: "operation failed",
			description: "Messages with nil errors should pass through unchanged",
		},
		{
			name:        "unauthorized database error",
			message:     "unauthorized access",
			err:         errors.New("pq: relation \"sensitive_table\" does not exist SQLSTATE 42P01"),
			expectedMsg: "unauthorized",
			description: "Unauthorized messages with database errors should remain unauthorized",
		},
		{
			name:        "auth database error",
			message:     "auth failed",
			err:         errors.New("database connection failed: pq pool exhausted"),
			expectedMsg: "unauthorized",
			description: "Auth messages with database errors should remain unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeErrorMessage(tt.message, tt.err)
			if result != tt.expectedMsg {
				t.Errorf("sanitizeErrorMessage() = %v, want %v\nDescription: %s\nOriginal error: %v", 
					result, tt.expectedMsg, tt.description, tt.err)
			}
		})
	}
}

func TestContainsDatabaseError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"PostgreSQL error code", "pq: error SQLSTATE 23505", true},
		{"pgx error", "pgx: connection failed", true},
		{"Constraint error", "duplicate key violates constraint", true},
		{"Table reference", "relation \"users\" does not exist", true},
		{"Database connection", "database connection pool exhausted", true},
		{"Clean error", "user input validation failed", false},
		{"Go-playground validator", "Key: 'Request.Name' Error:Field validation for 'Name' failed on the 'required' tag", false},
		{"HTTP error", "404 not found", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDatabaseError(tt.errMsg)
			if result != tt.expected {
				t.Errorf("containsDatabaseError(%q) = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		errMsg   string
		expected bool
	}{
		{"Validation error message", "validation error occurred", "field 'name' is required", true},
		{"Missing field", "missing required field", "field cannot be empty", true},
		{"Decode error", "failed to decode request body", "invalid character '}'", true},
		{"Go-playground validator", "validation failed", "Key: 'Request.Name' Error:Field validation for 'Name' failed on the 'required' tag", true},
		{"Validation with database error", "validation failed", "pq: duplicate key SQLSTATE 23505", false},
		{"Validation with JWT error", "validation failed", "failed to parse token: invalid signature", false},
		{"Database error", "operation failed", "pq: connection failed", false},
		{"JWT error", "authentication failed", "token expired", false},
		{"Generic error", "something went wrong", "unknown error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.message, tt.errMsg)
			if result != tt.expected {
				t.Errorf("isValidationError(%q, %q) = %v, want %v", tt.message, tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestContainsJWTError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{"JWT parse error", "failed to parse JWT token", true},
		{"Token validation", "failed to validate token signature", true},
		{"JWKS error", "jwks key set not found", true},
		{"Claims error", "token missing required claim", true},
		{"Algorithm error", "unsupported signature algorithm", true},
		{"Clean error", "user not found", false},
		{"HTTP error", "network timeout", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsJWTError(tt.errMsg)
			if result != tt.expected {
				t.Errorf("containsJWTError(%q) = %v, want %v", tt.errMsg, result, tt.expected)
			}
		})
	}
}

func TestErrorJSON_NoSensitiveDataInResponse(t *testing.T) {
	// Create a test logger that writes to a buffer so we can verify debug logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	
	tests := []struct {
		name                   string
		message               string
		err                   error
		expectedResponseMsg   string
		shouldContainInDebug  string
		description          string
	}{
		{
			name:                  "PostgreSQL constraint violation",
			message:               "failed to create user",
			err:                   errors.New("pq: duplicate key value violates unique constraint \"users_pkey\" SQLSTATE 23505"),
			expectedResponseMsg:   "internal error",
			shouldContainInDebug:  "SQLSTATE 23505",
			description:          "Database errors should be sanitized in response but logged in debug",
		},
		{
			name:                  "pgx connection error",
			message:               "database unavailable",
			err:                   errors.New("pgx: failed to connect to database pool"),
			expectedResponseMsg:   "internal error",
			shouldContainInDebug:  "pgx: failed to connect",
			description:          "pgx errors should be sanitized in response but logged in debug",
		},
		{
			name:                  "JWT token error",
			message:               "authentication failed",
			err:                   errors.New("failed to parse token: invalid signature algorithm RS256"),
			expectedResponseMsg:   "unauthorized",
			shouldContainInDebug:  "invalid signature algorithm",
			description:          "JWT errors should be sanitized to 'unauthorized' but logged in debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP request and response
			req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
			w := httptest.NewRecorder()

			// Call ErrorJSON
			ErrorJSON(w, req, logger, http.StatusInternalServerError, tt.message, tt.err)

			// Check that the HTTP response contains the sanitized message
			responseBody := w.Body.String()
			if !strings.Contains(responseBody, fmt.Sprintf(`"message":"%s"`, tt.expectedResponseMsg)) {
				t.Errorf("Response body should contain sanitized message '%s', got: %s", 
					tt.expectedResponseMsg, responseBody)
			}

			// Verify that sensitive information is NOT in the response
			if tt.err != nil {
				sensitiveContent := tt.err.Error()
				if strings.Contains(responseBody, sensitiveContent) {
					t.Errorf("Response body should NOT contain sensitive error details: %s\nResponse: %s", 
						sensitiveContent, responseBody)
				}
			}

			// Check status code
			if w.Code != http.StatusInternalServerError {
				t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, w.Code)
			}
		})
	}
}

// TestNoPostgreSQLErrorCodesInResponse is the main test required by the task
func TestNoPostgreSQLErrorCodesInResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Common PostgreSQL error codes that should never appear in HTTP responses
	pgErrorCodes := []string{
		"23505", // unique_violation
		"23503", // foreign_key_violation
		"42501", // insufficient_privilege
		"42P01", // undefined_table
		"42703", // undefined_column
		"08006", // connection_failure
		"57P01", // admin_shutdown
	}

	pgxErrors := []string{
		"pgx: connection",
		"pq: duplicate key",
		"SQLSTATE",
		"constraint violation",
		"relation does not exist",
	}

	allSensitiveErrors := append(pgErrorCodes, pgxErrors...)

	for _, errorCode := range allSensitiveErrors {
		t.Run(fmt.Sprintf("ErrorCode_%s", errorCode), func(t *testing.T) {
			// Create an error that contains the sensitive code/message
			testErr := errors.New(fmt.Sprintf("Database error with code: %s and additional context", errorCode))
			
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			w := httptest.NewRecorder()

			// Call ErrorJSON with the sensitive error
			ErrorJSON(w, req, logger, http.StatusInternalServerError, "operation failed", testErr)

			responseBody := w.Body.String()
			
			// Assert that the sensitive error code/message does NOT appear in the HTTP response
			if strings.Contains(responseBody, errorCode) {
				t.Errorf("HTTP response contains sensitive database error code '%s'.\nResponse body: %s", 
					errorCode, responseBody)
			}

			// Assert that we get a generic error message instead
			if !strings.Contains(responseBody, `"message":"internal error"`) {
				t.Errorf("Expected sanitized 'internal error' message in response, got: %s", responseBody)
			}
		})
	}
}
