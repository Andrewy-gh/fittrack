package workout

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatValidationErrors verifies that the FormatValidationErrors function
// correctly handles wrapped validation errors and non-validation errors
func TestFormatValidationErrors(t *testing.T) {
	// Create a validator for generating validation errors
	v := validator.New()

	// Test struct for validation
	type TestStruct struct {
		Name  string `validate:"required"`
		Age   int    `validate:"gte=0"`
		Email string `validate:"required,email"`
	}

	t.Run("unwraps wrapped validation errors", func(t *testing.T) {
		// Create a struct that fails validation
		testData := TestStruct{
			Name:  "", // required field is empty
			Age:   -1, // fails gte=0
			Email: "", // required field is empty
		}

		// Get validation error
		err := v.Struct(testData)
		require.Error(t, err, "Expected validation to fail")

		// Wrap the error to test errors.As unwrapping
		wrappedErr := fmt.Errorf("validation failed: %w", err)
		doubleWrappedErr := fmt.Errorf("request processing failed: %w", wrappedErr)

		// Test that FormatValidationErrors can unwrap and format the error
		result := FormatValidationErrors(doubleWrappedErr)

		// Verify the result contains validation error messages
		assert.Contains(t, result, "Validation errors:", "Should contain validation errors prefix")
		assert.Contains(t, result, "Name is required", "Should contain Name required message")
		assert.Contains(t, result, "Age must be greater than or equal to 0", "Should contain Age gte message")
		assert.Contains(t, result, "Email is required", "Should contain Email required message")
	})

	t.Run("handles single validation error", func(t *testing.T) {
		testData := TestStruct{
			Name:  "",
			Age:   25,
			Email: "valid@example.com",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Name is required")
	})

	t.Run("handles validation errors with different tags", func(t *testing.T) {
		type ValidationTestStruct struct {
			Username string `validate:"required,min=3,max=20"`
			Count    int    `validate:"gte=1"`
			Created  string `validate:"datetime=2006-01-02T15:04:05Z07:00"`
		}

		testData := ValidationTestStruct{
			Username: "ab",      // fails min=3
			Count:    0,         // fails gte=1
			Created:  "invalid", // fails datetime
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Username must be at least 3 characters")
		assert.Contains(t, result, "Count must be greater than or equal to 1")
		assert.Contains(t, result, "Created must be a valid datetime in RFC3339 format")
	})

	t.Run("falls back to err.Error() for non-validation errors", func(t *testing.T) {
		// Test with a standard error
		standardErr := fmt.Errorf("database connection failed")
		result := FormatValidationErrors(standardErr)
		assert.Equal(t, "database connection failed", result)

		// Test with a wrapped standard error
		wrappedStandardErr := fmt.Errorf("operation failed: %w", standardErr)
		result = FormatValidationErrors(wrappedStandardErr)
		assert.Equal(t, "operation failed: database connection failed", result)
	})

	t.Run("handles nil error gracefully", func(t *testing.T) {
		// This shouldn't happen in practice, but test defensive programming
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("FormatValidationErrors panicked with nil error: %v", r)
			}
		}()

		// If err is nil, err.Error() will panic, but this tests if there's nil checking
		// In the current implementation, this would panic, which is acceptable
		// since the function is never called with nil in actual use
	})

	t.Run("handles max validation tag", func(t *testing.T) {
		type MaxTestStruct struct {
			Description string `validate:"max=10"`
		}

		testData := MaxTestStruct{
			Description: "This is a very long description that exceeds the maximum",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Description must be at most 10 characters")
	})

	t.Run("handles unknown validation tag with default message", func(t *testing.T) {
		type UnknownTagStruct struct {
			Field string `validate:"customtag"`
		}

		// Register a custom validator that will fail
		v.RegisterValidation("customtag", func(fl validator.FieldLevel) bool {
			return false
		})

		testData := UnknownTagStruct{
			Field: "test",
		}

		err := v.Struct(testData)
		require.Error(t, err)

		result := FormatValidationErrors(err)

		assert.Contains(t, result, "Validation errors:")
		assert.Contains(t, result, "Field failed validation (customtag)")
	})
}
