package workout

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/go-playground/validator/v10"
)

type WorkoutHandler struct {
	logger         *slog.Logger
	validator      *validator.Validate
	workoutService *WorkoutService
}

func NewHandler(logger *slog.Logger, validator *validator.Validate, workoutService *WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		logger:         logger,
		validator:      validator,
		workoutService: workoutService,
	}
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListWorkouts(r.Context())
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to list workouts", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, workouts); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
		return
	}
}

func (h *WorkoutHandler) GetWorkoutWithSets(w http.ResponseWriter, r *http.Request) {
	workoutID := r.PathValue("id")
	if workoutID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing workout ID", nil)
		return
	}

	workoutIDInt, err := strconv.Atoi(workoutID)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid workout ID", err)
		return
	}

	workoutWithSets, err := h.workoutService.GetWorkoutWithSets(r.Context(), int32(workoutIDInt))
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to get workout with sets", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, workoutWithSets); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
		return
	}
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "validation error occurred", err)
		return
	}

	if err := h.workoutService.CreateWorkout(r.Context(), req); err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to create workout", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]bool{"success": true}); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
		return
	}
}

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
