package exercise

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/go-playground/validator/v10"
)

// ExerciseHandler handles exercise HTTP requests
type ExerciseHandler struct {
	logger          *slog.Logger
	validator       *validator.Validate
	exerciseService *ExerciseService
}

func NewHandler(logger *slog.Logger, validator *validator.Validate, exerciseService *ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		logger:          logger,
		validator:       validator,
		exerciseService: exerciseService,
	}
}

// MARK: ListExercises
// ListExercises godoc
// @Summary List exercises
// @Description Get all exercises for the authenticated user
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Success 200 {array} exercise.ExerciseResponse
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises [get]
func (h *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.exerciseService.ListExercises(r.Context())
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to list exercises", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, exercises); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

// MARK: GetExerciseWithSets
// GetExerciseWithSets godoc
// @Summary Get exercise with sets
// @Description Get a specific exercise with all its sets from workouts. Returns empty array when exercise exists but has no sets.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Success 200 {array} exercise.ExerciseWithSetsResponse "Success (may be empty array if exercise has no sets)"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Not Found - Exercise not found or doesn't belong to user"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id} [get]
func (h *ExerciseHandler) GetExerciseWithSets(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return
	}

	exerciseIDInt, err := strconv.Atoi(exerciseID)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return
	}

	req := GetExerciseWithSetsRequest{
		ExerciseID: int32(exerciseIDInt),
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID: must be positive", err)
		return
	}

	exerciseWithSets, err := h.exerciseService.GetExerciseWithSets(r.Context(), req.ExerciseID)
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		var errNotFound *ErrNotFound

		switch {
		case errors.As(err, &errUnauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		case errors.As(err, &errNotFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to get exercise with sets", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, exerciseWithSets); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

// MARK: GetOrCreateExercise
// GetOrCreateExercise godoc
// @Summary Get or create exercise
// @Description Get an existing exercise by name or create it if it doesn't exist
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param request body exercise.CreateExerciseRequest true "Exercise data"
// @Success 200 {object} exercise.CreateExerciseResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises [post]
func (h *ExerciseHandler) GetOrCreateExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Failed to decode request body", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Validation failed", err)
		return
	}

	exercise, err := h.exerciseService.GetOrCreateExercise(r.Context(), req.Name)
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to get or create exercise", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, exercise); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

// MARK: GetRecentSetsForExercise
// GetRecentSetsForExercise godoc
// @Summary Get recent sets for exercise
// @Description Get the 3 most recent sets for a specific exercise. Returns empty array when exercise has no sets.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Success 200 {array} exercise.RecentSetsResponse "Success (may be empty array)"
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id}/recent-sets [get]
func (h *ExerciseHandler) GetRecentSetsForExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return
	}

	exerciseIDInt, err := strconv.Atoi(exerciseID)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return
	}

	req := GetRecentSetsRequest{
		ExerciseID: int32(exerciseIDInt),
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID: must be positive", err)
		return
	}

	sets, err := h.exerciseService.GetRecentSetsForExercise(r.Context(), req.ExerciseID)
	if err != nil {
		var errUnauthorized *ErrUnauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to get recent sets for exercise", err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, sets); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

// MARK: UpdateExerciseName
// UpdateExerciseName godoc
// @Summary Update an exercise name
// @Description Update the name of an exercise by ID (must belong to authenticated user)
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param body body UpdateExerciseNameRequest true "Exercise name update request"
// @Success 204 "No Content - Exercise name updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request - Invalid exercise ID or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid token"
// @Failure 404 {object} response.ErrorResponse "Not Found - Exercise not found or doesn't belong to user"
// @Failure 409 {object} response.ErrorResponse "Conflict - Exercise name already exists"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id} [patch]
func (h *ExerciseHandler) UpdateExerciseName(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return
	}

	exerciseIDInt, err := strconv.ParseInt(exerciseID, 10, 32)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return
	}

	var req UpdateExerciseNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Failed to decode request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := h.exerciseService.UpdateExerciseName(r.Context(), int32(exerciseIDInt), req.Name); err != nil {
		var errUnauthorized *ErrUnauthorized
		var errNotFound *ErrNotFound

		switch {
		case errors.As(err, &errUnauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		case errors.As(err, &errNotFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
		case db.IsUniqueConstraintError(err):
			response.ErrorJSON(w, r, h.logger, http.StatusConflict, "Exercise name already exists", nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to update exercise name", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MARK: DeleteExercise
// DeleteExercise godoc
// @Summary Delete an exercise
// @Description Delete a specific exercise and all its associated sets. Only the owner of the exercise can delete it.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Success 204 "No Content - Exercise deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request - Invalid exercise ID"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid token"
// @Failure 404 {object} response.ErrorResponse "Not Found - Exercise not found or doesn't belong to user"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id} [delete]
func (h *ExerciseHandler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return
	}

	exerciseIDInt, err := strconv.ParseInt(exerciseID, 10, 32)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return
	}

	if err := h.exerciseService.DeleteExercise(r.Context(), int32(exerciseIDInt)); err != nil {
		var errUnauthorized *ErrUnauthorized
		var errNotFound *ErrNotFound

		switch {
		case errors.As(err, &errUnauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Message, nil)
		case errors.As(err, &errNotFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Message, nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to delete exercise", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
