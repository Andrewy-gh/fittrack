package account

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubService struct {
	err error
}

func (s stubService) DeleteCurrentUser(ctx context.Context) error {
	return s.err
}

func TestHandlerDeleteAccount_ReturnsNoContent(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubService{})
	req := httptest.NewRequest(http.MethodDelete, "/api/account", nil)
	rr := httptest.NewRecorder()

	handler.DeleteAccount(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Body.String())
}

func TestHandlerDeleteAccount_ServiceFailureReturnsInternalServerError(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubService{err: assert.AnError})
	req := httptest.NewRequest(http.MethodDelete, "/api/account", nil)
	rr := httptest.NewRecorder()

	handler.DeleteAccount(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed to delete account")
}
