package billing

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/response"
)

const maxWebhookPayloadBytes = 1 << 20

type billingService interface {
	CreateCheckoutSession(ctx context.Context) (*CheckoutSessionResponse, error)
	CreateCustomerPortalSession(ctx context.Context, destination PortalReturnDestination) (*CustomerPortalSessionResponse, error)
	CreateSubscriptionCancelPortalSession(ctx context.Context) (*CustomerPortalSessionResponse, error)
	CurrentStatus(ctx context.Context) (*StatusResponse, error)
	HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error
}

type Handler struct {
	logger  *slog.Logger
	service billingService
}

func NewHandler(logger *slog.Logger, service billingService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.CreateCheckoutSession(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) CreateCustomerPortalSession(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.CreateCustomerPortalSession(r.Context(), parsePortalReturnDestination(r))
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to create billing portal session")
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func parsePortalReturnDestination(r *http.Request) PortalReturnDestination {
	if r.URL.Query().Get("return_to") == string(PortalReturnSettings) {
		return PortalReturnSettings
	}
	return PortalReturnChat
}

func (h *Handler) CreateSubscriptionCancelPortalSession(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.CreateSubscriptionCancelPortalSession(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to create subscription cancellation portal session")
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) CurrentStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.CurrentStatus(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err, http.StatusInternalServerError, "failed to load billing status")
		return
	}

	if err := response.JSON(w, http.StatusOK, resp); err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "failed to write response", err)
	}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes))
	if err != nil {
		response.ErrorJSON(w, r, h.logger, http.StatusBadRequest, "failed to read stripe webhook payload", err)
		return
	}

	if err := h.service.HandleWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		h.writeServiceError(w, r, err, http.StatusBadRequest, "failed to process stripe webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error, unexpectedStatus int, unexpectedMessage string) {
	var errUnauthorized *apperrors.Unauthorized

	switch {
	case errors.As(err, &errUnauthorized):
		response.ErrorJSON(w, r, h.logger, http.StatusUnauthorized, errUnauthorized.Error(), nil)
	case errors.Is(err, ErrBillingNotConfigured):
		response.ErrorJSON(w, r, h.logger, http.StatusServiceUnavailable, "billing is not configured", nil)
	case errors.Is(err, ErrBillingCustomerMissing):
		response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "billing customer was not found", nil)
	case errors.Is(err, ErrBillingSubscriptionMissing):
		response.ErrorJSON(w, r, h.logger, http.StatusNotFound, "billing subscription was not found", nil)
	case errors.Is(err, ErrBillingSubscriptionNotCancelable):
		response.ErrorJSON(w, r, h.logger, http.StatusConflict, "billing subscription cannot be canceled", nil)
	case errors.Is(err, ErrTrialPromptLimitExceeded):
		response.ErrorJSON(w, r, h.logger, http.StatusForbidden, "ai chat trial prompt limit reached", nil)
	default:
		response.ErrorJSON(w, r, h.logger, unexpectedStatus, unexpectedMessage, err)
	}
}
