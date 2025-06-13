package exercise

import (
	"encoding/json"
	"net/http"
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
