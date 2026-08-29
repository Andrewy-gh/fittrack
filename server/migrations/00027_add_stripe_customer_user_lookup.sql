-- +goose Up
-- +goose StatementBegin
-- Stripe webhooks are unauthenticated application requests. This narrow definer
-- function resolves the tenant before the repository establishes that tenant's RLS
-- context; it does not expose the rest of the customer row.
CREATE OR REPLACE FUNCTION lookup_stripe_customer_user_id(target_customer_id TEXT)
RETURNS TEXT
LANGUAGE SQL
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT sc.user_id
    FROM public.stripe_customers AS sc
    WHERE sc.stripe_customer_id = target_customer_id
    LIMIT 1
$$;

REVOKE ALL PRIVILEGES ON FUNCTION lookup_stripe_customer_user_id(TEXT) FROM PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS lookup_stripe_customer_user_id(TEXT);
-- +goose StatementEnd
