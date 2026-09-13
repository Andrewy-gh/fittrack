package recommendation

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// Get godoc
// @Summary Recommend today's sets, reps, and weight for an owned exercise
// @Tags recommendations
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param readiness query string false "normal, great, or sluggish"
// @Success 200 {object} recommendation.Result
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /exercises/{id}/recommendation [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r)
	if !ok {
		return
	}
	readiness := r.URL.Query().Get("readiness")
	if readiness == "" {
		readiness = "normal"
	}
	result, err := h.service.Get(r.Context(), id, readiness)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	if err := response.JSON(w, http.StatusOK, result); err != nil {
		h.fail(w, r, err)
	}
}

func (h *Handler) id(w http.ResponseWriter, r *http.Request) (int32, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "invalid ID", nil)
		return 0, false
	}
	return int32(id), true
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	status, message := http.StatusInternalServerError, "recommendation request failed"
	switch {
	case errors.Is(err, ErrUnauthorized):
		status, message = http.StatusUnauthorized, err.Error()
	case errors.Is(err, ErrNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, ErrInvalid):
		status, message = http.StatusBadRequest, err.Error()
	}
	response.ErrorJSON(w, r, h.logger, status, message, err)
}
