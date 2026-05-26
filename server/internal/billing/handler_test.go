package billing

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubBillingService struct {
	createCheckoutSession       func(context.Context) (*CheckoutSessionResponse, error)
	createCustomerPortalSession func(context.Context) (*CustomerPortalSessionResponse, error)
	currentStatus               func(context.Context) (*StatusResponse, error)
	handleWebhook               func(context.Context, []byte, string) error
}

func (s stubBillingService) CreateCheckoutSession(ctx context.Context) (*CheckoutSessionResponse, error) {
	if s.createCheckoutSession != nil {
		return s.createCheckoutSession(ctx)
	}
	return &CheckoutSessionResponse{URL: "https://checkout.stripe.test/session"}, nil
}

func (s stubBillingService) CreateCustomerPortalSession(ctx context.Context) (*CustomerPortalSessionResponse, error) {
	if s.createCustomerPortalSession != nil {
		return s.createCustomerPortalSession(ctx)
	}
	return &CustomerPortalSessionResponse{URL: "https://billing.stripe.test/session"}, nil
}

func (s stubBillingService) CurrentStatus(ctx context.Context) (*StatusResponse, error) {
	if s.currentStatus != nil {
		return s.currentStatus(ctx)
	}
	return &StatusResponse{FeatureKey: FeatureKeyAIChatbot}, nil
}

func (s stubBillingService) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	if s.handleWebhook != nil {
		return s.handleWebhook(ctx, payload, signatureHeader)
	}
	return nil
}

func TestHandlerCreateCustomerPortalSession_ReturnsPortalURL(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/customer-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateCustomerPortalSession(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"url":"https://billing.stripe.test/session"}`, rr.Body.String())
}

func TestHandlerCreateCustomerPortalSession_NotConfiguredReturns503(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{
		createCustomerPortalSession: func(context.Context) (*CustomerPortalSessionResponse, error) {
			return nil, ErrBillingNotConfigured
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/customer-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateCustomerPortalSession(rr, req)

	require.Equal(t, http.StatusServiceUnavailable, rr.Code)
	assert.Contains(t, rr.Body.String(), "billing is not configured")
}

func TestHandlerCreateCustomerPortalSession_MissingCustomerReturns404(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{
		createCustomerPortalSession: func(context.Context) (*CustomerPortalSessionResponse, error) {
			return nil, ErrBillingCustomerMissing
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/customer-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateCustomerPortalSession(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "billing customer was not found")
}
