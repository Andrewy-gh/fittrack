# Stripe Billing

FitTrack uses Stripe subscriptions to grant premium access to the AI chat feature. Stripe is the billing source of truth; FitTrack stores the Stripe customer and latest subscription snapshot so the app can answer access checks quickly.

## Product Behavior

- The paid feature key is `ai_chatbot`.
- Checkout creates a Stripe subscription with a 7-day trial.
- `trialing` and `active` subscriptions grant access when the subscription is for `STRIPE_PREMIUM_PRICE_ID` and the current billing period has not ended.
- Trialing subscriptions consume AI chat prompts against `AI_CHAT_TRIAL_PROMPT_CAP`.
- `past_due`, `unpaid`, `canceled`, `incomplete`, and `incomplete_expired` subscriptions do not grant access.
- Checkout redirects users back to `/chat?checkout=success` or `/chat?checkout=cancelled`.
- The chat UI gates composer access on the same active `ai_chatbot` feature grant that the AI chat backend enforces.
- After Checkout success, the chat UI polls access briefly because Stripe redirects can arrive before subscription webhooks finish.
- Users with `past_due` or `unpaid` subscriptions are sent to Stripe's billing portal to update payment details.

## Required Stripe Setup

Create a recurring Stripe price for FitTrack premium AI chat and configure these server variables:

```env
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PREMIUM_PRICE_ID=price_...
APP_BASE_URL=http://localhost:5173
AI_CHAT_TRIAL_PROMPT_CAP=30
```

`APP_BASE_URL` must be the browser-facing app origin because Stripe uses it for checkout redirects.

## Backend Endpoints

- `POST /api/billing/checkout-session`: authenticated endpoint that creates or reuses the current user's Stripe customer, then returns a hosted checkout URL.
- `POST /api/billing/customer-portal-session`: authenticated endpoint that creates a Stripe billing portal session for the current customer, then returns a hosted billing management URL.
- `GET /api/billing/status`: authenticated endpoint that returns the current user's AI chat billing access, subscription details, and trial prompt usage when applicable.
- `POST /stripe/webhook`: unauthenticated Stripe webhook endpoint. Signature verification uses `STRIPE_WEBHOOK_SECRET`.

If checkout variables are missing, checkout returns `503 billing is not configured`. If only webhook configuration is missing, webhook processing returns the same configuration error.
Stripe billing portal sessions require a Stripe Billing Portal configuration in the Stripe account.

## Webhook Events

FitTrack handles these Stripe events:

- `checkout.session.completed`: records the mapping between the FitTrack user and Stripe customer.
- `customer.subscription.created`: stores the subscription snapshot and grants feature access when eligible.
- `customer.subscription.updated`: refreshes subscription status, period dates, cancellation state, and feature access.
- `customer.subscription.deleted`: records cancellation and revokes Stripe-backed feature access.

Unsupported Stripe events are ignored after signature verification.

Webhook processing is idempotent. Processed Stripe event IDs are stored in `stripe_webhook_events`, so Stripe retries should not duplicate access changes.

## Stored Data

Billing state is stored in these tables:

- `stripe_customers`: maps FitTrack user IDs to Stripe customer IDs.
- `stripe_subscriptions`: stores the latest subscription snapshot from Stripe.
- `stripe_webhook_events`: records processed webhook events for idempotency.
- `ai_chat_trial_prompt_usage`: tracks trial prompt consumption per user and subscription.

The Stripe-backed feature grant is written through the existing feature access table with source reference set to the Stripe subscription ID.

## Local Webhook Testing

Use the Stripe CLI to forward events to the local API:

```bash
stripe listen --forward-to localhost:8080/stripe/webhook
```

Copy the printed webhook signing secret into `STRIPE_WEBHOOK_SECRET`, then run the backend normally:

```bash
cd server
make dev
```

Trigger or complete a test checkout, then verify the user's status:

```bash
curl -H "x-stack-access-token: $STACK_ACCESS_TOKEN" \
  http://localhost:8080/api/billing/status
```

## Operational Notes

- Webhook subscription payloads must include `fittrack_user_id` metadata, or FitTrack must already know the Stripe customer from checkout.
- Subscription events older than the stored `stripe_event_created_at` are ignored so stale Stripe retries cannot roll back newer access state.
- Access is only granted for the configured premium price ID; subscriptions for another price are stored but do not unlock AI chat.
- If support needs to inspect a user's state, check the Stripe customer mapping first, then the latest `stripe_subscriptions` row and feature access source reference.
