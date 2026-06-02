package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type routeBillingService struct{}

func (routeBillingService) CreateCheckoutSession(context.Context) (*billing.CheckoutSessionResponse, error) {
	return &billing.CheckoutSessionResponse{URL: "https://checkout.stripe.test/session"}, nil
}

func (routeBillingService) CreateCustomerPortalSession(context.Context) (*billing.CustomerPortalSessionResponse, error) {
	return &billing.CustomerPortalSessionResponse{URL: "https://billing.stripe.test/session"}, nil
}

func (routeBillingService) CurrentStatus(context.Context) (*billing.StatusResponse, error) {
	return &billing.StatusResponse{FeatureKey: billing.FeatureKeyAIChatbot}, nil
}

func (routeBillingService) HandleWebhook(context.Context, []byte, string) error {
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

	_ = api.routes(wh, eh, fh, hh, ah, nil, nil)
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil)
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil)

	for _, path := range []string{"/api/ai/chat/validate", "/api/ai/chat/validate/stream"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"prompt":"prove the slice"}`))
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code == http.StatusOK {
			t.Fatalf("%s should not expose a product prompt endpoint", path)
		}
	}
}

func TestRoutes_PublicMetricsRequiresBasicAuthWhenConfigured(t *testing.T) {
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

	mux := api.routes(wh, eh, fh, hh, ah, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected public /metrics to require auth, got %d", rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("metrics-user", "metrics-password")
	rr = httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected authorized public /metrics request to succeed, got %d", rr.Code)
	}
}

func TestMetricsHandler_AllowsInternalScrapeWithoutBasicAuth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := &api{
		logger: logger,
		cfg: &config.Config{
			MetricsUsername: "metrics-user",
			MetricsPassword: "metrics-password",
		},
		pool: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	api.metricsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected internal metrics handler to allow scrape without auth, got %d", rr.Code)
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

	mux := api.routes(wh, eh, fh, hh, ah, bh, nil)
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
