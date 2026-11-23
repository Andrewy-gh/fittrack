package aichat

import (
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Reconstructed from the synthetic run in 3f84324:service_test.go using that
// revision's workout model JSON tags and run writer (json.Marshal). This is
// repository-history evidence, not a production row.
//
//go:embed testdata/workout_draft_3f84324.json
var historicalWorkoutDraft string

// Deliberately invalid for today's save; this is not the historical fixture.
const unsavableStoredDraft = `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Squat","sets":[{"reps":0,"setType":"working"}]}],"extra":{"preserveInStorage":true}}`

func TestParseStoredWorkoutDraftDistinguishesMissingAndMalformed(t *testing.T) {
	draft, err := parseStoredWorkoutDraft(nil)
	require.NoError(t, err)
	require.Nil(t, draft)

	for _, tc := range []struct{ name, payload string }{
		{"malformed JSON", `{"date":`},
		{"JSON null", `null`},
		{"non-object", `[]`},
		{"incompatible field type", `{"exercises":"not an array"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			draft, err := parseStoredWorkoutDraft([]byte(tc.payload))
			require.ErrorContains(t, err, "decode stored workout draft")
			require.Nil(t, draft)
		})
	}
}

func TestPersistedDraftReadAndSaveCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed draft compatibility test in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	defer cleanup()
	const userID = "aichat-stored-draft-compatibility"
	ctx := user.WithContext(context.Background(), userID)
	seedAIChatRepositoryTestUser(t, pool, userID)
	defer func() {
		_, err := pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, err)
	}()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)
	saver := workout.NewTxSaver(logger, exercise.NewRepository(logger, queries, pool))
	access := new(mockFeatureAccessService)
	access.On("HasCurrentUserFeatureAccess", mock.Anything, featureKeyAIChatbot).Return(true, nil)
	service := NewService(logger, access, nil, repo, saver)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Draft a workout", "synthetic", "stored-draft-compatibility")
	require.NoError(t, err)
	// Finish a real run without a draft, then inject raw JSONB independently of
	// today's draft writer, like rows from an earlier release.
	_, run, err := repo.CompleteRun(ctx, prepared, "Workout draft", nil, time.Now())
	require.NoError(t, err)
	installDraft := func(t *testing.T, payload string) {
		t.Helper()
		_, err := pool.Exec(ctx, `UPDATE ai_chat_conversation
			SET latest_workout_draft = $1::jsonb, latest_workout_draft_source_run_id = $2
			WHERE id = $3`, payload, run.ID, conversation.ID)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `UPDATE ai_chat_run SET workout_draft = $1::jsonb WHERE id = $2`, payload, run.ID)
		require.NoError(t, err)
	}
	type storageState struct {
		conversationDraft, runDraft, conversationRevision, runRevision string
		savedWorkoutID                                                 *int32
		savedAt                                                        *time.Time
		workouts, exercises, sets                                      int
	}
	snapshot := func(t *testing.T) storageState {
		t.Helper()
		var state storageState
		require.NoError(t, pool.QueryRow(ctx, `SELECT c.latest_workout_draft::text, r.workout_draft::text,
			c.xmin::text, r.xmin::text, c.latest_workout_draft_saved_workout_id, c.latest_workout_draft_saved_at,
			(SELECT count(*) FROM workout WHERE user_id = c.user_id),
			(SELECT count(*) FROM exercise WHERE user_id = c.user_id),
			(SELECT count(*) FROM "set" WHERE user_id = c.user_id)
			FROM ai_chat_conversation c JOIN ai_chat_run r ON r.id = c.latest_workout_draft_source_run_id
			WHERE c.id = $1`, conversation.ID).Scan(&state.conversationDraft, &state.runDraft,
			&state.conversationRevision, &state.runRevision, &state.savedWorkoutID, &state.savedAt,
			&state.workouts, &state.exercises, &state.sets))
		return state
	}
	assertReadable := func(t *testing.T, expected string) {
		t.Helper()
		detail, err := service.GetConversation(ctx, conversation.ID)
		require.NoError(t, err)
		resume, err := repo.LoadPreparedRunForResume(ctx, conversation.ID, run.ID, userID, 0)
		require.NoError(t, err)
		for _, draft := range []*workout.CreateWorkoutRequest{detail.Conversation.LatestWorkoutDraft, resume.Run.WorkoutDraft} {
			require.NotNil(t, draft)
			raw, err := json.Marshal(draft)
			require.NoError(t, err)
			require.JSONEq(t, expected, string(raw))
		}
	}

	installDraft(t, unsavableStoredDraft)
	before := snapshot(t)
	assertReadable(t, `{"date":"2026-01-01T12:00:00Z","exercises":[{"name":"Squat","sets":[{"reps":0,"setType":"working"}]}]}`)
	require.Equal(t, before, snapshot(t), "reads must not rewrite either row")
	failed, err := service.SaveLatestWorkoutDraft(ctx, conversation.ID)
	require.ErrorContains(t, err, "validate workout draft before saving")
	require.ErrorContains(t, err, "Reps")
	require.Nil(t, failed)
	after := snapshot(t)
	require.Nil(t, after.savedAt)
	require.Nil(t, after.savedWorkoutID)
	require.Zero(t, after.workouts)
	require.Zero(t, after.exercises)
	require.Zero(t, after.sets)
	require.Equal(t, before, after, "rejected save must leave storage unchanged")

	installDraft(t, historicalWorkoutDraft)
	before = snapshot(t)
	assertReadable(t, historicalWorkoutDraft)
	require.Equal(t, before, snapshot(t), "historical reads must not rewrite either row")
	saved, err := service.SaveLatestWorkoutDraft(ctx, conversation.ID)
	require.NoError(t, err)
	require.Positive(t, saved.WorkoutID)

	// Simulate an already-saved draft that became invalid when rules changed.
	// Keep the real saved outcome (including workout and marker) in place.
	installDraft(t, unsavableStoredDraft)
	before = snapshot(t)
	require.Equal(t, &saved.WorkoutID, before.savedWorkoutID)
	require.NotNil(t, before.savedAt)
	require.Equal(t, 1, before.workouts)
	repeated, err := service.SaveLatestWorkoutDraft(ctx, conversation.ID)
	require.NoError(t, err)
	require.Equal(t, saved.WorkoutID, repeated.WorkoutID)
	require.True(t, repeated.Conversation.LatestWorkoutDraftStatus.IsSaved)
	require.Equal(t, before, snapshot(t), "idempotent save must not rewrite or add rows")
}
