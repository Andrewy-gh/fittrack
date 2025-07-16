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
