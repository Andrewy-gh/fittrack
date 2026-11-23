package aichat

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/stretchr/testify/require"
)

func TestSaveDraftRejectsUnrepresentableSetAtomically(t *testing.T) {
	if testing.Short() {
		t.Skip("database integration")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		t.Fatal("database unavailable")
	}
	defer cleanup()
	ctx := context.Background()
	const owner = "aichat-numeric-user"
	seedAIChatRepositoryTestUser(t, pool, owner)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	q := db.New(pool)
	repo := NewRepository(logger, q, pool)
	saver := workout.NewTxSaver(logger, exercise.NewRepository(logger, q, pool))
	conv, err := repo.CreateConversation(ctx, owner)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, conv.ID, owner, "Build a workout", defaultModelName, "numeric-draft")
	require.NoError(t, err)
	draft := &workout.CreateWorkoutRequest{Date: "2026-04-22T12:00:00Z", Exercises: []workout.ExerciseInput{{Name: "Numeric draft", Sets: []workout.SetInput{{Reps: 10, SetType: "working"}, {Reps: 2147483648, SetType: "working"}}}}}
	_, _, err = repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Now())
	require.NoError(t, err)
	_, err = repo.SaveLatestWorkoutDraft(ctx, SaveLatestWorkoutDraftRequest{ConversationID: conv.ID, UserID: owner, SavedAt: time.Now(), SaveWorkout: saver.SaveWorkoutTx})
	require.ErrorIs(t, err, workout.ErrInvalidSetNumeric)
	var workouts, sets int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM workout WHERE user_id=$1`, owner).Scan(&workouts))
	require.Zero(t, workouts)
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM "set" WHERE user_id=$1`, owner).Scan(&sets))
	require.Zero(t, sets)
	var saved bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_workout_draft_saved_workout_id IS NOT NULL FROM ai_chat_conversation WHERE id=$1`, conv.ID).Scan(&saved))
	require.False(t, saved)
}
