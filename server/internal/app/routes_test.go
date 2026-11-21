package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/account"
	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type routeBillingService struct{}

type routeAccountService struct{}

func (routeBillingService) CreateCheckoutSession(context.Context) (*billing.CheckoutSessionResponse, error) {
	return &billing.CheckoutSessionResponse{URL: "https://checkout.stripe.test/session"}, nil
}

func (routeBillingService) CreateCustomerPortalSession(context.Context, billing.PortalReturnDestination) (*billing.CustomerPortalSessionResponse, error) {
	return &billing.CustomerPortalSessionResponse{URL: "https://billing.stripe.test/session"}, nil
}

func (routeBillingService) CreateSubscriptionCancelPortalSession(context.Context) (*billing.CustomerPortalSessionResponse, error) {
	return &billing.CustomerPortalSessionResponse{URL: "https://billing.stripe.test/cancel-session"}, nil
}

func (routeBillingService) CurrentStatus(context.Context) (*billing.StatusResponse, error) {
	return &billing.StatusResponse{FeatureKey: billing.FeatureKeyAIChatbot}, nil
}

func (routeBillingService) HandleWebhook(context.Context, []byte, string) error {
	return nil
}

func (routeAccountService) DeleteCurrentUser(context.Context) error {
	return nil
}

func TestRoutes_RegistersPutForInngestHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	called := false
	api := &api{
		logger: logger,
		cfg:    &config.Config{},
		pool:   nil,
		inngestHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			if r.Method != http.MethodPut {
				t.Fatalf("expected PUT request, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}),
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPut, "/inngest", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected inngest handler to be called for PUT /inngest")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestRoutes_DoesNotExposeAIChatValidationEndpoints(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg:    &config.Config{},
		pool:   nil,
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, nil, nil)

	for _, path := range []string{"/api/ai/chat/validate", "/api/ai/chat/validate/stream"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"prompt":"prove the slice"}`))
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			t.Fatalf("%s should not expose a product prompt endpoint", path)
		}
	}
}

func TestRoutes_PublicMetricsAuthentication(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cases := []struct {
		name, username, password, requestUser, requestPassword string
		auth                                                   bool
		status                                                 int
	}{
		{"complete missing auth", "user", "password", "", "", false, http.StatusUnauthorized},
		{"complete wrong user", "user", "password", "wrong", "password", true, http.StatusUnauthorized},
		{"complete wrong password", "user", "password", "user", "wrong", true, http.StatusUnauthorized},
		{"complete correct auth", "user", "password", "user", "password", true, http.StatusOK},
		{"missing both anonymous", "", "", "", "", false, http.StatusNotFound},
		{"missing username anonymous", "", "password", "", "", false, http.StatusNotFound},
		{"missing password anonymous", "user", "", "", "", false, http.StatusNotFound},
		{"missing both", "", "", "user", "password", true, http.StatusNotFound},
		{"missing username", "", "password", "user", "password", true, http.StatusNotFound},
		{"missing password", "user", "", "user", "password", true, http.StatusNotFound},
	}
	for _, tc := range cases {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			t.Run(tc.name+"/"+method, func(t *testing.T) {
				cfg := &config.Config{MetricsUsername: tc.username, MetricsPassword: tc.password}
				api := &api{logger: logger, cfg: cfg}
				mux := api.routes(&workout.WorkoutHandler{}, &exercise.ExerciseHandler{}, &featureaccess.Handler{}, health.NewHandler(logger, nil), aichat.NewHandler(logger, nil), nil, nil, nil, nil)
				req := httptest.NewRequest(method, "/metrics", nil)
				req.Header.Set("Accept", "text/html")
				if tc.auth {
					req.SetBasicAuth(tc.requestUser, tc.requestPassword)
				}
				rr := httptest.NewRecorder()
				mux.ServeHTTP(rr, req)
				if rr.Code != tc.status {
					t.Fatalf("status = %d, want %d", rr.Code, tc.status)
				}
				if tc.status == http.StatusUnauthorized && rr.Header().Get("WWW-Authenticate") != `Basic realm="Metrics"` {
					t.Fatalf("missing Basic challenge: %v", rr.Header())
				}
				if tc.status == http.StatusOK {
					if !strings.Contains(rr.Header().Get("Content-Type"), "text/plain") {
						t.Fatalf("unexpected content type: %v", rr.Header())
					}
					if method == http.MethodGet && !strings.Contains(rr.Body.String(), "go_goroutines") {
						t.Fatal("missing Prometheus exposition")
					}
				} else if strings.Contains(rr.Body.String(), "go_goroutines") || strings.Contains(rr.Header().Get("Content-Type"), "text/html") || strings.Contains(strings.ToLower(rr.Body.String()), "<html") {
					t.Fatalf("denial exposed metrics or HTML: %q", rr.Body.String())
				}
			})
		}
	}
}

func TestMetricsServer_InternalScrapeIgnoresPublicCredentials(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, cfg := range []*config.Config{
		{MetricsUsername: "user", MetricsPassword: "password"},
		{MetricsUsername: "user"},
		{MetricsPassword: "password"},
		{},
	} {
		api := &api{logger: logger, cfg: cfg}
		server := newMetricsServer(cfg, api.metricsHandler(), logger)
		rr := httptest.NewRecorder()
		server.Handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Header().Get("Content-Type"), "text/plain") || !strings.Contains(rr.Body.String(), "go_goroutines") {
			t.Fatalf("internal scrape: status %d, content type %q, expected Prometheus exposition", rr.Code, rr.Header().Get("Content-Type"))
		}
	}
}

func TestRoutes_RegistersBillingSubscriptionCancelPortalSession(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg:    &config.Config{},
		pool:   nil,
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)
	bh := billing.NewHandler(logger, routeBillingService{})

	mux := api.routes(wh, eh, fh, hh, ah, bh, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/billing/subscription-cancel-portal-session", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "https://billing.stripe.test/cancel-session") {
		t.Fatalf("expected billing cancel portal URL response, got %s", rr.Body.String())
	}
}

func TestRoutes_RegistersAccountDeletion(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg:    &config.Config{},
		pool:   nil,
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)
	accountHandler := account.NewHandler(logger, routeAccountService{})

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, accountHandler, nil)
	req := httptest.NewRequest(http.MethodDelete, "/api/account", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNoContent, rr.Code, rr.Body.String())
	}
}

func TestStaticFiles_SetCacheHeadersForPWAUpdates(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("switch to temporary static file directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	writeStaticTestFile(t, "index.html", "<!doctype html>")
	writeStaticTestFile(t, "sw.js", "self.skipWaiting()")
	writeStaticTestFile(t, "manifest.webmanifest", "{}")
	writeStaticTestFile(t, "robots.txt", "User-agent: *\nDisallow:\n")
	writeStaticTestFile(t, "sitemap.xml", "<urlset></urlset>")
	writeStaticTestFile(t, "llms.txt", "# FitTrack\n")
	writeStaticTestFile(t, filepath.Join("assets", "index-abc123.js"), "console.log('ok')")
	writeStaticTestFile(t, filepath.Join("guide", "index.html"), "guide index")

	api := &api{}
	handler := api.handleStaticFiles()

	tests := []struct {
		name       string
		path       string
		accept     string
		wantStatus int
		wantCache  string
		wantBody   string
	}{
		{
			name:       "app shell fallback is revalidated",
			path:       "/workouts",
			accept:     "text/html",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache",
		},
		{
			name:       "service worker is revalidated",
			path:       "/sw.js",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache",
		},
		{
			name:       "manifest is revalidated",
			path:       "/manifest.webmanifest",
			wantStatus: http.StatusOK,
			wantCache:  "no-cache",
		},
		{
			name:       "robots policy is served as a static asset",
			path:       "/robots.txt",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=3600",
			wantBody:   "User-agent: *\nDisallow:\n",
		},
		{
			name:       "sitemap is served as a static asset",
			path:       "/sitemap.xml",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=3600",
			wantBody:   "<urlset></urlset>",
		},
		{
			name:       "llms document is served as a static asset",
			path:       "/llms.txt",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=3600",
			wantBody:   "# FitTrack\n",
		},
		{
			name:       "hashed assets are immutable",
			path:       "/assets/index-abc123.js",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=31536000, immutable",
		},
		{
			name:       "directories do not expose listings",
			path:       "/assets/",
			accept:     "text/html",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "directory without trailing slash does not redirect to a listing",
			path:       "/assets",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "directory index remains available",
			path:       "/guide/",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=3600",
			wantBody:   "guide index",
		},
		{
			name:       "missing old assets are still not rewritten to html",
			path:       "/assets/missing-old-build.js",
			wantStatus: http.StatusNotFound,
			wantCache:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d with body %s", tt.wantStatus, rr.Code, rr.Body.String())
			}
			if got := rr.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Fatalf("expected Cache-Control %q, got %q", tt.wantCache, got)
			}
			if tt.wantBody != "" && rr.Body.String() != tt.wantBody {
				t.Fatalf("expected body %q, got %q", tt.wantBody, rr.Body.String())
			}
		})
	}
}

func TestStaticFiles_RootWithoutIndexDoesNotList(t *testing.T) {
	withStaticTestDirectory(t)
	writeStaticTestFile(t, "asset.js", "asset")

	rr := httptest.NewRecorder()
	(&api{}).handleStaticFiles().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusNotFound || strings.Contains(rr.Body.String(), "asset.js") {
		t.Fatalf("root without index: status %d, body %q", rr.Code, rr.Body.String())
	}
}

func writeStaticTestFile(t *testing.T, relativePath string, contents string) {
	t.Helper()

	path := filepath.Join("dist", relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create static test directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write static test file: %v", err)
	}
}
