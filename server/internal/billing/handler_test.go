package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v83/webhook"
)

type stubBillingService struct {
	createCheckoutSession                 func(context.Context) (*CheckoutSessionResponse, error)
	createCustomerPortalSession           func(context.Context, PortalReturnDestination) (*CustomerPortalSessionResponse, error)
	createSubscriptionCancelPortalSession func(context.Context) (*CustomerPortalSessionResponse, error)
	currentStatus                         func(context.Context) (*StatusResponse, error)
	handleWebhook                         func(context.Context, []byte, string) error
}

func (s stubBillingService) CreateCheckoutSession(ctx context.Context) (*CheckoutSessionResponse, error) {
	if s.createCheckoutSession != nil {
		return s.createCheckoutSession(ctx)
	}
	return &CheckoutSessionResponse{URL: "https://checkout.stripe.test/session"}, nil
}

func (s stubBillingService) CreateCustomerPortalSession(ctx context.Context, destination PortalReturnDestination) (*CustomerPortalSessionResponse, error) {
	if s.createCustomerPortalSession != nil {
		return s.createCustomerPortalSession(ctx, destination)
	}
	return &CustomerPortalSessionResponse{URL: "https://billing.stripe.test/session"}, nil
}

func (s stubBillingService) CreateSubscriptionCancelPortalSession(ctx context.Context) (*CustomerPortalSessionResponse, error) {
	if s.createSubscriptionCancelPortalSession != nil {
		return s.createSubscriptionCancelPortalSession(ctx)
	}
	return &CustomerPortalSessionResponse{URL: "https://billing.stripe.test/cancel-session"}, nil
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

func TestHandlerWebhook_UsesSignatureFailureTypeNotErrorText(t *testing.T) {
	var event struct{}
	decodeErr := json.Unmarshal([]byte("[]"), &event)
	require.Error(t, decodeErr)

	for _, tc := range []struct {
		name    string
		err     error
		message string
	}{
		{"missing signature", webhook.ErrNotSigned, "unauthorized"},
		{"invalid header", webhook.ErrInvalidHeader, "unauthorized"},
		{"invalid signature", webhook.ErrNoValidSignature, "unauthorized"},
		{"stale timestamp", webhook.ErrTooOld, "failed to process stripe webhook"},
		{"other error mentioning signature", errors.New("invalid signature field in event body"), "failed to process stripe webhook"},
		{"provider event decode failure", decodeErr, "failed to process stripe webhook"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			handler := NewHandler(slog.New(slog.NewJSONHandler(&logs, nil)), stubBillingService{
				handleWebhook: func(context.Context, []byte, string) error {
					return fmt.Errorf("webhook processing: %w", tc.err)
				},
			})
			req := httptest.NewRequest(http.MethodPost, "/stripe/webhook", nil)
			rec := httptest.NewRecorder()
			handler.Webhook(rec, req)
			require.Equal(t, http.StatusBadRequest, rec.Code)
			require.JSONEq(t, `{"message":"`+tc.message+`"}`, rec.Body.String())

			var entry map[string]any
			require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
			require.Equal(t, "error response", entry["msg"])
			require.Equal(t, float64(http.StatusBadRequest), entry["status"])
			require.Equal(t, "internal", entry["error_category"])
			require.Equal(t, tc.message, entry["response_message"])
			require.NotContains(t, logs.String(), tc.err.Error())
		})
	}
}

func TestHandlerCreateCustomerPortalSession_NotConfiguredReturns503(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{
		createCustomerPortalSession: func(context.Context, PortalReturnDestination) (*CustomerPortalSessionResponse, error) {
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
		createCustomerPortalSession: func(context.Context, PortalReturnDestination) (*CustomerPortalSessionResponse, error) {
			return nil, ErrBillingCustomerMissing
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/customer-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateCustomerPortalSession(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "billing customer was not found")
}

func TestHandlerCreateSubscriptionCancelPortalSession_ReturnsPortalURL(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/subscription-cancel-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateSubscriptionCancelPortalSession(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"url":"https://billing.stripe.test/cancel-session"}`, rr.Body.String())
}

func TestHandlerCreateSubscriptionCancelPortalSession_NotCancelableReturns409(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)), stubBillingService{
		createSubscriptionCancelPortalSession: func(context.Context) (*CustomerPortalSessionResponse, error) {
			return nil, ErrBillingSubscriptionNotCancelable
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/billing/subscription-cancel-portal-session", nil)
	rr := httptest.NewRecorder()

	handler.CreateSubscriptionCancelPortalSession(rr, req)

	require.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "billing subscription cannot be canceled")
}
