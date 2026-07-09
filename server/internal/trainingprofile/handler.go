package trainingprofile

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type profileService interface {
	Get(ctx context.Context) (*ProfileResponse, error)
	Upsert(ctx context.Context, req UpdateProfileRequest) (*ProfileResponse, error)
}

type Handler struct {
	logger  *slog.Logger
	service profileService
}

func NewHandler(logger *slog.Logger, service profileService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// Get godoc
// @Summary Get training profile
// @Description Returns the authenticated user's durable AI training profile. First-time users receive an empty profile shape.
// @Tags training-profile
// @Produce json
// @Security StackAuth
// @Success 200 {object} trainingprofile.ProfileResponse
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /training-profile [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.Get(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, "failed to get training profile")
		return
	}

	if err := response.JSON(w, http.StatusOK, profile); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

// Upsert godoc
// @Summary Update training profile
// @Description Replaces the authenticated user's durable AI training profile with the full submitted document.
// @Tags training-profile
// @Accept json
// @Produce json
// @Security StackAuth
// @Param request body trainingprofile.UpdateProfileRequest true "Training profile"
// @Success 200 {object} trainingprofile.ProfileResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /training-profile [put]
func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := decodeStrictJSON(w, r, &req); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to decode request body", err)
		return
	}

	profile, err := h.service.Upsert(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, r, err, "failed to update training profile")
		return
	}

	if err := response.JSON(w, http.StatusOK, profile); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	var errUnauthorized *apperrors.Unauthorized
	var errValidation *ValidationError

	switch {
	case errors.As(err, &errUnauthorized):
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
	case errors.As(err, &errValidation):
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, errValidation.Error(), errValidation)
	default:
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, fallback, err)
	}
}
