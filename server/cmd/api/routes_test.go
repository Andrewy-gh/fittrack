package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

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

	_ = api.routes(wh, eh, fh, hh, ah)
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

	mux := api.routes(wh, eh, fh, hh, ah)
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
