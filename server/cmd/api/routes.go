package main

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func (api *api) routes(wh *workout.WorkoutHandler, eh *exercise.ExerciseHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	mux.HandleFunc("GET /api/workouts/{id}", wh.GetWorkoutWithSets)
	mux.HandleFunc("GET /api/exercises", eh.ListExercises)
	mux.HandleFunc("GET /api/exercises/{id}", eh.GetExercise)
	return mux
}
