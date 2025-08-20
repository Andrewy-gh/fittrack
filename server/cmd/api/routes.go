package main

import (
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (api *api) routes(wh *workout.WorkoutHandler, eh *exercise.ExerciseHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	mux.HandleFunc("POST /api/workouts", wh.CreateWorkout)
	mux.HandleFunc("GET /api/workouts/{id}", wh.GetWorkoutWithSets)
	mux.HandleFunc("PUT /api/workouts/{id}", wh.UpdateWorkout)
	mux.HandleFunc("DELETE /api/workouts/{id}", wh.DeleteWorkout)
	mux.HandleFunc("GET /api/exercises", eh.ListExercises)
	mux.HandleFunc("POST /api/exercises", eh.GetOrCreateExercise)
	mux.HandleFunc("GET /api/exercises/{id}", eh.GetExerciseWithSets)
	mux.HandleFunc("GET /api/exercises/{id}/recent-sets", eh.GetRecentSetsForExercise)
	// Swagger documentation
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
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
