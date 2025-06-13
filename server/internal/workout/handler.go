package workout

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type WorkoutHandler struct {
	workoutService *WorkoutService
}

func NewHandler(workoutService *WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		workoutService: workoutService,
	}
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListWorkouts(r.Context())
	if err != nil {
		h.workoutService.logger.Error("failed to list workouts", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(workouts)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func (h *WorkoutHandler) GetWorkoutWithSets(w http.ResponseWriter, r *http.Request) {
	workoutID := r.PathValue("id")
	if workoutID == "" {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutIDInt, err := strconv.ParseInt(workoutID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid workout ID", http.StatusBadRequest)
		return
	}

	workoutWithSets, err := h.workoutService.GetWorkoutWithSets(r.Context(), int32(workoutIDInt))
	if err != nil {
		h.workoutService.logger.Error("failed to get workout with sets", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(workoutWithSets)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
