package exercise

import (
	"context"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestGoalCheckQueryOwnershipAndWindow(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("requires isolated test database")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	owner := exerciseTestUserA
	id := setupTestExercise(t, pool, owner, "Goal check bench")
	ctx := testutils.SetTestUserContext(context.Background(), t, pool, owner)
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	window := goalCheckWindow(now, time.UTC)
	for i, day := range []time.Time{window.Start, now.AddDate(0, 0, -1), now, window.Start.Add(-time.Second), now.Add(time.Second)} {
		workout := insertMetricsHistoryWorkout(t, ctx, pool, owner, day)
		for _, kind := range []string{"working", "warmup"} {
			_, err := pool.Exec(ctx, `INSERT INTO "set" (exercise_id,workout_id,user_id,reps,weight,set_type,exercise_order,set_order) VALUES ($1,$2,$3,5,NULL,$4,0,$5)`, id, workout, owner, kind, i)
			require.NoError(t, err)
		}
	}
	queries := db.New(pool)
	params := db.GetExerciseGoalCheckParams{ExerciseID: id, UserID: owner, WindowStart: pgtype.Timestamptz{Time: window.Start, Valid: true}, AsOf: pgtype.Timestamptz{Time: now, Valid: true}}
	rows, err := queries.GetExerciseGoalCheck(ctx, params)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	for _, row := range rows {
		require.True(t, row.SetID.Valid)
		require.False(t, row.Weight.Valid)
	}
	// Explicit ownership holds even when the DB connection has broader access.
	params.UserID = exerciseTestUserB
	rows, err = queries.GetExerciseGoalCheck(ctx, params)
	require.NoError(t, err)
	require.Empty(t, rows)
	empty := setupTestExercise(t, pool, owner, "Empty goal exercise")
	params.UserID = owner
	params.ExerciseID = empty
	rows, err = queries.GetExerciseGoalCheck(ctx, params)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.False(t, rows[0].SetID.Valid)
}
