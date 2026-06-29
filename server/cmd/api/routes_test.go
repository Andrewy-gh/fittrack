package main

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

func TestRoutes_AllowsInngestHandlerAlongsideStaticFallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg:    &config.Config{},
		// nil pool is fine here because we only care about route registration.
		pool:           nil,
		inngestHandler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("route registration should not panic, got %v", r)
		}
	}()

	_ = api.routes(wh, eh, fh, hh, ah, nil, nil, nil)
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, nil)
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, nil)

	for _, path := range []string{"/api/ai/chat/validate", "/api/ai/chat/validate/stream"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"prompt":"prove the slice"}`))
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			t.Fatalf("%s should not expose a product prompt endpoint", path)
		}
	}
}

func TestRoutes_ProtectsPublicMetricsWhenCredentialsConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg: &config.Config{
			MetricsUsername: "metrics-user",
			MetricsPassword: "metrics-password",
		},
		pool: nil,
	}

	wh := &workout.WorkoutHandler{}
	eh := &exercise.ExerciseHandler{}
	fh := &featureaccess.Handler{}
	hh := health.NewHandler(logger, nil)
	ah := aichat.NewHandler(logger, nil)

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRoutes_RegistersBillingCustomerPortalSession(t *testing.T) {
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

	mux := api.routes(wh, eh, fh, hh, ah, bh, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/billing/customer-portal-session", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "https://billing.stripe.test/session") {
		t.Fatalf("expected billing portal URL response, got %s", rr.Body.String())
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

	mux := api.routes(wh, eh, fh, hh, ah, bh, nil, nil)
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, accountHandler, nil)
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
	writeStaticTestFile(t, filepath.Join("assets", "index-abc123.js"), "console.log('ok')")

	api := &api{}
	handler := api.handleStaticFiles()

	tests := []struct {
		name       string
		path       string
		accept     string
		wantStatus int
		wantCache  string
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
			name:       "hashed assets are immutable",
			path:       "/assets/index-abc123.js",
			wantStatus: http.StatusOK,
			wantCache:  "public, max-age=31536000, immutable",
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
		})
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
