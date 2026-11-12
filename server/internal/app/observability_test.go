package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/auth"
	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
	"github.com/Andrewy-gh/fittrack/server/internal/health"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
)

type observationTokens struct{}

func (observationTokens) GetUserIDFromToken(token string) (string, error) {
	if token == "private-token" {
		return "private-user", nil
	}
	if token == "failed-user-token" {
		return "failed-user", nil
	}
	return "", errors.New("invalid token")
}

type observationUsers struct{}

func (observationUsers) EnsureUser(ctx context.Context, id string) (db.Users, error) {
	if got, ok := user.Current(ctx); !ok || got != id {
		return db.Users{}, errors.New("identity not available before EnsureUser")
	}
	if id == "failed-user" {
		return db.Users{}, errors.New("private database details")
	}
	return db.Users{UserID: id}, nil
}

type observationWorkouts struct{ workout.WorkoutRepository }

func (observationWorkouts) GetWorkout(ctx context.Context, id int32, userID string) (db.Workout, error) {
	if got, ok := user.Current(ctx); !ok || got != userID || userID != "private-user" {
		return db.Workout{}, errors.New("wrong request identity")
	}
	return db.Workout{}, nil
}
func (observationWorkouts) GetWorkoutWithSets(context.Context, int32, string) ([]db.GetWorkoutWithSetsRow, error) {
	return nil, nil
}

func TestApplicationHTTPObservations(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	cfg := &config.Config{RateLimitRPM: 10000}
	api := &api{logger: logger, cfg: cfg}
	router := api.routes(workout.NewHandler(logger, nil, workout.NewService(logger, observationWorkouts{})), &exercise.ExerciseHandler{}, &featureaccess.Handler{}, health.NewHandler(logger, nil), aichat.NewHandler(logger, nil), billing.NewHandler(logger, routeBillingService{}), nil, nil, nil)
	handler := newAPIHandler(cfg, logger, router, auth.NewAuthenticator(logger, observationTokens{}, observationUsers{}))
	cases := []struct {
		path, token, route string
		status             int
		present            bool
	}{
		{"/api/workouts/1900123001?prompt=private-prompt", "private-token", "/api/workouts/{id}", 200, true},
		{"/api/workouts/1900123002", "private-token", "/api/workouts/{id}", 200, true},
		{"/api/workouts/private-invalid-id", "private-token", "/api/workouts/{id}", 400, true},
		{"/api/billing/status", "private-token", "/api/billing/status", 200, true},
		{"/api/workouts/private-id-three", "", "/api/workouts/{id}", 401, false},
		{"/api/workouts/private-id-four", "bad-token", "/api/workouts/{id}", 401, false},
		{"/api/workouts/private-id-five", "failed-user-token", "/api/workouts/{id}", 500, true},
		{"/api/missing-private-id", "private-token", "/", 404, true},
		// Neither missing routes nor canonicalization may bypass authentication.
		{"/api/missing-private-id", "", "/", 401, false},
		{"/api/workouts//private-redirect-id", "", "/api/workouts/{id}", 401, false},
		{"/.well-known/api-catalog", "", "/.well-known/api-catalog", 200, false},
	}
	scrapeCounts := func() map[string]float64 {
		rr := httptest.NewRecorder()
		api.metricsHandler().ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
		counts := map[string]float64{}
		for _, line := range strings.Split(rr.Body.String(), "\n") {
			if !strings.HasPrefix(line, "http_requests_total{") {
				continue
			}
			label, value, ok := strings.Cut(line, " ")
			if !ok {
				t.Fatalf("invalid metric %q", line)
			}
			count, err := strconv.ParseFloat(value, 64)
			if err != nil {
				t.Fatal(err)
			}
			counts[label] = count
		}
		return counts
	}
	before := scrapeCounts()
	// Concurrent requests share the application but must not share identity or route state.
	var wg sync.WaitGroup
	for i, tc := range cases {
		wg.Add(1)
		go func(i int, tc struct {
			path, token, route string
			status             int
			present            bool
		}) {
			defer wg.Done()
			req := httptest.NewRequest("GET", tc.path, nil)
			req.Header.Set("X-Request-ID", fmt.Sprintf("observation-%d", i))
			req.Header.Set("x-stack-access-token", tc.token)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tc.status {
				t.Errorf("case %d status=%d body=%s", i, rr.Code, rr.Body.String())
			}
		}(i, tc)
	}
	wg.Wait()
	logs := map[string]map[string]any{}
	logOutput := output.String()
	scanner := bufio.NewScanner(strings.NewReader(logOutput))
	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatal(err)
		}
		if entry["msg"] == "request completed" {
			logs[entry["request_id"].(string)] = entry
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(logs) != len(cases) {
		t.Fatalf("completion logs=%d want %d", len(logs), len(cases))
	}
	for i, tc := range cases {
		entry := logs[fmt.Sprintf("observation-%d", i)]
		if entry["route"] != tc.route || entry["user_present"] != tc.present || entry["status"] != float64(tc.status) {
			t.Errorf("case %d observation=%v", i, entry)
		}
	}
	after := scrapeCounts()
	expected := map[string]float64{}
	for _, tc := range cases {
		label := fmt.Sprintf(`http_requests_total{method="GET",path="%s",status="%d"}`, tc.route, tc.status)
		expected[label]++
	}
	for label, want := range expected {
		if got := after[label] - before[label]; got != want {
			t.Errorf("metric %s delta=%v want %v", label, got, want)
		}
	}
	for _, secret := range []string{"1900123001", "1900123002", "private-invalid-id", "private-id", "private-user", "private-token", "private-prompt", "failed-user", "private-redirect-id", "bad-token", "private database details"} {
		if strings.Contains(logOutput, secret) {
			t.Errorf("logs leaked %q", secret)
		}
		if strings.Contains(fmt.Sprint(after), secret) {
			t.Errorf("metrics leaked %q", secret)
		}
	}
}
