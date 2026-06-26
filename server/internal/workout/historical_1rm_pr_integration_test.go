package workout

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkout_Historical1RM_PR_Lifecycle_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)

	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutRepo := NewRepository(logger, queries, pool, exerciseRepo)
	workoutService := NewService(logger, workoutRepo)

	userID := "test-user-a"
	ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
	ctx = user.WithContext(ctx, userID)

	// 1) Create workout with a PR.
	createReq := CreateWorkoutRequest{
		Date:  "2026-02-10T10:00:00Z",
		Notes: stringPtr("PR workout"),
		Exercises: []ExerciseInput{
			{
				Name: "Bench Press",
				Sets: []SetInput{
					{Weight: float64Ptr(200), Reps: 1, SetType: "working"}, // e1rm ~= 206.67
				},
			},
		},
	}

	require.NoError(t, workoutService.CreateWorkout(ctx, createReq))

	var workoutID int32
	err := pool.QueryRow(ctx, "SELECT id FROM workout WHERE user_id = $1 ORDER BY id DESC LIMIT 1", userID).
		Scan(&workoutID)
	require.NoError(t, err)

	var exerciseID int32
	err = pool.QueryRow(ctx, "SELECT id FROM exercise WHERE user_id = $1 AND name = $2", userID, "Bench Press").
		Scan(&exerciseID)
	require.NoError(t, err)

	var hist1rm float64
	var srcWorkoutID int32
	err = pool.QueryRow(ctx, "SELECT historical_1rm, historical_1rm_source_workout_id FROM exercise WHERE id = $1 AND user_id = $2", exerciseID, userID).
		Scan(&hist1rm, &srcWorkoutID)
	require.NoError(t, err)
	assert.InDelta(t, 206.67, hist1rm, 0.02)
	assert.Equal(t, workoutID, srcWorkoutID)

	// 2) Update workout to reduce the best e1rm; since the PR is sourced from this workout, it must recompute down.
	updateReq := UpdateWorkoutRequest{
		Date:  "2026-02-10T10:00:00Z",
		Notes: stringPtr("PR reduced"),
		Exercises: []UpdateExercise{
			{
				Name: "Bench Press",
				Sets: []UpdateSet{
					{Weight: float64Ptr(100), Reps: 1, SetType: "working"}, // e1rm ~= 103.33
				},
			},
		},
	}
	require.NoError(t, workoutService.UpdateWorkout(ctx, workoutID, updateReq))

	err = pool.QueryRow(ctx, "SELECT historical_1rm, historical_1rm_source_workout_id FROM exercise WHERE id = $1 AND user_id = $2", exerciseID, userID).
		Scan(&hist1rm, &srcWorkoutID)
	require.NoError(t, err)
	assert.InDelta(t, 103.33, hist1rm, 0.02)
	assert.Equal(t, workoutID, srcWorkoutID)

	// 3) Create a second workout with a higher PR; auto-update should promote it.
	createReq2 := CreateWorkoutRequest{
		Date:  "2026-02-11T10:00:00Z",
		Notes: stringPtr("new PR"),
		Exercises: []ExerciseInput{
			{
				Name: "Bench Press",
				Sets: []SetInput{
					{Weight: float64Ptr(250), Reps: 1, SetType: "working"}, // e1rm ~= 258.33
				},
			},
		},
	}
	require.NoError(t, workoutService.CreateWorkout(ctx, createReq2))

	var workoutID2 int32
	err = pool.QueryRow(ctx, "SELECT id FROM workout WHERE user_id = $1 ORDER BY id DESC LIMIT 1", userID).
		Scan(&workoutID2)
	require.NoError(t, err)
	require.NotEqual(t, workoutID, workoutID2)

	err = pool.QueryRow(ctx, "SELECT historical_1rm, historical_1rm_source_workout_id FROM exercise WHERE id = $1 AND user_id = $2", exerciseID, userID).
		Scan(&hist1rm, &srcWorkoutID)
	require.NoError(t, err)
	assert.InDelta(t, 258.33, hist1rm, 0.02)
	assert.Equal(t, workoutID2, srcWorkoutID)

	// 4) Delete the PR workout; historical should recompute to the remaining best (workout 1).
	require.NoError(t, workoutService.DeleteWorkout(ctx, workoutID2))

	err = pool.QueryRow(ctx, "SELECT historical_1rm, historical_1rm_source_workout_id FROM exercise WHERE id = $1 AND user_id = $2", exerciseID, userID).
		Scan(&hist1rm, &srcWorkoutID)
	require.NoError(t, err)
	assert.InDelta(t, 103.33, hist1rm, 0.02)
	assert.Equal(t, workoutID, srcWorkoutID)
}
