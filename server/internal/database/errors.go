package db

import "strings"

// IsUniqueConstraintError checks if the error is a PostgreSQL or SQLite unique constraint violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key value violates unique constraint") ||
		strings.Contains(errMsg, "unique constraint") ||
		strings.Contains(errMsg, "UNIQUE constraint failed")
}
