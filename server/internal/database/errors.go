package db

import (
	"errors"
	"regexp"
	"strings"
)

// IsUniqueConstraintError checks if an error is a unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	// PostgreSQL unique constraint violation
	if strings.Contains(msg, "SQLSTATE 23505") {
		return true
	}

	// SQLite unique constraint violation
	if strings.Contains(msg, "UNIQUE constraint failed") {
		return true
	}

	// Generic check for duplicate key violations
	if strings.Contains(msg, "duplicate key") && strings.Contains(msg, "violates") {
		return true
	}

	return false
}

// IsForeignKeyConstraintError checks if an error is a foreign key constraint violation
func IsForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	// PostgreSQL foreign key constraint violation
	if strings.Contains(msg, "SQLSTATE 23503") {
		return true
	}

	// SQLite foreign key constraint violation
	sqliteFKRegex := regexp.MustCompile(`(?i)foreign key constraint failed`)
	return sqliteFKRegex.MatchString(msg)
}

// IsRowLevelSecurityError checks if an error is related to row level security
// When RLS blocks access, PostgreSQL typically returns empty result sets
// rather than explicit permission errors, so this function may need to be
// used in conjunction with application-level checks
func IsRowLevelSecurityError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	// Check for explicit RLS permission errors (rare but possible)
	if strings.Contains(msg, "SQLSTATE 42501") {
		return true
	}

	// Check for insufficient privilege errors
	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "insufficient privilege") {
		return true
	}

	// Check for application-level RLS errors
	return errors.Is(err, ErrRowLevelSecurity)
}

// IsRLSContextError checks if an error is related to missing or invalid RLS context
func IsRLSContextError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	// Check for session variable related errors
	if strings.Contains(msg, "app.current_user_id") {
		return true
	}

	// Check for set_config related errors
	if strings.Contains(msg, "set_config") {
		return true
	}

	return errors.Is(err, ErrRLSContext)
}

// RLSLogData contains structured data for RLS-related logging
type RLSLogData struct {
	Operation   string `json:"operation"`
	UserID      string `json:"user_id"`
	TableName   string `json:"table_name,omitempty"`
	RecordID    any    `json:"record_id,omitempty"`
	ResultCount int    `json:"result_count,omitempty"`
	ContextSet  bool   `json:"rls_context_set"`
}

// ErrRowLevelSecurity is returned when a database operation is blocked by RLS
var ErrRowLevelSecurity = errors.New("access denied by row level security policy")

// ErrRLSContext is returned when there's an issue with RLS context setup
var ErrRLSContext = errors.New("failed to set RLS context")
