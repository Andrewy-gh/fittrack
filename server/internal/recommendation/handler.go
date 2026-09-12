package recommendation

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

type PrescriptionRequest struct {
	Baseline *Range `json:"baseline" extensions:"x-nullable"`
}

// Get godoc
// @Summary Recommend today's working sets for an owned exercise
// @Tags recommendations
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param readiness query string false "normal, great, or sluggish"
// @Success 200 {object} recommendation.Result
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

// Prescribe godoc
// @Summary Save or clear an optional user-entered working-set baseline
// @Tags recommendations
// @Accept json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param request body recommendation.PrescriptionRequest true "Baseline; null clears it"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /exercises/{id}/prescription [put]
func (h *Handler) Prescribe(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r)
	if !ok {
		return
	}
	var envelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		h.fail(w, r, ErrInvalid)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		h.fail(w, r, ErrInvalid)
		return
	}
	if len(envelope.Baseline) == 0 {
		h.fail(w, r, ErrInvalid)
		return
	}
	var req PrescriptionRequest
	baselineDecoder := json.NewDecoder(strings.NewReader(string(envelope.Baseline)))
	baselineDecoder.DisallowUnknownFields()
	if err := baselineDecoder.Decode(&req.Baseline); err != nil {
		h.fail(w, r, ErrInvalid)
		return
	}
	if err := h.service.Prescribe(r.Context(), id, req.Baseline); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Saved godoc
// @Summary Read the recommendation context recorded when a workout was saved
// @Tags recommendations
// @Produce json
// @Security StackAuth
// @Param id path int true "Workout ID"
// @Success 200 {array} recommendation.Snapshot
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /workouts/{id}/recommendations [get]
func (h *Handler) Saved(w http.ResponseWriter, r *http.Request) {
	id, ok := h.id(w, r)
	if !ok {
		return
	}
	result, err := h.service.Saved(r.Context(), id)
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
		response.ErrorJSON(w, r, h.logger, 400, "invalid ID", nil)
		return 0, false
	}
	return int32(id), true
}
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	status, message := 500, "recommendation request failed"
	switch {
	case errors.Is(err, ErrUnauthorized):
		status, message = 401, err.Error()
	case errors.Is(err, ErrNotFound):
		status, message = 404, err.Error()
	case errors.Is(err, ErrInvalid):
		status, message = 400, err.Error()
	}
	response.ErrorJSON(w, r, h.logger, status, message, err)
}
