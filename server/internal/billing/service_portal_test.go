package billing

import (
	"context"
	"io"
	"log/slog"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
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
		assert.Equal(t, "http://localhost:5173/chat", *params.ReturnURL)
		return &stripe.BillingPortalSession{URL: "https://billing.stripe.test/session"}, nil
	}

	resp, err := service.CreateCustomerPortalSession(ctx)

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

	resp, err := service.CreateCustomerPortalSession(ctx)

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

	resp, err := service.CreateCustomerPortalSession(ctx)

	require.ErrorIs(t, err, ErrBillingCustomerMissing)
	assert.Nil(t, resp)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "UpsertStripeCustomer", mock.Anything, mock.Anything, mock.Anything)
}
