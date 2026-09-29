package middleware

import (
	"bytes"
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
