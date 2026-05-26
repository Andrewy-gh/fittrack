package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	apperrors "github.com/Andrewy-gh/fittrack/server/internal/errors"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v83"
	portalsession "github.com/stripe/stripe-go/v83/billingportal/session"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/webhook"
)

type Service struct {
	logger                *slog.Logger
	repo                  Repository
	stripeSecretKey       string
	webhookSecret         string
	premiumPriceID        string
	appBaseURL            string
	trialPromptCap        int32
	createCustomer        func(*stripe.CustomerParams) (*stripe.Customer, error)
	createCheckoutSession func(*stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error)
	createPortalSession   func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error)
	constructEvent        func([]byte, string, string) (stripe.Event, error)
}

func NewService(logger *slog.Logger, repo Repository, stripeSecretKey string, webhookSecret string, premiumPriceID string, appBaseURL string, trialPromptCap int) *Service {
	return &Service{
		logger:                logger,
		repo:                  repo,
		stripeSecretKey:       strings.TrimSpace(stripeSecretKey),
		webhookSecret:         strings.TrimSpace(webhookSecret),
		premiumPriceID:        strings.TrimSpace(premiumPriceID),
		appBaseURL:            strings.TrimRight(strings.TrimSpace(appBaseURL), "/"),
		trialPromptCap:        int32(trialPromptCap),
		createCustomer:        customer.New,
		createCheckoutSession: session.New,
		createPortalSession:   portalsession.New,
		constructEvent: func(payload []byte, header string, secret string) (stripe.Event, error) {
			return webhook.ConstructEventWithOptions(payload, header, secret, webhook.ConstructEventOptions{
				IgnoreAPIVersionMismatch: true,
			})
		},
	}
}

func (s *Service) CreateCheckoutSession(ctx context.Context) (*CheckoutSessionResponse, error) {
	if !s.configuredForCheckout() {
		return nil, ErrBillingNotConfigured
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	stripe.Key = s.stripeSecretKey
	customerID, err := s.ensureStripeCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{
		Mode:                    stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		Customer:                stripe.String(customerID),
		ClientReferenceID:       stripe.String(userID),
		SuccessURL:              stripe.String(s.appBaseURL + "/chat?checkout=success"),
		CancelURL:               stripe.String(s.appBaseURL + "/chat?checkout=cancelled"),
		PaymentMethodCollection: stripe.String(string(stripe.CheckoutSessionPaymentMethodCollectionAlways)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(s.premiumPriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			TrialPeriodDays: stripe.Int64(7),
			Metadata: map[string]string{
				"fittrack_user_id": userID,
				"feature_key":      FeatureKeyAIChatbot,
			},
		},
		Metadata: map[string]string{
			"fittrack_user_id": userID,
			"feature_key":      FeatureKeyAIChatbot,
		},
	}

	checkoutSession, err := s.createCheckoutSession(params)
	if err != nil {
		return nil, fmt.Errorf("create stripe checkout session: %w", err)
	}
	if strings.TrimSpace(checkoutSession.URL) == "" {
		return nil, fmt.Errorf("stripe checkout session did not include a hosted url")
	}

	return &CheckoutSessionResponse{URL: checkoutSession.URL}, nil
}

func (s *Service) CreateCustomerPortalSession(ctx context.Context) (*CustomerPortalSessionResponse, error) {
	if !s.configuredForPortal() {
		return nil, ErrBillingNotConfigured
	}

	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	stripe.Key = s.stripeSecretKey
	customerID, err := s.existingStripeCustomer(ctx, userID)
	if err != nil {
		return nil, err
	}

	portalSession, err := s.createPortalSession(&stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(s.appBaseURL + "/chat"),
	})
	if err != nil {
		return nil, fmt.Errorf("create stripe billing portal session: %w", err)
	}
	if strings.TrimSpace(portalSession.URL) == "" {
		return nil, fmt.Errorf("stripe billing portal session did not include a hosted url")
	}

	return &CustomerPortalSessionResponse{URL: portalSession.URL}, nil
}

func (s *Service) CurrentStatus(ctx context.Context) (*StatusResponse, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}

	resp := &StatusResponse{
		FeatureKey: FeatureKeyAIChatbot,
	}

	subscription, err := s.repo.GetCurrentSubscriptionByUserID(ctx, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return resp, nil
	case err != nil:
		return nil, fmt.Errorf("get current billing subscription: %w", err)
	}

	resp.HasAccess = s.subscriptionRowGrantsAccess(subscription) && subscriptionAccessPeriodOpen(subscription, time.Now().UTC())
	resp.Subscription = subscriptionView(subscription)

	if statusConsumesTrialPrompts(subscription.Status) {
		resp.TrialUsage = &TrialUsageView{Limit: s.trialPromptCap}
		usage, err := s.repo.GetTrialPromptUsage(ctx, userID, subscription.StripeSubscriptionID)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return nil, fmt.Errorf("get trial prompt usage: %w", err)
		default:
			resp.TrialUsage.Used = usage.PromptCount
		}
	}

	return resp, nil
}

func (s *Service) EnsureAIChatPromptAllowed(ctx context.Context) error {
	userID, err := currentUserID(ctx)
	if err != nil {
		return err
	}

	subscription, err := s.repo.GetCurrentSubscriptionByUserID(ctx, userID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return nil
	case err != nil:
		return fmt.Errorf("get current billing subscription: %w", err)
	}

	if !statusConsumesTrialPrompts(subscription.Status) {
		return nil
	}
	if s.trialPromptCap <= 0 {
		return ErrTrialPromptLimitExceeded
	}

	_, err = s.repo.ConsumeTrialPrompt(ctx, userID, subscription.StripeSubscriptionID, s.trialPromptCap)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signatureHeader string) error {
	if strings.TrimSpace(s.webhookSecret) == "" {
		return ErrBillingNotConfigured
	}

	event, err := s.constructEvent(payload, signatureHeader, s.webhookSecret)
	if err != nil {
		return fmt.Errorf("verify stripe webhook: %w", err)
	}

	processed, err := s.repo.HasProcessedWebhookEvent(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("check stripe webhook idempotency: %w", err)
	}
	if processed {
		return nil
	}

	if err := s.applyWebhookEvent(ctx, event); err != nil {
		return err
	}

	if err := s.repo.MarkWebhookEventProcessed(ctx, event.ID, string(event.Type)); err != nil {
		return fmt.Errorf("mark stripe webhook processed: %w", err)
	}
	return nil
}

func (s *Service) ApplySubscriptionSnapshot(ctx context.Context, snapshot StripeSubscriptionSnapshot) error {
	if err := validateSnapshot(snapshot); err != nil {
		return err
	}
	_, err := s.repo.UpsertSubscriptionFromWebhook(ctx, snapshot)
	return err
}

func (s *Service) applyWebhookEvent(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		var checkoutSession stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
			return fmt.Errorf("decode stripe checkout session: %w", err)
		}
		return s.recordCheckoutCustomer(ctx, checkoutSession)
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return fmt.Errorf("decode stripe subscription: %w", err)
		}
		return s.applySubscriptionEvent(ctx, subscription, unixPtr(event.Created))
	default:
		s.logger.Debug("ignoring unsupported stripe webhook event", "event_type", event.Type, "event_id", event.ID)
		return nil
	}
}

func (s *Service) applySubscriptionEvent(ctx context.Context, subscription stripe.Subscription, eventCreatedAt *time.Time) error {
	snapshot := subscriptionSnapshot(subscription)
	snapshot.StripeEventCreatedAt = eventCreatedAt
	if snapshot.UserID == "" && snapshot.StripeCustomerID != "" {
		customerRow, err := s.repo.GetStripeCustomerByCustomerID(ctx, snapshot.StripeCustomerID)
		if err != nil {
			return fmt.Errorf("find user for stripe customer: %w", err)
		}
		snapshot.UserID = customerRow.UserID
	}
	snapshot.GrantAIChatAccess = s.snapshotGrantsAccess(snapshot)
	return s.ApplySubscriptionSnapshot(ctx, snapshot)
}

func (s *Service) recordCheckoutCustomer(ctx context.Context, checkoutSession stripe.CheckoutSession) error {
	userID := strings.TrimSpace(checkoutSession.ClientReferenceID)
	if userID == "" {
		userID = strings.TrimSpace(checkoutSession.Metadata["fittrack_user_id"])
	}
	if userID == "" {
		return ErrStripeUserMissing
	}
	if checkoutSession.Customer == nil || strings.TrimSpace(checkoutSession.Customer.ID) == "" {
		return ErrStripeCustomerMissing
	}

	_, err := s.repo.UpsertStripeCustomer(ctx, userID, checkoutSession.Customer.ID)
	return err
}

func (s *Service) ensureStripeCustomer(ctx context.Context, userID string) (string, error) {
	row, err := s.repo.GetStripeCustomerByUserID(ctx, userID)
	switch {
	case err == nil:
		return row.StripeCustomerID, nil
	case errors.Is(err, pgx.ErrNoRows):
	default:
		return "", fmt.Errorf("get stripe customer: %w", err)
	}

	customer, err := s.createCustomer(&stripe.CustomerParams{
		Metadata: map[string]string{
			"fittrack_user_id": userID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}
	if strings.TrimSpace(customer.ID) == "" {
		return "", fmt.Errorf("stripe customer did not include an id")
	}

	if _, err := s.repo.UpsertStripeCustomer(ctx, userID, customer.ID); err != nil {
		return "", fmt.Errorf("save stripe customer: %w", err)
	}

	return customer.ID, nil
}

func (s *Service) existingStripeCustomer(ctx context.Context, userID string) (string, error) {
	row, err := s.repo.GetStripeCustomerByUserID(ctx, userID)
	switch {
	case err == nil:
	case errors.Is(err, pgx.ErrNoRows):
		return "", ErrBillingCustomerMissing
	default:
		return "", fmt.Errorf("get stripe customer: %w", err)
	}

	customerID := strings.TrimSpace(row.StripeCustomerID)
	if customerID == "" {
		return "", ErrBillingCustomerMissing
	}

	return customerID, nil
}

func (s *Service) configuredForCheckout() bool {
	return strings.TrimSpace(s.stripeSecretKey) != "" &&
		strings.TrimSpace(s.premiumPriceID) != "" &&
		strings.TrimSpace(s.appBaseURL) != ""
}

func (s *Service) configuredForPortal() bool {
	return strings.TrimSpace(s.stripeSecretKey) != "" &&
		strings.TrimSpace(s.appBaseURL) != ""
}

func subscriptionSnapshot(subscription stripe.Subscription) StripeSubscriptionSnapshot {
	userID := strings.TrimSpace(subscription.Metadata["fittrack_user_id"])
	priceID := ""
	var currentPeriodStart *time.Time
	var currentPeriodEnd *time.Time
	if subscription.Items != nil && len(subscription.Items.Data) > 0 {
		item := subscription.Items.Data[0]
		if item.Price != nil {
			priceID = strings.TrimSpace(item.Price.ID)
		}
		currentPeriodStart = unixPtr(item.CurrentPeriodStart)
		currentPeriodEnd = unixPtr(item.CurrentPeriodEnd)
	}

	customerID := ""
	if subscription.Customer != nil {
		customerID = strings.TrimSpace(subscription.Customer.ID)
	}

	return StripeSubscriptionSnapshot{
		StripeSubscriptionID: strings.TrimSpace(subscription.ID),
		UserID:               userID,
		StripeCustomerID:     customerID,
		StripePriceID:        priceID,
		Status:               string(subscription.Status),
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd,
		CurrentPeriodStart:   currentPeriodStart,
		CurrentPeriodEnd:     currentPeriodEnd,
		TrialStart:           unixPtr(subscription.TrialStart),
		TrialEnd:             unixPtr(subscription.TrialEnd),
	}
}

func validateSnapshot(snapshot StripeSubscriptionSnapshot) error {
	if strings.TrimSpace(snapshot.StripeSubscriptionID) == "" {
		return fmt.Errorf("stripe subscription id is required")
	}
	if strings.TrimSpace(snapshot.UserID) == "" {
		return ErrStripeUserMissing
	}
	if strings.TrimSpace(snapshot.StripeCustomerID) == "" {
		return ErrStripeCustomerMissing
	}
	if snapshot.StripeEventCreatedAt == nil || snapshot.StripeEventCreatedAt.IsZero() {
		return fmt.Errorf("stripe event created time is required")
	}
	if strings.TrimSpace(snapshot.Status) == "" {
		return fmt.Errorf("stripe subscription status is required")
	}
	return nil
}

func (s *Service) snapshotGrantsAccess(snapshot StripeSubscriptionSnapshot) bool {
	return statusAllowsAccess(snapshot.Status) && strings.TrimSpace(snapshot.StripePriceID) == s.premiumPriceID
}

func (s *Service) subscriptionRowGrantsAccess(subscription db.StripeSubscriptions) bool {
	return statusAllowsAccess(subscription.Status) && textFromPg(subscription.StripePriceID) == s.premiumPriceID
}

func subscriptionAccessPeriodOpen(subscription db.StripeSubscriptions, now time.Time) bool {
	if subscription.CurrentPeriodEnd.Valid && !subscription.CurrentPeriodEnd.Time.After(now) {
		return false
	}
	return true
}

func subscriptionView(row db.StripeSubscriptions) *SubscriptionView {
	return &SubscriptionView{
		StripeSubscriptionID: row.StripeSubscriptionID,
		Status:               row.Status,
		CancelAtPeriodEnd:    row.CancelAtPeriodEnd,
		CurrentPeriodEnd:     timePtrFromPg(row.CurrentPeriodEnd),
		TrialEnd:             timePtrFromPg(row.TrialEnd),
	}
}

func unixPtr(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	t := time.Unix(value, 0).UTC()
	return &t
}

func timePtrFromPg(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func textFromPg(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func currentUserID(ctx context.Context) (string, error) {
	userID, ok := user.Current(ctx)
	if !ok || userID == "" {
		return "", apperrors.NewUnauthorized("billing", "")
	}
	return userID, nil
}
