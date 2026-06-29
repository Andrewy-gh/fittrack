package aichat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

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

func (h *Handler) decodeResumeStreamQuery(w http.ResponseWriter, r *http.Request) (int32, int32, bool) {
	rawRunID := strings.TrimSpace(r.URL.Query().Get("runId"))
	if rawRunID == "" {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "runId is required", nil)
		return 0, 0, false
	}

	parsedRunID, err := strconv.ParseInt(rawRunID, 10, 32)
	if err != nil || parsedRunID <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "runId must be a positive integer", err)
		return 0, 0, false
	}

	rawAfterSequence := strings.TrimSpace(r.URL.Query().Get("afterSequence"))
	if rawAfterSequence == "" {
		return int32(parsedRunID), 0, true
	}

	parsedAfterSequence, err := strconv.ParseInt(rawAfterSequence, 10, 32)
	if err != nil || parsedAfterSequence < 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "afterSequence must be a non-negative integer", err)
		return 0, 0, false
	}

	return int32(parsedRunID), int32(parsedAfterSequence), true
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
	case errors.Is(err, billing.ErrTrialPromptLimitExceeded):
		response.ErrorJSON(w, r, h.logger, http.StatusForbidden, "ai chat trial prompt limit reached", nil)
	case errors.Is(err, ErrRuntimeUnavailable):
		response.ErrorJSON(w, r, h.logger, http.StatusServiceUnavailable, "ai chat runtime is not configured", nil)
	case errors.Is(err, ErrRecoveryUnavailable):
		response.ErrorJSON(w, r, h.logger, http.StatusServiceUnavailable, "ai chat recovery is not configured", nil)
	case errors.Is(err, ErrConversationBusy):
		response.ErrorJSON(w, r, h.logger, http.StatusConflict, "ai chat conversation already has an active run", nil)
	case errors.Is(err, ErrLatestWorkoutDraftUnavailable):
		response.ErrorJSON(w, r, h.logger, http.StatusConflict, "ai chat conversation does not have a latest workout draft to save", nil)
	case errors.Is(err, ErrLatestWorkoutDraftSuperseded):
		response.ErrorJSON(w, r, h.logger, http.StatusConflict, "ai chat latest workout draft changed before it could be saved", nil)
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
	case errors.Is(err, billing.ErrTrialPromptLimitExceeded):
		event.Message = "ai chat trial prompt limit reached"
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
	h.logger.Error(message,
		"path", r.URL.Path,
		"method", r.Method,
		"request_id", request.GetRequestID(r.Context()),
		"error_category", "stream_write",
		"error_present", err != nil,
		"error_type", fmt.Sprintf("%T", err))
}

func (h *Handler) logServiceFailure(message string, r *http.Request, err error, attrs ...any) {
	logAttrs := []any{
		"path", r.URL.Path,
		"method", r.Method,
		"request_id", request.GetRequestID(r.Context()),
		"error_category", "service",
		"error_present", err != nil,
		"error_type", fmt.Sprintf("%T", err),
	}
	logAttrs = append(logAttrs, attrs...)
	h.logger.Error(message, logAttrs...)
}
