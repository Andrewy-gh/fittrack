package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
)

func TestUserTrainingProfileSchemaAndRLS(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, getTestDatabaseURL())
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))
	requireUserTrainingProfileTable(t, pool)

	cleanupUserTrainingProfileTestData(t, pool)
	defer cleanupUserTrainingProfileTestData(t, pool)

	requireUserTrainingProfilePolicies(t, pool)

	userA := "training-profile-test-user-a"
	userB := "training-profile-test-user-b"
	seedUserTrainingProfileTestUser(t, pool, userA)
	seedUserTrainingProfileTestUser(t, pool, userB)

	connA, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer connA.Release()
	ctxA := testutils.SetTestUserContext(ctx, t, connA, userA)

	sourceConversationA := insertUserTrainingProfileTestConversation(t, connA, ctxA, userA)
	sourceMessageA := insertUserTrainingProfileTestMessage(t, connA, ctxA, userA, sourceConversationA)
	otherConversationA := insertUserTrainingProfileTestConversation(t, connA, ctxA, userA)
	otherMessageA := insertUserTrainingProfileTestMessage(t, connA, ctxA, userA, otherConversationA)

	connB, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer connB.Release()
	ctxB := testutils.SetTestUserContext(ctx, t, connB, userB)

	sourceConversationB := insertUserTrainingProfileTestConversation(t, connB, ctxB, userB)
	sourceMessageB := insertUserTrainingProfileTestMessage(t, connB, ctxB, userB, sourceConversationB)

	_, err = connA.Exec(ctxA, `
		INSERT INTO user_training_profile (
			user_id,
			primary_goal,
			experience_level,
			preferred_session_duration_minutes,
			usual_training_location,
			available_equipment,
			avoided_exercises,
			movement_limitations,
			source_conversation_id,
			source_message_id
		) VALUES (
			$1,
			'strength',
			'intermediate',
			60,
			'gym',
			'["barbell", "dumbbells"]'::jsonb,
			'["burpees"]'::jsonb,
			'["sensitive knee"]'::jsonb,
			$2,
			$3
		)
	`, userA, sourceConversationA, sourceMessageA)
	require.NoError(t, err)

	var visibleToOwner int
	err = connA.QueryRow(ctxA, "SELECT COUNT(*) FROM user_training_profile WHERE user_id = $1", userA).Scan(&visibleToOwner)
	require.NoError(t, err)
	require.Equal(t, 1, visibleToOwner)

	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "unsupported primary goal", `
		UPDATE user_training_profile
		SET primary_goal = 'unsupported_goal'
		WHERE user_id = $1
	`, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "unsupported experience level", `
		UPDATE user_training_profile
		SET experience_level = 'expert'
		WHERE user_id = $1
	`, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "too-short duration", `
		UPDATE user_training_profile
		SET preferred_session_duration_minutes = 5
		WHERE user_id = $1
	`, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "non-array equipment", `
		UPDATE user_training_profile
		SET available_equipment = '{}'::jsonb
		WHERE user_id = $1
	`, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "cross-user source conversation", `
		UPDATE user_training_profile
		SET source_conversation_id = $1,
			source_message_id = NULL
		WHERE user_id = $2
	`, sourceConversationB, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "cross-user source message", `
		UPDATE user_training_profile
		SET source_conversation_id = $1,
			source_message_id = $2
		WHERE user_id = $3
	`, sourceConversationA, sourceMessageB, userA)
	assertUserTrainingProfileConstraintRejects(t, connA, ctxA, "same-user mismatched conversation and message", `
		UPDATE user_training_profile
		SET source_conversation_id = $1,
			source_message_id = $2
		WHERE user_id = $3
	`, sourceConversationA, otherMessageA, userA)

	if isCurrentDatabaseUserSuperuser(t, pool) {
		t.Log("Skipping cross-user visibility assertion because PostgreSQL superusers bypass RLS")
		return
	}

	var visibleToOtherUser int
	err = connB.QueryRow(ctxB, "SELECT COUNT(*) FROM user_training_profile WHERE user_id = $1", userA).Scan(&visibleToOtherUser)
	require.NoError(t, err)
	require.Equal(t, 0, visibleToOtherUser)
}

func requireUserTrainingProfileTable(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var exists bool
	err := pool.QueryRow(context.Background(), "SELECT to_regclass('public.user_training_profile') IS NOT NULL").Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "run database migrations before executing this test")
}

func requireUserTrainingProfilePolicies(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	var rlsEnabled bool
	err := pool.QueryRow(context.Background(), `
		SELECT relrowsecurity
		FROM pg_class
		WHERE relname = 'user_training_profile'
	`).Scan(&rlsEnabled)
	require.NoError(t, err)
	require.True(t, rlsEnabled)

	var policyCount int
	err = pool.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM pg_policies
		WHERE tablename = 'user_training_profile'
		  AND policyname IN (
			'user_training_profile_select_policy',
			'user_training_profile_insert_policy',
			'user_training_profile_update_policy',
			'user_training_profile_delete_policy'
		  )
	`).Scan(&policyCount)
	require.NoError(t, err)
	require.Equal(t, 4, policyCount)
}

func seedUserTrainingProfileTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	ctx = testutils.SetTestUserContext(ctx, t, conn, userID)

	_, err = conn.Exec(ctx, "INSERT INTO users (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	require.NoError(t, err)
}

func insertUserTrainingProfileTestConversation(t *testing.T, db testutils.DBTX, ctx context.Context, userID string) int32 {
	t.Helper()

	var conversationID int32
	err := db.QueryRow(ctx, `
		INSERT INTO ai_chat_conversation (user_id, title)
		VALUES ($1, 'Training profile source test')
		RETURNING id
	`, userID).Scan(&conversationID)
	require.NoError(t, err)
	return conversationID
}

func insertUserTrainingProfileTestMessage(t *testing.T, db testutils.DBTX, ctx context.Context, userID string, conversationID int32) int32 {
	t.Helper()

	var messageID int32
	err := db.QueryRow(ctx, `
		INSERT INTO ai_chat_message (
			conversation_id,
			user_id,
			role,
			content,
			status
		) VALUES (
			$1,
			$2,
			'user',
			'Use this as durable training profile source context.',
			'completed'
		)
		RETURNING id
	`, conversationID, userID).Scan(&messageID)
	require.NoError(t, err)
	return messageID
}

func assertUserTrainingProfileConstraintRejects(t *testing.T, db testutils.DBTX, ctx context.Context, name string, query string, args ...interface{}) {
	t.Helper()

	_, err := db.Exec(ctx, query, args...)
	require.Error(t, err, name)
}

func isCurrentDatabaseUserSuperuser(t *testing.T, pool *pgxpool.Pool) bool {
	t.Helper()

	var isSuperuser bool
	err := pool.QueryRow(context.Background(), "SELECT usesuper FROM pg_user WHERE usename = current_user").Scan(&isSuperuser)
	require.NoError(t, err)
	return isSuperuser
}

func cleanupUserTrainingProfileTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	for _, userID := range []string{"training-profile-test-user-a", "training-profile-test-user-b"} {
		ctx := context.Background()
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)

		ctx = testutils.SetTestUserContext(ctx, t, conn, userID)

		_, err = conn.Exec(ctx, "DELETE FROM user_training_profile WHERE user_id = $1", userID)
		require.NoError(t, err)

		_, err = conn.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)
		require.NoError(t, err)

		conn.Release()
	}
}
