package aichat

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/billing"
	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/exercise"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCompleteRun_PersistsWorkoutDraftJSONB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-complete-run-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Build me a structured workout draft for today.", defaultModelName, "req-complete-run-jsonb")
	require.NoError(t, err)

	workoutFocus := "full body"
	weight := 50.0
	draft := &workout.CreateWorkoutRequest{
		Date:         "2026-04-22T12:00:00Z",
		WorkoutFocus: &workoutFocus,
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Weight: &weight, Reps: 10, SetType: "working"},
				},
			},
		},
	}

	_, run, err := repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, run)
	require.NotNil(t, run.WorkoutDraft)
	assert.Equal(t, draft.Date, run.WorkoutDraft.Date)

	storedConversation, err := repo.GetConversation(ctx, conversation.ID, userID)
	require.NoError(t, err)
	require.NotNil(t, storedConversation.LatestWorkoutDraft)
	assert.Equal(t, draft.Date, storedConversation.LatestWorkoutDraft.Date)
	require.NotNil(t, storedConversation.LatestWorkoutDraft.WorkoutFocus)
	assert.Equal(t, workoutFocus, *storedConversation.LatestWorkoutDraft.WorkoutFocus)

	expectedJSON, err := json.Marshal(draft)
	require.NoError(t, err)

	var storedRunDraft []byte
	err = pool.QueryRow(ctx, "SELECT workout_draft FROM ai_chat_run WHERE id = $1 AND user_id = $2", run.ID, userID).Scan(&storedRunDraft)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(storedRunDraft))

	var storedLatestDraft []byte
	err = pool.QueryRow(ctx, "SELECT latest_workout_draft FROM ai_chat_conversation WHERE id = $1 AND user_id = $2", conversation.ID, userID).Scan(&storedLatestDraft)
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), string(storedLatestDraft))
}

func TestRepositoryCompleteRun_AllowsNilWorkoutDraft(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-complete-run-nil-draft-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "How should I warm up today?", defaultModelName, "req-complete-run-nil-draft")
	require.NoError(t, err)

	message, run, err := repo.CompleteRun(ctx, prepared, "Start with five easy minutes and a few ramp-up sets.", nil, time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, message)
	require.NotNil(t, run)
	assert.Nil(t, run.WorkoutDraft)
	assert.Equal(t, statusCompleted, run.Status)

	storedConversation, err := repo.GetConversation(ctx, conversation.ID, userID)
	require.NoError(t, err)
	assert.Nil(t, storedConversation.LatestWorkoutDraft)

	var storedRunDraft []byte
	err = pool.QueryRow(ctx, "SELECT workout_draft FROM ai_chat_run WHERE id = $1 AND user_id = $2", run.ID, userID).Scan(&storedRunDraft)
	require.NoError(t, err)
	assert.Nil(t, storedRunDraft)
}

func TestRepositoryInterruptRun_RequiresLoadedGenerationSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-interrupt-snapshot-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Keep generating while recovery races.", defaultModelName, "req-interrupt-snapshot")
	require.NoError(t, err)

	claimAt := time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC)
	owner := newRunOwner(runOwnerKindAPI, "stale-owner")
	err = repo.ClaimRunGeneration(ctx, prepared.Run, owner, claimAt)
	require.NoError(t, err)

	_, err = repo.AppendStreamChunk(ctx, prepared, "partial", "partial", claimAt.Add(time.Second))
	require.NoError(t, err)

	renewedLease := prepared.Run.LeaseExpiresAt.Add(generationLeaseDuration)
	heartbeatAt := claimAt.Add(5 * time.Second)
	_, err = pool.Exec(ctx, `
		UPDATE ai_chat_run
		SET generation_lease_expires_at = $1,
		    generation_heartbeat_at = $2,
		    updated_at = $2
		WHERE id = $3 AND user_id = $4
	`, renewedLease, heartbeatAt, prepared.Run.ID, userID)
	require.NoError(t, err)

	err = repo.InterruptRun(ctx, prepared, "partial", interruptionReasonStalePartial, claimAt.Add(10*time.Second))
	require.Error(t, err)
	assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected stale interrupt to miss the renewed run, got %v", err)

	var runStatus, generationStatus string
	err = pool.QueryRow(ctx, `
		SELECT status, generation_status
		FROM ai_chat_run
		WHERE id = $1 AND user_id = $2
	`, prepared.Run.ID, userID).Scan(&runStatus, &generationStatus)
	require.NoError(t, err)
	assert.Equal(t, statusStreaming, runStatus)
	assert.Equal(t, generationStatusGenerating, generationStatus)

	var messageStatus, content string
	err = pool.QueryRow(ctx, `
		SELECT status, content
		FROM ai_chat_message
		WHERE id = $1 AND user_id = $2
	`, prepared.AssistantMessage.ID, userID).Scan(&messageStatus, &content)
	require.NoError(t, err)
	assert.Equal(t, statusStreaming, messageStatus)
	assert.Equal(t, "partial", content)
}

func TestRepositoryCompleteRun_PersistsWorkoutDraftWithNullByteText(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-complete-run-null-byte-draft-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Build me a structured workout draft for today.", defaultModelName, "req-complete-run-null-byte-draft")
	require.NoError(t, err)

	notes := "Move well\x00and stop with two reps in reserve."
	draft := &workout.CreateWorkoutRequest{
		Date:  "2026-04-22T12:00:00Z",
		Notes: &notes,
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}

	_, run, err := repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	require.NotNil(t, run)
	require.NotNil(t, run.WorkoutDraft)
	require.NotNil(t, run.WorkoutDraft.Notes)
	assert.NotContains(t, *run.WorkoutDraft.Notes, "\x00")
}

func TestRepositorySaveLatestWorkoutDraft_ConcurrentCallsCreateOneWorkout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-save-draft-race-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	exerciseRepo := exercise.NewRepository(logger, queries, pool)
	workoutSaver := workout.NewTxSaver(logger, exerciseRepo)
	repo := NewRepository(logger, queries, pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "Build me a structured workout draft for today.", defaultModelName, "req-save-draft-race")
	require.NoError(t, err)

	draft := &workout.CreateWorkoutRequest{
		Date: "2026-04-22T12:00:00Z",
		Exercises: []workout.ExerciseInput{
			{
				Name: "Goblet Squat",
				Sets: []workout.SetInput{
					{Reps: 10, SetType: "working"},
				},
			},
		},
	}

	_, _, err = repo.CompleteRun(ctx, prepared, workoutDraftSummaryMessage, draft, time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC))
	require.NoError(t, err)

	start := make(chan struct{})
	results := make(chan int32, 2)
	errs := make(chan error, 2)

	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			resp, saveErr := repo.SaveLatestWorkoutDraft(context.Background(), SaveLatestWorkoutDraftRequest{
				ConversationID: conversation.ID,
				UserID:         userID,
				SavedAt:        time.Now().UTC(),
				SaveWorkout:    workoutSaver.SaveWorkoutTx,
			})
			if saveErr != nil {
				errs <- saveErr
				return
			}
			results <- resp.WorkoutID
		}()
	}

	close(start)
	wg.Wait()
	close(results)
	close(errs)

	for saveErr := range errs {
		require.NoError(t, saveErr)
	}

	var savedWorkoutIDs []int32
	for workoutID := range results {
		savedWorkoutIDs = append(savedWorkoutIDs, workoutID)
	}
	require.Len(t, savedWorkoutIDs, 2)
	assert.Equal(t, savedWorkoutIDs[0], savedWorkoutIDs[1])

	var workoutCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM workout WHERE user_id = $1", userID).Scan(&workoutCount)
	require.NoError(t, err)
	assert.Equal(t, 1, workoutCount)
}

func TestRepositoryListConversations_OrdersByRecentActivityAndScopesToUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-list-conversations-user"
	const otherUserID = "aichat-list-conversations-other-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestUser(t, pool, otherUserID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)
	now := time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC)

	newEmptyConversationID := insertAIChatRepositoryTestConversation(t, pool, userID, "Updated but no messages", now, now.Add(3*time.Hour), nil)
	olderActivityID := insertAIChatRepositoryTestConversation(t, pool, userID, "Older activity", now, now.Add(30*time.Minute), ptrTime(now.Add(30*time.Minute)))
	sameMessageOlderUpdateID := insertAIChatRepositoryTestConversation(t, pool, userID, "Same message, older update", now, now.Add(40*time.Minute), ptrTime(now.Add(45*time.Minute)))
	sameMessageNewerUpdateID := insertAIChatRepositoryTestConversation(t, pool, userID, "Same message, newer update", now, now.Add(50*time.Minute), ptrTime(now.Add(45*time.Minute)))
	newestActivityID := insertAIChatRepositoryTestConversation(t, pool, userID, "Newest activity", now, now.Add(time.Hour), ptrTime(now.Add(time.Hour)))
	insertAIChatRepositoryTestConversation(t, pool, otherUserID, "Other user's chat", now, now.Add(2*time.Hour), ptrTime(now.Add(2*time.Hour)))

	conversations, err := repo.ListConversations(ctx, userID, 50)

	require.NoError(t, err)
	require.Len(t, conversations, 5)
	assert.Equal(t, []int32{newEmptyConversationID, newestActivityID, sameMessageNewerUpdateID, sameMessageOlderUpdateID, olderActivityID}, []int32{
		conversations[0].ID,
		conversations[1].ID,
		conversations[2].ID,
		conversations[3].ID,
		conversations[4].ID,
	})
	assert.Nil(t, conversations[0].LastMessageAt)
}

func TestRepositoryDeleteConversation_CascadesAndRejectsActiveRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-delete-conversation-user"
	const otherUserID = "aichat-delete-conversation-other-user"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestUser(t, pool, otherUserID)
	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "hello", defaultModelName, "req-delete")
	require.NoError(t, err)

	assert.ErrorIs(t, repo.DeleteConversation(ctx, conversation.ID, userID), ErrConversationBusy)
	_, err = repo.GetConversation(ctx, conversation.ID, userID)
	require.NoError(t, err)

	require.NoError(t, repo.FailRun(ctx, prepared, "partial", errors.New("finished before delete"), time.Now().UTC()))
	assert.ErrorIs(t, repo.DeleteConversation(ctx, conversation.ID, otherUserID), pgx.ErrNoRows)
	require.NoError(t, repo.DeleteConversation(ctx, conversation.ID, userID))

	_, err = repo.GetConversation(ctx, conversation.ID, userID)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	for table, column := range map[string]string{
		"ai_chat_message": "conversation_id",
		"ai_chat_run":     "conversation_id",
	} {
		var count int
		require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table+" WHERE "+column+" = $1", conversation.ID).Scan(&count))
		assert.Zero(t, count, table)
	}
	var chunkCount int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM ai_chat_stream_chunk WHERE run_id = $1", prepared.Run.ID).Scan(&chunkCount))
	assert.Zero(t, chunkCount)

}

func TestRepositoryListConversations_AppliesLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-list-conversations-limit-user"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool)
	now := time.Date(2026, 4, 22, 16, 0, 0, 0, time.UTC)
	var oldestIDs []int32
	for i := 0; i < 52; i++ {
		lastMessageAt := now.Add(time.Duration(i) * time.Minute)
		id := insertAIChatRepositoryTestConversation(t, pool, userID, "Limited chat", now, lastMessageAt, &lastMessageAt)
		if i < 2 {
			oldestIDs = append(oldestIDs, id)
		}
	}

	conversations, err := repo.ListConversations(ctx, userID, 50)

	require.NoError(t, err)
	require.Len(t, conversations, 50)
	for _, conversation := range conversations {
		assert.NotContains(t, oldestIDs, conversation.ID)
	}
}

func TestRepositoryPrepareMessageStream_TrialCapAllowsTwoStartsAndBlocksThird(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-trial-cap-user"
	const subscriptionID = "sub_aichat_trial_cap"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestTrialSubscription(t, pool, userID, subscriptionID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool, 2)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	first, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "first prompt", defaultModelName, "req-trial-cap-1")
	require.NoError(t, err)
	_, _, err = repo.CompleteRun(ctx, first, "first answer", nil, time.Now().UTC())
	require.NoError(t, err)

	second, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "second prompt", defaultModelName, "req-trial-cap-2")
	require.NoError(t, err)
	_, _, err = repo.CompleteRun(ctx, second, "second answer", nil, time.Now().UTC())
	require.NoError(t, err)

	third, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "third prompt", defaultModelName, "req-trial-cap-3")
	require.ErrorIs(t, err, billing.ErrTrialPromptLimitExceeded)
	assert.Nil(t, third)

	assertTrialPromptUsage(t, pool, userID, subscriptionID, 2)
	assertAIChatRunCount(t, pool, userID, conversation.ID, 2)
}

func TestRepositoryPrepareMessageStream_DoesNotConsumeTrialPromptWhenStartFails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}

	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-trial-failed-start-user"
	const subscriptionID = "sub_aichat_trial_failed_start"

	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestTrialSubscription(t, pool, userID, subscriptionID)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries, pool, 2)

	missing, err := repo.PrepareMessageStream(ctx, 999999, userID, "missing conversation", defaultModelName, "req-missing")
	require.Error(t, err)
	assert.Nil(t, missing)
	assertNoTrialPromptUsage(t, pool, userID, subscriptionID)

	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)

	invalidModel, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "invalid model", "", "req-invalid-model")
	require.Error(t, err)
	assert.Nil(t, invalidModel)
	assertNoTrialPromptUsage(t, pool, userID, subscriptionID)

	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "started prompt", defaultModelName, "req-started")
	require.NoError(t, err)
	assertTrialPromptUsage(t, pool, userID, subscriptionID, 1)

	busy, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "busy prompt", defaultModelName, "req-busy")
	require.ErrorIs(t, err, ErrConversationBusy)
	assert.Nil(t, busy)
	assertTrialPromptUsage(t, pool, userID, subscriptionID, 1)

	err = repo.FailRun(ctx, prepared, "", errors.New("provider failed"), time.Now().UTC())
	require.NoError(t, err)
	assertTrialPromptUsage(t, pool, userID, subscriptionID, 1)
}

func setupAIChatRepositoryTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dbURL := getAIChatRepositoryTestDatabaseURL()
	if dbURL == "" {
		t.Skip("Skipping database-backed AI chat repository test without DATABASE_URL")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping database-backed AI chat repository test: %v", err)
	}

	tables := []string{
		"users",
		"stripe_customers",
		"stripe_subscriptions",
		"ai_chat_trial_prompt_usage",
		"ai_chat_conversation",
		"ai_chat_message",
		"ai_chat_run",
		"ai_chat_stream_chunk",
		"workout",
		"exercise",
		"set",
	}
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), "ALTER TABLE "+table+" DISABLE ROW LEVEL SECURITY")
		require.NoError(t, err)
	}

	cleanupAIChatRepositoryTestUsers(t, pool)

	cleanup := func() {
		ctx := context.Background()
		cleanupAIChatRepositoryTestUsers(t, pool)

		for _, table := range tables {
			_, err := pool.Exec(ctx, "ALTER TABLE "+table+" ENABLE ROW LEVEL SECURITY")
			require.NoError(t, err)
		}

		pool.Close()
	}

	return pool, cleanup
}

func cleanupAIChatRepositoryTestUsers(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	testUserIDs := []string{
		"aichat-complete-run-user",
		"aichat-complete-run-nil-draft-user",
		"aichat-complete-run-null-byte-draft-user",
		"aichat-interrupt-snapshot-user",
		"aichat-save-draft-race-user",
		"aichat-list-conversations-user",
		"aichat-list-conversations-other-user",
		"aichat-list-conversations-limit-user",
		"aichat-trial-cap-user",
		"aichat-trial-failed-start-user",
		"aichat-data-reader-user",
		"aichat-data-reader-other-user",
		"aichat-data-reader-limit-user",
		"aichat-exercise-stats-user",
		"aichat-exercise-stats-other-user",
		"aichat-profile-user",
		"aichat-profile-update-user",
	}
	for _, userID := range testUserIDs {
		_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, err)
	}
}

func seedAIChatRepositoryTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	_, err := pool.Exec(context.Background(), "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	require.NoError(t, err)
}

func insertAIChatRepositoryTestConversation(t *testing.T, pool *pgxpool.Pool, userID string, title string, createdAt time.Time, updatedAt time.Time, lastMessageAt *time.Time) int32 {
	t.Helper()

	var id int32
	err := pool.QueryRow(context.Background(), `
		INSERT INTO ai_chat_conversation (user_id, title, created_at, updated_at, last_message_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, title, createdAt, updatedAt, lastMessageAt).Scan(&id)
	require.NoError(t, err)

	return id
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func seedAIChatRepositoryTestTrialSubscription(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string) {
	t.Helper()

	now := time.Now().UTC()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO stripe_customers (user_id, stripe_customer_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET stripe_customer_id = EXCLUDED.stripe_customer_id
	`, userID, "cus_"+subscriptionID)
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), `
		INSERT INTO stripe_subscriptions (
			stripe_subscription_id,
			user_id,
			stripe_customer_id,
			stripe_price_id,
			stripe_event_created_at,
			status,
			cancel_at_period_end,
			current_period_start,
			current_period_end,
			trial_start,
			trial_end
		)
		VALUES ($1, $2, $3, $4, $5, 'trialing', false, $5, $6, $5, $6)
		ON CONFLICT (stripe_subscription_id) DO UPDATE
		SET status = 'trialing',
			stripe_event_created_at = EXCLUDED.stripe_event_created_at,
			current_period_end = EXCLUDED.current_period_end,
			trial_end = EXCLUDED.trial_end
	`, subscriptionID, userID, "cus_"+subscriptionID, "price_test", now, now.Add(24*time.Hour))
	require.NoError(t, err)
}

func assertTrialPromptUsage(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string, want int32) {
	t.Helper()

	var got int32
	err := pool.QueryRow(context.Background(), `
		SELECT prompt_count
		FROM ai_chat_trial_prompt_usage
		WHERE user_id = $1 AND stripe_subscription_id = $2
	`, userID, subscriptionID).Scan(&got)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func assertNoTrialPromptUsage(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string) {
	t.Helper()

	var count int
	err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM ai_chat_trial_prompt_usage
		WHERE user_id = $1 AND stripe_subscription_id = $2
	`, userID, subscriptionID).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func assertAIChatRunCount(t *testing.T, pool *pgxpool.Pool, userID string, conversationID int32, want int) {
	t.Helper()

	var got int
	err := pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM ai_chat_run
		WHERE user_id = $1 AND conversation_id = $2
	`, userID, conversationID).Scan(&got)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func getAIChatRepositoryTestDatabaseURL() string {
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}

	_ = godotenv.Load(".env", "../.env", "../../.env")
	if os.Getenv("DATABASE_URL") != "" {
		return os.Getenv("DATABASE_URL")
	}

	return "postgres://postgres:password@localhost:5432/fittrack_test?sslmode=disable"
}
