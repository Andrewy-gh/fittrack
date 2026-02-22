package exercise

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

// MARK: UpdateExerciseHistorical1RM
// UpdateExerciseHistorical1RM godoc
// @Summary Update exercise historical 1RM
// @Description Set a manual historical 1RM (clears source workout) or recompute from best working-set e1RM across history.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param body body UpdateExerciseHistorical1RMRequest true "Historical 1RM update request"
// @Success 204 "No Content - Exercise historical 1RM updated successfully"
// @Failure 400 {object} response.ErrorResponse "Bad Request - Invalid exercise ID or validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized - Invalid token"
// @Failure 404 {object} response.ErrorResponse "Not Found - Exercise not found or doesn't belong to user"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /exercises/{id}/historical-1rm [patch]
func (h *ExerciseHandler) UpdateExerciseHistorical1RM(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateExerciseHistorical1RMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Failed to decode request body", err)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := h.exerciseService.UpdateExerciseHistorical1RM(r.Context(), int32(exerciseIDInt), req); err != nil {
		var errUnauthorized *apperrors.Unauthorized
		var errNotFound *apperrors.NotFound

		switch {
		case errors.As(err, &errUnauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
		case errors.As(err, &errNotFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Error(), nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to update exercise historical 1RM", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
