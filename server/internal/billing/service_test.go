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
	now := time.Date(2026, 5, 18, 16, 0, 0, 0, time.UTC)
	periodEnd := now.Add(30 * 24 * time.Hour)
	raw := subscriptionEventPayload(t, "sub_123", "trialing", false, now, periodEnd)

	service.constructEvent = func(payload []byte, header string, secret string) (stripe.Event, error) {
		assert.Equal(t, []byte("payload"), payload)
		assert.Equal(t, "sig", header)
		assert.Equal(t, "whsec_123", secret)
		return stripe.Event{
			ID:   "evt_123",
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{Raw: raw},
		}, nil
	}

	expectedSnapshot := StripeSubscriptionSnapshot{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        "price_premium",
		Status:               "trialing",
		CancelAtPeriodEnd:    false,
		CurrentPeriodStart:   &now,
		CurrentPeriodEnd:     &periodEnd,
		TrialStart:           &now,
		TrialEnd:             &periodEnd,
	}

	repo.On("HasProcessedWebhookEvent", mock.Anything, "evt_123").Return(false, nil).Once()
	repo.On("UpsertSubscriptionFromWebhook", mock.Anything, expectedSnapshot).Return(db.StripeSubscriptions{}, nil).Once()
	repo.On("MarkWebhookEventProcessed", mock.Anything, "evt_123", "customer.subscription.updated").Return(nil).Once()

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

func TestServiceCurrentStatus_AccessRules(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	now := time.Now().UTC()

	tests := []struct {
		name       string
		status     string
		canceling  bool
		periodEnd  time.Time
		wantAccess bool
	}{
		{
			name:       "active has access",
			status:     "active",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: true,
		},
		{
			name:       "cancel at period end keeps access until period end",
			status:     "active",
			canceling:  true,
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: true,
		},
		{
			name:       "ended cancellation blocks after period end",
			status:     "active",
			canceling:  true,
			periodEnd:  now.Add(-time.Hour),
			wantAccess: false,
		},
		{
			name:       "past_due blocks immediately",
			status:     "past_due",
			periodEnd:  now.Add(24 * time.Hour),
			wantAccess: false,
		},
		{
			name:       "canceled blocks",
			status:     "canceled",
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
				StripePriceID:        "price_premium",
				Status:               tt.status,
				CancelAtPeriodEnd:    tt.canceling,
				CurrentPeriodEnd:     pgtype.Timestamptz{Time: tt.periodEnd, Valid: true},
			}, nil).Once()

			resp, err := service.CurrentStatus(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.wantAccess, resp.HasAccess)
			require.NotNil(t, resp.Subscription)
			assert.Equal(t, tt.canceling, resp.Subscription.CancelAtPeriodEnd)
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
			"data": []map[string]any{
				{
					"id":                   "si_123",
					"object":               "subscription_item",
					"current_period_start": periodStart.Unix(),
					"current_period_end":   periodEnd.Unix(),
					"price": map[string]any{
						"id":     "price_premium",
						"object": "price",
					},
				},
			},
		},
	}

	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return raw
}
