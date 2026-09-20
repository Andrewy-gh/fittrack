package workout

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/require"
)

func TestWorkoutCatalogSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	service := NewService(logger, NewRepository(logger, queries, pool, exerciseRepo))
	ctx := user.WithContext(context.Background(), "test-user-a")
	catalogID := "Pushups"
	req := CreateWorkoutRequest{Date: "2026-09-19T12:00:00Z", Exercises: []ExerciseInput{{Name: "My pushups", CatalogID: &catalogID, Sets: []SetInput{{Reps: 10, SetType: "working"}}}, {Name: "Bench Press", Sets: []SetInput{{Reps: 5, SetType: "working"}}}}}
	id, err := service.CreateWorkoutWithID(ctx, req)
	require.NoError(t, err)
	row, err := queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{Name: "My pushups", UserID: "test-user-a"})
	require.NoError(t, err)
	require.Equal(t, catalogID, row.CatalogID.String)
	unknown, err := queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{Name: "Bench Press", UserID: "test-user-a"})
	require.NoError(t, err)
	require.False(t, unknown.CatalogID.Valid, "ambiguous names stay unclassified")
	// Even an explicit workout selection cannot reclassify an existing same-name exercise.
	changed := "Dumbbell_Bench_Press"
	req.Exercises[0].CatalogID = &changed
	require.NoError(t, service.CreateWorkout(ctx, req))
	same, err := queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{Name: "My pushups", UserID: "test-user-a"})
	require.NoError(t, err)
	require.Equal(t, row, same)
	// Adding a catalog exercise through the edit flow keeps the original exercise link.
	update := UpdateWorkoutRequest{Date: req.Date, Exercises: []UpdateExercise{{Name: "My pushups", Sets: []UpdateSet{{Reps: 10, SetType: "working"}}}, {Name: "New dumbbell press", CatalogID: &changed, Sets: []UpdateSet{{Reps: 8, SetType: "working"}}}}}
	require.NoError(t, service.UpdateWorkout(ctx, id, update))
	added, err := queries.GetExerciseByName(ctx, db.GetExerciseByNameParams{Name: "New dumbbell press", UserID: "test-user-a"})
	require.NoError(t, err)
	require.Equal(t, changed, added.CatalogID.String)
	// Invalid or contradictory selections fail before any workout is written.
	var before, after int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM workout WHERE user_id='test-user-a'").Scan(&before))
	invalid := "unknown"
	req.Exercises[0].CatalogID = &invalid
	require.ErrorIs(t, service.CreateWorkout(ctx, req), ErrInvalidCatalog)
	req.Exercises[0].CatalogID = &catalogID
	req.Exercises = append(req.Exercises, ExerciseInput{Name: "My pushups", CatalogID: &changed, Sets: []SetInput{{Reps: 5, SetType: "working"}}})
	require.ErrorIs(t, service.CreateWorkout(ctx, req), ErrInvalidCatalog)
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM workout WHERE user_id='test-user-a'").Scan(&after))
	require.Equal(t, before, after)
}
