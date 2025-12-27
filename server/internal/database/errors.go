package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueConstraintError checks if an error is a unique constraint violation.
// PostgreSQL error code 23505: unique_violation
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL unique constraint violation (23505)
		return pgErr.Code == "23505"
	}

	return false
}

// IsForeignKeyConstraintError checks if an error is a foreign key constraint violation.
// PostgreSQL error code 23503: foreign_key_violation
func IsForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL foreign key constraint violation (23503)
		return pgErr.Code == "23503"
	}

	return false
}

// IsRowLevelSecurityError checks if an error is related to row level security.
// When RLS blocks access, PostgreSQL typically returns empty result sets
// rather than explicit permission errors, so this function may need to be
// used in conjunction with application-level checks.
// PostgreSQL error code 42501: insufficient_privilege
func IsRowLevelSecurityError(err error) bool {
	if err == nil {
		return false
	}

	// Check for application-level RLS errors first
	if errors.Is(err, ErrRowLevelSecurity) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL insufficient privilege error (42501)
		// This can indicate RLS policy violations
		return pgErr.Code == "42501"
	}

	return false
}

// IsRLSContextError checks if an error is related to missing or invalid RLS context.
// This primarily relies on the application-level ErrRLSContext sentinel error.
func IsRLSContextError(err error) bool {
	if err == nil {
		return false
	}

	// Check for application-level RLS context errors
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
