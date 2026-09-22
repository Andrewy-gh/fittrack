package workout

import (
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type MuscleContributionHandler struct {
	service *MuscleContributionService
	logger  *slog.Logger
}

func NewMuscleContributionHandler(service *MuscleContributionService, logger *slog.Logger) *MuscleContributionHandler {
	return &MuscleContributionHandler{service: service, logger: logger}
}

// Get godoc
// @Summary Summarize reviewed weekly muscle contributions from logged working sets
// @Tags analytics
// @Produce json
// @Security StackAuth
// @Param startDate query string true "First local date, YYYY-MM-DD"
// @Param timezone query string true "IANA timezone"
// @Success 200 {object} workout.MuscleContributions
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /analytics/muscle-contributions [get]
func (h *MuscleContributionHandler) Get(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Summarize(r.Context(), TrainingEvidenceRequest{StartDate: r.URL.Query().Get("startDate"), Timezone: r.URL.Query().Get("timezone")})
	if err != nil {
		status, message := http.StatusInternalServerError, "Could not load training assessment"
		var unauthorized *apperrors.Unauthorized
		switch {
		case errors.As(err, &unauthorized):
			status, message = http.StatusUnauthorized, "Unauthorized"
		case errors.Is(err, ErrInvalidTrainingEvidencePeriod), errors.Is(err, ErrFutureTrainingEvidencePeriod):
			status, message = http.StatusBadRequest, "Invalid assessment period"
		}
		response.ErrorJSON(w, r, h.logger, status, message, err)
		return
	}
	if err := response.JSON(w, http.StatusOK, result); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Could not write assessment", err)
	}
}
