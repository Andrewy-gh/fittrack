package main

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

func (api *api) routes(wh *workout.WorkoutHandler, eh *exercise.ExerciseHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	mux.HandleFunc("POST /api/workouts", wh.CreateWorkout)
	mux.HandleFunc("GET /api/workouts/{id}", wh.GetWorkoutWithSets)
	mux.HandleFunc("GET /api/exercises", auth.WithAuth(eh.ListExercises))
	mux.HandleFunc("POST /api/exercises", eh.GetOrCreateExercise)
	mux.HandleFunc("GET /api/exercises/{id}", eh.GetExerciseWithSets)
	mux.HandleFunc("GET /", api.handleStaticFiles())

	return mux
}

func (api *api) handleStaticFiles() http.HandlerFunc {
	fs := http.FileServer(http.Dir("./dist"))

	return func(w http.ResponseWriter, r *http.Request) {
		_, err := http.Dir("./dist").Open(r.URL.Path)
		if err != nil {
			http.ServeFile(w, r, "./dist/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	}
}
