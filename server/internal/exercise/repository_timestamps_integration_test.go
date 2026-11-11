package exercise

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseListAndCreateReturnPersistedTimestamps(t *testing.T) {
	if testing.Short() && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed exercise timestamp regression without DATABASE_URL")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(logger, db.New(pool), pool)
	service := NewService(logger, repo)
	userID := exerciseTestUserA
	ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)

	created, err := service.GetOrCreateExercise(ctx, "Timestamped exercise")
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.True(t, created.CreatedAt.Valid)
	assert.True(t, created.UpdatedAt.Valid)
	assert.False(t, created.CreatedAt.Time.IsZero())
	assert.False(t, created.UpdatedAt.Time.IsZero())

	listed, err := service.ListExercises(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.True(t, listed[0].CreatedAt.Valid)
	assert.True(t, listed[0].UpdatedAt.Valid)
	assert.Equal(t, created.CreatedAt.Time, listed[0].CreatedAt.Time)
	assert.Equal(t, created.UpdatedAt.Time, listed[0].UpdatedAt.Time)

	detail, err := service.GetExerciseWithSets(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, detail.Exercise.CreatedAt, listed[0].CreatedAt.Time)
	assert.Equal(t, detail.Exercise.UpdatedAt, listed[0].UpdatedAt.Time)
}
