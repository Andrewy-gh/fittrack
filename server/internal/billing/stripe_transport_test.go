package billing

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	db "github.com/Andrewy-gh/fittrack/server/internal/database"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/stripe/stripe-go/v83"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type stripeTransportRepo struct {
	Repository
	id      string
	created atomic.Bool
	ready   *sync.WaitGroup
}

func (r *stripeTransportRepo) GetStripeCustomerByUserID(context.Context, string) (db.StripeCustomers, error) {
	if !r.created.Load() {
		r.ready.Done()
		r.ready.Wait()
		return db.StripeCustomers{}, pgx.ErrNoRows
	}
	return db.StripeCustomers{StripeCustomerID: "cus_" + r.id}, nil
}
func (r *stripeTransportRepo) UpsertStripeCustomer(_ context.Context, _, id string) (db.StripeCustomers, error) {
	r.created.Store(true)
	return db.StripeCustomers{StripeCustomerID: id}, nil
}
func (r *stripeTransportRepo) GetCurrentSubscriptionByUserID(context.Context, string) (db.StripeSubscriptions, error) {
	return db.StripeSubscriptions{StripeSubscriptionID: "sub_" + r.id, StripePriceID: textToPg("price_premium"), Status: subscriptionStatusActive}, nil
}

// This test is serial because stripe.Key is process-global. Check each entry point
// independently so a future write cannot hide behind the SDK transport test.
func TestBillingDoesNotMutateGlobalStripeKey(t *testing.T) {
	const sentinel = "synthetic_global_sentinel"
	previous := stripe.Key
	stripe.Key = sentinel
	t.Cleanup(func() { stripe.Key = previous })

	repo := &stripeTransportRepo{id: "a"}
	repo.created.Store(true)
	s := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, "synthetic_service", "", "price_premium", "https://example.invalid", 30)
	if stripe.Key != sentinel {
		t.Fatal("constructing a service changed the global Stripe key")
	}
	s.createCheckoutSession = func(*stripe.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
		return &stripe.CheckoutSession{URL: "https://example.invalid"}, nil
	}
	s.createPortalSession = func(*stripe.BillingPortalSessionParams) (*stripe.BillingPortalSession, error) {
		return &stripe.BillingPortalSession{URL: "https://example.invalid"}, nil
	}
	s.cancelSubscription = func(string, *stripe.SubscriptionCancelParams) (*stripe.Subscription, error) {
		return &stripe.Subscription{}, nil
	}
	ctx := user.WithContext(context.Background(), "a")
	operations := []struct {
		name string
		run  func() error
	}{
		{"checkout", func() error { _, err := s.CreateCheckoutSession(ctx); return err }},
		{"portal", func() error { _, err := s.CreateCustomerPortalSession(ctx, PortalReturnChat); return err }},
		{"cancel portal", func() error { _, err := s.CreateSubscriptionCancelPortalSession(ctx); return err }},
		{"immediate cancellation", func() error { return s.CancelCurrentSubscriptionImmediately(ctx) }},
	}
	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			stripe.Key = sentinel
			if err := op.run(); err != nil {
				t.Fatalf("operation failed: %v", err)
			}
			if stripe.Key != sentinel {
				t.Error("operation changed the global Stripe key")
			}
		})
	}
}

// Exercises the actual SDK serialization and authentication, without network access.
func TestStripeRequestsAreScopedToEachService(t *testing.T) {
	var mu sync.Mutex
	var failures []string
	var requests int
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		id := ""
		switch r.URL.Path {
		case "/v1/customers":
			id = form.Get("metadata[fittrack_user_id]")
		case "/v1/checkout/sessions", "/v1/billing_portal/sessions":
			id = strings.TrimPrefix(form.Get("customer"), "cus_")
		default:
			id = strings.TrimPrefix(r.URL.Path, "/v1/subscriptions/sub_")
		}
		mu.Lock()
		requests++
		if (id != "a" && id != "b") || r.Header.Get("Authorization") != "Bearer synthetic_"+id || r.Header.Get("Stripe-Account") != "" {
			failures = append(failures, fmt.Sprintf("incorrect authentication or account for %s (%s)", r.URL.Path, id))
		}
		if r.Method == http.MethodPost && r.Header.Get("Idempotency-Key") == "" {
			failures = append(failures, "missing SDK idempotency key for "+r.URL.Path)
		}
		if r.URL.Path == "/v1/checkout/sessions" && (r.Method != http.MethodPost || form.Get("mode") != "subscription" || form.Get("line_items[0][price]") != "price_premium" || form.Get("subscription_data[trial_period_days]") != "7" || form.Get("client_reference_id") != id) {
			failures = append(failures, "incorrect checkout request")
		}
		if r.URL.Path == "/v1/billing_portal/sessions" && form.Get("flow_data[type]") != "" && (form.Get("flow_data[type]") != "subscription_cancel" || form.Get("flow_data[subscription_cancel][subscription]") != "sub_"+id) {
			failures = append(failures, "incorrect cancellation portal request")
		}
		if strings.HasPrefix(r.URL.Path, "/v1/subscriptions/") && r.Method != http.MethodDelete {
			failures = append(failures, "incorrect immediate cancellation method")
		}
		mu.Unlock()
		response := `{"id":"cus_` + id + `","url":"https://example.invalid/session"}`
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: r}, nil
	})}
	old := stripe.GetBackend(stripe.APIBackend)
	stripe.SetBackend(stripe.APIBackend, stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{HTTPClient: client}))
	t.Cleanup(func() { stripe.SetBackend(stripe.APIBackend, old) })

	var ready sync.WaitGroup
	ready.Add(2)
	var workers sync.WaitGroup
	for _, id := range []string{"a", "b"} {
		workers.Add(1)
		go func(id string) {
			defer workers.Done()
			repo := &stripeTransportRepo{id: id, ready: &ready}
			s := NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, "synthetic_"+id, "", "price_premium", "https://example.invalid", 30)
			ctx := user.WithContext(context.Background(), id)
			_, err := s.CreateCheckoutSession(ctx)
			if err == nil {
				_, err = s.CreateCustomerPortalSession(ctx, PortalReturnChat)
			}
			if err == nil {
				_, err = s.CreateSubscriptionCancelPortalSession(ctx)
			}
			if err == nil {
				err = s.CancelCurrentSubscriptionImmediately(ctx)
			}
			if err != nil {
				mu.Lock()
				failures = append(failures, err.Error())
				mu.Unlock()
			}
		}(id)
	}
	workers.Wait()
	for _, failure := range failures {
		t.Error(failure)
	}
	if requests != 10 {
		t.Errorf("got %d requests, want 10", requests)
	}
}
