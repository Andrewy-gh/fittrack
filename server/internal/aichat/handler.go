package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type chatService interface {
	Validate(ctx context.Context, prompt string) (*ValidateResponse, error)
	StreamValidate(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error)
}

type Handler struct {
	logger  *slog.Logger
	service chatService
}

func NewHandler(logger *slog.Logger, service chatService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	req, ok := h.decodeRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.service.Validate(r.Context(), req.Prompt)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) StreamValidate(w http.ResponseWriter, r *http.Request) {
	req, ok := h.decodeRequest(w, r)
	if !ok {
		return
	}

	sse := newSSEWriter(w)
	sse.prepareHeaders()

	if err := sse.write("start", StreamEvent{
		Type:      "start",
		RequestID: request.GetRequestID(r.Context()),
	}); err != nil {
		h.logger.Error("failed to start ai chat stream", "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
		return
	}

	done, err := h.service.StreamValidate(r.Context(), req.Prompt, func(delta string) error {
		return sse.write("delta", StreamEvent{
			Type:  "delta",
			Delta: delta,
		})
	})
	if err != nil {
		if !h.writeStreamError(sse, r, err) {
			h.logger.Error("failed to stream ai chat response", "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
		}
		return
	}

	if err := sse.write("done", StreamEvent{
		Type:  "done",
		Model: done.Model,
		Text:  done.Text,
	}); err != nil {
		h.logger.Error("failed to finish ai chat stream", "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
	}
}

func (h *Handler) decodeRequest(w http.ResponseWriter, r *http.Request) (*ValidateRequest, bool) {
	var req ValidateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return nil, false
	}

	req.Prompt = strings.TrimSpace(req.Prompt)
	if req.Prompt == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "prompt is required", nil)
		return nil, false
	}

	return &req, true
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var errUnauthorized *apperrors.Unauthorized
	switch {
	case errors.As(err, &errUnauthorized):
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
	case errors.Is(err, ErrFeatureDisabled):
		response.ErrorJSON(w, r, h.logger, http.StatusForbidden, "ai chat feature is not enabled for this user", nil)
	case errors.Is(err, ErrRuntimeUnavailable):
		response.ErrorJSON(w, r, h.logger, http.StatusServiceUnavailable, "ai chat runtime is not configured", nil)
	default:
		response.ErrorJSON(w, r, h.logger, http.StatusBadGateway, "failed to generate ai chat validation", err)
	}
}

func (h *Handler) writeStreamError(sse *sseWriter, r *http.Request, err error) bool {
	var errUnauthorized *apperrors.Unauthorized

	event := StreamEvent{Type: "error"}
	switch {
	case errors.As(err, &errUnauthorized):
		event.Message = errUnauthorized.Error()
	case errors.Is(err, ErrFeatureDisabled):
		event.Message = "ai chat feature is not enabled for this user"
	case errors.Is(err, ErrRuntimeUnavailable):
		event.Message = "ai chat runtime is not configured"
	default:
		event.Message = "failed to generate ai chat validation"
	}

	return sse.write("error", event) == nil
}
