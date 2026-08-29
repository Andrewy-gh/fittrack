-- Run with the same owning/admin role used by Goose, never the runtime credential.
-- This script intentionally does not configure LOGIN or a password.

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'fittrack_app') THEN
        CREATE ROLE fittrack_app
            NOLOGIN
            NOSUPERUSER
            NOCREATEDB
            NOCREATEROLE
            NOINHERIT
            NOREPLICATION
            NOBYPASSRLS;
    ELSE
        -- Preserve externally managed LOGIN/password settings.
        ALTER ROLE fittrack_app
            NOSUPERUSER
            NOCREATEDB
            NOCREATEROLE
            NOINHERIT
            NOREPLICATION
            NOBYPASSRLS;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_roles r ON r.oid = c.relowner
        WHERE r.rolname = 'fittrack_app'
          AND c.relrowsecurity
    ) THEN
        RAISE EXCEPTION 'fittrack_app must not own an RLS-protected table';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_roles inherited
        JOIN pg_roles app ON app.rolname = 'fittrack_app'
        WHERE inherited.oid <> app.oid
          AND pg_has_role(app.oid, inherited.oid, 'MEMBER')
          AND (
              inherited.rolsuper
              OR inherited.rolbypassrls
              OR EXISTS (
                  SELECT 1
                  FROM pg_class owned
                  WHERE owned.relowner = inherited.oid
                    AND owned.relrowsecurity
              )
          )
    ) THEN
        RAISE EXCEPTION 'fittrack_app must not inherit or assume a role that can bypass RLS';
    END IF;

    EXECUTE format('GRANT CONNECT ON DATABASE %I TO fittrack_app', current_database());
END $$;

-- Establish a deny-by-default baseline for every existing public-schema object,
-- including migration metadata and objects not known to the application.
REVOKE CREATE ON SCHEMA public FROM PUBLIC, fittrack_app;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM PUBLIC, fittrack_app;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC, fittrack_app;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public FROM PUBLIC, fittrack_app;

-- New objects created by the current migration owner also start deny-by-default.
DO $$
DECLARE
    migration_owner TEXT := current_user;
BEGIN
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON TABLES FROM PUBLIC', migration_owner);
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON TABLES FROM fittrack_app', migration_owner);
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON SEQUENCES FROM PUBLIC', migration_owner);
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON SEQUENCES FROM fittrack_app', migration_owner);
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON FUNCTIONS FROM PUBLIC', migration_owner);
    EXECUTE format('ALTER DEFAULT PRIVILEGES FOR ROLE %I IN SCHEMA public REVOKE ALL PRIVILEGES ON FUNCTIONS FROM fittrack_app', migration_owner);
END $$;

GRANT USAGE ON SCHEMA public TO fittrack_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE
    users, workout, exercise, "set", user_feature_access,
    ai_chat_conversation, ai_chat_message, ai_chat_run,
    user_training_profile
TO fittrack_app;
GRANT SELECT, INSERT, DELETE ON ai_chat_stream_chunk TO fittrack_app;
GRANT SELECT, INSERT, UPDATE ON TABLE
    stripe_customers, stripe_subscriptions, ai_chat_trial_prompt_usage
TO fittrack_app;

-- Webhook idempotency is reachable only through the owner-backed functions below.
REVOKE ALL PRIVILEGES ON TABLE public.stripe_webhook_events FROM fittrack_app;

GRANT USAGE ON SEQUENCE
    users_id_seq, workout_id_seq, exercise_id_seq, set_id_seq,
    user_feature_access_id_seq, ai_chat_conversation_id_seq,
    ai_chat_message_id_seq, ai_chat_run_id_seq
TO fittrack_app;

GRANT EXECUTE ON FUNCTION current_user_id() TO fittrack_app;
DO $$
DECLARE
    api_role TEXT;
BEGIN
    IF to_regprocedure('public.lookup_stripe_customer_user_id(text)') IS NOT NULL THEN
        GRANT EXECUTE ON FUNCTION lookup_stripe_customer_user_id(TEXT) TO fittrack_app;
    END IF;

    IF to_regprocedure('public.has_processed_stripe_webhook_event(text)') IS NOT NULL
       AND to_regprocedure('public.record_stripe_webhook_event(text, text)') IS NOT NULL THEN
        FOREACH api_role IN ARRAY ARRAY['anon', 'authenticated', 'service_role']
        LOOP
            IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = api_role) THEN
                EXECUTE format(
                    'REVOKE ALL PRIVILEGES ON FUNCTION public.has_processed_stripe_webhook_event(TEXT), public.record_stripe_webhook_event(TEXT, TEXT) FROM %I',
                    api_role
                );
            END IF;
        END LOOP;

        GRANT EXECUTE ON FUNCTION public.has_processed_stripe_webhook_event(TEXT) TO fittrack_app;
        GRANT EXECUTE ON FUNCTION public.record_stripe_webhook_event(TEXT, TEXT) TO fittrack_app;
    END IF;
END $$;
