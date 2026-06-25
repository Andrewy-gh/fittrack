package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
)

func TestNewHTTPServerUsesBoundedWriteTimeout(t *testing.T) {
	srv := newHTTPServer(&config.Config{Port: 8080}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	if srv.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout = %v, want %v", srv.WriteTimeout, 10*time.Second)
	}
}

func TestNewMetricsServerUsesConfiguredPortAndMetricsPath(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := newMetricsServer(&config.Config{MetricsPort: 9092}, handler)

	if srv.Addr != ":9092" {
		t.Fatalf("Addr = %q, want %q", srv.Addr, ":9092")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/not-metrics", nil)
	rr = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}
