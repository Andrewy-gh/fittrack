package exercise

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/exercisecatalog"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
)

type UpdateCatalogRequest struct {
	// Empty string explicitly clears classification. The field must be present.
	CatalogID *string `json:"catalog_id" validate:"required"`
}

var errInvalidCatalog = errors.New("unknown catalog ID")

func (es *ExerciseService) UpdateCatalog(ctx context.Context, id int32, catalogID string) error {
	userID, ok := user.Current(ctx)
	if !ok {
		return &apperrors.Unauthorized{Resource: "exercise"}
	}
	var link *string
	if catalogID != "" {
		if exercisecatalog.Find(catalogID) == nil {
			return errInvalidCatalog
		}
		link = &catalogID
	}
	err := es.repo.UpdateExerciseCatalog(ctx, id, userID, link)
	if errors.Is(err, pgx.ErrNoRows) {
		return &apperrors.NotFound{Resource: "exercise", ID: fmt.Sprint(id)}
	}
	return err
}

// ListCatalog godoc
// @Summary List reviewed exercise catalog
// @Tags exercises
// @Produce json
// @Security StackAuth
// @Success 200 {array} exercisecatalog.Entry
// @Failure 401 {object} response.ErrorResponse
// @Router /exercise-catalog [get]
func (h *ExerciseHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	if _, ok := user.Current(r.Context()); !ok {
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	if err := response.JSON(w, http.StatusOK, exercisecatalog.All()); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to write catalog", err)
	}
}

// UpdateCatalog godoc
// @Summary Classify an exercise or clear its classification
// @Description Changes only the catalog link. An empty catalog_id clears it; name and history are preserved.
// @Tags exercises
// @Accept json
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param request body exercise.UpdateCatalogRequest true "Catalog link"
// @Success 204 "Updated"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /exercises/{id}/catalog [put]
func (h *ExerciseHandler) UpdateCatalog(w http.ResponseWriter, r *http.Request) {
	id, ok := h.decodeExerciseID(w, r)
	if !ok {
		return
	}
	var req UpdateCatalogRequest
	if err := decodeStrictJSON(w, r, &req); err != nil || req.CatalogID == nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "catalog_id is required; use an empty string to clear", err)
		return
	}
	if err := h.exerciseService.UpdateCatalog(r.Context(), id, *req.CatalogID); err != nil {
		var unauthorized *apperrors.Unauthorized
		var notFound *apperrors.NotFound
		switch {
		case errors.As(err, &unauthorized):
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, "Unauthorized", nil)
		case errors.As(err, &notFound):
			response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "Exercise not found", nil)
		case errors.Is(err, errInvalidCatalog):
			response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, err.Error(), nil)
		default:
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to update classification", err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
