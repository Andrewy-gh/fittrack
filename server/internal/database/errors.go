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
	// RLS doesn't typically return specific error codes for blocked access
	// This function is a placeholder for where you might add such checks
	// if your application implements additional RLS-related error handling
	return errors.Is(err, ErrRowLevelSecurity)
}

// ErrRowLevelSecurity is returned when a database operation is blocked by RLS
var ErrRowLevelSecurity = errors.New("access denied by row level security policy")
