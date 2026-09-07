package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
)

func TestValidateRLSCompatibleDatabaseURL(t *testing.T) {
	tests := []struct {
		name                string
		url                 string
		enforcementRequired bool
		wantErr             bool
	}{
		{
			name:                "rejects Supabase transaction pooler",
			url:                 "postgres://app:secret@aws-0-us-east-2.pooler.supabase.com:6543/postgres",
			enforcementRequired: true,
			wantErr:             true,
		},
		{
			name:                "accepts Supabase session pooler",
			url:                 "postgres://app:secret@aws-0-us-east-2.pooler.supabase.com:5432/postgres",
			enforcementRequired: true,
		},
		{
			name:                "rejects unrecognized endpoint in enforcement mode",
			url:                 "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable",
			enforcementRequired: true,
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRLSCompatibleDatabaseURL(tt.url, tt.enforcementRequired)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestValidateRLSCompatibleDatabaseURLAllowsStagedCutover(t *testing.T) {
	err := validateRLSCompatibleDatabaseURL(
		"postgres://postgres:secret@aws-0-us-east-2.pooler.supabase.com:6543/postgres",
		false,
	)
	if err != nil {
		t.Fatalf("staged cutover URL rejected before enforcement: %v", err)
	}
}

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
