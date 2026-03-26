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

	"github.com/Andrewy-gh/fittrack/server/internal/request"
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

func newStreamResponseRecorder() *streamResponseRecorder {
	return &streamResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (r *streamResponseRecorder) SetWriteDeadline(time.Time) error {
	return nil
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

func (m *mockChatService) StreamValidate(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error) {
	args := m.Called(ctx, prompt, onChunk)
	if args.Error(1) == nil {
		_ = onChunk("Phase 0 ")
		_ = onChunk("stream works.")
	}
	done, _ := args.Get(0).(*StreamDone)
	return done, args.Error(1)
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
