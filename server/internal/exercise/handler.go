package exercise

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
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

	responseJSON, err := json.Marshal(exercises)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
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

	responseJSON, err := json.Marshal(exercise)
	if err != nil {
		eh.logger.Error("failed to marshal response", "error", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
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

	responseJSON, err := json.Marshal(exercise)
	if err != nil {
		eh.logger.Error("failed to marshal response", "error", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
