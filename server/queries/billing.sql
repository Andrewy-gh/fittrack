-- name: ListActiveFeatureAccess :many
SELECT
    id,
    user_id,
    feature_key,
    source,
    source_reference,
    granted_by,
    note,
    starts_at,
    expires_at,
    revoked_at,
    created_at
FROM user_feature_access
WHERE user_id = $1
  AND revoked_at IS NULL
  AND starts_at <= NOW()
  AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY feature_key, starts_at DESC, id DESC;

-- name: HasActiveFeatureAccess :one
SELECT EXISTS (
    SELECT 1
    FROM user_feature_access
    WHERE user_id = $1
      AND feature_key = $2
      AND revoked_at IS NULL
      AND starts_at <= NOW()
      AND (expires_at IS NULL OR expires_at > NOW())
);

-- Billing queries

-- name: GetStripeCustomerByUserID :one
SELECT user_id, stripe_customer_id, created_at, updated_at
FROM stripe_customers
WHERE user_id = $1;

-- name: GetStripeCustomerUserIDByCustomerID :one
SELECT user_id::text
FROM (SELECT lookup_stripe_customer_user_id($1::text) AS user_id) AS lookup
WHERE user_id IS NOT NULL;

-- name: GetBillingUserForUpdate :one
SELECT id, user_id, created_at
FROM users
WHERE user_id = $1
FOR UPDATE;

-- name: UpsertStripeCustomer :one
INSERT INTO stripe_customers (
    user_id,
    stripe_customer_id
)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET stripe_customer_id = EXCLUDED.stripe_customer_id,
    updated_at = CURRENT_TIMESTAMP
RETURNING user_id, stripe_customer_id, created_at, updated_at;

-- name: UpsertStripeSubscription :one
INSERT INTO stripe_subscriptions (
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end
)
VALUES (
    sqlc.arg(stripe_subscription_id),
    sqlc.arg(user_id),
    sqlc.arg(stripe_customer_id),
    NULLIF(sqlc.arg(stripe_price_id)::text, ''),
    sqlc.arg(stripe_event_created_at),
    sqlc.arg(status),
    sqlc.arg(cancel_at_period_end),
    sqlc.arg(cancel_at),
    sqlc.arg(current_period_start),
    sqlc.arg(current_period_end),
    sqlc.arg(trial_start),
    sqlc.arg(trial_end)
)
ON CONFLICT (stripe_subscription_id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    stripe_customer_id = EXCLUDED.stripe_customer_id,
    stripe_price_id = EXCLUDED.stripe_price_id,
    stripe_event_created_at = EXCLUDED.stripe_event_created_at,
    status = EXCLUDED.status,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    cancel_at = EXCLUDED.cancel_at,
    current_period_start = EXCLUDED.current_period_start,
    current_period_end = EXCLUDED.current_period_end,
    trial_start = EXCLUDED.trial_start,
    trial_end = EXCLUDED.trial_end,
    updated_at = CURRENT_TIMESTAMP
WHERE stripe_subscriptions.stripe_event_created_at < EXCLUDED.stripe_event_created_at
   OR (
       stripe_subscriptions.stripe_event_created_at = EXCLUDED.stripe_event_created_at
       AND NOT (
           stripe_subscriptions.status IN ('past_due', 'unpaid', 'canceled', 'incomplete', 'incomplete_expired', 'paused')
           AND EXCLUDED.status IN ('trialing', 'active')
       )
   )
RETURNING
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end,
    created_at,
    updated_at;

-- name: GetCurrentStripeSubscriptionByUserID :one
SELECT
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end,
    created_at,
    updated_at
FROM stripe_subscriptions
WHERE user_id = $1
ORDER BY
    stripe_event_created_at DESC,
    updated_at DESC,
    created_at DESC
LIMIT 1;

-- name: HasProcessedStripeWebhookEvent :one
SELECT webhook_event.processed::boolean
FROM (
    SELECT public.has_processed_stripe_webhook_event(
        sqlc.arg(stripe_event_id)::TEXT
    ) AS processed
) AS webhook_event;

-- name: MarkStripeWebhookEventProcessed :exec
SELECT public.record_stripe_webhook_event(
    sqlc.arg(stripe_event_id)::TEXT,
    sqlc.arg(event_type)::TEXT
);

-- name: RevokeStripeFeatureAccess :exec
UPDATE user_feature_access
SET revoked_at = GREATEST(CURRENT_TIMESTAMP, starts_at)
WHERE user_id = $1
  AND feature_key = $2
  AND source = 'stripe'
  AND source_reference = $3
  AND revoked_at IS NULL;

-- name: GrantStripeFeatureAccess :exec
INSERT INTO user_feature_access (
    user_id,
    feature_key,
    source,
    source_reference,
    granted_by,
    note,
    starts_at,
    expires_at
)
VALUES ($1, $2, 'stripe', $3, 'stripe_webhook', $4, CURRENT_TIMESTAMP, $5);

-- name: GetAIChatTrialPromptUsage :one
SELECT user_id, stripe_subscription_id, prompt_count, created_at, updated_at
FROM ai_chat_trial_prompt_usage
WHERE user_id = $1
  AND stripe_subscription_id = $2;

-- name: ConsumeAIChatTrialPrompt :one
INSERT INTO ai_chat_trial_prompt_usage (
    user_id,
    stripe_subscription_id,
    prompt_count
)
VALUES ($1, $2, 1)
ON CONFLICT (user_id, stripe_subscription_id) DO UPDATE
SET prompt_count = ai_chat_trial_prompt_usage.prompt_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE ai_chat_trial_prompt_usage.prompt_count < $3
RETURNING user_id, stripe_subscription_id, prompt_count, created_at, updated_at;

-- name: ConsumeAIChatTrialPromptForCurrentSubscription :one
WITH current_subscription AS (
    SELECT stripe_subscription_id, status
    FROM stripe_subscriptions
    WHERE stripe_subscriptions.user_id = $1
    ORDER BY
        stripe_event_created_at DESC,
        updated_at DESC,
        created_at DESC
    LIMIT 1
),
trial_subscription AS (
    SELECT stripe_subscription_id
    FROM current_subscription
    WHERE status = 'trialing'
),
consumed AS (
    INSERT INTO ai_chat_trial_prompt_usage (
        user_id,
        stripe_subscription_id,
        prompt_count
    )
    SELECT $1, stripe_subscription_id, 1
    FROM trial_subscription
    WHERE sqlc.arg(prompt_cap)::integer > 0
    ON CONFLICT (user_id, stripe_subscription_id) DO UPDATE
    SET prompt_count = ai_chat_trial_prompt_usage.prompt_count + 1,
        updated_at = CURRENT_TIMESTAMP
    WHERE ai_chat_trial_prompt_usage.prompt_count < sqlc.arg(prompt_cap)::integer
    RETURNING 1
)
SELECT
    NOT EXISTS (SELECT 1 FROM trial_subscription)
    OR EXISTS (SELECT 1 FROM consumed) AS allowed;

-- UPDATE queries for PUT endpoint
