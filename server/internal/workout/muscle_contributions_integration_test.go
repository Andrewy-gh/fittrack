package workout

import (
	"context"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestMuscleContributionsIntegrationOwnershipAndPeriod(t *testing.T) {
	if testing.Short() || os.Getenv("DATABASE_URL") == "" {
		t.Skip("requires isolated test database")
	}
	pool, cleanup := setupTestDatabase(t)
	defer cleanup()
	owner, other := "test-user-a", "test-user-b"
	ownerCtx := setTestUserContext(context.Background(), t, pool, owner)
	exercise := insertEvidenceExercise(t, pool, ownerCtx, owner, "Unclassified press", nil)
	otherCtx := setTestUserContext(context.Background(), t, pool, other)
	foreignExercise := insertEvidenceExercise(t, pool, otherCtx, other, "Other press", nil)
	foreignWorkout := insertEvidenceWorkout(t, pool, otherCtx, other, time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC))
	ownerCtx = setTestUserContext(context.Background(), t, pool, owner)
	start := time.Date(2026, 9, 14, 4, 0, 0, 0, time.UTC) // Midnight New York.
	now := start.Add(48 * time.Hour)
	for _, date := range []time.Time{start.Add(-time.Second), start, now, now.Add(time.Second), start.AddDate(0, 0, 7)} {
		id := insertEvidenceWorkout(t, pool, ownerCtx, owner, date)
		insertEvidenceSet(t, pool, ownerCtx, owner, id, exercise, nil, 5, "working", 1, 1)
		insertEvidenceSet(t, pool, ownerCtx, owner, id, exercise, 0.0, 5, "working", 1, 2)
	}
	ownedWorkout := insertEvidenceWorkout(t, pool, ownerCtx, owner, start)
	insertEvidenceSet(t, pool, context.Background(), owner, ownedWorkout, foreignExercise, 20.0, 5, "working", 1, 1)
	// Legacy malformed joins must not leak or contribute even under migration-owner access.
	insertEvidenceSet(t, pool, context.Background(), owner, foreignWorkout, exercise, 20.0, 5, "working", 1, 1)
	result, err := NewMuscleContributionService(db.New(pool), func() time.Time { return now }).Summarize(user.WithContext(ownerCtx, owner), TrainingEvidenceRequest{StartDate: "2026-09-14", Timezone: "America/New_York"})
	require.NoError(t, err)
	require.Len(t, result.Exercises, 1)
	require.Equal(t, 4, result.WorkingSets)
	require.True(t, result.Period.Partial)
	require.Equal(t, start, result.Period.StartAt)
	require.Equal(t, "Unclassified press", result.Exercises[0].Name)
	require.Equal(t, 4, result.UnclassifiedSets)
	// Relinking changes the assessment without rewriting historical sets.
	_, err = pool.Exec(ownerCtx, "UPDATE exercise SET catalog_id = 'Barbell_Curl' WHERE id = $1 AND user_id = $2", exercise, owner)
	require.NoError(t, err)
	result, err = NewMuscleContributionService(db.New(pool), func() time.Time { return now }).Summarize(user.WithContext(ownerCtx, owner), TrainingEvidenceRequest{StartDate: "2026-09-14", Timezone: "America/New_York"})
	require.NoError(t, err)
	require.Equal(t, 4, result.Muscles[2].DirectSets)
	require.Zero(t, result.UnclassifiedSets)

}
