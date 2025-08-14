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

// ListWorkouts godoc
// @Summary List workouts
// @Description Get all workouts for the authenticated user
// @Tags workouts
// @Accept json
// @Produce json
// @Security StackAuth
// @Success 200 {array} workout.WorkoutResponse
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /workouts [get]
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

// GetWorkoutWithSets godoc
// @Summary Get workout with sets
// @Description Get a specific workout with all its sets and exercises
// @Tags workouts
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Workout ID"
// @Success 200 {array} workout.WorkoutWithSetsResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /workouts/{id} [get]
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

// CreateWorkout godoc
// @Summary Create a new workout
// @Description Create a new workout with exercises and sets
// @Tags workouts
// @Accept json
// @Produce json
// @Security StackAuth
// @Param request body workout.CreateWorkoutRequest true "Workout data"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /workouts [post]
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

// UpdateWorkout godoc
// @Summary Update an existing workout
// @Description Update workout metadata (date/notes) for the authenticated user. Returns 204 No Content on success.
// @Tags workouts
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Workout ID"
// @Param request body workout.UpdateWorkoutRequest true "Updated workout data"
// @Success 204 "No Content - Workout updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request - Invalid input or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid token"
// @Failure 404 {object} response.ErrorResponse "Not Found - Workout not found or doesn't belong to user"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /workouts/{id} [put]
func (h *WorkoutHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	// Extract and validate workout ID from path
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

	// Parse and decode the request body
	var req UpdateWorkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return
	}

	// Validate the request using struct validation tags
	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "validation error occurred", err)
		return
	}

	// Delegate to service layer for business logic
	if err := h.workoutService.UpdateWorkout(r.Context(), int32(workoutIDInt), req); err != nil {
		// Handle different error types with appropriate HTTP status codes
		var errUnauthorized *ErrUnauthorized
		var errNotFound *ErrNotFound

		switch {
		case errors.As(err, &errUnauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		case errors.As(err, &errNotFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to update workout", err)
		}
		return
	}

	// Success: Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
	// No body content for 204 response
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
