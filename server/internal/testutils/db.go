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

// BackfillSetOrderColumns populates exercise_order and set_order columns for test data
// This maintains the previous ordering behavior for tests:
// - Exercises ordered by name within each workout
// - Sets ordered by created_at then id within each exercise
func BackfillSetOrderColumns(ctx context.Context, t *testing.T, db DBTX, userID string) {
	t.Helper()

	// Check if columns exist before attempting backfill
	var hasColumns bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'set' 
			AND column_name IN ('exercise_order', 'set_order')
			HAVING COUNT(*) = 2
		)
	`).Scan(&hasColumns)
	if err != nil || !hasColumns {
		// Columns don't exist, skip backfill
		return
	}

	// Backfill order columns using the same logic as the backfill script
	backfillQuery := `
		WITH ranked AS (
			SELECT
				s.id,
				s.workout_id,
				s.exercise_id,
				s.user_id,
				-- Exercise order within the workout (by exercise name, then exercise id)
				DENSE_RANK() OVER (
					PARTITION BY s.workout_id 
					ORDER BY e.name, e.id
				) AS new_exercise_order,
				-- Set order within the exercise (by created_at, then id)
				ROW_NUMBER() OVER (
					PARTITION BY s.workout_id, s.exercise_id 
					ORDER BY s.created_at, s.id
				) AS new_set_order
			FROM "set" s
			JOIN exercise e ON e.id = s.exercise_id
			WHERE s.user_id = $1
			AND (s.exercise_order IS NULL OR s.set_order IS NULL)
		)
		UPDATE "set" s
		SET 
			exercise_order = r.new_exercise_order,
			set_order = r.new_set_order
		FROM ranked r
		WHERE r.id = s.id
	`

	_, err = db.Exec(ctx, backfillQuery, userID)
	if err != nil {
		t.Fatalf("Failed to backfill set order columns: %v", err)
	}
}
