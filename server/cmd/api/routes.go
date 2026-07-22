package main

import (
	"net/http"
	"path"
	"strings"

	"github.com/Andrewy-gh/fittrack/server/internal/account"
	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	"github.com/Andrewy-gh/fittrack/server/internal/e2eauth"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/middleware"
	"github.com/Andrewy-gh/fittrack/server/internal/trainingprofile"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (api *api) routes(wh *workout.WorkoutHandler, eh *exercise.ExerciseHandler, fh *featureaccess.Handler, hh *health.Handler, ah *aichat.Handler, bh *billing.Handler, tph *trainingprofile.Handler, accountHandler *account.Handler, e2eh *e2eauth.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	// Health endpoints (no authentication required)
	mux.HandleFunc("GET /health", hh.Health)
	mux.HandleFunc("GET /ready", hh.Ready)
	if bh != nil {
		mux.HandleFunc("POST /stripe/webhook", bh.Webhook)
	}

	// Wrap with basic auth if credentials are configured
	protectedMetrics := middleware.BasicAuth(api.cfg.MetricsUsername, api.cfg.MetricsPassword, api.logger)(api.metricsHandler())
	mux.Handle("GET /metrics", protectedMetrics)

	// API endpoints (authentication required)
	mux.HandleFunc("GET /api/workouts", wh.ListWorkouts)
	mux.HandleFunc("POST /api/workouts", wh.CreateWorkout)
	mux.HandleFunc("GET /api/workouts/{id}", wh.GetWorkoutWithSets)
	mux.HandleFunc("PUT /api/workouts/{id}", wh.UpdateWorkout)
	mux.HandleFunc("DELETE /api/workouts/{id}", wh.DeleteWorkout)
	mux.HandleFunc("GET /api/workouts/new-workout-context", wh.GetNewWorkoutContext)
	mux.HandleFunc("GET /api/workouts/focus-values", wh.ListWorkoutFocusValues)
	mux.HandleFunc("GET /api/workouts/contribution-data", wh.GetContributionData)
	mux.HandleFunc("GET /api/exercises", eh.ListExercises)
	mux.HandleFunc("GET /api/features/access", fh.ListActiveFeatureAccess)
	if tph != nil {
		mux.HandleFunc("GET /api/training-profile", tph.Get)
		mux.HandleFunc("PUT /api/training-profile", tph.Upsert)
	}
	if accountHandler != nil {
		mux.HandleFunc("DELETE /api/account", accountHandler.DeleteAccount)
	}
	if bh != nil {
		mux.HandleFunc("POST /api/billing/checkout-session", bh.CreateCheckoutSession)
		mux.HandleFunc("POST /api/billing/customer-portal-session", bh.CreateCustomerPortalSession)
		mux.HandleFunc("POST /api/billing/subscription-cancel-portal-session", bh.CreateSubscriptionCancelPortalSession)
		mux.HandleFunc("GET /api/billing/status", bh.CurrentStatus)
	}
	mux.HandleFunc("POST /api/exercises", eh.GetOrCreateExercise)
	mux.HandleFunc("GET /api/exercises/{id}", eh.GetExerciseWithSets)
	mux.HandleFunc("GET /api/exercises/{id}/recent-sets", eh.GetRecentSetsForExercise)
	mux.HandleFunc("GET /api/exercises/{id}/metrics-history", eh.GetExerciseMetricsHistory)
	mux.HandleFunc("PATCH /api/exercises/{id}", eh.UpdateExerciseName)
	mux.HandleFunc("PATCH /api/exercises/{id}/historical-1rm", eh.UpdateExerciseHistorical1RM)
	mux.HandleFunc("DELETE /api/exercises/{id}", eh.DeleteExercise)
	mux.HandleFunc("POST /api/ai/conversations", ah.CreateConversation)
	mux.HandleFunc("GET /api/ai/conversations", ah.ListConversations)
	mux.HandleFunc("DELETE /api/ai/conversations", ah.DeleteAllConversations)
	mux.HandleFunc("GET /api/ai/conversations/{id}", ah.GetConversation)
	mux.HandleFunc("DELETE /api/ai/conversations/{id}", ah.DeleteConversation)
	mux.HandleFunc("POST /api/ai/conversations/{id}/latest-workout-draft/save", ah.SaveLatestWorkoutDraft)
	mux.HandleFunc("POST /api/ai/conversations/{id}/messages/stream", ah.StreamMessage)
	mux.HandleFunc("GET /api/ai/conversations/{id}/messages/stream/resume", ah.ResumeMessageStream)
	mux.HandleFunc("POST /api/ai/conversations/{id}/messages/recover", ah.RecoverMessage)
	mux.HandleFunc("POST /api/ai/conversations/{id}/runs/{runID}/stop", ah.StopRun)
	mux.HandleFunc("POST /api/ai/chat/telemetry", ah.RecordTelemetry)
	if api.inngestHandler != nil {
		mux.Handle("GET /inngest", api.inngestHandler)
		mux.Handle("PUT /inngest", api.inngestHandler)
		mux.Handle("POST /inngest", api.inngestHandler)
	}
	if e2eh != nil {
		mux.HandleFunc("POST /dev/e2e/auth/bootstrap", e2eh.Bootstrap)
		mux.HandleFunc("POST /dev/e2e/ai-chat/conversations", e2eh.SeedConversation)
	}
	// Swagger documentation
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /", api.handleStaticFiles())

	return mux
}

func (api *api) metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.UpdateDatabaseMetrics(api.pool)
		promhttp.Handler().ServeHTTP(w, r)
	})
}

func (api *api) handleStaticFiles() http.HandlerFunc {
	fs := http.FileServer(http.Dir("./dist"))

	return func(w http.ResponseWriter, r *http.Request) {
		file, err := http.Dir("./dist").Open(r.URL.Path)
		if err == nil {
			_ = file.Close()
			setStaticCacheHeader(w, r.URL.Path)
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
			setStaticCacheHeader(w, "/index.html")
			http.ServeFile(w, r, "./dist/index.html")
			return
		}

		http.NotFound(w, r)
	}
}

func setStaticCacheHeader(w http.ResponseWriter, requestPath string) {
	switch {
	case requestPath == "/" || requestPath == "/index.html" || requestPath == "/sw.js" || requestPath == "/manifest.webmanifest" || requestPath == "/manifest.json":
		w.Header().Set("Cache-Control", "no-cache")
	case strings.HasPrefix(requestPath, "/assets/"):
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	default:
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
}
