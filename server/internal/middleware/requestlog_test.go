package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

func TestRequestLog_EmitsStructuredCompletionLog(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := RequestLog(logger)(nextHandler)
	req := httptest.NewRequest(http.MethodPost, "/api/workouts/123", nil)
	req.Pattern = "POST /api/workouts/{id}"
	ctx := request.WithRequestID(req.Context(), "req-123")
	ctx = user.WithContext(ctx, "user-123")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var logEntry map[string]any
	if err := json.Unmarshal(output.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to decode request log output: %v", err)
	}

	if got := logEntry["msg"]; got != "request completed" {
		t.Fatalf("expected request completed log message, got %v", got)
	}
	if got := logEntry["route"]; got != "/api/workouts/{id}" {
		t.Fatalf("expected normalized route, got %v", got)
	}
	if got := logEntry["request_id"]; got != "req-123" {
		t.Fatalf("expected request id req-123, got %v", got)
	}
	if got := logEntry["status"]; got != float64(http.StatusCreated) {
		t.Fatalf("expected status %d, got %v", http.StatusCreated, got)
	}
	if got := logEntry["user_present"]; got != true {
		t.Fatalf("expected user_present true, got %v", got)
	}
}

func TestRouteLabel_FallbackGroupsUnknownPaths(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "api path without pattern",
			req:  httptest.NewRequest(http.MethodGet, "/api/unknown/123", nil),
			want: "/api/*",
		},
		{
			name: "static asset path",
			req:  httptest.NewRequest(http.MethodGet, "/assets/app-123.js", nil),
			want: "static_asset",
		},
		{
			name: "static page path",
			req:  httptest.NewRequest(http.MethodGet, "/chat", nil),
			want: "static_page",
		},
		{
			name: "normalized pattern drops method prefix",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/workouts/1", nil)
				req.Pattern = "GET /api/workouts/{id}"
				return req
			}(),
			want: "/api/workouts/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routeLabel(tt.req.WithContext(context.Background())); got != tt.want {
				t.Fatalf("routeLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}
