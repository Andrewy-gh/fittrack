package app

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOpenDatabaseRejectsOwnerRoleWhenRLSIsRequired(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for database integration test")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || (parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1") {
		t.Skip("owner-role assertion is only for the repository's local test database")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("open local database: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping local database: %v", err)
	}

	err = validateProductionDatabaseRole(context.Background(), pool)
	if err == nil || !strings.Contains(err.Error(), "owns_rls_table=true") {
		t.Fatalf("expected owner-role rejection, got %v", err)
	}
}

func TestOpenDatabaseSetsRLSContextForEveryAcquisition(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for database integration test")
	}

	cfg := databaseTestConfig(databaseURL)
	pool, err := openDatabase(context.Background(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	for _, userID := range []string{"rls-context-user-a", "rls-context-user-b"} {
		ctx := user.WithContext(context.Background(), userID)
		var got string
		if err := pool.QueryRow(ctx, "SELECT current_user_id()").Scan(&got); err != nil {
			t.Fatalf("read current user for %s: %v", userID, err)
		}
		if got != userID {
			t.Fatalf("current user = %q, want %q", got, userID)
		}
	}

	var cleared string
	if err := pool.QueryRow(context.Background(), "SELECT current_user_id()").Scan(&cleared); err != nil {
		t.Fatalf("read cleared current user: %v", err)
	}
	if cleared != "" {
		t.Fatalf("unauthenticated acquisition retained user %q", cleared)
	}
}

func databaseTestConfig(databaseURL string) *config.Config {
	return &config.Config{
		DatabaseURL:   databaseURL,
		Environment:   "development",
		DBMaxConns:    2,
		DBMinConns:    0,
		DBMaxConnIdle: "30s",
		DBMaxConnLife: "30m",
		DBHealthCheck: "30s",
	}
}
