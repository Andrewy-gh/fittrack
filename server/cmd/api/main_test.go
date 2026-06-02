package main

import (
	"net/http"
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

func TestNewInternalMetricsServerUsesMetricsPort(t *testing.T) {
	srv := newInternalMetricsServer(&config.Config{MetricsPort: 9091}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	if srv.Addr != ":9091" {
		t.Fatalf("Addr = %q, want %q", srv.Addr, ":9091")
	}
}
