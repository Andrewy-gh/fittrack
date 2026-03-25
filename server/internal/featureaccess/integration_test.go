package featureaccess

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/testutils"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureAccessIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed feature access integration test without DATABASE_URL")
	}

	pool, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queries := db.New(pool)
	repo := NewRepository(logger, queries)
	service := NewService(logger, repo)

	userID := "feature-access-user"
	otherUserID := "feature-access-other-user"

	seedFeatureAccessData(t, pool, userID, otherUserID)

	ctx := testutils.SetTestUserContext(context.Background(), t, pool, userID)
	ctx = user.WithContext(ctx, userID)

	grants, err := service.ListCurrentUserAccess(ctx)
	require.NoError(t, err)
	require.Len(t, grants, 2)
	assert.Equal(t, "ai_chatbot", grants[0].FeatureKey)
	assert.Equal(t, "manual", grants[0].Source)
	assert.Equal(t, "analytics_beta", grants[1].FeatureKey)

	hasChatAccess, err := service.HasCurrentUserFeatureAccess(ctx, "ai_chatbot")
	require.NoError(t, err)
	assert.True(t, hasChatAccess)

	hasExpiredAccess, err := service.HasCurrentUserFeatureAccess(ctx, "expired_feature")
	require.NoError(t, err)
	assert.False(t, hasExpiredAccess)

	hasOtherUsersAccess, err := service.HasCurrentUserFeatureAccess(ctx, "other_user_feature")
	require.NoError(t, err)
	assert.False(t, hasOtherUsersAccess)
}

func setupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	dbURL := getTestDatabaseURL()
	pool, err := pgxpool.New(context.Background(), dbURL)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))

	cleanup := func() {
		ctx := context.Background()
		_, err := pool.Exec(ctx, `
			DELETE FROM user_feature_access
			WHERE user_id IN ('feature-access-user', 'feature-access-other-user')
		`)
		require.NoError(t, err)

		_, err = pool.Exec(ctx, `
			DELETE FROM users
			WHERE user_id IN ('feature-access-user', 'feature-access-other-user')
		`)
		require.NoError(t, err)

		pool.Close()
	}

	return pool, cleanup
}

func seedFeatureAccessData(t *testing.T, pool *pgxpool.Pool, userID string, otherUserID string) {
	t.Helper()

	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id) VALUES ($1), ($2)
		ON CONFLICT (user_id) DO NOTHING
	`, userID, otherUserID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		DELETE FROM user_feature_access
		WHERE user_id IN ($1, $2)
	`, userID, otherUserID)
	require.NoError(t, err)

	now := time.Now().UTC()
	_, err = pool.Exec(ctx, `
		INSERT INTO user_feature_access (
			user_id, feature_key, source, source_reference, granted_by, note, starts_at, expires_at, revoked_at
		)
		VALUES
			($1, 'ai_chatbot', 'manual', 'manual-dev-demo', 'andy', 'demo access', $3, NULL, NULL),
			($1, 'expired_feature', 'manual', NULL, 'andy', NULL, $4, $5, NULL),
			($1, 'revoked_feature', 'manual', NULL, 'andy', NULL, $3, NULL, $3),
			($1, 'analytics_beta', 'manual', NULL, 'andy', NULL, $3, NULL, NULL),
			($2, 'other_user_feature', 'manual', NULL, 'andy', NULL, $3, NULL, NULL)
	`, userID, otherUserID, now.Add(-time.Hour), now.Add(-48*time.Hour), now.Add(-24*time.Hour))
	require.NoError(t, err)
}

func getTestDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}
