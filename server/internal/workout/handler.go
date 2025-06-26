package workout

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/go-playground/validator/v10"
)

type WorkoutHandler struct {
	logger         *slog.Logger
	workoutService *WorkoutService
}

func NewHandler(logger *slog.Logger, workoutService *WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		logger:         logger,
		workoutService: workoutService,
	}
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListWorkouts(r.Context())
	if err != nil {
		h.logger.Error("failed to list workouts", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, workouts); err != nil {
		h.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *WorkoutHandler) GetWorkoutWithSets(w http.ResponseWriter, r *http.Request) {
	workoutID := r.PathValue("id")
	if workoutID == "" {
		h.logger.Error("missing workout ID", "error", "Missing workout ID")
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutIDInt, err := strconv.ParseInt(workoutID, 10, 32)
	if err != nil {
		h.logger.Error("invalid workout ID", "error", err)
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	workoutWithSets, err := h.workoutService.GetWorkoutWithSets(r.Context(), int32(workoutIDInt))
	if err != nil {
		h.logger.Error("failed to get workout with sets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, workoutWithSets); err != nil {
		h.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var request CreateWorkoutRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		h.logger.Error("validation error occurred", "error", err)
		http.Error(w, FormatValidationErrors(err), http.StatusBadRequest)
		return
	}

	// Call service with validated struct
	if err := h.workoutService.CreateWorkout(r.Context(), request); err != nil {
		h.logger.Error("failed to create workout", "error", err)
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]bool{"success": true}); err != nil {
		h.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
