package main

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func (api *api) routes(wh *workout.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	return mux
}
