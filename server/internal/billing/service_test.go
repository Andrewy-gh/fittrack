package billing

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetStripeCustomerByUserID(ctx context.Context, userID string) (db.StripeCustomers, error) {
	args := m.Called(ctx, userID)
	row, _ := args.Get(0).(db.StripeCustomers)
	return row, args.Error(1)
}

func (m *mockRepository) GetStripeCustomerByCustomerID(ctx context.Context, stripeCustomerID string) (db.StripeCustomers, error) {
	args := m.Called(ctx, stripeCustomerID)
	row, _ := args.Get(0).(db.StripeCustomers)
	return row, args.Error(1)
}

func (m *mockRepository) UpsertStripeCustomer(ctx context.Context, userID string, stripeCustomerID string) (db.StripeCustomers, error) {
	args := m.Called(ctx, userID, stripeCustomerID)
	row, _ := args.Get(0).(db.StripeCustomers)
	return row, args.Error(1)
}

func (m *mockRepository) GetCurrentSubscriptionByUserID(ctx context.Context, userID string) (db.StripeSubscriptions, error) {
	args := m.Called(ctx, userID)
	row, _ := args.Get(0).(db.StripeSubscriptions)
	return row, args.Error(1)
}

func (m *mockRepository) UpsertSubscriptionFromWebhook(ctx context.Context, snapshot StripeSubscriptionSnapshot) (db.StripeSubscriptions, error) {
	args := m.Called(ctx, snapshot)
	row, _ := args.Get(0).(db.StripeSubscriptions)
	return row, args.Error(1)
}

func (m *mockRepository) HasProcessedWebhookEvent(ctx context.Context, stripeEventID string) (bool, error) {
	args := m.Called(ctx, stripeEventID)
	return args.Bool(0), args.Error(1)
}

func (m *mockRepository) MarkWebhookEventProcessed(ctx context.Context, stripeEventID string, eventType string) error {
	args := m.Called(ctx, stripeEventID, eventType)
	return args.Error(0)
}

func (m *mockRepository) GetTrialPromptUsage(ctx context.Context, userID string, stripeSubscriptionID string) (db.AiChatTrialPromptUsage, error) {
	args := m.Called(ctx, userID, stripeSubscriptionID)
	row, _ := args.Get(0).(db.AiChatTrialPromptUsage)
	return row, args.Error(1)
}

func (m *mockRepository) ConsumeTrialPrompt(ctx context.Context, userID string, stripeSubscriptionID string, cap int32) (db.AiChatTrialPromptUsage, error) {
	args := m.Called(ctx, userID, stripeSubscriptionID, cap)
	row, _ := args.Get(0).(db.AiChatTrialPromptUsage)
	return row, args.Error(1)
}

func TestServiceHandleWebhook_SubscriptionUpdatedPersistsSnapshotAndEventID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	now := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	eventCreatedAt := now.Add(time.Minute)
	periodEnd := now.Add(30 * 24 * time.Hour)
	raw := subscriptionEventPayload(t, "sub_123", "trialing", false, now, periodEnd)

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		assert.Equal(t, []byte("payload"), payload)
		assert.Equal(t, "sig", header)
		assert.Equal(t, "whsec_123", secret)
		return stripe.Event{
			ID:      "evt_123",
			Type:    "customer.subscription.updated",
			Created: eventCreatedAt.Unix(),
			Data:    &stripe.EventData{Raw: raw},
		}, nil
	}

	expectedSnapshot := StripeSubscriptionSnapshot{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               "trialing",
		CancelAtPeriodEnd:    false,
		CurrentPeriodStart:   &now,
		CurrentPeriodEnd:     &periodEnd,
		TrialStart:           &now,
		TrialEnd:             &periodEnd,
		GrantAIChatAccess:    true,
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_123").Return(false, nil).Once()
	repo.On("UpsertSubscriptionFromWebhook", mock.Anything, expectedSnapshot).Return(db.StripeSubscriptions{}, nil).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_123", "customer.subscription.updated").Return(nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceHandleWebhook_DoesNotGrantAccessAfterAccessEnd(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	now := time.Now().UTC().Truncate(time.Second)
	periodStart := now.Add(-48 * time.Hour)
	cancelAt := now.Add(-time.Hour)
	periodEnd := now.Add(24 * time.Hour)
	eventCreatedAt := now
	raw := subscriptionEventPayloadWithCancelAt(t, "sub_cancel_expired", "active", cancelAt, periodStart, periodEnd)

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		return stripe.Event{
			ID:      "evt_cancel_expired",
			Type:    "customer.subscription.updated",
			Created: eventCreatedAt.Unix(),
			Data:    &stripe.EventData{Raw: raw},
		}, nil
	}

	expectedSnapshot := StripeSubscriptionSnapshot{
		StripeSubscriptionID: "sub_cancel_expired",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               "active",
		CancelAt:             &cancelAt,
		CurrentPeriodStart:   &periodStart,
		CurrentPeriodEnd:     &periodEnd,
		TrialStart:           &periodStart,
		TrialEnd:             &periodEnd,
		GrantAIChatAccess:    false,
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_cancel_expired").Return(false, nil).Once()
	repo.On("UpsertSubscriptionFromWebhook", mock.Anything, expectedSnapshot).Return(db.StripeSubscriptions{}, nil).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_cancel_expired", "customer.subscription.updated").Return(nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceHandleWebhook_SkipsAlreadyProcessedEvent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		return stripe.Event{
			ID:   "evt_done",
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{Raw: subscriptionEventPayload(t, "sub_123", "active", false, time.Now(), time.Now().Add(time.Hour))},
		}, nil
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_done").Return(true, nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertNotCalled(t, "UpsertSubscriptionFromWebhook", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "MarkWebhookEventProcessed", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestServiceHandleWebhook_DoesNotGrantAccessForWrongOrMissingPrice(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	periodEnd := now.Add(30 * 24 * time.Hour)
	eventCreatedAt := now.Add(time.Minute)

	tests := []struct {
		name          string
		priceID       *string
		wantPriceID   string
		subscription  string
		stripeEventID string
	}{
		{
			name:          "active subscription with wrong price",
			priceID:       stringPtr("price_other"),
			wantPriceID:   "price_other",
			subscription:  "sub_wrong_price",
			stripeEventID: "evt_wrong_price",
		},
		{
			name:          "trialing subscription with missing price",
			priceID:       nil,
			wantPriceID:   "",
			subscription:  "sub_missing_price",
			stripeEventID: "evt_missing_price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
			raw := subscriptionEventPayloadWithPrice(t, tt.subscription, "active", false, now, periodEnd, tt.priceID)

			service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
				return stripe.Event{
					ID:      tt.stripeEventID,
					Type:    "customer.subscription.updated",
					Created: eventCreatedAt.Unix(),
					Data:    &stripe.EventData{Raw: raw},
				}, nil
			}

			expectedSnapshot := StripeSubscriptionSnapshot{
				StripeSubscriptionID: tt.subscription,
				UserID:               "user-123",
				StripeCustomerID:     "cus_123",
				StripePriceID:        tt.wantPriceID,
				StripeEventCreatedAt: &eventCreatedAt,
				Status:               "active",
				CurrentPeriodStart:   &now,
				CurrentPeriodEnd:     &periodEnd,
				TrialStart:           &now,
				TrialEnd:             &periodEnd,
				GrantAIChatAccess:    false,
			}

			repo.On("HasProcessedWebhookEvent", mock.Anything, tt.stripeEventID).Return(false, nil).Once()
			repo.On("UpsertSubscriptionFromWebhook", mock.Anything, expectedSnapshot).Return(db.StripeSubscriptions{}, nil).Once()
			repo.On("MarkWebhookEventProcessed", mock.Anything, tt.stripeEventID, "customer.subscription.updated").Return(nil).Once()

			err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestServiceHandleWebhook_AcknowledgesDeletedAccountSubscriptionEvent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	now := time.Now().UTC().Truncate(time.Second)
	eventCreatedAt := now.Add(time.Minute)
	raw := subscriptionEventPayloadWithoutUserMetadata(t, "sub_deleted", "canceled", now, now.Add(24*time.Hour))

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		return stripe.Event{
			ID:      "evt_deleted_account_subscription",
			Type:    "customer.subscription.deleted",
			Created: eventCreatedAt.Unix(),
			Data:    &stripe.EventData{Raw: raw},
		}, nil
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_deleted_account_subscription").Return(false, nil).Once()
	repo.On("GetStripeCustomerByCustomerID", mock.Anything, "cus_123").Return(db.StripeCustomers{}, pgx.ErrNoRows).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_deleted_account_subscription", "customer.subscription.deleted").Return(nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertNotCalled(t, "UpsertSubscriptionFromWebhook", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "UpsertStripeCustomer", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestServiceHandleWebhook_AcknowledgesDeletedAccountSubscriptionEventWithUserMetadata(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	now := time.Now().UTC().Truncate(time.Second)
	eventCreatedAt := now.Add(time.Minute)
	periodEnd := now.Add(24 * time.Hour)
	raw := subscriptionEventPayload(t, "sub_deleted_with_metadata", "canceled", false, now, periodEnd)

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		return stripe.Event{
			ID:      "evt_deleted_account_subscription_with_metadata",
			Type:    "customer.subscription.deleted",
			Created: eventCreatedAt.Unix(),
			Data:    &stripe.EventData{Raw: raw},
		}, nil
	}

	expectedSnapshot := StripeSubscriptionSnapshot{
		StripeSubscriptionID: "sub_deleted_with_metadata",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        "price_premium",
		StripeEventCreatedAt: &eventCreatedAt,
		Status:               "canceled",
		CurrentPeriodStart:   &now,
		CurrentPeriodEnd:     &periodEnd,
		TrialStart:           &now,
		TrialEnd:             &periodEnd,
		GrantAIChatAccess:    false,
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_deleted_account_subscription_with_metadata").Return(false, nil).Once()
	repo.On("UpsertSubscriptionFromWebhook", mock.Anything, expectedSnapshot).Return(db.StripeSubscriptions{}, ErrBillingAccountDeleted).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_deleted_account_subscription_with_metadata", "customer.subscription.deleted").Return(nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertNotCalled(t, "UpsertStripeCustomer", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestServiceHandleWebhook_AcknowledgesDeletedAccountCheckoutEvent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		return stripe.Event{
			ID:   "evt_deleted_account_checkout",
			Type: "checkout.session.completed",
			Data: &stripe.EventData{Raw: []byte(`{
				"id": "cs_deleted",
				"object": "checkout.session",
				"client_reference_id": "user-123",
				"customer": "cus_123"
			}`)},
		}, nil
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_deleted_account_checkout").Return(false, nil).Once()
	repo.On("UpsertStripeCustomer", mock.Anything, "user-123", "cus_123").Return(db.StripeCustomers{}, ErrBillingAccountDeleted).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_deleted_account_checkout", "checkout.session.completed").Return(nil).Once()

	err := service.HandleWebhook(context.Background(), []byte("payload"), "sig")

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCurrentStatus_AccessRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := time.Now().UTC()

	tests := []struct {
		name       string
		status     string
		priceID    string
		canceling  bool
		cancelAt   *time.Time
		periodEnd  time.Time
		wantAccess bool
	}{
		{
			name:       "active has access",
			status:     "active",
			priceID:    "price_premium",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: true,
		},
		{
			name:       "cancel at period end keeps access until period end",
			status:     "active",
			priceID:    "price_premium",
			canceling:  true,
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: true,
		},
		{
			name:       "scheduled cancel_at keeps access until cancel_at",
			status:     "active",
			priceID:    "price_premium",
			cancelAt:   timePtr(now.Add(24 * time.Hour)),
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: true,
		},
		{
			name:       "expired scheduled cancel_at blocks before later period end",
			status:     "active",
			priceID:    "price_premium",
			cancelAt:   timePtr(now.Add(-time.Hour)),
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: false,
		},
		{
			name:       "expired period end blocks before later cancel_at",
			status:     "active",
			priceID:    "price_premium",
			cancelAt:   timePtr(now.Add(24 * time.Hour)),
			periodEnd:  now.Add(-time.Hour),
			wantAccess: false,
		},
		{
			name:       "ended cancellation blocks after period end",
			status:     "active",
			priceID:    "price_premium",
			canceling:  true,
			periodEnd:  now.Add(-time.Hour),
			wantAccess: false,
		},
		{
			name:       "active wrong price blocks",
			status:     "active",
			priceID:    "price_other",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: false,
		},
		{
			name:       "past_due blocks immediately",
			status:     "past_due",
			priceID:    "price_premium",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: false,
		},
		{
			name:       "canceled blocks",
			status:     "canceled",
			priceID:    "price_premium",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
			ctx := user.WithContext(context.Background(), "user-123")
			repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
				StripeSubscriptionID: "sub_123",
				UserID:               "user-123",
				StripeCustomerID:     "cus_123",
				StripePriceID:        textToPg(tt.priceID),
				Status:               tt.status,
				CancelAtPeriodEnd:    tt.canceling,
				CancelAt:             timePtrToPg(tt.cancelAt),
				CurrentPeriodEnd:     pgtype.Timestamptz{Time: tt.periodEnd, Valid: true},
			}, nil).Once()

			resp, err := service.CurrentStatus(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.wantAccess, resp.HasAccess)
			require.NotNil(t, resp.Subscription)
			assert.Equal(t, tt.canceling || tt.cancelAt != nil, resp.Subscription.CancellationScheduled)
			if tt.cancelAt != nil && tt.cancelAt.Before(tt.periodEnd) {
				require.NotNil(t, resp.Subscription.AccessEndsAt)
				assert.Equal(t, tt.cancelAt.UTC(), resp.Subscription.AccessEndsAt.UTC())
			} else {
				require.NotNil(t, resp.Subscription.AccessEndsAt)
				assert.Equal(t, tt.periodEnd.UTC(), resp.Subscription.AccessEndsAt.UTC())
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestServiceEnsureAIChatPromptAllowed_TrialCap(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("trialing consumes a prompt", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
		ctx := user.WithContext(context.Background(), "user-123")

		repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
			StripeSubscriptionID: "sub_123",
			UserID:               "user-123",
			Status:               "trialing",
		}, nil).Once()
		repo.On("ConsumeTrialPrompt", mock.Anything, "user-123", "sub_123", int32(30)).Return(db.AiChatTrialPromptUsage{
			PromptCount: 30,
		}, nil).Once()

		err := service.EnsureAIChatPromptAllowed(ctx)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("cap reached blocks", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
		ctx := user.WithContext(context.Background(), "user-123")

		repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
			StripeSubscriptionID: "sub_123",
			UserID:               "user-123",
			Status:               "trialing",
		}, nil).Once()
		repo.On("ConsumeTrialPrompt", mock.Anything, "user-123", "sub_123", int32(30)).Return(db.AiChatTrialPromptUsage{}, ErrTrialPromptLimitExceeded).Once()

		err := service.EnsureAIChatPromptAllowed(ctx)

		require.ErrorIs(t, err, ErrTrialPromptLimitExceeded)
		repo.AssertExpectations(t)
	})

	t.Run("cap two allows two prompts and blocks the third", func(t *testing.T) {
		repo := new(mockRepository)
		service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 2)
		ctx := user.WithContext(context.Background(), "user-123")
		subscription := db.StripeSubscriptions{
			StripeSubscriptionID: "sub_123",
			UserID:               "user-123",
			Status:               "trialing",
		}

		repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(subscription, nil).Times(3)
		repo.On("ConsumeTrialPrompt", mock.Anything, "user-123", "sub_123", int32(2)).
			Return(db.AiChatTrialPromptUsage{PromptCount: 1}, nil).
			Once()
		repo.On("ConsumeTrialPrompt", mock.Anything, "user-123", "sub_123", int32(2)).
			Return(db.AiChatTrialPromptUsage{PromptCount: 2}, nil).
			Once()
		repo.On("ConsumeTrialPrompt", mock.Anything, "user-123", "sub_123", int32(2)).
			Return(db.AiChatTrialPromptUsage{}, ErrTrialPromptLimitExceeded).
			Once()

		require.NoError(t, service.EnsureAIChatPromptAllowed(ctx))
		require.NoError(t, service.EnsureAIChatPromptAllowed(ctx))
		require.ErrorIs(t, service.EnsureAIChatPromptAllowed(ctx), ErrTrialPromptLimitExceeded)
		repo.AssertExpectations(t)
	})
}

func TestServiceCurrentStatus_NoSubscription(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{}, pgx.ErrNoRows).Once()

	resp, err := service.CurrentStatus(ctx)

	require.NoError(t, err)
	assert.False(t, resp.HasAccess)
	assert.Nil(t, resp.Subscription)
	repo.AssertExpectations(t)
}

func subscriptionEventPayload(t *testing.T, subscriptionID string, status string, cancelAtPeriodEnd bool, periodStart time.Time, periodEnd time.Time) []byte {
	t.Helper()
	return subscriptionEventPayloadWithPrice(t, subscriptionID, status, cancelAtPeriodEnd, periodStart, periodEnd, stringPtr("price_premium"))
}

func subscriptionEventPayloadWithPrice(t *testing.T, subscriptionID string, status string, cancelAtPeriodEnd bool, periodStart time.Time, periodEnd time.Time, priceID *string) []byte {
	t.Helper()

	item := map[string]any{
		"id":                   "si_123",
		"object":               "subscription_item",
		"current_period_start": periodStart.Unix(),
		"current_period_end":   periodEnd.Unix(),
	}
	if priceID != nil {
		item["price"] = map[string]any{
			"id":     *priceID,
			"object": "price",
		}
	}

	payload := map[string]any{
		"id":                   subscriptionID,
		"object":               "subscription",
		"status":               status,
		"cancel_at_period_end": cancelAtPeriodEnd,
		"customer":             "cus_123",
		"metadata": map[string]string{
			"fittrack_user_id": "user-123",
		},
		"trial_start": periodStart.Unix(),
		"trial_end":   periodEnd.Unix(),
		"items": map[string]any{
			"object": "list",
			"data":   []map[string]any{item},
		},
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return raw
}

func subscriptionEventPayloadWithCancelAt(t *testing.T, subscriptionID string, status string, cancelAt time.Time, periodStart time.Time, periodEnd time.Time) []byte {
	t.Helper()

	var payload map[string]any
	require.NoError(t, json.Unmarshal(subscriptionEventPayload(t, subscriptionID, status, false, periodStart, periodEnd), &payload))
	payload["cancel_at"] = cancelAt.Unix()

	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return raw
}

func subscriptionEventPayloadWithoutUserMetadata(t *testing.T, subscriptionID string, status string, periodStart time.Time, periodEnd time.Time) []byte {
	t.Helper()

	var payload map[string]any
	require.NoError(t, json.Unmarshal(subscriptionEventPayload(t, subscriptionID, status, false, periodStart, periodEnd), &payload))
	payload["metadata"] = map[string]string{}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return raw
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
