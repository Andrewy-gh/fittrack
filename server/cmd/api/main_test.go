package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
)

func TestNewHTTPServerUsesBoundedWriteTimeout(t *testing.T) {
	srv := newHTTPServer(&config.Config{Port: 8080}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	if srv.WriteTimeout != 90*time.Second {
		t.Fatalf("WriteTimeout = %v, want %v", srv.WriteTimeout, 90*time.Second)
	}
}
