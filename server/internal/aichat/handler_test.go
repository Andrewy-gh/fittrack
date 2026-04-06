package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockChatService struct {
	mock.Mock
}

type streamResponseRecorder struct {
	*httptest.ResponseRecorder
}

type failingStreamResponseRecorder struct {
	*streamResponseRecorder
	failWrites bool
}

func newStreamResponseRecorder() *streamResponseRecorder {
	return &streamResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func newFailingStreamResponseRecorder() *failingStreamResponseRecorder {
	return &failingStreamResponseRecorder{streamResponseRecorder: newStreamResponseRecorder(), failWrites: true}
}

func (r *streamResponseRecorder) SetWriteDeadline(time.Time) error {
	return nil
}

func (r *failingStreamResponseRecorder) Write(p []byte) (int, error) {
	if r.failWrites {
		return 0, errors.New("broken pipe")
	}
	return r.streamResponseRecorder.Write(p)
}

func (m *mockChatService) Validate(ctx context.Context, prompt string) (*ValidateResponse, error) {
	args := m.Called(ctx, prompt)
	resp, _ := args.Get(0).(*ValidateResponse)
	return resp, args.Error(1)
}

func (m *mockChatService) EnsureAllowed(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockChatService) CurrentUserHasFeatureAccess(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *mockChatService) StreamValidate(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error) {
	args := m.Called(ctx, prompt, onChunk)
	if args.Error(1) == nil {
		_ = onChunk("Phase 0 ")
		_ = onChunk("stream works.")
	}
	done, _ := args.Get(0).(*StreamDone)
	return done, args.Error(1)
}

func (m *mockChatService) CreateConversation(ctx context.Context) (*Conversation, error) {
	args := m.Called(ctx)
	conversation, _ := args.Get(0).(*Conversation)
	return conversation, args.Error(1)
}

func (m *mockChatService) GetConversation(ctx context.Context, conversationID int32) (*ConversationDetail, error) {
	args := m.Called(ctx, conversationID)
	detail, _ := args.Get(0).(*ConversationDetail)
	return detail, args.Error(1)
}

func (m *mockChatService) RequestMessageRecovery(ctx context.Context, conversationID int32, reason string) (*RecoverMessageResponse, error) {
	args := m.Called(ctx, conversationID, reason)
	resp, _ := args.Get(0).(*RecoverMessageResponse)
	return resp, args.Error(1)
}

func (m *mockChatService) PrepareMessageStream(ctx context.Context, conversationID int32, prompt string, requestID string) (*PreparedMessageStream, error) {
	args := m.Called(ctx, conversationID, prompt, requestID)
	prepared, _ := args.Get(0).(*PreparedMessageStream)
	return prepared, args.Error(1)
}

func (m *mockChatService) StreamMessage(ctx context.Context, prepared *PreparedMessageStream, onChunk func(string) error) (*StreamDone, error) {
	args := m.Called(ctx, prepared, onChunk)
	if args.Error(1) == nil {
		_ = onChunk("hello ")
		_ = onChunk("world")
	}
	done, _ := args.Get(0).(*StreamDone)
	return done, args.Error(1)
}

func (m *mockChatService) AbortPreparedMessageStream(ctx context.Context, prepared *PreparedMessageStream, failure error) error {
	args := m.Called(ctx, prepared, failure)
	return args.Error(0)
}

func TestHandlerValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("writes structured response", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("Validate", mock.Anything, "prove the slice").Return(&ValidateResponse{
			Model: defaultModelName,
			Output: &ValidationOutput{
				Summary:  "Viable.",
				NextStep: "Build phase 1.",
			},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate", strings.NewReader(`{"prompt":"prove the slice"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.Validate(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		var resp ValidateResponse
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
		assert.Equal(t, defaultModelName, resp.Model)
		require.NotNil(t, resp.Output)
		assert.Equal(t, "Viable.", resp.Output.Summary)
		service.AssertExpectations(t)
	})

	t.Run("maps runtime unavailable to 503", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("Validate", mock.Anything, "prove the slice").Return((*ValidateResponse)(nil), ErrRuntimeUnavailable).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate", strings.NewReader(`{"prompt":"prove the slice"}`))
		rr := httptest.NewRecorder()

		handler.Validate(rr, req)

		require.Equal(t, http.StatusServiceUnavailable, rr.Code)
		assert.Contains(t, rr.Body.String(), "ai chat runtime is not configured")
		service.AssertExpectations(t)
	})

	t.Run("rejects blank prompts", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate", strings.NewReader(`{"prompt":"   "}`))
		rr := httptest.NewRecorder()

		handler.Validate(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "prompt is required")
		service.AssertExpectations(t)
	})
}

func TestHandlerStreamValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("writes SSE frames with deltas and done event", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("EnsureAllowed", mock.Anything).Return(nil).Once()
		service.On("StreamValidate", mock.Anything, "prove streaming", mock.Anything).Return(&StreamDone{
			Model: defaultModelName,
			Text:  "Phase 0 stream works.",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		req = req.WithContext(request.WithRequestID(req.Context(), "req-123"))
		rr := newStreamResponseRecorder()

		handler.StreamValidate(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
		assert.True(t, rr.Flushed)
		body := rr.Body.String()
		assert.Contains(t, body, "event: start")
		assert.Contains(t, body, `"request_id":"req-123"`)
		assert.Contains(t, body, "event: delta")
		assert.Contains(t, body, `"delta":"Phase 0 "`)
		assert.Contains(t, body, "event: done")
		assert.Contains(t, body, `"text":"Phase 0 stream works."`)
		service.AssertExpectations(t)
	})

	t.Run("returns json error before opening stream when preflight fails", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("EnsureAllowed", mock.Anything).Return(ErrRuntimeUnavailable).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		rr := newStreamResponseRecorder()

		handler.StreamValidate(rr, req)

		require.Equal(t, http.StatusServiceUnavailable, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.NotContains(t, rr.Body.String(), "event: start")
		service.AssertNotCalled(t, "StreamValidate", mock.Anything, mock.Anything, mock.Anything)
		service.AssertExpectations(t)
	})

	t.Run("writes safe error event after stream starts", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("EnsureAllowed", mock.Anything).Return(nil).Once()
		service.On("StreamValidate", mock.Anything, "prove streaming", mock.Anything).Return((*StreamDone)(nil), errors.New("provider failed")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		rr := newStreamResponseRecorder()

		handler.StreamValidate(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `event: error`)
		assert.Contains(t, rr.Body.String(), `"message":"failed to generate ai chat validation"`)
		service.AssertExpectations(t)
	})

	t.Run("rejects malformed request before opening stream", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate/stream", strings.NewReader(`{"prompt":`))
		rr := httptest.NewRecorder()

		handler.StreamValidate(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		service.AssertExpectations(t)
	})
}

func TestHandlerWriteServiceError_DefaultsToBadGateway(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := new(mockChatService)
	handler := NewHandler(logger, service)
	service.On("Validate", mock.Anything, "prove the slice").Return((*ValidateResponse)(nil), errors.New("provider failed")).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/validate", strings.NewReader(`{"prompt":"prove the slice"}`))
	rr := httptest.NewRecorder()

	handler.Validate(rr, req)

	require.Equal(t, http.StatusBadGateway, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed to generate ai chat validation")
	service.AssertExpectations(t)
}

func TestHandlerCreateConversation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := new(mockChatService)
	handler := NewHandler(logger, service)
	now := time.Date(2026, 3, 26, 17, 0, 0, 0, time.UTC)
	service.On("CreateConversation", mock.Anything).Return(&Conversation{
		ID:        41,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations", nil)
	rr := httptest.NewRecorder()

	handler.CreateConversation(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var conversation Conversation
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&conversation))
	assert.Equal(t, int32(41), conversation.ID)
	service.AssertExpectations(t)
}

func TestHandlerGetConversation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns conversation detail", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		now := time.Date(2026, 3, 26, 17, 5, 0, 0, time.UTC)
		service.On("GetConversation", mock.Anything, int32(41)).Return(&ConversationDetail{
			Conversation: &Conversation{
				ID:        41,
				CreatedAt: now,
				UpdatedAt: now,
			},
			Messages: []ChatMessage{
				{
					ID:             1,
					ConversationID: 41,
					Role:           roleUser,
					Content:        "hello",
					Status:         statusCompleted,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/ai/conversations/41", nil)
		req.SetPathValue("id", "41")
		rr := httptest.NewRecorder()

		handler.GetConversation(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"conversation"`)
		assert.Contains(t, rr.Body.String(), `"messages"`)
		service.AssertExpectations(t)
	})

	t.Run("maps not found", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("GetConversation", mock.Anything, int32(41)).Return((*ConversationDetail)(nil), apperrors.NewNotFound("ai conversation", "41")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/ai/conversations/41", nil)
		req.SetPathValue("id", "41")
		rr := httptest.NewRecorder()

		handler.GetConversation(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "ai conversation with id 41 not found")
		service.AssertExpectations(t)
	})
}

func TestHandlerRecordTelemetry(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("accepts a valid stream telemetry event", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		aiChatClientOutcomesTotal.Reset()
		service.On("CurrentUserHasFeatureAccess", mock.Anything).Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/telemetry", strings.NewReader(`{"category":"stream","outcome":"transport_ended_pre_terminal","stage":"pre_start"}`))
		rr := httptest.NewRecorder()

		handler.RecordTelemetry(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
		assert.Equal(t, float64(1), testutil.ToFloat64(aiChatClientOutcomesTotal.WithLabelValues(
			telemetryCategoryStream,
			telemetryOutcomeTransportEndedPreTerminal,
			telemetryStreamStagePreStart,
			telemetryCohortBeta,
		)))
		service.AssertExpectations(t)
	})

	t.Run("normalizes padded telemetry labels before recording", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		aiChatClientOutcomesTotal.Reset()
		service.On("CurrentUserHasFeatureAccess", mock.Anything).Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/telemetry", strings.NewReader(`{"category":" stream ","outcome":" transport_ended_pre_terminal ","stage":" pre_start "}`))
		rr := httptest.NewRecorder()

		handler.RecordTelemetry(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
		assert.Equal(t, float64(1), testutil.ToFloat64(aiChatClientOutcomesTotal.WithLabelValues(
			telemetryCategoryStream,
			telemetryOutcomeTransportEndedPreTerminal,
			telemetryStreamStagePreStart,
			telemetryCohortBeta,
		)))
		assert.Equal(t, float64(0), testutil.ToFloat64(aiChatClientOutcomesTotal.WithLabelValues(
			" stream ",
			" transport_ended_pre_terminal ",
			" pre_start ",
			telemetryCohortBeta,
		)))
		service.AssertExpectations(t)
	})

	t.Run("rejects invalid telemetry outcomes", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)

		req := httptest.NewRequest(http.MethodPost, "/api/ai/chat/telemetry", strings.NewReader(`{"category":"stream","outcome":"nope","stage":"pre_start"}`))
		rr := httptest.NewRecorder()

		handler.RecordTelemetry(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid ai chat telemetry event")
		service.AssertNotCalled(t, "CurrentUserHasFeatureAccess", mock.Anything)
	})
}

func TestHandlerStreamMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("returns json error before opening stream when preflight fails", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("PrepareMessageStream", mock.Anything, int32(41), "prove streaming", "req-123").Return((*PreparedMessageStream)(nil), ErrRuntimeUnavailable).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations/41/messages/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		req = req.WithContext(request.WithRequestID(req.Context(), "req-123"))
		req.SetPathValue("id", "41")
		rr := newStreamResponseRecorder()

		handler.StreamMessage(rr, req)

		require.Equal(t, http.StatusServiceUnavailable, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.NotContains(t, rr.Body.String(), "event: start")
		service.AssertNotCalled(t, "StreamMessage", mock.Anything, mock.Anything, mock.Anything)
		service.AssertExpectations(t)
	})

	t.Run("writes SSE frames with start delta done sequence", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		prepared := preparedStreamFixture()
		service.On("PrepareMessageStream", mock.Anything, int32(41), "prove streaming", "req-123").Return(prepared, nil).Once()
		service.On("StreamMessage", mock.Anything, prepared, mock.Anything).Return(&StreamDone{
			ConversationID: 41,
			RunID:          51,
			MessageID:      61,
			Model:          defaultModelName,
			Text:           "hello world",
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations/41/messages/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		req = req.WithContext(request.WithRequestID(req.Context(), "req-123"))
		req.SetPathValue("id", "41")
		rr := newStreamResponseRecorder()

		handler.StreamMessage(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))
		body := rr.Body.String()
		assert.Contains(t, body, "event: start")
		assert.Contains(t, body, `"request_id":"req-123"`)
		assert.Contains(t, body, `"conversation_id":41`)
		assert.Contains(t, body, `"run_id":51`)
		assert.Contains(t, body, `"message_id":61`)
		assert.Contains(t, body, "event: delta")
		assert.Contains(t, body, `"delta":"hello "`)
		assert.Contains(t, body, "event: done")
		assert.Contains(t, body, `"text":"hello world"`)
		service.AssertExpectations(t)
	})

	t.Run("writes SSE error event after stream starts", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		prepared := preparedStreamFixture()
		service.On("PrepareMessageStream", mock.Anything, int32(41), "prove streaming", "req-123").Return(prepared, nil).Once()
		service.On("StreamMessage", mock.Anything, prepared, mock.Anything).Return((*StreamDone)(nil), errors.New("provider failed")).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations/41/messages/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		req = req.WithContext(request.WithRequestID(req.Context(), "req-123"))
		req.SetPathValue("id", "41")
		rr := newStreamResponseRecorder()

		handler.StreamMessage(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "event: start")
		assert.Contains(t, rr.Body.String(), `event: error`)
		assert.Contains(t, rr.Body.String(), `"message":"failed to generate ai chat response"`)
		service.AssertExpectations(t)
	})

	t.Run("aborts prepared run when start event write disconnects", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		prepared := preparedStreamFixture()
		service.On("PrepareMessageStream", mock.Anything, int32(41), "prove streaming", "req-123").Return(prepared, nil).Once()
		service.On("AbortPreparedMessageStream", mock.Anything, prepared, ErrStreamNotStarted).Return(nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations/41/messages/stream", strings.NewReader(`{"prompt":"prove streaming"}`))
		req = req.WithContext(request.WithRequestID(req.Context(), "req-123"))
		req.SetPathValue("id", "41")
		rr := newFailingStreamResponseRecorder()

		handler.StreamMessage(rr, req)

		service.AssertNotCalled(t, "StreamMessage", mock.Anything, mock.Anything, mock.Anything)
		service.AssertExpectations(t)
	})
}

func TestHandlerRecoverMessage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("queues recovery for an active run", func(t *testing.T) {
		service := new(mockChatService)
		handler := NewHandler(logger, service)
		service.On("RequestMessageRecovery", mock.Anything, int32(41), recoverReasonStreamReconnect).Return(&RecoverMessageResponse{
			ConversationID: 41,
			RunID:          51,
			Status:         recoverStatusQueued,
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/ai/conversations/41/messages/recover", nil)
		req.SetPathValue("id", "41")
		rr := httptest.NewRecorder()

		handler.RecoverMessage(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
		assert.Contains(t, rr.Body.String(), `"conversation_id":41`)
		assert.Contains(t, rr.Body.String(), `"run_id":51`)
		assert.Contains(t, rr.Body.String(), `"status":"queued"`)
		service.AssertExpectations(t)
	})
}

func preparedStreamFixture() *PreparedMessageStream {
	now := time.Date(2026, 3, 26, 17, 10, 0, 0, time.UTC)
	return &PreparedMessageStream{
		Conversation: &Conversation{
			ID:        41,
			UserID:    "user-123",
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserMessage: &ChatMessage{
			ID:             60,
			ConversationID: 41,
			UserID:         "user-123",
			Role:           roleUser,
			Content:        "prove streaming",
			Status:         statusCompleted,
			CreatedAt:      now,
			UpdatedAt:      now,
			CompletedAt:    &now,
		},
		AssistantMessage: &ChatMessage{
			ID:             61,
			ConversationID: 41,
			UserID:         "user-123",
			Role:           roleAssistant,
			Content:        "",
			Status:         statusStreaming,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		Run: &ChatRun{
			ID:                 51,
			ConversationID:     41,
			UserID:             "user-123",
			UserMessageID:      60,
			AssistantMessageID: 61,
			Model:              defaultModelName,
			Status:             statusStreaming,
			CreatedAt:          now,
			UpdatedAt:          now,
			StartedAt:          now,
		},
		Prompt: "prove streaming",
	}
}
