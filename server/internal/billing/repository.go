package billing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetStripeCustomerByUserID(ctx context.Context, userID string) (db.StripeCustomers, error)
	GetStripeCustomerByCustomerID(ctx context.Context, stripeCustomerID string) (db.StripeCustomers, error)
	UpsertStripeCustomer(ctx context.Context, userID string, stripeCustomerID string) (db.StripeCustomers, error)
	GetCurrentSubscriptionByUserID(ctx context.Context, userID string) (db.StripeSubscriptions, error)
	UpsertSubscriptionFromWebhook(ctx context.Context, snapshot StripeSubscriptionSnapshot) (db.StripeSubscriptions, error)
	HasProcessedWebhookEvent(ctx context.Context, stripeEventID string) (bool, error)
	MarkWebhookEventProcessed(ctx context.Context, stripeEventID string, eventType string) error
	GetTrialPromptUsage(ctx context.Context, userID string, stripeSubscriptionID string) (db.AiChatTrialPromptUsage, error)
	ConsumeTrialPrompt(ctx context.Context, userID string, stripeSubscriptionID string, cap int32) (db.AiChatTrialPromptUsage, error)
}

type repository struct {
	logger  *slog.Logger
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(logger *slog.Logger, queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &repository{
		logger:  logger,
		queries: queries,
		pool:    pool,
	}
}

func (r *repository) GetStripeCustomerByUserID(ctx context.Context, userID string) (db.StripeCustomers, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.GetStripeCustomerByUserID(ctx, userID)
}

func (r *repository) GetStripeCustomerByCustomerID(ctx context.Context, stripeCustomerID string) (db.StripeCustomers, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.GetStripeCustomerByCustomerID(ctx, stripeCustomerID)
}

func (r *repository) UpsertStripeCustomer(ctx context.Context, userID string, stripeCustomerID string) (db.StripeCustomers, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if r.pool == nil {
		return r.queries.UpsertStripeCustomer(ctx, db.UpsertStripeCustomerParams{
			UserID:           userID,
			StripeCustomerID: stripeCustomerID,
		})
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.StripeCustomers{}, fmt.Errorf("begin stripe customer transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := setCurrentUser(ctx, tx, userID); err != nil {
		return db.StripeCustomers{}, err
	}

	row, err := r.queries.WithTx(tx).UpsertStripeCustomer(ctx, db.UpsertStripeCustomerParams{
		UserID:           userID,
		StripeCustomerID: stripeCustomerID,
	})
	if err != nil {
		return db.StripeCustomers{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return db.StripeCustomers{}, fmt.Errorf("commit stripe customer transaction: %w", err)
	}

	return row, nil
}

func (r *repository) GetCurrentSubscriptionByUserID(ctx context.Context, userID string) (db.StripeSubscriptions, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.GetCurrentStripeSubscriptionByUserID(ctx, userID)
}

func (r *repository) UpsertSubscriptionFromWebhook(ctx context.Context, snapshot StripeSubscriptionSnapshot) (db.StripeSubscriptions, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.StripeSubscriptions{}, fmt.Errorf("begin stripe subscription transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := setCurrentUser(ctx, tx, snapshot.UserID); err != nil {
		return db.StripeSubscriptions{}, err
	}

	qtx := r.queries.WithTx(tx)
	if _, err := qtx.UpsertStripeCustomer(ctx, db.UpsertStripeCustomerParams{
		UserID:           snapshot.UserID,
		StripeCustomerID: snapshot.StripeCustomerID,
	}); err != nil {
		return db.StripeSubscriptions{}, fmt.Errorf("upsert stripe customer from subscription: %w", err)
	}

	row, err := qtx.UpsertStripeSubscription(ctx, db.UpsertStripeSubscriptionParams{
		StripeSubscriptionID: snapshot.StripeSubscriptionID,
		UserID:               snapshot.UserID,
		StripeCustomerID:     snapshot.StripeCustomerID,
		StripePriceID:        snapshot.StripePriceID,
		StripeEventCreatedAt: timePtrToPg(snapshot.StripeEventCreatedAt),
		Status:               snapshot.Status,
		CancelAtPeriodEnd:    snapshot.CancelAtPeriodEnd,
		CancelAt:             timePtrToPg(snapshot.CancelAt),
		CurrentPeriodStart:   timePtrToPg(snapshot.CurrentPeriodStart),
		CurrentPeriodEnd:     timePtrToPg(snapshot.CurrentPeriodEnd),
		TrialStart:           timePtrToPg(snapshot.TrialStart),
		TrialEnd:             timePtrToPg(snapshot.TrialEnd),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		r.logger.Debug("ignored stale stripe subscription event", "stripe_subscription_id", snapshot.StripeSubscriptionID)
		return db.StripeSubscriptions{}, nil
	}
	if err != nil {
		return db.StripeSubscriptions{}, fmt.Errorf("upsert stripe subscription: %w", err)
	}

	if err := qtx.RevokeStripeFeatureAccess(ctx, db.RevokeStripeFeatureAccessParams{
		UserID:          snapshot.UserID,
		FeatureKey:      FeatureKeyAIChatbot,
		SourceReference: textToPg(snapshot.StripeSubscriptionID),
	}); err != nil {
		return db.StripeSubscriptions{}, fmt.Errorf("revoke previous stripe feature access: %w", err)
	}

	if snapshot.GrantAIChatAccess {
		if err := qtx.GrantStripeFeatureAccess(ctx, db.GrantStripeFeatureAccessParams{
			UserID:          snapshot.UserID,
			FeatureKey:      FeatureKeyAIChatbot,
			SourceReference: textToPg(snapshot.StripeSubscriptionID),
			Note:            textToPg("FitTrack premium AI chat"),
			ExpiresAt:       timePtrToPg(snapshot.CurrentPeriodEnd),
		}); err != nil {
			return db.StripeSubscriptions{}, fmt.Errorf("grant stripe feature access: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.StripeSubscriptions{}, fmt.Errorf("commit stripe subscription transaction: %w", err)
	}

	return row, nil
}

func (r *repository) HasProcessedWebhookEvent(ctx context.Context, stripeEventID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.HasProcessedStripeWebhookEvent(ctx, stripeEventID)
}

func (r *repository) MarkWebhookEventProcessed(ctx context.Context, stripeEventID string, eventType string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.MarkStripeWebhookEventProcessed(ctx, db.MarkStripeWebhookEventProcessedParams{
		StripeEventID: stripeEventID,
		EventType:     eventType,
	})
}

func (r *repository) GetTrialPromptUsage(ctx context.Context, userID string, stripeSubscriptionID string) (db.AiChatTrialPromptUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return r.queries.GetAIChatTrialPromptUsage(ctx, db.GetAIChatTrialPromptUsageParams{
		UserID:               userID,
		StripeSubscriptionID: stripeSubscriptionID,
	})
}

func (r *repository) ConsumeTrialPrompt(ctx context.Context, userID string, stripeSubscriptionID string, cap int32) (db.AiChatTrialPromptUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	row, err := r.queries.ConsumeAIChatTrialPrompt(ctx, db.ConsumeAIChatTrialPromptParams{
		UserID:               userID,
		StripeSubscriptionID: stripeSubscriptionID,
		PromptCount:          cap,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.AiChatTrialPromptUsage{}, ErrTrialPromptLimitExceeded
	}
	return row, err
}

func setCurrentUser(ctx context.Context, tx pgx.Tx, userID string) error {
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID)
	if err != nil {
		return fmt.Errorf("set billing rls user: %w", err)
	}
	return nil
}

func timePtrToPg(value *time.Time) pgtype.Timestamptz {
	if value == nil || value.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func textToPg(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

var _ Repository = (*repository)(nil)
