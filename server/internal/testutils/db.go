// Package testutils provides utilities for testing
package testutils

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is an interface that matches the database transaction interface
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// SetTestUserContext sets the current user ID in the database session for testing
func SetTestUserContext(ctx context.Context, t *testing.T, db DBTX, userID string) context.Context {
	t.Helper()

	_, err := db.Exec(ctx, "SELECT set_config('app.current_user_id', $1, false)", userID)
	if err != nil {
		t.Fatalf("Failed to set test user context: %v", err)
	}

	return ctx
}
