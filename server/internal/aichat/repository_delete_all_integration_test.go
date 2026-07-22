package aichat

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryDeleteAllConversations_OwnerScopedCompleteAndIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-delete-all-owner"
	const otherUserID = "aichat-delete-all-other"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	seedAIChatRepositoryTestUser(t, pool, otherUserID)
	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)

	var sourceConversationID int32
	for index := 0; index < 52; index++ {
		conversation, err := repo.CreateConversation(ctx, userID)
		require.NoError(t, err)
		if index == 0 {
			sourceConversationID = conversation.ID
		}
	}
	otherConversation, err := repo.CreateConversation(ctx, otherUserID)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, sourceConversationID, userID, "private prompt", defaultModelName, "req-delete-all")
	require.NoError(t, err)

	var sourceMessageID int32
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT id FROM ai_chat_message
		WHERE conversation_id = $1 AND user_id = $2
		ORDER BY id LIMIT 1
	`, sourceConversationID, userID).Scan(&sourceMessageID))
	_, err = pool.Exec(ctx, `
		INSERT INTO user_training_profile (user_id, primary_goal, source_conversation_id, source_message_id)
		VALUES ($1, 'strength', $2, $3)
	`, userID, sourceConversationID, sourceMessageID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO workout (date, user_id) VALUES (NOW(), $1)", userID)
	require.NoError(t, err)

	result, err := repo.DeleteAllConversations(ctx, userID, time.Now().UTC())

	require.NoError(t, err)
	assert.Equal(t, int64(52), result.ConversationsDeleted)
	assert.Equal(t, []int32{prepared.Run.ID}, result.StoppedRunIDs)
	assert.Equal(t, 0, countAIChatRowsForUser(t, pool, "ai_chat_conversation", userID))
	assert.Equal(t, 0, countAIChatRowsForUser(t, pool, "ai_chat_message", userID))
	assert.Equal(t, 0, countAIChatRowsForUser(t, pool, "ai_chat_run", userID))
	assert.Equal(t, 1, countAIChatRowsForUser(t, pool, "ai_chat_conversation", otherUserID))
	_, err = repo.GetConversation(ctx, otherConversation.ID, otherUserID)
	require.NoError(t, err)

	var primaryGoal string
	var sourceConversation *int32
	var sourceMessage *int32
	require.NoError(t, pool.QueryRow(ctx, `
		SELECT primary_goal, source_conversation_id, source_message_id
		FROM user_training_profile WHERE user_id = $1
	`, userID).Scan(&primaryGoal, &sourceConversation, &sourceMessage))
	assert.Equal(t, "strength", primaryGoal)
	assert.Nil(t, sourceConversation)
	assert.Nil(t, sourceMessage)
	assert.Equal(t, 1, countAIChatRowsForUser(t, pool, "workout", userID))

	repeated, err := repo.DeleteAllConversations(ctx, userID, time.Now().UTC())
	require.NoError(t, err)
	assert.Zero(t, repeated.ConversationsDeleted)
	assert.Empty(t, repeated.StoppedRunIDs)
}

func TestRepositoryDeleteAllConversations_RollsBackStoppedRunsAndProvenanceOnDeleteFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-delete-all-rollback"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "keep me on rollback", defaultModelName, "req-delete-all-rollback")
	require.NoError(t, err)
	var sourceMessageID int32
	require.NoError(t, pool.QueryRow(ctx, `SELECT id FROM ai_chat_message WHERE conversation_id = $1 ORDER BY id LIMIT 1`, conversation.ID).Scan(&sourceMessageID))
	_, err = pool.Exec(ctx, `
		INSERT INTO user_training_profile (user_id, primary_goal, source_conversation_id, source_message_id)
		VALUES ($1, 'mobility', $2, $3)
	`, userID, conversation.ID, sourceMessageID)
	require.NoError(t, err)

	guardTable := fmt.Sprintf("aichat_delete_all_guard_%d", time.Now().UnixNano())
	_, err = pool.Exec(ctx, fmt.Sprintf(`CREATE TABLE %s (conversation_id INTEGER REFERENCES ai_chat_conversation(id) ON DELETE RESTRICT)`, guardTable))
	require.NoError(t, err)
	defer func() { _, _ = pool.Exec(context.Background(), fmt.Sprintf("DROP TABLE IF EXISTS %s", guardTable)) }()
	_, err = pool.Exec(ctx, fmt.Sprintf("INSERT INTO %s (conversation_id) VALUES ($1)", guardTable), conversation.ID)
	require.NoError(t, err)

	_, err = repo.DeleteAllConversations(ctx, userID, time.Now().UTC())
	require.Error(t, err)

	var runStatus string
	require.NoError(t, pool.QueryRow(ctx, "SELECT status FROM ai_chat_run WHERE id = $1", prepared.Run.ID).Scan(&runStatus))
	assert.Equal(t, statusStreaming, runStatus)
	var sourceConversationID *int32
	require.NoError(t, pool.QueryRow(ctx, "SELECT source_conversation_id FROM user_training_profile WHERE user_id = $1", userID).Scan(&sourceConversationID))
	require.NotNil(t, sourceConversationID)
	assert.Equal(t, conversation.ID, *sourceConversationID)
	_, err = repo.GetConversation(ctx, conversation.ID, userID)
	require.NoError(t, err)
}

func TestRepositoryDeleteAllConversations_SerializesWithActiveRunWriter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database-backed AI chat repository test in short mode")
	}
	pool, cleanup := setupAIChatRepositoryTestDatabase(t)
	if pool == nil {
		return
	}
	defer cleanup()

	const userID = "aichat-delete-all-active-run-race"
	ctx := context.Background()
	seedAIChatRepositoryTestUser(t, pool, userID)
	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	conversation, err := repo.CreateConversation(ctx, userID)
	require.NoError(t, err)
	prepared, err := repo.PrepareMessageStream(ctx, conversation.ID, userID, "race prompt", defaultModelName, "req-delete-all-race")
	require.NoError(t, err)

	writerTx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer writerTx.Rollback(ctx)
	_, err = writerTx.Exec(ctx, "SELECT id FROM ai_chat_run WHERE id = $1 FOR UPDATE", prepared.Run.ID)
	require.NoError(t, err)

	type deletionOutcome struct {
		result *DeleteAllConversationsResult
		err    error
	}
	outcome := make(chan deletionOutcome, 1)
	go func() {
		result, deleteErr := repo.DeleteAllConversations(context.Background(), userID, time.Now().UTC())
		outcome <- deletionOutcome{result: result, err: deleteErr}
	}()

	select {
	case premature := <-outcome:
		t.Fatalf("history deletion returned before the active writer released its run lock: result=%v err=%v", premature.result, premature.err)
	case <-time.After(200 * time.Millisecond):
	}
	require.NoError(t, writerTx.Commit(ctx))

	select {
	case completed := <-outcome:
		require.NoError(t, completed.err)
		assert.Equal(t, int64(1), completed.result.ConversationsDeleted)
		assert.Equal(t, []int32{prepared.Run.ID}, completed.result.StoppedRunIDs)
	case <-time.After(10 * time.Second):
		t.Fatal("history deletion did not finish after the active writer released its run lock")
	}
}

func countAIChatRowsForUser(t *testing.T, pool *pgxpool.Pool, table string, userID string) int {
	t.Helper()
	var count int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table+" WHERE user_id = $1", userID).Scan(&count))
	return count
}
