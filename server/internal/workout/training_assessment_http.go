package workout

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

type AssessmentHandler struct {
	service *AssessmentService
	logger  *slog.Logger
}

func NewAssessmentHandler(service *AssessmentService, logger *slog.Logger) *AssessmentHandler {
	return &AssessmentHandler{service: service, logger: logger}
}

// Get godoc
// @Summary Compare saved exercise sessions with a general strength set reference
// @Tags exercises
// @Produce json
// @Security StackAuth
// @Param id path int true "Exercise ID"
// @Param startDate query string true "First local date, YYYY-MM-DD"
// @Param timezone query string true "IANA timezone"
// @Success 200 {object} workout.StrengthAssessment
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /exercises/{id}/assessment [get]
func (h *AssessmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || id <= 0 {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "Invalid exercise ID", nil)
		return
	}
	result, err := h.service.Assess(r.Context(), int32(id), TrainingEvidenceRequest{StartDate: r.URL.Query().Get("startDate"), Timezone: r.URL.Query().Get("timezone")})
	if err != nil {
		status, message := http.StatusInternalServerError, "Could not load training assessment"
		var unauthorized *apperrors.Unauthorized
		var notFound *apperrors.NotFound
		switch {
		case errors.As(err, &unauthorized):
			status, message = http.StatusUnauthorized, "Unauthorized"
		case errors.As(err, &notFound):
			status, message = http.StatusNotFound, "Exercise not found"
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
