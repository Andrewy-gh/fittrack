-- +goose Up
-- +goose StatementBegin
CREATE TABLE stripe_customers (
    user_id VARCHAR(256) PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(256) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT stripe_customers_customer_id_not_empty CHECK (btrim(stripe_customer_id) <> '')
);

CREATE TABLE stripe_subscriptions (
    stripe_subscription_id VARCHAR(256) PRIMARY KEY,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(256) NOT NULL,
    stripe_price_id VARCHAR(256),
    stripe_event_created_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(64) NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    trial_start TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT stripe_subscriptions_customer_id_not_empty CHECK (btrim(stripe_customer_id) <> ''),
    CONSTRAINT stripe_subscriptions_status_not_empty CHECK (btrim(status) <> '')
);

CREATE INDEX idx_stripe_subscriptions_user_updated
    ON stripe_subscriptions (user_id, stripe_event_created_at DESC);

CREATE INDEX idx_stripe_subscriptions_customer
    ON stripe_subscriptions (stripe_customer_id);

CREATE TABLE stripe_webhook_events (
    stripe_event_id VARCHAR(256) PRIMARY KEY,
    event_type VARCHAR(128) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT stripe_webhook_events_event_type_not_empty CHECK (btrim(event_type) <> '')
);

CREATE TABLE ai_chat_trial_prompt_usage (
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(256) NOT NULL REFERENCES stripe_subscriptions(stripe_subscription_id) ON DELETE CASCADE,
    prompt_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, stripe_subscription_id),
    CONSTRAINT ai_chat_trial_prompt_usage_count_non_negative CHECK (prompt_count >= 0)
);

ALTER TABLE stripe_customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE stripe_subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_chat_trial_prompt_usage ENABLE ROW LEVEL SECURITY;

CREATE POLICY stripe_customers_select_policy ON stripe_customers
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY stripe_customers_insert_policy ON stripe_customers
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY stripe_customers_update_policy ON stripe_customers
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY stripe_subscriptions_select_policy ON stripe_subscriptions
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY stripe_subscriptions_insert_policy ON stripe_subscriptions
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY stripe_subscriptions_update_policy ON stripe_subscriptions
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY ai_chat_trial_prompt_usage_select_policy ON ai_chat_trial_prompt_usage
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY ai_chat_trial_prompt_usage_insert_policy ON ai_chat_trial_prompt_usage
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY ai_chat_trial_prompt_usage_update_policy ON ai_chat_trial_prompt_usage
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

GRANT SELECT, INSERT, UPDATE ON stripe_customers TO PUBLIC;
GRANT SELECT, INSERT, UPDATE ON stripe_subscriptions TO PUBLIC;
GRANT SELECT, INSERT, UPDATE ON stripe_webhook_events TO PUBLIC;
GRANT SELECT, INSERT, UPDATE ON ai_chat_trial_prompt_usage TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS ai_chat_trial_prompt_usage_update_policy ON ai_chat_trial_prompt_usage;
DROP POLICY IF EXISTS ai_chat_trial_prompt_usage_insert_policy ON ai_chat_trial_prompt_usage;
DROP POLICY IF EXISTS ai_chat_trial_prompt_usage_select_policy ON ai_chat_trial_prompt_usage;
DROP POLICY IF EXISTS stripe_subscriptions_update_policy ON stripe_subscriptions;
DROP POLICY IF EXISTS stripe_subscriptions_insert_policy ON stripe_subscriptions;
DROP POLICY IF EXISTS stripe_subscriptions_select_policy ON stripe_subscriptions;
DROP POLICY IF EXISTS stripe_customers_update_policy ON stripe_customers;
DROP POLICY IF EXISTS stripe_customers_insert_policy ON stripe_customers;
DROP POLICY IF EXISTS stripe_customers_select_policy ON stripe_customers;

ALTER TABLE ai_chat_trial_prompt_usage DISABLE ROW LEVEL SECURITY;
ALTER TABLE stripe_subscriptions DISABLE ROW LEVEL SECURITY;
ALTER TABLE stripe_customers DISABLE ROW LEVEL SECURITY;

REVOKE ALL ON ai_chat_trial_prompt_usage FROM PUBLIC;
REVOKE ALL ON stripe_webhook_events FROM PUBLIC;
REVOKE ALL ON stripe_subscriptions FROM PUBLIC;
REVOKE ALL ON stripe_customers FROM PUBLIC;

DROP TABLE IF EXISTS ai_chat_trial_prompt_usage;
DROP TABLE IF EXISTS stripe_webhook_events;
DROP TABLE IF EXISTS stripe_subscriptions;
DROP TABLE IF EXISTS stripe_customers;
-- +goose StatementEnd
