package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestValidateRLSCompatibleDatabaseConfig(t *testing.T) {
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
			name:                "accepts resolved safe query overrides",
			url:                 "postgres://app:secret@localhost:6543/postgres?host=aws-0.pooler.supabase.com&port=5432",
			enforcementRequired: true,
		},
		{
			name:                "accepts safe multi-host fallbacks",
			url:                 "postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=aws-0.pooler.supabase.com,aws-1.pooler.supabase.com&port=5432,5432",
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
			poolConfig, err := pgxpool.ParseConfig(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			err = validateRLSCompatibleDatabaseConfig(poolConfig.ConnConfig, tt.enforcementRequired)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestNewRejectsUnsafeProductionURLWithoutOptIn(t *testing.T) {
	// Call the application boundary directly, bypassing config.Load. Rejection
	// must occur before a connection attempt, even with a canceled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, databaseURL := range []string{
		"postgres://postgres:secret@aws-0-us-east-2.pooler.supabase.com:6543/postgres",
		"postgres://postgres:secret@localhost:5432/fittrack_test",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?port=6543",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=localhost",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=%2Ftmp%2Fevil.pooler.supabase.com",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=aws-0.pooler.supabase.com,%2Ftmp%2Fevil.pooler.supabase.com",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=C%3A%5Ctmp%5Cevil.pooler.supabase.com",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=aws-0.pooler.supabase.com,localhost",
		"postgres://app:secret@aws-0.pooler.supabase.com:5432/postgres?host=aws-0.pooler.supabase.com,aws-1.pooler.supabase.com&port=5432,6543",
	} {
		cfg := &config.Config{Environment: "production", DatabaseURL: databaseURL}
		application, err := New(ctx, cfg)
		if application != nil {
			application.Close()
			t.Fatal("unsafe production configuration constructed an application")
		}
		if err == nil || !strings.Contains(err.Error(), "session pooler") {
			t.Fatalf("expected session-pooler rejection, got %v", err)
		}
	}
}

func TestValidateRLSCompatibleDatabaseConfigAllowsNonProductionOptOut(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig("postgres://postgres:secret@aws-0-us-east-2.pooler.supabase.com:5432/postgres?host=localhost&port=6543")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRLSCompatibleDatabaseConfig(poolConfig.ConnConfig, false); err != nil {
		t.Fatalf("non-production URL rejected with enforcement disabled: %v", err)
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
