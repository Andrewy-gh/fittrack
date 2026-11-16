package db

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueConstraintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "postgres unique constraint violation",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint",
			},
			expected: true,
		},
		{
			name: "wrapped postgres unique constraint violation",
			err: fmt.Errorf("failed to insert user: %w", &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key value violates unique constraint \"users_email_key\"",
			}),
			expected: true,
		},
		{
			name: "postgres foreign key violation (wrong code)",
			err: &pgconn.PgError{
				Code:    "23503",
				Message: "foreign key constraint violation",
			},
			expected: false,
		},
		{
			name:     "non-postgres error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUniqueConstraintError(tt.err)
			assert.Equal(t, tt.expected, result, "IsUniqueConstraintError() = %v, want %v", result, tt.expected)
		})
	}
}

func TestIsRowLevelSecurityError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "application-level RLS error",
			err:      ErrRowLevelSecurity,
			expected: true,
		},
		{
			name:     "wrapped application-level RLS error",
			err:      fmt.Errorf("access denied: %w", ErrRowLevelSecurity),
			expected: true,
		},
		{
			name: "postgres insufficient privilege error",
			err: &pgconn.PgError{
				Code:    "42501",
				Message: "permission denied for table workout",
			},
			expected: true,
		},
		{
			name: "wrapped postgres insufficient privilege error",
			err: fmt.Errorf("query failed: %w", &pgconn.PgError{
				Code:    "42501",
				Message: "insufficient privilege",
			}),
			expected: true,
		},
		{
			name: "postgres other error code",
			err: &pgconn.PgError{
				Code:    "23505",
				Message: "duplicate key",
			},
			expected: false,
		},
		{
			name:     "non-postgres error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRowLevelSecurityError(tt.err)
			assert.Equal(t, tt.expected, result, "IsRowLevelSecurityError() = %v, want %v", result, tt.expected)
		})
	}
}
