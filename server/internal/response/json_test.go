package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		err         error
		expectedMsg string
		description string
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

func TestErrorJSON_WritesSanitizedHTTPResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req = req.WithContext(request.WithRequestID(req.Context(), "req-http-response"))
	w := httptest.NewRecorder()

	ErrorJSON(
		w,
		req,
		logger,
		http.StatusInternalServerError,
		"failed to create user",
		errors.New("pq: duplicate key value violates unique constraint \"users_pkey\" SQLSTATE 23505"),
	)

	if got, want := w.Code, http.StatusInternalServerError; got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
	if got, want := w.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := w.Body.String(), `{"message":"internal error","request_id":"req-http-response"}`; got != want {
		t.Errorf("serialized response = %s, want %s", got, want)
	}
}

func TestErrorJSON_DoesNotLeakSensitiveData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tests := []struct {
		name             string
		message          string
		err              error
		expectedResponse string
		forbidden        []string
	}{
		{
			name:    "PostgreSQL constraint details",
			message: "failed to create user",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"users_pkey\" DETAIL: Key (user_id)=(test) already exists",
			},
			expectedResponse: "internal error",
			forbidden:        []string{"duplicate key", "users_pkey", "SQLSTATE", "23505"},
		},
		{
			name:             "pgx connection details",
			message:          "database unavailable",
			err:              errors.New("pgx: failed to connect to database pool"),
			expectedResponse: "internal error",
			forbidden:        []string{"pgx:", "connect", "database pool"},
		},
		{
			name:             "relation and table details",
			message:          "operation failed",
			err:              errors.New("pq: relation \"users\" does not exist"),
			expectedResponse: "internal error",
			forbidden:        []string{"pq:", "relation", "users"},
		},
		{
			name:             "JWT token details",
			message:          "authentication failed",
			err:              errors.New("failed to parse token: invalid signature algorithm RS256 token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"),
			expectedResponse: "unauthorized",
			forbidden:        []string{"token:", "signature", "RS256", "eyJhbGci"},
		},
	}

	for _, code := range []string{
		"23503", "42501", "42P01", "42703", "08006", "57P01",
	} {
		tests = append(tests, struct {
			name             string
			message          string
			err              error
			expectedResponse string
			forbidden        []string
		}{
			name:    "PostgreSQL status code " + code,
			message: "operation failed",
			err: &pgconn.PgError{
				Code:    code,
				Message: "query failed",
			},
			expectedResponse: "internal error",
			forbidden:        []string{code, "SQLSTATE"},
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/test", nil)

			ErrorJSON(
				w,
				req,
				logger,
				http.StatusInternalServerError,
				tt.message,
				tt.err,
			)

			if got, want := w.Code, http.StatusInternalServerError; got != want {
				t.Errorf("status = %d, want %d", got, want)
			}

			var response Error
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if got := response.Message; got != tt.expectedResponse {
				t.Errorf("response message = %q, want %q", got, tt.expectedResponse)
			}

			for _, forbidden := range tt.forbidden {
				if strings.Contains(w.Body.String(), forbidden) {
					t.Errorf("response leaked sensitive detail %q: %s", forbidden, w.Body.String())
				}
			}
		})
	}
}

func TestErrorJSON_LogsSafeErrorSummary(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
	req = req.WithContext(request.WithRequestID(req.Context(), "req-safe-logs"))
	w := httptest.NewRecorder()
	sensitiveErr := errors.New("pq: duplicate key value violates unique constraint \"users_email_key\" DETAIL: Key (email)=(andy@example.com) already exists. SQLSTATE 23505")

	ErrorJSON(w, req, logger, http.StatusInternalServerError, "failed to create user", sensitiveErr)

	logOutput := output.String()
	for _, forbidden := range []string{
		"pq:",
		"SQLSTATE",
		"23505",
		"users_email_key",
		"andy@example.com",
		"duplicate key",
	} {
		if strings.Contains(logOutput, forbidden) {
			t.Fatalf("error log contains sensitive raw error detail %q: %s", forbidden, logOutput)
		}
	}

	var logEntry map[string]any
	if err := json.Unmarshal(output.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to decode error log output: %v", err)
	}

	if got := logEntry["msg"]; got != "error response" {
		t.Fatalf("expected safe log message, got %v", got)
	}
	if got := logEntry["status"]; got != float64(http.StatusInternalServerError) {
		t.Fatalf("expected status field, got %v", got)
	}
	if got := logEntry["request_id"]; got != "req-safe-logs" {
		t.Fatalf("expected request_id field, got %v", got)
	}
	if got := logEntry["response_message"]; got != "internal error" {
		t.Fatalf("expected sanitized response_message field, got %v", got)
	}
	if got := logEntry["error_category"]; got != errorCategoryDatabase {
		t.Fatalf("expected database error_category field, got %v", got)
	}
	if got := logEntry["error_type"]; got != "*errors.errorString" {
		t.Fatalf("expected stable error_type field, got %v", got)
	}
}
