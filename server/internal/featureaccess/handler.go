package featureaccess

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type listService interface {
	ListCurrentUserAccess(ctx context.Context) ([]FeatureAccessGrant, error)
}

type Handler struct {
	logger  *slog.Logger
	service listService
}

func NewHandler(logger *slog.Logger, service listService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// ListActiveFeatureAccess godoc
// @Summary List active feature access grants
// @Description Get all active feature access grants for the authenticated user
// @Tags feature-access
// @Accept json
// @Produce json
// @Security StackAuth
// @Success 200 {array} featureaccess.FeatureAccessResponse
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /features/access [get]
func (h *Handler) ListActiveFeatureAccess(w http.ResponseWriter, r *http.Request) {
	grants, err := h.service.ListCurrentUserAccess(r.Context())
	if err != nil {
		var errUnauthorized *apperrors.Unauthorized
		if errors.As(err, &errUnauthorized) {
			response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
		} else {
			response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to list feature access", err)
		}
		return
	}

	if grants == nil {
		grants = []FeatureAccessGrant{}
	}

	if err := response.JSON(w, http.StatusOK, grants); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}
