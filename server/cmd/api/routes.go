package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"path"
	"strconv"
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
	// Public API discovery and documentation
	mux.HandleFunc("GET /.well-known/api-catalog", api.handleAPICatalog)
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

type linkTarget struct {
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

type apiCatalogLinkset struct {
	Anchor      string       `json:"anchor"`
	ServiceDesc []linkTarget `json:"service-desc"`
	ServiceDoc  []linkTarget `json:"service-doc"`
	Status      []linkTarget `json:"status"`
}

type apiCatalogDocument struct {
	Linkset []apiCatalogLinkset `json:"linkset"`
}

func (api *api) handleAPICatalog(w http.ResponseWriter, r *http.Request) {
	setAPIDiscoveryLinks(w)
	w.Header().Set("Content-Type", `application/linkset+json; profile="https://www.rfc-editor.org/info/rfc9727"`)

	document := apiCatalogDocument{Linkset: []apiCatalogLinkset{{
		Anchor: "/api",
		ServiceDesc: []linkTarget{{
			Href: "/swagger/doc.json",
			Type: "application/json",
		}},
		ServiceDoc: []linkTarget{{
			Href: "/swagger/",
			Type: "text/html",
		}},
		Status: []linkTarget{{
			Href: "/health",
			Type: "application/json",
		}},
	}}}

	if err := json.NewEncoder(w).Encode(document); err != nil {
		api.logger.Error("failed to encode API catalog", "error", err)
	}
}

func (api *api) handleStaticFiles() http.HandlerFunc {
	fs := http.FileServer(http.Dir("./dist"))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			setAPIDiscoveryLinks(w)
			w.Header().Add("Vary", "Accept")
			if prefersMarkdown(r.Header.Get("Accept")) {
				setStaticCacheHeader(w, "/")
				w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
				_, _ = w.Write([]byte(fitTrackMarkdownOverview))
				return
			}
		}

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

const fitTrackMarkdownOverview = `# FitTrack

FitTrack is a fitness tracking application for recording workouts, exercises, sets, and training progress.

## Public API resources

- [API description](/swagger/doc.json)
- [Interactive API documentation](/swagger/)
- [Service health](/health)
- [API catalog](/.well-known/api-catalog)

Product data under /api requires authentication; this public overview does not expose user data.
`

func setAPIDiscoveryLinks(w http.ResponseWriter) {
	w.Header().Add("Link", `</.well-known/api-catalog>; rel="api-catalog"; type="application/linkset+json"`)
	w.Header().Add("Link", `</swagger/doc.json>; rel="service-desc"; type="application/json"`)
	w.Header().Add("Link", `</swagger/>; rel="service-doc"; type="text/html"`)
}

func prefersMarkdown(accept string) bool {
	markdownQuality, markdownExplicit := representationQuality(accept, "text/markdown")
	if !markdownExplicit || markdownQuality <= 0 {
		return false
	}

	htmlQuality, _ := representationQuality(accept, "text/html")
	return markdownQuality >= htmlQuality
}

func representationQuality(accept, representation string) (float64, bool) {
	quality := 0.0
	bestSpecificity := -1
	explicit := false
	representationParts := strings.SplitN(representation, "/", 2)

	for _, item := range strings.Split(accept, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(item))
		if err != nil {
			continue
		}
		mediaParts := strings.SplitN(mediaType, "/", 2)
		if len(mediaParts) != 2 || (mediaParts[0] != "*" && mediaParts[0] != representationParts[0]) || (mediaParts[1] != "*" && mediaParts[1] != representationParts[1]) {
			continue
		}

		specificity := 0
		if mediaParts[0] != "*" {
			specificity++
		}
		if mediaParts[1] != "*" {
			specificity++
		}
		if mediaType == representation {
			explicit = true
		}
		if specificity < bestSpecificity {
			continue
		}

		itemQuality := 1.0
		if rawQuality, ok := params["q"]; ok {
			parsedQuality, err := strconv.ParseFloat(rawQuality, 64)
			if err != nil || parsedQuality < 0 || parsedQuality > 1 {
				continue
			}
			itemQuality = parsedQuality
		}
		bestSpecificity = specificity
		quality = itemQuality
	}

	return quality, explicit
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
