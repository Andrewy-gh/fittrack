package app

import (
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

// Exercise startup configuration and real HTTP composition, including responses
// rejected by authentication before routing. The fixtures perform no DB/JWKS IO.
func TestApplicationSecurityHeaders(t *testing.T) {
	for _, environment := range []string{"production", "development", "staging"} {
		for _, directTLS := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/TLS=%t", environment, directTLS), func(t *testing.T) {
				t.Setenv("ENVIRONMENT", environment)
				t.Setenv("DATABASE_URL", "postgres://unused:unused@localhost/unused")
				t.Setenv("PROJECT_ID", "unused")
				t.Setenv("RLS_ENFORCEMENT_REQUIRED", "")
				cfg, err := config.Load()
				if err != nil {
					t.Fatal(err)
				}
				cfg.RateLimitRPM = 10000
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				api := &api{logger: logger, cfg: cfg}
				router := api.routes(&workout.WorkoutHandler{}, &exercise.ExerciseHandler{}, &featureaccess.Handler{}, health.NewHandler(logger, nil), aichat.NewHandler(logger, nil), nil, nil, nil, nil)
				handler := newAPIHandler(cfg, logger, router, auth.NewAuthenticator(logger, observationTokens{}, observationUsers{}))
				for _, forwarded := range []struct{ name, proto, standard string }{
					{"absent", "", ""},
					{"spoofed_https", "https", "for=192.0.2.1;proto=https"},
					{"spoofed_http", "http", "for=192.0.2.1;proto=http"},
					{"malformed", "https, http", "proto=\"https;proto=http"},
				} {
					t.Run(forwarded.name, func(t *testing.T) {
						for _, route := range []struct {
							path, token string
							status      int
						}{
							{"/.well-known/api-catalog", "", http.StatusOK},
							{"/swagger/index.html", "", http.StatusOK},
							{"/api/workouts", "", http.StatusUnauthorized},
							{"/api/workouts", "bad-token", http.StatusUnauthorized},
							{"/api/workouts", "failed-user-token", http.StatusInternalServerError},
							{"/assets/missing-hsts-test.js", "", http.StatusNotFound},
						} {
							req := httptest.NewRequest(http.MethodGet, "http://localhost"+route.path, nil)
							if directTLS {
								req.TLS = &tls.ConnectionState{}
							}
							req.Header.Set("X-Forwarded-Proto", forwarded.proto)
							req.Header.Set("Forwarded", forwarded.standard)
							req.Header.Set("x-stack-access-token", route.token)
							rr := httptest.NewRecorder()
							handler.ServeHTTP(rr, req)
							if rr.Code != route.status {
								t.Fatalf("%s: status=%d, want %d", route.path, rr.Code, route.status)
							}
							wantHSTS := ""
							if environment == "production" || directTLS {
								wantHSTS = "max-age=31536000; includeSubDomains"
							}
							frame := "DENY"
							if strings.HasPrefix(route.path, "/swagger/") {
								frame = "SAMEORIGIN"
							}
							for name, want := range map[string]string{
								"Strict-Transport-Security": wantHSTS,
								"X-Content-Type-Options":    "nosniff",
								"X-Frame-Options":           frame,
								"X-XSS-Protection":          "1; mode=block",
							} {
								if got := rr.Result().Header.Get(name); got != want {
									t.Errorf("%s: %s=%q, want %q", route.path, name, got, want)
								}
							}
						}
					})
				}
			})
		}
	}
}
