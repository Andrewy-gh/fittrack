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
		eh.logger.Error("failed to list exercises", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercises); err != nil {
		eh.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (eh *ExerciseHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")
	if exerciseID == "" {
		eh.logger.Error("missing exercise ID", "error", "Missing exercise ID")
		http.Error(w, "Missing exercise ID", http.StatusBadRequest)
		return
	}

	exerciseIDInt, err := strconv.ParseInt(exerciseID, 10, 32)
	if err != nil {
		eh.logger.Error("invalid exercise ID", "error", err)
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	exercise, err := eh.exerciseService.GetExercise(r.Context(), int32(exerciseIDInt))
	if err != nil {
		eh.logger.Error("failed to get exercise", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercise); err != nil {
		eh.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (eh *ExerciseHandler) GetOrCreateExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("failed to decode request body", "error", "Failed to decode request body")
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	exercise, err := eh.exerciseService.GetOrCreateExercise(r.Context(), req.Name)
	if err != nil {
		eh.logger.Error("failed to get or create exercise", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := response.JSON(w, http.StatusOK, exercise); err != nil {
		eh.logger.Error("failed to write response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
