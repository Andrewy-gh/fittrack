package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics_UsesNormalizedRouteLabels(t *testing.T) {
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := Metrics()(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/api/workouts/123", nil)
	req.Pattern = "GET /api/workouts/{id}"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got := testutil.ToFloat64(
		httpRequestsTotal.WithLabelValues(http.MethodGet, "/api/workouts/{id}", "200"),
	); got != 1 {
		t.Fatalf("expected normalized route label counter to be 1, got %v", got)
	}

	if got := testutil.ToFloat64(
		httpRequestsTotal.WithLabelValues(http.MethodGet, "/api/workouts/123", "200"),
	); got != 0 {
		t.Fatalf("expected raw path label counter to remain 0, got %v", got)
	}
}

func TestMetrics_PrometheusEndpointFormat(t *testing.T) {
	// Reset all metrics
	httpRequestsTotal.Reset()
	httpRequestDuration.Reset()
	dbConnectionsActive.Set(0)
	dbConnectionsIdle.Set(0)

	// Simulate some requests to generate metrics
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := Metrics()(nextHandler)

	// Make a few requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Update database metrics
	UpdateDatabaseMetrics(nil) // nil pool is safe

	// Test the Prometheus endpoint format
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	promhttp.Handler().ServeHTTP(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Check for expected metric names in the output
	expectedMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"db_connections_active",
		"db_connections_idle",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(bodyStr, metric) {
			t.Errorf("Expected metric %q in output, but it was not found", metric)
		}
	}

	// Verify Content-Type header
	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected Content-Type to contain 'text/plain', got %q", contentType)
	}
}

func TestResponseWriter_MultipleWriteHeaderCalls(t *testing.T) {
	// Test that multiple WriteHeader calls only use the first status code
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusInternalServerError) // Should be ignored
		w.Write([]byte("response"))
	})

	handler := Metrics()(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should record the first status code (200), not the second (500)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}
}

func TestMetrics_PreservesFlusher(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatal("expected wrapped response writer to preserve http.Flusher")
		}

		_, _ = w.Write([]byte("data: hello\n\n"))
		w.(http.Flusher).Flush()
	})

	handler := Metrics()(nextHandler)
	req := httptest.NewRequest(http.MethodGet, "/api/ai/chat/validate/stream", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !rr.Flushed {
		t.Fatal("expected response recorder to be flushed")
	}
}
