package billing

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryUpsertSubscriptionFromWebhook_SourceScopedFeatureAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	userID := "billing-source-scope-user"
	subscriptionID := "sub_source_scope"
	customerID := "cus_source_scope"
	cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	defer cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)

	seedLocalE2EFeatureAccess(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	periodStart := time.Now().UTC().Add(-time.Hour)
	periodEnd := periodStart.Add(30 * 24 * time.Hour)
	activeEventCreatedAt := periodStart.Add(time.Minute)
	canceledEventCreatedAt := activeEventCreatedAt.Add(time.Minute)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &activeEventCreatedAt,
		Status:               subscriptionStatusActive,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		GrantAIChatAccess:    true,
	})
	require.NoError(t, err)

	activeSources := activeAIChatSources(t, pool, userID)
	assert.ElementsMatch(t, []string{"local_e2e_auth", "stripe"}, activeSources)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &canceledEventCreatedAt,
		Status:               subscriptionStatusCanceled,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
	})
	require.NoError(t, err)

	activeSources = activeAIChatSources(t, pool, userID)
	assert.ElementsMatch(t, []string{"local_e2e_auth"}, activeSources)
	assert.Equal(t, int64(1), revokedStripeGrantCount(t, pool, userID, subscriptionID))
}

func TestRepositoryUpsertSubscriptionFromWebhook_GrantExpiresAtCancelAtWhenEarlier(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	userID := "billing-cancel-at-expiry-user"
	subscriptionID := "sub_cancel_at_expiry"
	customerID := "cus_cancel_at_expiry"
	cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	defer cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	seedBillingUser(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	periodStart := time.Now().UTC().Add(-time.Hour)
	cancelAt := periodStart.Add(24 * time.Hour)
	periodEnd := periodStart.Add(30 * 24 * time.Hour)
	eventCreatedAt := periodStart.Add(time.Minute)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               subscriptionStatusActive,
		CancelAt:             &cancelAt,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		GrantAIChatAccess:    true,
	})
	require.NoError(t, err)

	assert.WithinDuration(t, cancelAt, stripeGrantExpiresAt(t, pool, userID, subscriptionID), time.Second)
}

func TestRepositoryUpsertSubscriptionFromWebhook_IgnoresOlderGrantingEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	tests := []struct {
		name         string
		latestStatus string
		olderStatus  string
	}{
		{
			name:         "canceled beats older active event",
			latestStatus: subscriptionStatusCanceled,
			olderStatus:  subscriptionStatusActive,
		},
		{
			name:         "past due beats older trialing event",
			latestStatus: subscriptionStatusPastDue,
			olderStatus:  subscriptionStatusTrialing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := "billing-stale-" + tt.latestStatus
			subscriptionID := "sub_stale_" + tt.latestStatus
			customerID := "cus_stale_" + tt.latestStatus
			cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
			defer cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
			seedBillingUser(t, pool, userID)

			repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
			periodStart := time.Now().UTC().Add(-time.Hour)
			periodEnd := periodStart.Add(30 * 24 * time.Hour)
			olderEventCreatedAt := periodStart.Add(time.Minute)
			latestEventCreatedAt := olderEventCreatedAt.Add(time.Minute)

			_, err := repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
				StripeSubscriptionID: subscriptionID,
				UserID:               userID,
				StripeCustomerID:     customerID,
				StripePriceID:        "price_premium",
				StripeEventCreatedAt: &latestEventCreatedAt,
				Status:               tt.latestStatus,
				CurrentPeriodStart:   &periodStart,
				CurrentPeriodEnd:     &periodEnd,
			})
			require.NoError(t, err)

			_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
				StripeSubscriptionID: subscriptionID,
				UserID:               userID,
				StripeCustomerID:     customerID,
				StripePriceID:        "price_premium",
				StripeEventCreatedAt: &olderEventCreatedAt,
				Status:               tt.olderStatus,
				CurrentPeriodStart:   &periodStart,
				CurrentPeriodEnd:     &periodEnd,
				GrantAIChatAccess:    true,
			})
			require.NoError(t, err)

			assert.Empty(t, activeAIChatSources(t, pool, userID))
			assert.Equal(t, tt.latestStatus, currentStripeSubscriptionStatus(t, pool, userID, subscriptionID))
		})
	}
}

func TestRepositoryUpsertSubscriptionFromWebhook_AppliesSameSecondEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	userID := "billing-same-second-user"
	subscriptionID := "sub_same_second"
	customerID := "cus_same_second"
	cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	defer cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	seedBillingUser(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	periodStart := time.Now().UTC().Add(-time.Hour)
	periodEnd := periodStart.Add(30 * 24 * time.Hour)
	eventCreatedAt := periodStart.Add(time.Minute).Truncate(time.Second)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               subscriptionStatusActive,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		GrantAIChatAccess:    true,
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"stripe"}, activeAIChatSources(t, pool, userID))

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               subscriptionStatusCanceled,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
	})
	require.NoError(t, err)

	assert.Empty(t, activeAIChatSources(t, pool, userID))
	assert.Equal(t, subscriptionStatusCanceled, currentStripeSubscriptionStatus(t, pool, userID, subscriptionID))
}

func TestRepositoryUpsertSubscriptionFromWebhook_SameSecondRevocationBeatsGrantingEvent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	userID := "billing-same-second-revoke-user"
	subscriptionID := "sub_same_second_revoke"
	customerID := "cus_same_second_revoke"
	cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	defer cleanupBillingSourceScopeTest(t, pool, userID, subscriptionID, customerID)
	seedBillingUser(t, pool, userID)

	repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
	periodStart := time.Now().UTC().Add(-time.Hour)
	periodEnd := periodStart.Add(30 * 24 * time.Hour)
	eventCreatedAt := periodStart.Add(time.Minute).Truncate(time.Second)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               subscriptionStatusCanceled,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
	})
	require.NoError(t, err)

	_, err = repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
		StripeSubscriptionID: subscriptionID,
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               subscriptionStatusActive,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		GrantAIChatAccess:    true,
	})
	require.NoError(t, err)

	assert.Empty(t, activeAIChatSources(t, pool, userID))
	assert.Equal(t, subscriptionStatusCanceled, currentStripeSubscriptionStatus(t, pool, userID, subscriptionID))
}

func TestRepositoryUpsertSubscriptionFromWebhook_DoesNotGrantWithoutPremiumPrice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping database-backed billing integration test without DATABASE_URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	require.NoError(t, pool.Ping(ctx))

	tests := []struct {
		name         string
		userID       string
		subscription string
		customerID   string
		priceID      string
	}{
		{
			name:         "wrong price",
			userID:       "billing-wrong-price-user",
			subscription: "sub_wrong_price",
			customerID:   "cus_wrong_price",
			priceID:      "price_other",
		},
		{
			name:         "missing price",
			userID:       "billing-missing-price-user",
			subscription: "sub_missing_price",
			customerID:   "cus_missing_price",
			priceID:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupBillingSourceScopeTest(t, pool, tt.userID, tt.subscription, tt.customerID)
			defer cleanupBillingSourceScopeTest(t, pool, tt.userID, tt.subscription, tt.customerID)
			seedBillingUser(t, pool, tt.userID)

			repo := NewRepository(slog.New(slog.NewTextHandler(io.Discard, nil)), db.New(pool), pool)
			periodStart := time.Now().UTC().Add(-time.Hour)
			periodEnd := periodStart.Add(30 * 24 * time.Hour)
			eventCreatedAt := periodStart.Add(time.Minute)

			_, err := repo.UpsertSubscriptionFromWebhook(ctx, StripeSubscriptionSnapshot{
				StripeSubscriptionID: tt.subscription,
				UserID:               tt.userID,
				StripeCustomerID:     tt.customerID,
				StripePriceID:        tt.priceID,
				StripeEventCreatedAt: &eventCreatedAt,
				Status:               subscriptionStatusActive,
				CurrentPeriodStart:   &periodStart,
				CurrentPeriodEnd:     &periodEnd,
				GrantAIChatAccess:    false,
			})
			require.NoError(t, err)

			assert.Empty(t, activeAIChatSources(t, pool, tt.userID))
		})
	}
}

func seedLocalE2EFeatureAccess(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	seedBillingUser(t, pool, userID)

	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		_, err := tx.Exec(context.Background(), `
			INSERT INTO user_feature_access (
				user_id, feature_key, source, source_reference, granted_by, note
			)
			VALUES (
				$1, 'ai_chatbot', 'local_e2e_auth', 'local-e2e-bootstrap',
				'fittrack-local-e2e', 'Local development E2E bootstrap access'
			)
		`, userID)
		require.NoError(t, err)
	})
}

func seedBillingUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	t.Helper()

	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		_, err := tx.Exec(context.Background(), `
			INSERT INTO users (user_id)
			VALUES ($1)
			ON CONFLICT (user_id) DO NOTHING
		`, userID)
		require.NoError(t, err)
	})
}

func activeAIChatSources(t *testing.T, pool *pgxpool.Pool, userID string) []string {
	t.Helper()

	var sources []string
	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		rows, err := tx.Query(context.Background(), `
			SELECT source
			FROM user_feature_access
			WHERE user_id = $1
			  AND feature_key = 'ai_chatbot'
			  AND revoked_at IS NULL
			  AND starts_at <= NOW()
			  AND (expires_at IS NULL OR expires_at > NOW())
			ORDER BY source
		`, userID)
		require.NoError(t, err)
		defer rows.Close()

		for rows.Next() {
			var source string
			require.NoError(t, rows.Scan(&source))
			sources = append(sources, source)
		}
		require.NoError(t, rows.Err())
	})

	return sources
}

func revokedStripeGrantCount(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string) int64 {
	t.Helper()

	var count int64
	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		err := tx.QueryRow(context.Background(), `
			SELECT COUNT(*)
			FROM user_feature_access
			WHERE user_id = $1
			  AND feature_key = 'ai_chatbot'
			  AND source = 'stripe'
			  AND source_reference = $2
			  AND revoked_at IS NOT NULL
		`, userID, subscriptionID).Scan(&count)
		require.NoError(t, err)
	})

	return count
}

func currentStripeSubscriptionStatus(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string) string {
	t.Helper()

	var status string
	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		err := tx.QueryRow(context.Background(), `
			SELECT status
			FROM stripe_subscriptions
			WHERE user_id = $1
			  AND stripe_subscription_id = $2
		`, userID, subscriptionID).Scan(&status)
		require.NoError(t, err)
	})

	return status
}

func stripeGrantExpiresAt(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string) time.Time {
	t.Helper()

	var expiresAt time.Time
	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		err := tx.QueryRow(context.Background(), `
			SELECT expires_at
			FROM user_feature_access
			WHERE user_id = $1
			  AND feature_key = 'ai_chatbot'
			  AND source = 'stripe'
			  AND source_reference = $2
			  AND revoked_at IS NULL
		`, userID, subscriptionID).Scan(&expiresAt)
		require.NoError(t, err)
	})

	return expiresAt.UTC()
}

func cleanupBillingSourceScopeTest(t *testing.T, pool *pgxpool.Pool, userID string, subscriptionID string, customerID string) {
	t.Helper()

	withBillingTestUser(t, pool, userID, func(tx pgx.Tx) {
		_, err := tx.Exec(context.Background(), `
			DELETE FROM ai_chat_trial_prompt_usage
			WHERE user_id = $1 OR stripe_subscription_id = $2
		`, userID, subscriptionID)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), `
			DELETE FROM user_feature_access
			WHERE user_id = $1
		`, userID)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), `
			DELETE FROM stripe_subscriptions
			WHERE user_id = $1 OR stripe_subscription_id = $2 OR stripe_customer_id = $3
		`, userID, subscriptionID, customerID)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), `
			DELETE FROM stripe_customers
			WHERE user_id = $1 OR stripe_customer_id = $2
		`, userID, customerID)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), `
			DELETE FROM users
			WHERE user_id = $1
		`, userID)
		require.NoError(t, err)
	})
}

func withBillingTestUser(t *testing.T, pool *pgxpool.Pool, userID string, fn func(tx pgx.Tx)) {
	t.Helper()

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "SELECT set_config('app.current_user_id', $1, true)", userID)
	require.NoError(t, err)

	fn(tx)

	require.NoError(t, tx.Commit(context.Background()))
}
