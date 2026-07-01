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

	_, err = connA.Exec(ctxA, `
		INSERT INTO user_training_profile (
			user_id,
			primary_goal,
			experience_level,
			preferred_session_duration_minutes,
			usual_training_location,
			available_equipment,
			avoided_exercises,
			movement_limitations
		) VALUES (
			$1,
			'strength',
			'intermediate',
			60,
			'gym',
			'["barbell", "dumbbells"]'::jsonb,
			'["burpees"]'::jsonb,
			'["sensitive knee"]'::jsonb
		)
	`, userA)
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

	if isCurrentDatabaseUserSuperuser(t, pool) {
		t.Log("Skipping cross-user visibility assertion because PostgreSQL superusers bypass RLS")
		return
	}

	connB, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer connB.Release()
	ctxB := testutils.SetTestUserContext(ctx, t, connB, userB)

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

func assertUserTrainingProfileConstraintRejects(t *testing.T, db testutils.DBTX, ctx context.Context, name string, query string, userID string) {
	t.Helper()

	_, err := db.Exec(ctx, query, userID)
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
