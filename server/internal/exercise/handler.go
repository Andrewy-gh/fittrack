package exercise

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

// ExerciseHandler handles exercise HTTP requests
type ExerciseHandler struct {
	logger          *slog.Logger
	exerciseService *ExerciseService
}

func NewHandler(logger *slog.Logger, exerciseService *ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		logger:          logger,
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

func (eh *ExerciseHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, "Missing exercise ID", nil)
		return
	}

	exerciseIDInt, err := strconv.ParseInt(exerciseID, 10, 32)
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusBadRequest, "Invalid exercise ID", err)
		return
	}

	exercise, err := eh.exerciseService.GetExercise(r.Context(), int32(exerciseIDInt))
	if err != nil {
		response.ErrorJSON(w, r, eh.logger, http.StatusInternalServerError, "Failed to get exercise", err)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercise); err != nil {
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
