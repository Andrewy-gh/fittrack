package e2eauth

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBootstrapService struct {
	mock.Mock
}

func (m *mockBootstrapService) Bootstrap(ctx context.Context) (*BootstrapResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*BootstrapResponse), args.Error(1)
}

func (m *mockBootstrapService) SeedConversation(ctx context.Context, request SeedConversationRequest) (*SeedConversationResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*SeedConversationResponse), args.Error(1)
}

func TestHandler_Bootstrap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &mockBootstrapService{}
	service.On("Bootstrap", mock.Anything).Return(&BootstrapResponse{
		UserID:      "local-e2e-user",
		Email:       "local-e2e-user@example.test",
		DisplayName: "Local E2E User",
		FeatureKeys: []string{FeatureKeyAIChatbot},
	}, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/dev/e2e/auth/bootstrap", nil)
	rr := httptest.NewRecorder()

	NewHandler(logger, service).Bootstrap(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "local-e2e-user")
	service.AssertExpectations(t)
}

func TestHandler_SeedConversation_ValidatesDraft(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &mockBootstrapService{}

	req := httptest.NewRequest(
		http.MethodPost,
		"/dev/e2e/ai-chat/conversations",
		bytes.NewBufferString(`{"latest_workout_draft":{"date":"","exercises":[]}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	NewHandler(logger, service).SeedConversation(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "latest_workout_draft.date is required")
	service.AssertExpectations(t)
}
