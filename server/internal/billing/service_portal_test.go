package billing

import (
	"context"
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

func TestServiceCreateCustomerPortalSession(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetStripeCustomerByUserID", mock.Anything, "user-123").Return(db.StripeCustomers{
		UserID:           "user-123",
		StripeCustomerID: "cus_123",
	}, nil).Once()
	service.createPortalSession = func(params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		require.NotNil(t, params.Customer)
		require.NotNil(t, params.ReturnURL)
		assert.Equal(t, "cus_123", *params.Customer)
		assert.Equal(t, "http://localhost:5173/chat?billing=portal-return", *params.ReturnURL)
		return &stripe.BillingPortalSession{URL: "https://billing.stripe.test/session"}, nil
	}

	resp, err := service.CreateCustomerPortalSession(ctx, PortalReturnChat)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://billing.stripe.test/session", resp.URL)
	repo.AssertExpectations(t)
}

func TestServiceCreateCustomerPortalSession_SettingsReturnDestination(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetStripeCustomerByUserID", mock.Anything, "user-123").Return(db.StripeCustomers{
		UserID:           "user-123",
		StripeCustomerID: "cus_123",
	}, nil).Once()
	service.createPortalSession = func(params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		require.NotNil(t, params.ReturnURL)
		assert.Equal(t, "http://localhost:5173/settings", *params.ReturnURL)
		return &stripe.BillingPortalSession{URL: "https://billing.stripe.test/session"}, nil
	}

	resp, err := service.CreateCustomerPortalSession(ctx, PortalReturnSettings)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://billing.stripe.test/session", resp.URL)
	repo.AssertExpectations(t)
}

func TestServiceCreateCustomerPortalSession_NotConfigured(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	resp, err := service.CreateCustomerPortalSession(ctx, PortalReturnChat)

	require.ErrorIs(t, err, ErrBillingNotConfigured)
	assert.Nil(t, resp)
	repo.AssertNotCalled(t, "GetStripeCustomerByUserID", mock.Anything, mock.Anything)
}

func TestServiceCreateCustomerPortalSession_MissingCustomerDoesNotCreateStripeCustomer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetStripeCustomerByUserID", mock.Anything, "user-123").Return(db.StripeCustomers{}, pgx.ErrNoRows).Once()
	service.createCustomer = func(*stripe.CustomerParams) (*stripe.Customer, error) {
		t.Fatal("portal session should not create a Stripe customer")
		return nil, nil
	}
	service.createPortalSession = func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		t.Fatal("portal session should not open without an existing customer")
		return nil, nil
	}

	resp, err := service.CreateCustomerPortalSession(ctx, PortalReturnChat)

	require.ErrorIs(t, err, ErrBillingCustomerMissing)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "UpsertStripeCustomer", mock.Anything, mock.Anything, mock.Anything)
}

func TestServiceCreateSubscriptionCancelPortalSession(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        textToPg("price_premium"),
		Status:               subscriptionStatusActive,
	}, nil).Once()
	repo.On("GetStripeCustomerByUserID", mock.Anything, "user-123").Return(db.StripeCustomers{
		UserID:           "user-123",
		StripeCustomerID: "cus_123",
	}, nil).Once()
	service.createPortalSession = func(params *stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		require.NotNil(t, params.Customer)
		require.NotNil(t, params.ReturnURL)
		require.NotNil(t, params.FlowData)
		require.NotNil(t, params.FlowData.AfterCompletion)
		require.NotNil(t, params.FlowData.AfterCompletion.Redirect)
		require.NotNil(t, params.FlowData.SubscriptionCancel)
		require.NotNil(t, params.FlowData.SubscriptionCancel.Subscription)
		assert.Equal(t, "cus_123", *params.Customer)
		assert.Equal(t, "http://localhost:5173/chat?billing=portal-return", *params.ReturnURL)
		assert.Equal(t, string(stripe.BillingPortalSessionFlowTypeSubscriptionCancel), *params.FlowData.Type)
		assert.Equal(t, string(stripe.BillingPortalSessionFlowAfterCompletionTypeRedirect), *params.FlowData.AfterCompletion.Type)
		assert.Equal(t, "http://localhost:5173/chat?billing=cancelled", *params.FlowData.AfterCompletion.Redirect.ReturnURL)
		assert.Equal(t, "sub_123", *params.FlowData.SubscriptionCancel.Subscription)
		return &stripe.BillingPortalSession{URL: "https://billing.stripe.test/cancel-session"}, nil
	}

	resp, err := service.CreateSubscriptionCancelPortalSession(ctx)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "https://billing.stripe.test/cancel-session", resp.URL)
	repo.AssertExpectations(t)
}

func TestServiceCreateSubscriptionCancelPortalSession_MissingSubscription(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{}, pgx.ErrNoRows).Once()
	service.createPortalSession = func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		t.Fatal("portal session should not open without a cancelable subscription")
		return nil, nil
	}

	resp, err := service.CreateSubscriptionCancelPortalSession(ctx)

	require.ErrorIs(t, err, ErrBillingSubscriptionMissing)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetStripeCustomerByUserID", mock.Anything, mock.Anything)
}

func TestServiceCancelCurrentSubscriptionRenewal_SchedulesPeriodEndCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		Status:               subscriptionStatusActive,
	}, nil).Once()
	service.updateSubscription = func(id string, params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_123", id)
		require.NotNil(t, params.CancelAtPeriodEnd)
		assert.True(t, *params.CancelAtPeriodEnd)
		return &stripe.Subscription{ID: id}, nil
	}

	err := service.CancelCurrentSubscriptionRenewal(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionRenewal_NoSubscriptionNoops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{}, pgx.ErrNoRows).Once()
	service.updateSubscription = func(string, *stripe.SubscriptionParams) (*stripe.Subscription, error) {
		t.Fatal("missing subscriptions should not call Stripe")
		return nil, nil
	}

	err := service.CancelCurrentSubscriptionRenewal(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionRenewal_AlreadyCancelingNoops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		Status:               subscriptionStatusActive,
		CancelAtPeriodEnd:    true,
	}, nil).Once()
	service.updateSubscription = func(string, *stripe.SubscriptionParams) (*stripe.Subscription, error) {
		t.Fatal("already canceling subscriptions should not call Stripe")
		return nil, nil
	}

	err := service.CancelCurrentSubscriptionRenewal(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionRenewal_InactiveSubscriptionNoops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		Status:               subscriptionStatusCanceled,
	}, nil).Once()
	service.updateSubscription = func(string, *stripe.SubscriptionParams) (*stripe.Subscription, error) {
		t.Fatal("inactive subscriptions should not call Stripe")
		return nil, nil
	}

	err := service.CancelCurrentSubscriptionRenewal(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionRenewal_NotConfiguredWhenCancellationRequired(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		Status:               subscriptionStatusTrialing,
	}, nil).Once()
	service.updateSubscription = func(string, *stripe.SubscriptionParams) (*stripe.Subscription, error) {
		t.Fatal("unconfigured billing should not call Stripe")
		return nil, nil
	}

	err := service.CancelCurrentSubscriptionRenewal(ctx)

	require.ErrorIs(t, err, ErrBillingNotConfigured)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionImmediately_CancelsActiveSubscription(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		Status:               subscriptionStatusActive,
	}, nil).Once()
	service.cancelSubscription = func(id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
		assert.Equal(t, "sub_123", id)
		require.NotNil(t, params)
		return &stripe.Subscription{ID: id, Status: stripe.SubscriptionStatusCanceled}, nil
	}

	err := service.CancelCurrentSubscriptionImmediately(ctx)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCancelCurrentSubscriptionImmediately_CancelsRecoverableSubscription(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, status := range []string{
		subscriptionStatusPastDue,
		subscriptionStatusIncomplete,
		subscriptionStatusPaused,
	} {
		t.Run(status, func(t *testing.T) {
			repo := new(mockRepository)
			service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
			ctx := user.WithContext(context.Background(), "user-123")

			repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
				StripeSubscriptionID: "sub_cancelable",
				UserID:               "user-123",
				StripeCustomerID:     "cus_123",
				Status:               status,
			}, nil).Once()
			service.cancelSubscription = func(id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
				assert.Equal(t, "sub_cancelable", id)
				require.NotNil(t, params)
				return &stripe.Subscription{ID: id, Status: stripe.SubscriptionStatusCanceled}, nil
			}

			err := service.CancelCurrentSubscriptionImmediately(ctx)

			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestServiceCancelCurrentSubscriptionImmediately_NonBillableSubscriptionNoops(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, status := range []string{
		subscriptionStatusCanceled,
		subscriptionStatusIncompleteExpired,
		subscriptionStatusUnpaid,
	} {
		t.Run(status, func(t *testing.T) {
			repo := new(mockRepository)
			service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
			ctx := user.WithContext(context.Background(), "user-123")

			repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
				StripeSubscriptionID: "sub_non_billable",
				UserID:               "user-123",
				StripeCustomerID:     "cus_123",
				Status:               status,
			}, nil).Once()
			service.cancelSubscription = func(string, *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
				t.Fatal("non-billable subscriptions should not call Stripe")
				return nil, nil
			}

			err := service.CancelCurrentSubscriptionImmediately(ctx)

			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestServiceCancelCurrentSubscriptionImmediately_StripeFailureBlocksAccountDeletion(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		Status:               subscriptionStatusActive,
	}, nil).Once()
	service.cancelSubscription = func(id string, params *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
		return nil, assert.AnError
	}

	err := service.CancelCurrentSubscriptionImmediately(ctx)

	require.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}

func TestServiceCreateSubscriptionCancelPortalSession_AlreadyCanceling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        textToPg("price_premium"),
		Status:               subscriptionStatusActive,
		CancelAtPeriodEnd:    true,
	}, nil).Once()
	service.createPortalSession = func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		t.Fatal("portal session should not open for a subscription already set to cancel")
		return nil, nil
	}

	resp, err := service.CreateSubscriptionCancelPortalSession(ctx)

	require.ErrorIs(t, err, ErrBillingSubscriptionNotCancelable)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetStripeCustomerByUserID", mock.Anything, mock.Anything)
}

func TestServiceCreateSubscriptionCancelPortalSession_AlreadyCancelingWithCancelAt(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := new(mockRepository)
	service := NewService(logger, repo, "sk_test_123", "whsec_123", "price_premium", "http://localhost:5173", 30)
	ctx := user.WithContext(context.Background(), "user-123")

	repo.On("GetCurrentSubscriptionByUserID", mock.Anything, "user-123").Return(db.StripeSubscriptions{
		StripeSubscriptionID: "sub_123",
		UserID:               "user-123",
		StripeCustomerID:     "cus_123",
		StripePriceID:        textToPg("price_premium"),
		Status:               subscriptionStatusActive,
		CancelAt:             pgtype.Timestamptz{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true},
	}, nil).Once()
	service.createPortalSession = func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		t.Fatal("portal session should not open for a subscription already set to cancel")
		return nil, nil
	}

	resp, err := service.CreateSubscriptionCancelPortalSession(ctx)

	require.ErrorIs(t, err, ErrBillingSubscriptionNotCancelable)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetStripeCustomerByUserID", mock.Anything, mock.Anything)
}
