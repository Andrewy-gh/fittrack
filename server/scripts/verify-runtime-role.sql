-- Run with the candidate runtime DATABASE_URL. This script is read-only apart from
-- changing the current session's app.current_user_id setting.

DO $$
DECLARE
    role_superuser BOOLEAN;
    role_bypassrls BOOLEAN;
    owns_rls_table BOOLEAN;
    can_assume_bypass_role BOOLEAN;
    visible_without_user BIGINT;
    webhook_events REGCLASS;
    webhook_rls_enabled BOOLEAN;
    webhook_rls_forced BOOLEAN;
    webhook_check_function REGPROCEDURE;
    webhook_record_function REGPROCEDURE;
BEGIN
    SELECT
        r.rolsuper,
        r.rolbypassrls,
        EXISTS (
            SELECT 1 FROM pg_class c
            WHERE c.relowner = r.oid AND c.relrowsecurity
        ),
        EXISTS (
            SELECT 1
            FROM pg_roles inherited
            WHERE inherited.oid <> r.oid
              AND pg_has_role(r.oid, inherited.oid, 'MEMBER')
              AND (
                  inherited.rolsuper
                  OR inherited.rolbypassrls
                  OR EXISTS (
                      SELECT 1 FROM pg_class owned
                      WHERE owned.relowner = inherited.oid AND owned.relrowsecurity
                  )
              )
        )
    INTO role_superuser, role_bypassrls, owns_rls_table, can_assume_bypass_role
    FROM pg_roles r
    WHERE r.rolname = current_user;

    IF role_superuser OR role_bypassrls OR owns_rls_table OR can_assume_bypass_role THEN
        RAISE EXCEPTION 'runtime role can bypass RLS';
    END IF;

    IF has_schema_privilege(current_user, 'public', 'CREATE') THEN
        RAISE EXCEPTION 'runtime role can create public-schema objects';
    END IF;
    IF has_table_privilege(current_user, 'goose_db_version', 'SELECT')
       OR has_table_privilege(current_user, 'goose_db_version', 'UPDATE') THEN
        RAISE EXCEPTION 'runtime role can access Goose migration metadata';
    END IF;

    webhook_events := to_regclass('public.stripe_webhook_events');
    IF webhook_events IS NULL THEN
        RAISE EXCEPTION 'stripe_webhook_events is missing';
    END IF;

    SELECT c.relrowsecurity, c.relforcerowsecurity
    INTO webhook_rls_enabled, webhook_rls_forced
    FROM pg_class AS c
    WHERE c.oid = webhook_events;

    IF webhook_rls_enabled IS DISTINCT FROM TRUE
       OR webhook_rls_forced IS DISTINCT FROM FALSE THEN
        RAISE EXCEPTION 'stripe_webhook_events must enable RLS without forcing the table owner';
    END IF;

    IF has_table_privilege(current_user, webhook_events, 'SELECT')
       OR has_table_privilege(current_user, webhook_events, 'INSERT')
       OR has_table_privilege(current_user, webhook_events, 'UPDATE')
       OR has_table_privilege(current_user, webhook_events, 'DELETE') THEN
        RAISE EXCEPTION 'runtime role can access stripe_webhook_events directly';
    END IF;

    webhook_check_function := to_regprocedure('public.has_processed_stripe_webhook_event(text)');
    webhook_record_function := to_regprocedure('public.record_stripe_webhook_event(text, text)');
    IF webhook_check_function IS NULL OR webhook_record_function IS NULL THEN
        RAISE EXCEPTION 'stripe webhook idempotency functions are missing';
    END IF;

    IF NOT has_function_privilege(current_user, webhook_check_function, 'EXECUTE')
       OR NOT has_function_privilege(current_user, webhook_record_function, 'EXECUTE') THEN
        RAISE EXCEPTION 'runtime role is missing Stripe webhook idempotency function privileges';
    END IF;

    IF has_function_privilege('public', webhook_check_function, 'EXECUTE')
       OR has_function_privilege('public', webhook_record_function, 'EXECUTE') THEN
        RAISE EXCEPTION 'Stripe webhook idempotency functions are executable by PUBLIC';
    END IF;

    IF NOT has_table_privilege(current_user, 'workout', 'SELECT')
       OR NOT has_function_privilege(current_user, 'current_user_id()', 'EXECUTE')
       OR NOT has_function_privilege(current_user, 'lookup_stripe_customer_user_id(text)', 'EXECUTE') THEN
        RAISE EXCEPTION 'runtime role is missing required privileges';
    END IF;

    PERFORM set_config('app.current_user_id', '__fittrack_no_such_user__', false);
    SELECT count(*) INTO visible_without_user FROM workout;
    IF visible_without_user <> 0 THEN
        RAISE EXCEPTION 'nonexistent user context can see % workout rows', visible_without_user;
    END IF;
END $$;

SELECT current_user AS verified_runtime_role;
