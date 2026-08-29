-- +goose Up
-- +goose StatementBegin
-- The webhook idempotency log has no direct RLS policies. Only these narrow,
-- owner-backed functions may read or write it.
ALTER TABLE public.stripe_webhook_events ENABLE ROW LEVEL SECURITY;

CREATE OR REPLACE FUNCTION public.has_processed_stripe_webhook_event(
    target_stripe_event_id TEXT
)
RETURNS BOOLEAN
LANGUAGE SQL
STABLE
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    SELECT EXISTS (
        SELECT 1
        FROM public.stripe_webhook_events AS webhook_event
        WHERE webhook_event.stripe_event_id = target_stripe_event_id
    )
$$;

CREATE OR REPLACE FUNCTION public.record_stripe_webhook_event(
    target_stripe_event_id TEXT,
    target_event_type TEXT
)
RETURNS VOID
LANGUAGE SQL
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
    INSERT INTO public.stripe_webhook_events (
        stripe_event_id,
        event_type
    )
    VALUES (
        target_stripe_event_id,
        target_event_type
    )
    ON CONFLICT (stripe_event_id) DO NOTHING
$$;

REVOKE ALL PRIVILEGES ON FUNCTION public.has_processed_stripe_webhook_event(TEXT) FROM PUBLIC;
REVOKE ALL PRIVILEGES ON FUNCTION public.record_stripe_webhook_event(TEXT, TEXT) FROM PUBLIC;

-- Supabase may grant public-schema functions directly to its Data API roles through
-- role-specific default privileges. Those grants are independent of PUBLIC.
DO $$
DECLARE
    api_role TEXT;
BEGIN
    FOREACH api_role IN ARRAY ARRAY['anon', 'authenticated', 'service_role']
    LOOP
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = api_role) THEN
            EXECUTE format(
                'REVOKE ALL PRIVILEGES ON FUNCTION public.has_processed_stripe_webhook_event(TEXT), public.record_stripe_webhook_event(TEXT, TEXT) FROM %I',
                api_role
            );
        END IF;
    END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.record_stripe_webhook_event(TEXT, TEXT);
DROP FUNCTION IF EXISTS public.has_processed_stripe_webhook_event(TEXT);
ALTER TABLE public.stripe_webhook_events DISABLE ROW LEVEL SECURITY;
-- +goose StatementEnd
