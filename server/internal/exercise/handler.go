package exercise

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// ExerciseHandler handles exercise HTTP requests
type ExerciseHandler struct {
	exerciseService *ExerciseService
}

func NewHandler(exerciseService *ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		exerciseService: exerciseService,
	}
}

func (eh *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := eh.exerciseService.ListExercises(r.Context())
	if err != nil {
		eh.exerciseService.logger.Error("failed to list exercises", "error", err)
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
		http.Error(w, "Missing exercise ID", http.StatusBadRequest)
		return
	}

	exerciseIDInt, err := strconv.ParseInt(exerciseID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid exercise ID", http.StatusBadRequest)
		return
	}

	exercise, err := eh.exerciseService.GetExercise(r.Context(), int32(exerciseIDInt))
	if err != nil {
		eh.exerciseService.logger.Error("failed to get exercise", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(exercise)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
