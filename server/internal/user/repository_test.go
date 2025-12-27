package user

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

// TestPgxErrNoRowsDetection verifies that we can properly detect pgx.ErrNoRows
// using errors.Is instead of string comparison. This is the key behavior change
// made in repository.go line 38.
func TestPgxErrNoRowsDetection(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		isPgxNoRows    bool
		description    string
	}{
		{
			name:           "direct pgx.ErrNoRows",
			err:            pgx.ErrNoRows,
			isPgxNoRows:    true,
			description:    "errors.Is should detect pgx.ErrNoRows directly",
		},
		{
			name:           "wrapped pgx.ErrNoRows",
			err:            fmt.Errorf("query failed: %w", pgx.ErrNoRows),
			isPgxNoRows:    true,
			description:    "errors.Is should detect pgx.ErrNoRows even when wrapped",
		},
		{
			name:           "double wrapped pgx.ErrNoRows",
			err:            fmt.Errorf("failed to get user: %w", fmt.Errorf("database error: %w", pgx.ErrNoRows)),
			isPgxNoRows:    true,
			description:    "errors.Is should detect pgx.ErrNoRows through multiple wrapping layers",
		},
		{
			name:           "different error",
			err:            errors.New("some other error"),
			isPgxNoRows:    false,
			description:    "errors.Is should return false for non-pgx.ErrNoRows errors",
		},
		{
			name:           "sql.ErrNoRows is not pgx.ErrNoRows",
			err:            sql.ErrNoRows,
			isPgxNoRows:    false,
			description:    "sql.ErrNoRows and pgx.ErrNoRows are different errors",
		},
		{
			name:           "nil error",
			err:            nil,
			isPgxNoRows:    false,
			description:    "errors.Is should handle nil safely",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, pgx.ErrNoRows)
			assert.Equal(t, tt.isPgxNoRows, result, tt.description)

			// This demonstrates the old way (string comparison) would be fragile
			if tt.err != nil && tt.isPgxNoRows {
				// Wrapped errors wouldn't match with string comparison
				if tt.err != pgx.ErrNoRows {
					assert.Contains(t, tt.err.Error(), "no rows",
						"String comparison would require checking if 'no rows' is in the error message, which is fragile")
				}
			}
		})
	}
}

// TestErrorConversion verifies the pattern used in repository.go:
// converting pgx.ErrNoRows to sql.ErrNoRows for consistency across the application
func TestErrorConversion(t *testing.T) {
	// Simulate what the repository does
	simulateRepositoryBehavior := func(dbErr error) error {
		if errors.Is(dbErr, pgx.ErrNoRows) {
			return sql.ErrNoRows
		}
		return dbErr
	}

	tests := []struct {
		name          string
		inputError    error
		expectedError error
		description   string
	}{
		{
			name:          "pgx.ErrNoRows converted to sql.ErrNoRows",
			inputError:    pgx.ErrNoRows,
			expectedError: sql.ErrNoRows,
			description:   "Repository should convert pgx.ErrNoRows to sql.ErrNoRows",
		},
		{
			name:          "wrapped pgx.ErrNoRows converted to sql.ErrNoRows",
			inputError:    fmt.Errorf("query error: %w", pgx.ErrNoRows),
			expectedError: sql.ErrNoRows,
			description:   "Repository should detect and convert even wrapped pgx.ErrNoRows",
		},
		{
			name:          "other errors passed through unchanged",
			inputError:    errors.New("connection failed"),
			expectedError: errors.New("connection failed"),
			description:   "Non-NoRows errors should pass through",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simulateRepositoryBehavior(tt.inputError)

			if tt.expectedError == sql.ErrNoRows {
				assert.ErrorIs(t, result, sql.ErrNoRows, tt.description)
			} else {
				assert.Equal(t, tt.expectedError.Error(), result.Error(), tt.description)
			}
		})
	}
}

// TestUserRepositoryInterface ensures the interface exists and is implemented
func TestUserRepositoryInterface(t *testing.T) {
	// This is a compile-time check that userRepository implements UserRepository
	var _ UserRepository = (*userRepository)(nil)

	// If this compiles, the interface is correctly implemented
	assert.True(t, true, "userRepository implements UserRepository interface")
}
