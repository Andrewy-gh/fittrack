-- Run with the candidate runtime DATABASE_URL. This script is read-only apart from
-- changing the current session's app.current_user_id setting.

DO $$
DECLARE
    role_superuser BOOLEAN;
    role_bypassrls BOOLEAN;
    owns_rls_table BOOLEAN;
    can_assume_bypass_role BOOLEAN;
    visible_without_user BIGINT;
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
