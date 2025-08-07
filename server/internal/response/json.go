package response

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type Error struct {
	Message string `json:"message"`
}

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

func ErrorJSON(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message string, err error) {
	// Log the full error details at Error level for operational monitoring
	logger.Error(message, "error", err, "path", r.URL.Path, "method", r.Method, "status", status)
	
	// If there's an underlying error, log the raw details at Debug level for troubleshooting
	if err != nil {
		logger.Debug("raw error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "path", r.URL.Path)
	}

	// Create a sanitized error response - never include raw error details in HTTP responses
	resp := Error{
		Message: sanitizeErrorMessage(message, err),
	}

	if jsonErr := JSON(w, status, resp); jsonErr != nil {
		logger.Error("failed to write error response", "error", jsonErr, "path", r.URL.Path)
	}
}
