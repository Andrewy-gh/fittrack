package billing

import (
	"errors"
	"time"
)

const (
	FeatureKeyAIChatbot = "ai_chatbot"

	subscriptionStatusTrialing          = "trialing"
	subscriptionStatusActive            = "active"
	subscriptionStatusPastDue           = "past_due"
	subscriptionStatusUnpaid            = "unpaid"
	subscriptionStatusCanceled          = "canceled"
	subscriptionStatusIncomplete        = "incomplete"
	subscriptionStatusIncompleteExpired = "incomplete_expired"
)

var (
	ErrBillingNotConfigured             = errors.New("billing is not configured")
	ErrBillingCustomerMissing           = errors.New("billing customer is missing")
	ErrBillingSubscriptionMissing       = errors.New("billing subscription is missing")
	ErrBillingSubscriptionNotCancelable = errors.New("billing subscription cannot be canceled")
	ErrTrialPromptLimitExceeded         = errors.New("ai chat trial prompt limit exceeded")
	ErrStripeUserMissing                = errors.New("stripe payload is missing fittrack user metadata")
	ErrStripeCustomerMissing            = errors.New("stripe payload is missing customer")
	ErrUnsupportedStripeEvent           = errors.New("unsupported stripe webhook event")
)

type CheckoutSessionResponse struct {
	URL string `json:"url"`
}

type CustomerPortalSessionResponse struct {
	URL string `json:"url"`
}

type StatusResponse struct {
	FeatureKey   string            `json:"feature_key"`
	HasAccess    bool              `json:"has_access"`
	Subscription *SubscriptionView `json:"subscription,omitempty"`
	TrialUsage   *TrialUsageView   `json:"trial_usage,omitempty"`
}

type SubscriptionView struct {
	StripeSubscriptionID  string     `json:"stripe_subscription_id"`
	Status                string     `json:"status"`
	CancellationScheduled bool       `json:"cancellation_scheduled"`
	AccessEndsAt          *time.Time `json:"access_ends_at,omitempty"`
	TrialEnd              *time.Time `json:"trial_end,omitempty"`
}

type TrialUsageView struct {
	Used  int32 `json:"used"`
	Limit int32 `json:"limit"`
}

type StripeSubscriptionSnapshot struct {
	StripeSubscriptionID string
	UserID               string
	StripeCustomerID     string
	StripePriceID        string
	StripeEventCreatedAt *time.Time
	Status               string
	CancelAtPeriodEnd    bool
	CancelAt             *time.Time
	CurrentPeriodStart   *time.Time
	CurrentPeriodEnd     *time.Time
	TrialStart           *time.Time
	TrialEnd             *time.Time
	GrantAIChatAccess    bool
}

func statusAllowsAccess(status string) bool {
	return status == subscriptionStatusTrialing || status == subscriptionStatusActive
}

func statusConsumesTrialPrompts(status string) bool {
	return status == subscriptionStatusTrialing
}
