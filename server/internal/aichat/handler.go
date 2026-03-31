package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type chatService interface {
	Validate(ctx context.Context, prompt string) (*ValidateResponse, error)
	EnsureAllowed(ctx context.Context) error
	CurrentUserHasFeatureAccess(ctx context.Context) (bool, error)
	StreamValidate(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error)
	CreateConversation(ctx context.Context) (*Conversation, error)
	GetConversation(ctx context.Context, conversationID int32) (*ConversationDetail, error)
	PrepareMessageStream(ctx context.Context, conversationID int32, prompt string, requestID string) (*PreparedMessageStream, error)
	StreamMessage(ctx context.Context, prepared *PreparedMessageStream, onChunk func(string) error) (*StreamDone, error)
	AbortPreparedMessageStream(ctx context.Context, prepared *PreparedMessageStream, failure error) error
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

// Validate godoc
// @Summary Validate AI chat architecture
// @Description Phase-0 validation endpoint for FitTrack AI chat architecture.
// @Tags ai-chat
// @Accept json
// @Produce json
// @Security StackAuth
// @Param request body aichat.ValidateRequest true "Validation request"
// @Success 200 {object} aichat.ValidateResponse
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 503 {object} response.Error
// @Failure 502 {object} response.Error
// @Router /ai/chat/validate [post]
func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	req, ok := h.decodePromptRequest(w, r)
	if !ok {
		return
	}

	resp, err := h.service.Validate(r.Context(), req.Prompt)
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusBadGateway, "failed to generate ai chat validation")
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

// StreamValidate godoc
// @Summary Stream AI chat architecture validation
// @Description Phase-0 validation stream endpoint. Preflight failures return JSON; successful requests upgrade to SSE.
// @Tags ai-chat
// @Accept json
// @Produce text/event-stream
// @Security StackAuth
// @Param request body aichat.ValidateRequest true "Validation request"
// @Success 200 {object} aichat.StreamEvent
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 503 {object} response.Error
// @Failure 502 {object} response.Error
// @Router /ai/chat/validate/stream [post]
func (h *Handler) StreamValidate(w http.ResponseWriter, r *http.Request) {
	req, ok := h.decodePromptRequest(w, r)
	if !ok {
		return
	}

	if err := h.service.EnsureAllowed(r.Context()); err != nil {
		h.writeServiceError(w, r, err, http.StatusBadGateway, "failed to start ai chat validation stream")
		return
	}

	sse, err := h.startSSE(w)
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to start ai chat validation stream", err)
		return
	}

	if err := sse.write("start", StreamEvent{
		Type:      "start",
		RequestID: request.GetRequestID(r.Context()),
	}); err != nil {
		h.logStreamWriteFailure("failed to start ai chat validation stream", r, err)
		return
	}

	done, err := h.service.StreamValidate(r.Context(), req.Prompt, func(delta string) error {
		return sse.write("delta", StreamEvent{
			Type:  "delta",
			Delta: delta,
		})
	})
	if err != nil {
		if !h.writeStreamError(sse, r, 0, 0, 0, "failed to generate ai chat validation", err) {
			h.logger.Error("failed to stream ai chat validation response", "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
		}
		return
	}

	if err := sse.write("done", StreamEvent{
		Type:  "done",
		Model: done.Model,
		Text:  done.Text,
	}); err != nil {
		h.logStreamWriteFailure("failed to finish ai chat validation stream", r, err)
	}
}

// CreateConversation godoc
// @Summary Create AI chat conversation
// @Description Creates an empty persisted AI chat conversation for the authenticated user.
// @Tags ai-chat
// @Produce json
// @Security StackAuth
// @Success 201 {object} aichat.Conversation
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /ai/conversations [post]
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	conversation, err := h.service.CreateConversation(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to create ai chat conversation")
		return
	}

	if err := response.JSON(w, http.StatusCreated, conversation); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

// GetConversation godoc
// @Summary Get AI chat conversation
// @Description Returns a persisted AI chat conversation with its messages.
// @Tags ai-chat
// @Produce json
// @Security StackAuth
// @Param id path int true "Conversation ID"
// @Success 200 {object} aichat.ConversationDetail
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /ai/conversations/{id} [get]
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := h.decodeConversationID(w, r)
	if !ok {
		return
	}

	detail, err := h.service.GetConversation(r.Context(), conversationID)
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to load ai chat conversation")
		return
	}

	if err := response.JSON(w, http.StatusOK, detail); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

// RecordTelemetry godoc
// @Summary Record AI chat telemetry
// @Description Records authenticated client-observed AI chat outcomes for observability and rollout gating.
// @Tags ai-chat
// @Accept json
// @Produce json
// @Security StackAuth
// @Param request body aichat.ClientTelemetryEvent true "Telemetry event"
// @Success 202
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /ai/chat/telemetry [post]
func (h *Handler) RecordTelemetry(w http.ResponseWriter, r *http.Request) {
	var event ClientTelemetryEvent

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return
	}

	event = normalizeClientTelemetryEvent(event)
	if err := validateClientTelemetryEvent(event); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "invalid ai chat telemetry event", err)
		return
	}

	hasFeatureAccess, err := h.service.CurrentUserHasFeatureAccess(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to record ai chat telemetry")
		return
	}

	recordClientTelemetry(hasFeatureAccess, event)
	w.WriteHeader(http.StatusAccepted)
}

// StreamMessage godoc
// @Summary Stream AI chat message
// @Description Persists a user message and run, then streams normalized assistant events over SSE. Preflight failures stay JSON.
// @Tags ai-chat
// @Accept json
// @Produce text/event-stream
// @Security StackAuth
// @Param id path int true "Conversation ID"
// @Param request body aichat.SendMessageRequest true "Send message request"
// @Success 200 {object} aichat.StreamEvent
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 403 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 409 {object} response.Error
// @Failure 503 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /ai/conversations/{id}/messages/stream [post]
func (h *Handler) StreamMessage(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := h.decodeConversationID(w, r)
	if !ok {
		return
	}

	req, ok := h.decodeSendMessageRequest(w, r)
	if !ok {
		return
	}

	prepared, err := h.service.PrepareMessageStream(r.Context(), conversationID, req.Prompt, request.GetRequestID(r.Context()))
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to start ai chat stream")
		return
	}

	sse, err := h.startSSE(w)
	if err != nil {
		if abortErr := h.service.AbortPreparedMessageStream(r.Context(), prepared, ErrStreamNotStarted); abortErr != nil {
			h.logger.Error("failed to abort ai chat run after stream preflight error", "error", abortErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to start ai chat stream", err)
		return
	}

	if err := sse.write("start", StreamEvent{
		Type:           "start",
		RequestID:      request.GetRequestID(r.Context()),
		ConversationID: prepared.Conversation.ID,
		RunID:          prepared.Run.ID,
		MessageID:      prepared.AssistantMessage.ID,
		Model:          prepared.Run.Model,
	}); err != nil {
		if abortErr := h.service.AbortPreparedMessageStream(r.Context(), prepared, ErrStreamNotStarted); abortErr != nil {
			h.logger.Error("failed to abort ai chat run after start event write failure", "error", abortErr, "conversation_id", prepared.Conversation.ID, "run_id", prepared.Run.ID)
		}
		h.logStreamWriteFailure("failed to start ai chat stream", r, err)
		return
	}

	done, err := h.service.StreamMessage(r.Context(), prepared, func(delta string) error {
		return sse.write("delta", StreamEvent{
			Type:  "delta",
			Delta: delta,
		})
	})
	if err != nil {
		if !h.writeStreamError(sse, r, prepared.Conversation.ID, prepared.Run.ID, prepared.AssistantMessage.ID, "failed to generate ai chat response", err) {
			h.logger.Error("failed to stream ai chat response", "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
		}
		return
	}

	if err := sse.write("done", StreamEvent{
		Type:           "done",
		ConversationID: done.ConversationID,
		RunID:          done.RunID,
		MessageID:      done.MessageID,
		Model:          done.Model,
		Text:           done.Text,
	}); err != nil {
		h.logStreamWriteFailure("failed to finish ai chat stream", r, err)
	}
}

func (h *Handler) decodePromptRequest(w http.ResponseWriter, r *http.Request) (*ValidateRequest, bool) {
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

func (h *Handler) decodeSendMessageRequest(w http.ResponseWriter, r *http.Request) (*SendMessageRequest, bool) {
	var req SendMessageRequest

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

func (h *Handler) decodeConversationID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	raw := strings.TrimSpace(r.PathValue("id"))
	if raw == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "conversation id is required", nil)
		return 0, false
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "conversation id must be a positive integer", err)
		return 0, false
	}

	return int32(parsed), true
}

func (h *Handler) startSSE(w http.ResponseWriter) (*sseWriter, error) {
	sse := newSSEWriter(w)
	sse.prepareHeaders()
	if err := sse.disableWriteTimeout(); err != nil {
		return nil, err
	}
	return sse, nil
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error, unexpectedStatus int, unexpectedMessage string) {
	var errUnauthorized *apperrors.Unauthorized
	var errNotFound *apperrors.NotFound

	switch {
	case errors.As(err, &errUnauthorized):
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
	case errors.As(err, &errNotFound):
		response.ErrorJSON(w, r, h.logger, http.StatusNotFound, errNotFound.Error(), nil)
	case errors.Is(err, ErrFeatureDisabled):
		response.ErrorJSON(w, r, h.logger, http.StatusForbidden, "ai chat feature is not enabled for this user", nil)
	case errors.Is(err, ErrRuntimeUnavailable):
		response.ErrorJSON(w, r, h.logger, http.StatusServiceUnavailable, "ai chat runtime is not configured", nil)
	case errors.Is(err, ErrConversationBusy):
		response.ErrorJSON(w, r, h.logger, http.StatusConflict, "ai chat conversation already has an active run", nil)
	case errors.Is(err, ErrGenerationTimeout):
		response.ErrorJSON(w, r, h.logger, http.StatusGatewayTimeout, "ai chat generation timed out", nil)
	default:
		response.ErrorJSON(w, r, h.logger, unexpectedStatus, unexpectedMessage, err)
	}
}

func (h *Handler) writeStreamError(sse *sseWriter, r *http.Request, conversationID int32, runID int32, messageID int32, fallbackMessage string, err error) bool {
	var errUnauthorized *apperrors.Unauthorized

	event := StreamEvent{
		Type:           "error",
		ConversationID: conversationID,
		RunID:          runID,
		MessageID:      messageID,
	}

	switch {
	case errors.As(err, &errUnauthorized):
		event.Message = errUnauthorized.Error()
	case errors.Is(err, ErrFeatureDisabled):
		event.Message = "ai chat feature is not enabled for this user"
	case errors.Is(err, ErrRuntimeUnavailable):
		event.Message = "ai chat runtime is not configured"
	case errors.Is(err, ErrGenerationTimeout):
		event.Message = "ai chat generation timed out; retry with a shorter prompt"
	default:
		event.Message = fallbackMessage
	}

	return sse.write("error", event) == nil
}

func (h *Handler) logStreamWriteFailure(message string, r *http.Request, err error) {
	h.logger.Error(message, "error", err, "path", r.URL.Path, "request_id", request.GetRequestID(r.Context()))
}
