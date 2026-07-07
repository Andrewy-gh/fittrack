package aichat

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryChatDataReader_FiltersByUserDateAndExercise(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-data-reader-user"
	const otherUserID = "aichat-data-reader-other-user"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestUser(t, pool, otherUserID)

	seedAIChatDataWorkout(t, pool, userID, "2026-06-30", "lower body", []chatDataExerciseSeed{
		{name: "Back Squat", sets: []chatDataSetSeed{{weight: ptrFloat64(225), reps: 5, setType: "working"}, {weight: ptrFloat64(235), reps: 3, setType: "working"}}},
	})
	seedAIChatDataWorkout(t, pool, userID, "2026-06-15", "pull", []chatDataExerciseSeed{
		{name: "Deadlift", sets: []chatDataSetSeed{{weight: ptrFloat64(275), reps: 3, setType: "working"}}},
	})
	seedAIChatDataWorkout(t, pool, userID, "2026-05-20", "lower body", []chatDataExerciseSeed{
		{name: "Back Squat", sets: []chatDataSetSeed{{weight: ptrFloat64(205), reps: 5, setType: "working"}}},
	})
	seedAIChatDataWorkout(t, pool, otherUserID, "2026-06-29", "lower body", []chatDataExerciseSeed{
		{name: "Back Squat", sets: []chatDataSetSeed{{weight: ptrFloat64(315), reps: 1, setType: "working"}}},
	})

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)
	workouts, err := repo.ListWorkoutsWithSets(ctx, userID, WorkoutHistoryFilter{
		LastN:        10,
		StartDate:    &start,
		EndDate:      &end,
		ExerciseName: "Back Squat",
	})

	require.NoError(t, err)
	require.Len(t, workouts, 1)
	assert.Equal(t, "2026-06-30", workouts[0].Date)
	assert.Equal(t, "lower body", workouts[0].Focus)
	require.Len(t, workouts[0].Exercises, 1)
	assert.Equal(t, "Back Squat", workouts[0].Exercises[0].Name)
	assert.Equal(t, []string{"225x5 working", "235x3 working"}, workouts[0].Exercises[0].Sets)
}

func TestRepositoryChatDataReader_ResolveNamesAndSnapshotAreUserScoped(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-data-reader-user"
	const otherUserID = "aichat-data-reader-other-user"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestUser(t, pool, otherUserID)

	seedAIChatDataWorkout(t, pool, userID, time.Now().UTC().AddDate(0, 0, -2).Format("2006-01-02"), "push", []chatDataExerciseSeed{
		{name: "Bench Press", sets: []chatDataSetSeed{{weight: ptrFloat64(185), reps: 5, setType: "working"}}},
		{name: "Barbell Row", sets: []chatDataSetSeed{{weight: ptrFloat64(155), reps: 8, setType: "working"}}},
	})
	seedAIChatDataWorkout(t, pool, userID, time.Now().UTC().AddDate(0, 0, -8).Format("2006-01-02"), "pull", []chatDataExerciseSeed{
		{name: "Bench Press", sets: []chatDataSetSeed{{weight: ptrFloat64(175), reps: 6, setType: "working"}}},
	})
	seedAIChatDataWorkout(t, pool, otherUserID, time.Now().UTC().Format("2006-01-02"), "push", []chatDataExerciseSeed{
		{name: "Bench Machine", sets: []chatDataSetSeed{{weight: ptrFloat64(250), reps: 4, setType: "working"}}},
	})

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	names, err := repo.ResolveExerciseNames(ctx, userID, "bench")
	require.NoError(t, err)
	assert.Equal(t, []string{"Bench Press"}, names)

	snapshot, err := repo.TrainingSnapshot(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.NotEmpty(t, snapshot.LastWorkoutDate)
	assert.EqualValues(t, 2, snapshot.WorkoutsLast30D)
	assert.Contains(t, snapshot.TopExercises, "Bench Press")
	assert.NotContains(t, snapshot.TopExercises, "Bench Machine")
}

func TestRepositoryChatDataReader_CapsWorkoutLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-data-reader-limit-user"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	for i := 0; i < 25; i++ {
		date := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, -i).Format("2006-01-02")
		seedAIChatDataWorkout(t, pool, userID, date, "general", []chatDataExerciseSeed{
			{name: "Bench Press", sets: []chatDataSetSeed{{weight: ptrFloat64(135), reps: 5, setType: "working"}}},
		})
	}

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	workouts, err := repo.ListWorkoutsWithSets(ctx, userID, WorkoutHistoryFilter{LastN: 99})

	require.NoError(t, err)
	require.Len(t, workouts, 20)
	assert.Equal(t, "2026-07-01", workouts[0].Date)
	assert.Equal(t, "2026-06-12", workouts[19].Date)
}

type chatDataExerciseSeed struct {
	name string
	sets []chatDataSetSeed
}

type chatDataSetSeed struct {
	weight  *float64
	reps    int32
	setType string
}

func seedAIChatDataWorkout(t *testing.T, pool *pgxpool.Pool, userID string, date string, focus string, exercises []chatDataExerciseSeed) {
	t.Helper()

	ctx := context.Background()
	var workoutID int32
	err := pool.QueryRow(ctx, `
		INSERT INTO workout (date, workout_focus, user_id)
		VALUES ($1::date, $2, $3)
		RETURNING id
	`, date, focus, userID).Scan(&workoutID)
	require.NoError(t, err)

	for exerciseIndex, exercise := range exercises {
		var exerciseID int32
		err = pool.QueryRow(ctx, `
			INSERT INTO exercise (name, user_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, exercise.name, userID).Scan(&exerciseID)
		require.NoError(t, err)

		for setIndex, set := range exercise.sets {
			_, err = pool.Exec(ctx, `
				INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, exerciseID, workoutID, set.weight, set.reps, set.setType, userID, exerciseIndex+1, setIndex+1)
			require.NoError(t, err)
		}
	}
}

func ptrFloat64(value float64) *float64 {
	return &value
}
