package e2eauth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type bootstrapService interface {
	Bootstrap(ctx context.Context) (*BootstrapResponse, error)
	SeedConversation(ctx context.Context, request SeedConversationRequest) (*SeedConversationResponse, error)
}

type Handler struct {
	logger  *slog.Logger
	service bootstrapService
}

func NewHandler(logger *slog.Logger, service bootstrapService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.Bootstrap(r.Context())
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to bootstrap local e2e auth", err)
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) SeedConversation(w http.ResponseWriter, r *http.Request) {
	var req SeedConversationRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return
	}

	if strings.TrimSpace(req.LatestWorkoutDraft.Date) == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "latest_workout_draft.date is required", nil)
		return
	}
	if len(req.LatestWorkoutDraft.Exercises) == 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "latest_workout_draft.exercises must include at least one exercise", nil)
		return
	}

	resp, err := h.service.SeedConversation(r.Context(), req)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to seed local ai chat conversation", err)
		return
	}

	if err := response.JSON(w, http.StatusCreated, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}
