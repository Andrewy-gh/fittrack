package aichat

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStreamInputAdmission(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	prefix := `{"prompt":"x"}`
	tests := []struct {
		name, body string
		status     int
	}{
		{"prompt exact ASCII", `{"prompt":"` + strings.Repeat("a", 8192) + `"}`, http.StatusConflict},
		{"prompt too long ASCII", `{"prompt":"` + strings.Repeat("a", 8193) + `"}`, http.StatusBadRequest},
		{"prompt exact UTF8", `{"prompt":"` + strings.Repeat("é", 4096) + `"}`, http.StatusConflict},
		{"prompt too long UTF8", `{"prompt":"` + strings.Repeat("é", 4096) + `a"}`, http.StatusBadRequest},
		{"trimmed prompt", `{"prompt":"  ` + strings.Repeat("a", 8192) + `  "}`, http.StatusConflict},
		{"body exact", prefix + strings.Repeat(" ", 65536-len(prefix)), http.StatusConflict},
		{"body over", prefix + strings.Repeat(" ", 65537-len(prefix)), http.StatusRequestEntityTooLarge},
		{"invalid wire encoding", "{\"prompt\":\"\xff\"}", http.StatusBadRequest},
		{"valid replacement character", `{"prompt":"�"}`, http.StatusConflict},
		{"second value", prefix + ` {}`, http.StatusBadRequest},
		{"trailing garbage", prefix + ` x`, http.StatusBadRequest},
		{"unknown field", `{"prompt":"x","other":1}`, http.StatusBadRequest},
		{"escaped prompt over byte limit", `{"prompt":"` + strings.Repeat(`\u00e9`, 4097) + `"}`, http.StatusBadRequest},
		{"empty prompt", `{"prompt":"  "}`, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := new(mockChatService)
			h := NewHandler(logger, svc)
			if tc.status == http.StatusConflict {
				svc.On("PrepareMessageStream", mock.Anything, int32(1), mock.Anything, mock.Anything).Return((*PreparedMessageStream)(nil), ErrConversationBusy).Once()
			}
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			// Exercise streaming bodies without relying on Content-Length.
			req.ContentLength = -1
			req.SetPathValue("id", "1")
			rr := httptest.NewRecorder()
			h.StreamMessage(rr, req)
			require.Equal(t, tc.status, rr.Code, rr.Body.String())
			if tc.status != http.StatusConflict {
				svc.AssertNotCalled(t, "PrepareMessageStream", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
			svc.AssertExpectations(t)
		})
	}
}

func TestServiceRejectsOversizePromptBeforeAdmission(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	runtime := new(mockRuntime)
	access := new(mockFeatureAccessService)
	service := NewService(logger, access, runtime, repo, nil)
	ctx := user.WithContext(context.Background(), "user-123")
	runtime.On("Available").Return(true)
	access.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil)
	runtime.On("ModelName").Return(defaultModelName)
	// Without the guard, admission reaches the repository instead of rejecting the prompt.
	repo.On("PrepareMessageStream", mock.Anything, int32(1), "user-123", mock.Anything, defaultModelName, "request").
		Return((*PreparedMessageStream)(nil), ErrConversationBusy).Maybe()

	result, err := service.PrepareMessageStream(ctx, 1, strings.Repeat("é", 4097), "request")
	require.ErrorIs(t, err, errPromptTooLarge)
	require.Nil(t, result)
	repo.AssertNumberOfCalls(t, "PrepareMessageStream", 0)
}
