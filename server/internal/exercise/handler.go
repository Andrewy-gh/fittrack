package exercise

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
	"github.com/go-playground/validator/v10"
)

// ExerciseHandler handles exercise HTTP requests
type ExerciseHandler struct {
	logger          *slog.Logger
	validator       *validator.Validate
	exerciseService *ExerciseService
}

func NewHandler(logger *slog.Logger, validator *validator.Validate, exerciseService *ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		logger:          logger,
		validator:       validator,
		exerciseService: exerciseService,
	}
}

func (eh *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := eh.exerciseService.ListExercises(r.Context())
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to list exercises", err)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercises); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

func (eh *ExerciseHandler) GetExerciseWithSets(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	exerciseIDInt, err := request.ValidatePathID(exerciseID, "exercise ID")
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, err.Error(), nil)
		return
	}

	req := GetExerciseWithSetsRequest{
		ExerciseID: int32(exerciseIDInt),
	}

	if err := eh.validator.Struct(req); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, "Invalid exercise ID: must be positive", err)
		return
	}

	sets, err := eh.exerciseService.GetExerciseWithSets(r.Context(), req.ExerciseID)
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to get exercise with sets", err)
		return
	}

	if len(sets) == 0 {
		response.ErrorJSON(w, r, eh.logger, http.StatusNotFound, "No sets found for this exercise", nil)
		return
	}

	if err := response.JSON(w, http.StatusOK, sets); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}

func (eh *ExerciseHandler) GetOrCreateExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, "Failed to decode request body", err)
		return
	}

	if err := eh.validator.Struct(&req); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, "Validation failed", err)
		return
	}

	exercise, err := eh.exerciseService.GetOrCreateExercise(r.Context(), req.Name)
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to get or create exercise", err)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercise); err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to write response", err)
		return
	}
}
