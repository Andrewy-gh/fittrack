package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/models"
	"github.com/Andrewy-gh/fittrack/server/internal/validation"
)

// ListWorkouts handles GET /api/workouts
func (h *Handler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	workouts, err := h.workoutService.ListWorkouts(r.Context())
	if err != nil {
		log.Printf("Error listing workouts: %v", err)
		http.Error(w, "Failed to retrieve workouts", http.StatusInternalServerError)
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

func (h *Handler) GetWorkoutWithSets(w http.ResponseWriter, r *http.Request) {
	workoutID := r.PathValue("id")
	if workoutID == "" {
		http.Error(w, "Missing workout ID", http.StatusBadRequest)
		return
	}

	workoutIDInt, err := validation.ValidateWorkoutID(workoutID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	workout, err := h.workoutService.GetWorkoutWithSets(r.Context(), workoutIDInt)
	if err != nil {
		log.Printf("Error getting workout with sets: %v", err)
		http.Error(w, "Failed to retrieve workout with sets", http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(workout)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// CreateWorkout handles POST /api/workouts
func (h *Handler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	// Read the raw JSON body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	log.Printf("Received workout JSON: %s\n", string(body))

	// Parse JSON into our request structure
	var workoutReq models.WorkoutRequest
	if err := json.Unmarshal(body, &workoutReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the input
	if err := validation.ValidateWorkoutRequest(workoutReq); err != nil {
		http.Error(w, validation.FormatValidationErrors(err), http.StatusBadRequest)
		return
	}

	// Save to database
	response, err := h.workoutService.CreateWorkout(r.Context(), workoutReq)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Failed to save workout", http.StatusInternalServerError)
		return
	}

	// Convert response to JSON
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	log.Printf("Saved workout with ID: %d", response.WorkoutID)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseJSON)
}
