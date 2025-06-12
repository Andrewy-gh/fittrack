package api

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/db"
	"github.com/Andrewy-gh/fittrack/server/internal/service"
)

// Handler holds dependencies for API handlers
type Handler struct {
	workoutService *service.WorkoutService
}

// NewHandler creates a new API handler with dependencies
func NewHandler(queries *db.Queries) *Handler {
	return &Handler{
		workoutService: service.NewWorkoutService(queries),
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /api/workouts", h.ListWorkouts)
	router.HandleFunc("POST /api/workouts", h.CreateWorkout)
	router.HandleFunc("GET /api/workouts/{id}", h.GetWorkoutWithSets)
	// Add other routes here as you build them
	// router.HandleFunc("GET /api/workouts/{id}", h.GetWorkout)
}
