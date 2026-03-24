package exercise

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseRepository_GetExerciseMetricsHistory_PreservesSameDaySessions(t *testing.T) {
	if testing.Short() && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed metrics history regression without DATABASE_URL")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(logger, db.New(pool), pool)

	userID := "test-user-a"
	exerciseID := setupTestExercise(t, pool, userID, "Bench Press")
	ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)

	workoutA := insertMetricsHistoryWorkout(
		t,
		ctx,
		pool,
		userID,
		time.Date(2026, 3, 24, 8, 0, 0, 0, time.UTC),
	)
	workoutB := insertMetricsHistoryWorkout(
		t,
		ctx,
		pool,
		userID,
		time.Date(2026, 3, 24, 18, 0, 0, 0, time.UTC),
	)

	insertMetricsHistorySet(t, ctx, pool, userID, exerciseID, workoutA, 185, 5, 925)
	insertMetricsHistorySet(t, ctx, pool, userID, exerciseID, workoutB, 205, 3, 615)

	points, bucket, err := repo.GetExerciseMetricsHistory(ctx, GetExerciseMetricsHistoryRequest{
		ExerciseID: exerciseID,
		Range:      "Y",
	}, userID)
	require.NoError(t, err)

	assert.Equal(t, MetricsHistoryBucketWorkout, bucket)
	assert.Len(t, points, 2)
	assert.Equal(t, []int32{workoutA, workoutB}, []int32{
		*points[0].WorkoutID,
		*points[1].WorkoutID,
	})
	assert.Equal(t, []string{"2026-03-24", "2026-03-24"}, []string{
		points[0].Date.Format("2006-01-02"),
		points[1].Date.Format("2006-01-02"),
	})
}

func insertMetricsHistoryWorkout(
	t *testing.T,
	ctx context.Context,
	pool db.DBTX,
	userID string,
	date time.Time,
) int32 {
	t.Helper()

	var workoutID int32
	err := pool.QueryRow(
		ctx,
		`INSERT INTO workout (date, user_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id`,
		date,
		userID,
	).Scan(&workoutID)
	require.NoError(t, err)

	return workoutID
}

func insertMetricsHistorySet(
	t *testing.T,
	ctx context.Context,
	pool db.DBTX,
	userID string,
	exerciseID int32,
	workoutID int32,
	weight float64,
	reps int32,
	volume float64,
) {
	t.Helper()

	_, err := pool.Exec(
		ctx,
		`INSERT INTO "set"
			(exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'working', $5, 1, 1, NOW(), NOW())`,
		exerciseID,
		workoutID,
		weight,
		reps,
		userID,
	)
	require.NoError(t, err)

	// Sanity check the inserted volume footprint so the session metric is non-zero.
	var totalVolume float64
	err = pool.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(COALESCE(weight, 0) * reps), 0)::float8
		FROM "set"
		WHERE workout_id = $1 AND exercise_id = $2 AND user_id = $3`,
		workoutID,
		exerciseID,
		userID,
	).Scan(&totalVolume)
	require.NoError(t, err)
	assert.InDelta(t, volume, totalVolume, 0.001)
}
