package main

import (
	"net/http"
	"path"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/middleware"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (api *api) routes(wh *workout.WorkoutHandler, eh *exercise.ExerciseHandler, hh *health.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoints (no authentication required)
	mux.HandleFunc("GET /health", hh.Health)
	mux.HandleFunc("GET /ready", hh.Ready)

	// Metrics endpoint (basic auth if configured)
	metricsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Update database metrics before serving
		middleware.UpdateDatabaseMetrics(api.pool)
		promhttp.Handler().ServeHTTP(w, r)
	})

	// Wrap with basic auth if credentials are configured
	protectedMetrics := middleware.BasicAuth(api.cfg.MetricsUsername, api.cfg.MetricsPassword, api.logger)(metricsHandler)
	mux.Handle("GET /metrics", protectedMetrics)

	// API endpoints (authentication required)
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	mux.HandleFunc("POST /api/workouts", wh.CreateWorkout)
	mux.HandleFunc("GET /api/workouts/{id}", wh.GetWorkoutWithSets)
	mux.HandleFunc("PUT /api/workouts/{id}", wh.UpdateWorkout)
	mux.HandleFunc("DELETE /api/workouts/{id}", wh.DeleteWorkout)
	mux.HandleFunc("GET /api/workouts/focus-values", wh.ListWorkoutFocusValues)
	mux.HandleFunc("GET /api/workouts/contribution-data", wh.GetContributionData)
	mux.HandleFunc("GET /api/exercises", eh.ListExercises)
	mux.HandleFunc("POST /api/exercises", eh.GetOrCreateExercise)
	mux.HandleFunc("GET /api/exercises/{id}", eh.GetExerciseWithSets)
	mux.HandleFunc("GET /api/exercises/{id}/recent-sets", eh.GetRecentSetsForExercise)
	mux.HandleFunc("GET /api/exercises/{id}/metrics-history", eh.GetExerciseMetricsHistory)
	mux.HandleFunc("PATCH /api/exercises/{id}", eh.UpdateExerciseName)
	mux.HandleFunc("PATCH /api/exercises/{id}/historical-1rm", eh.UpdateExerciseHistorical1RM)
	mux.HandleFunc("DELETE /api/exercises/{id}", eh.DeleteExercise)
	// Swagger documentation
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /", api.handleStaticFiles())

	return mux
}

func (api *api) handleStaticFiles() http.HandlerFunc {
	fs := http.FileServer(http.Dir("./dist"))

	return func(w http.ResponseWriter, r *http.Request) {
		_, err := http.Dir("./dist").Open(r.URL.Path)
		if err == nil {
			fs.ServeHTTP(w, r)
			return
		}

		// For missing static assets (e.g. old hashed JS chunks), return 404.
		// Serving index.html for asset paths causes MIME errors like:
		// "text/html is not a valid JavaScript MIME type".
		if path.Ext(r.URL.Path) != "" {
			http.NotFound(w, r)
			return
		}

		// SPA fallback only for browser navigation requests.
		accept := r.Header.Get("Accept")
		if r.Method == http.MethodGet && strings.Contains(accept, "text/html") {
			http.ServeFile(w, r, "./dist/index.html")
			return
		}

		http.NotFound(w, r)
	}
}
