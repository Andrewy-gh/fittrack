package workout

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/recommendation"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestWorkoutService_CreateWorkoutPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	const owner = "test-user-a"
	ctx := user.WithContext(context.Background(), owner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	exercises := exercise.NewRepository(logger, queries, pool)
	service := NewService(logger, NewRepository(logger, queries, pool, exercises))

	press, err := exercises.GetOrCreateExerciseTx(ctx, queries, "Persistence press", owner)
	require.NoError(t, err)
	const date = "2026-10-08T12:00:00Z"
	snapshot := recommendation.Snapshot{
		ExerciseName: press.Name,
		Recommendation: recommendation.Result{
			ExerciseID: press.ID, Readiness: "normal", Explanation: "Keep the current load.", PolicyVersion: recommendation.PolicyVersion,
		},
	}
	zero := 0.0
	id, err := service.CreateWorkoutWithID(ctx, CreateWorkoutRequest{
		Date: date,
		Exercises: []ExerciseInput{
			{Name: press.Name, Sets: []SetInput{{Reps: 5, SetType: "warmup"}}},
			{Name: "Persistence row", Sets: []SetInput{{Reps: 10, SetType: "working"}}},
			{Name: press.Name, Sets: []SetInput{{Weight: &zero, Reps: 8, SetType: "working"}}},
		},
		Recommendations: []recommendation.Snapshot{snapshot},
	})
	require.NoError(t, err)
	reloaded, err := service.GetWorkoutWithSets(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []recommendation.Snapshot{snapshot}, reloaded.Recommendations)
	sets := reloaded.Sets
	require.Len(t, sets, 3)
	require.Equal(t, []string{press.Name, press.Name, "Persistence row"}, []string{sets[0].ExerciseName, sets[1].ExerciseName, sets[2].ExerciseName})
	require.Equal(t, []int32{5, 8, 10}, []int32{sets[0].Reps, sets[1].Reps, sets[2].Reps})
	require.Nil(t, sets[0].Weight)
	require.Equal(t, &zero, sets[1].Weight)
	require.Less(t, sets[0].SetID, sets[1].SetID)
	require.Less(t, sets[1].SetID, sets[2].SetID, "manual creation allocates set IDs in exercise order")

	// The recommendation is valid by itself, but points to a different exercise.
	// Its write is checked after the new workout, exercise and set are inserted.
	const rejected = "Persistence rejected movement"
	snapshot.ExerciseName = rejected
	failedID, err := service.CreateWorkoutWithID(ctx, CreateWorkoutRequest{
		Date:            date,
		Exercises:       []ExerciseInput{{Name: rejected, Sets: []SetInput{{Reps: 8, SetType: "working"}}}},
		Recommendations: []recommendation.Snapshot{snapshot},
	})
	require.ErrorIs(t, err, recommendation.ErrInvalidContext)
	require.Zero(t, failedID)
	workouts, err := service.ListWorkouts(ctx)
	require.NoError(t, err)
	require.Len(t, workouts, 1)
	require.Equal(t, id, workouts[0].ID)
	_, err = queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{Name: rejected, UserID: owner})
	require.ErrorIs(t, err, pgx.ErrNoRows, "new exercise must roll back with the workout")
}
