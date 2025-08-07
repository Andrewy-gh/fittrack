package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsUniqueConstraintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "postgres duplicate key error",
			err:      errors.New("ERROR: duplicate key value violates unique constraint \"users_pkey\" (SQLSTATE 23505)"),
			expected: true,
		},
		{
			name:     "postgres unique constraint error",
			err:      errors.New("ERROR: duplicate key violates unique constraint \"users_email_key\" (SQLSTATE 23505)"),
			expected: true,
		},
		{
			name:     "sqlite unique constraint error",
			err:      errors.New("UNIQUE constraint failed: users.email"),
			expected: true,
		},
		{
			name:     "other error",
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

func TestIsForeignKeyConstraintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "postgres foreign key error",
			err:      errors.New("ERROR: insert or update on table \"workout\" violates foreign key constraint \"workout_user_id_fkey\" (SQLSTATE 23503)"),
			expected: true,
		},
		{
			name:     "sqlite foreign key error",
			err:      errors.New("FOREIGN KEY constraint failed"),
			expected: true,
		},
		{
			name:     "other error",
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
			result := IsForeignKeyConstraintError(tt.err)
			assert.Equal(t, tt.expected, result, "IsForeignKeyConstraintError() = %v, want %v", result, tt.expected)
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
			name:     "RLS error",
			err:      ErrRowLevelSecurity,
			expected: true,
		},
		{
			name:     "wrapped RLS error",
			err:      errors.Join(errors.New("other error"), ErrRowLevelSecurity),
			expected: true,
		},
		{
			name:     "postgres permission denied error",
			err:      errors.New("ERROR: permission denied for table workout (SQLSTATE 42501)"),
			expected: true,
		},
		{
			name:     "insufficient privilege error",
			err:      errors.New("ERROR: insufficient privilege to access table"),
			expected: true,
		},
		{
			name:     "other error",
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

func TestIsRLSContextError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "RLS context error",
			err:      ErrRLSContext,
			expected: true,
		},
		{
			name:     "set_config error",
			err:      errors.New("ERROR: failed to set_config('app.current_user_id', 'user123', false)"),
			expected: true,
		},
		{
			name:     "session variable error",
			err:      errors.New("ERROR: invalid session variable app.current_user_id"),
			expected: true,
		},
		{
			name:     "other error",
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
			result := IsRLSContextError(tt.err)
			assert.Equal(t, tt.expected, result, "IsRLSContextError() = %v, want %v", result, tt.expected)
		})
	}
}
