package account

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type accountService interface {
	DeleteCurrentUser(ctx context.Context) error
}

type Handler struct {
	logger  *slog.Logger
	service accountService
}

func NewHandler(logger *slog.Logger, service accountService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// DeleteAccount godoc
// @Summary Delete the current user's application data
// @Description Cancels applicable billing subscriptions and removes FitTrack data; does not delete the authentication provider identity.
// @Tags account
// @Security StackAuth
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /account [delete]
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteCurrentUser(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var errUnauthorized *apperrors.Unauthorized

	switch {
	case errors.As(err, &errUnauthorized):
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
	default:
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to delete account", err)
	}
}
