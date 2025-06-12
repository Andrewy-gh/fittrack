package validation

import (
	"fmt"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/models"
	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateWorkoutRequest validates a workout request
func ValidateWorkoutRequest(req models.WorkoutRequest) error {
	return validate.Struct(req)
}

func ValidateWorkoutID(idParam string) (int32, error) {
	if idParam == "" {
		return 0, fmt.Errorf("workout ID is required")
	}

	id, err := strconv.ParseInt(idParam, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid workout ID format")
	}
	if id <= 0 {
		return 0, fmt.Errorf("workout ID must be a positive number")
	}
	return int32(id), nil
}

// FormatValidationErrors formats validation errors in a user-friendly way
func FormatValidationErrors(err error) string {
	if validationErrors, ok := err.(*validator.ValidationErrors); ok {
		var messages []string
		for _, fieldError := range *validationErrors {
			switch fieldError.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("%s is required", fieldError.Field()))
			case "min":
				messages = append(messages, fmt.Sprintf("%s must be at least %s characters", fieldError.Field(), fieldError.Param()))
			case "max":
				messages = append(messages, fmt.Sprintf("%s must be at most %s characters", fieldError.Field(), fieldError.Param()))
			case "gte":
				messages = append(messages, fmt.Sprintf("%s must be greater than or equal to %s", fieldError.Field(), fieldError.Param()))
			case "datetime":
				messages = append(messages, fmt.Sprintf("%s must be a valid datetime in RFC3339 format", fieldError.Field()))
			default:
				messages = append(messages, fmt.Sprintf("%s failed validation (%s)", fieldError.Field(), fieldError.Tag()))
			}
		}
		return fmt.Sprintf("Validation errors: %v", messages)
	}
	return err.Error()
}
