package main

import (
	"io"
	"log/slog"
	"net/http"
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
