package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
)

const (
	errorCategoryNone       = "none"
	errorCategoryDatabase   = "database"
	errorCategoryJWT        = "jwt"
	errorCategoryValidation = "validation"
	errorCategoryInternal   = "internal"
)

type Error struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON writes a JSON response to the HTTP response writer.
//
// Note on error handling: JSON encoding errors are extremely rare in practice and
// typically indicate serious programming bugs (attempting to marshal channels, functions,
// or cyclic data structures). If encoding fails after headers are written, the response
// will be partial. Callers should log encoding errors for debugging. Adding metrics for
// these errors would add complexity without significant value unless errors are observed
// in production logs.
func JSON(w http.ResponseWriter, status int, data any) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

// sanitizeErrorMessage removes sensitive database and JWT error details from public messages
func sanitizeErrorMessage(message string, err error) string {
	if err == nil {
		return message
	}

	errMsg := err.Error()

	// Allow validation errors to pass through - they are safe for users
	if isValidationError(message, errMsg) {
		return message
	}

	// Check for database error codes that should be hidden
	if containsDatabaseError(errMsg) {
		// For database errors, return generic messages based on context
		if strings.Contains(strings.ToLower(message), "unauthorized") ||
			strings.Contains(strings.ToLower(message), "auth") {
			return "unauthorized"
		}
		return "internal error"
	}

	// Check for JWT errors that should be hidden
	if containsJWTError(errMsg) {
		return "unauthorized"
	}

	// Return the original message if no sensitive content detected
	return message
}

// containsDatabaseError checks if an error message contains sensitive database information
func containsDatabaseError(errMsg string) bool {
	lowerMsg := strings.ToLower(errMsg)

	// Check for go-playground/validator patterns first (these are safe)
	if strings.Contains(errMsg, "Key: '") && strings.Contains(errMsg, "Error:Field validation") {
		return false
	}

	sensitivePatterns := []string{
		"pq:", "pgx", "SQLSTATE", "ERROR:",
		"23505", "23503", "42501", // PostgreSQL error codes
		"duplicate key", "foreign key", "constraint",
		"relation", "column", "table",
		"database", "connection", "pool",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerMsg, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// containsJWTError checks if an error message contains sensitive JWT information
func containsJWTError(errMsg string) bool {
	jwtPatterns := []string{
		"jwt", "token", "jwks", "claim", "signature",
		"failed to parse", "failed to validate",
		"key set", "algorithm",
	}

	lowerMsg := strings.ToLower(errMsg)
	for _, pattern := range jwtPatterns {
		if strings.Contains(lowerMsg, pattern) {
			return true
		}
	}
	return false
}

// isValidationError checks if this is a user input validation error that is safe to show
func isValidationError(message, errMsg string) bool {
	lowerMsg := strings.ToLower(message)
	lowerErr := strings.ToLower(errMsg)

	// Common validation message patterns that are safe to show users
	validationPatterns := []string{
		"validation", "required", "invalid format", "missing", "empty",
		"too short", "too long", "out of range", "invalid date",
		"field", "parameter", "input", "decode",
	}

	// Check if the message indicates a validation error
	for _, pattern := range validationPatterns {
		if strings.Contains(lowerMsg, pattern) {
			// Make sure it doesn't also contain database error patterns
			if !containsDatabaseError(errMsg) && !containsJWTError(errMsg) {
				return true
			}
		}
	}

	// Check for go-playground/validator specific errors
	if strings.Contains(lowerErr, "validation failed") ||
		strings.Contains(lowerErr, "field validation") ||
		strings.Contains(lowerErr, "key:") || strings.Contains(lowerErr, "tag:") {
		return !containsDatabaseError(errMsg) && !containsJWTError(errMsg)
	}

	return false
}

func safeErrorCategory(message string, err error) string {
	if err == nil {
		return errorCategoryNone
	}

	errMsg := err.Error()
	switch {
	case isValidationError(message, errMsg):
		return errorCategoryValidation
	case containsDatabaseError(errMsg):
		return errorCategoryDatabase
	case containsJWTError(errMsg):
		return errorCategoryJWT
	default:
		return errorCategoryInternal
	}
}

func ErrorJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message string, err error) {
	requestID := request.GetRequestID(r.Context())
	safeMessage := sanitizeErrorMessage(message, err)
	logAttrs := []any{
		"path", r.URL.Path,
		"method", r.Method,
		"status", status,
		"request_id", requestID,
		"response_message", safeMessage,
		"error_category", safeErrorCategory(message, err),
		"error_present", err != nil,
	}
	if err != nil {
		logAttrs = append(logAttrs, "error_type", fmt.Sprintf("%T", err))
	}

	logger.Error("error response", logAttrs...)

	// Create a sanitized error response - never include raw error details in HTTP responses
	resp := Error{
		Message:   safeMessage,
		RequestID: requestID,
	}

	if jsonErr := JSON(w, status, resp); jsonErr != nil {
		logger.Error("failed to write error response",
			"path", r.URL.Path,
			"method", r.Method,
			"status", status,
			"request_id", requestID,
			"error_type", fmt.Sprintf("%T", jsonErr),
		)
	}
}
