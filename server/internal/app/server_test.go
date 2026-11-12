package app

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"log/slog"
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
	oldLogger, oldWriter, oldFlags := slog.Default(), log.Writer(), log.Flags()
	t.Cleanup(func() { slog.SetDefault(oldLogger); log.SetOutput(oldWriter); log.SetFlags(oldFlags) })
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

func TestServerPanicDiagnosticsSurviveLogFiltering(t *testing.T) {
	oldLogger, oldWriter, oldFlags := slog.Default(), log.Writer(), log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(oldLogger)
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
	})
	for _, level := range []slog.Level{slog.LevelWarn, slog.LevelError} {
		for _, kind := range []string{"api", "metrics"} {
			t.Run(level.String()+"/"+kind, func(t *testing.T) {
				var output bytes.Buffer
				logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: level}))
				slog.SetDefault(logger)
				panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					panic("diagnostic-test-panic")
				})
				cfg := &config.Config{}
				srv := newHTTPServer(cfg, panicHandler, logger)
				path := "/panic"
				if kind == "metrics" {
					srv = newMetricsServer(cfg, panicHandler, logger)
					path = "/metrics"
				}
				local := httptest.NewUnstartedServer(nil)
				local.Config = srv
				local.Start()
				t.Cleanup(local.Close)
				client := local.Client()
				client.Timeout = 5 * time.Second
				resp, err := client.Get(local.URL + path)
				if resp != nil {
					resp.Body.Close()
				}
				local.Close() // Join the real handler before inspecting its diagnostic.
				if err == nil {
					t.Fatal("expected the panicking handler to terminate the connection")
				}
				var entry struct {
					Level string `json:"level"`
					Msg   string `json:"msg"`
				}
				if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
					t.Fatalf("missing JSON panic diagnostic: %v; output=%q", err, output.String())
				}
				if entry.Level != "ERROR" || !strings.Contains(entry.Msg, "http: panic serving") || !strings.Contains(entry.Msg, "diagnostic-test-panic") {
					t.Fatalf("unexpected panic diagnostic: %+v", entry)
				}
			})
		}
	}
}

func TestNewHTTPServerUsesBoundedWriteTimeout(t *testing.T) {
	srv := newHTTPServer(&config.Config{Port: 8080}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), slog.Default())

	if srv.WriteTimeout != 10*time.Second {
		t.Fatalf("WriteTimeout = %v, want %v", srv.WriteTimeout, 10*time.Second)
	}
}
