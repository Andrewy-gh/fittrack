-- Disposable test database only. Connect as fittrack_app after migrations and
-- provisioning; run with psql -X -v ON_ERROR_STOP=1. Fixture writes roll back,
-- but sequence values may advance. Never run against production.
BEGIN;

DO $$
DECLARE
    user_a CONSTANT TEXT := '__runtime_smoke_a__';
    user_b CONSTANT TEXT := '__runtime_smoke_b__';
    workout_a INTEGER;
    workout_b INTEGER;
    exercise_a INTEGER;
    visible_count BIGINT;
BEGIN
    IF current_user <> 'fittrack_app' THEN
        RAISE EXCEPTION 'connect as fittrack_app, not the migration owner';
    END IF;

    PERFORM set_config('app.current_user_id', user_a, true);
    INSERT INTO users (user_id) VALUES (user_a);
    INSERT INTO workout (user_id, date) VALUES (user_a, now()) RETURNING id INTO workout_a;
    INSERT INTO exercise (user_id, name) VALUES (user_a, 'Runtime prescription test') RETURNING id INTO exercise_a;
    INSERT INTO exercise_prescription (exercise_id, user_id, min_sets, max_sets) VALUES (exercise_a, user_a, 2, 4);
    UPDATE workout SET recommendation_context = '[{"readiness":"normal"}]'::jsonb WHERE id = workout_a;
    INSERT INTO stripe_customers (user_id, stripe_customer_id)
        VALUES (user_a, 'cus_runtime_smoke_a');

    PERFORM set_config('app.current_user_id', user_b, true);
    INSERT INTO users (user_id) VALUES (user_b);
    INSERT INTO workout (user_id, date) VALUES (user_b, now()) RETURNING id INTO workout_b;

    SELECT count(*) INTO visible_count FROM workout;
    IF visible_count <> 1 OR NOT EXISTS (SELECT 1 FROM workout WHERE id = workout_b) THEN
        RAISE EXCEPTION 'user B must see only their own workout';
    END IF;
    IF EXISTS (SELECT 1 FROM stripe_customers WHERE user_id = user_a) THEN
        RAISE EXCEPTION 'user B can read user A billing data';
    END IF;
    IF EXISTS (SELECT 1 FROM exercise_prescription WHERE exercise_id = exercise_a) THEN
        RAISE EXCEPTION 'user B can read user A prescription';
    END IF;
    UPDATE exercise_prescription SET min_sets = 1 WHERE exercise_id = exercise_a;
    IF FOUND THEN RAISE EXCEPTION 'user B can update user A prescription'; END IF;
    BEGIN
        INSERT INTO exercise_prescription(exercise_id, user_id, min_sets, max_sets) VALUES(exercise_a, user_b, 1, 2);
        RAISE EXCEPTION 'user B can prescribe for user A exercise';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;

    UPDATE workout SET notes = 'cross-user update' WHERE id = workout_a;
    IF FOUND THEN
        RAISE EXCEPTION 'user B can update user A workout';
    END IF;
    DELETE FROM workout WHERE id = workout_a;
    IF FOUND THEN
        RAISE EXCEPTION 'user B can delete user A workout';
    END IF;
    BEGIN
        INSERT INTO workout (user_id, date) VALUES (user_a, now());
        RAISE EXCEPTION 'user B can insert a workout for user A';
    EXCEPTION WHEN insufficient_privilege THEN
        NULL; -- Expected RLS WITH CHECK rejection.
    END;

    PERFORM set_config('app.current_user_id', user_a, true);
    UPDATE exercise_prescription SET min_sets = 3 WHERE exercise_id = exercise_a;
    IF NOT FOUND THEN RAISE EXCEPTION 'owner cannot update prescription'; END IF;
    DELETE FROM exercise_prescription WHERE exercise_id = exercise_a;
    IF NOT FOUND THEN RAISE EXCEPTION 'owner cannot clear prescription'; END IF;
    SELECT count(*) INTO visible_count FROM workout;
    IF visible_count <> 1 OR NOT EXISTS (SELECT 1 FROM workout WHERE id = workout_a) THEN
        RAISE EXCEPTION 'user A must see only their own workout';
    END IF;
    UPDATE workout SET notes = 'own update' WHERE id = workout_a;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'user A cannot update their own workout';
    END IF;

    -- Unauthenticated webhooks may use narrow functions, not direct table access.
    PERFORM set_config('app.current_user_id', '', true);
    IF EXISTS (SELECT 1 FROM workout) OR EXISTS (SELECT 1 FROM stripe_customers) THEN
        RAISE EXCEPTION 'missing user context can read user data';
    END IF;
    IF public.lookup_stripe_customer_user_id('cus_runtime_smoke_a') IS DISTINCT FROM user_a THEN
        RAISE EXCEPTION 'metadata-free Stripe customer lookup failed';
    END IF;
    IF public.lookup_stripe_customer_user_id('cus_runtime_smoke_missing') IS NOT NULL THEN
        RAISE EXCEPTION 'unknown Stripe customer resolved to a user';
    END IF;
    IF public.has_processed_stripe_webhook_event('evt_runtime_smoke') THEN
        RAISE EXCEPTION 'unexpected existing webhook fixture';
    END IF;
    PERFORM public.record_stripe_webhook_event('evt_runtime_smoke', 'customer.subscription.updated');
    PERFORM public.record_stripe_webhook_event('evt_runtime_smoke', 'customer.subscription.updated');
    IF NOT public.has_processed_stripe_webhook_event('evt_runtime_smoke') THEN
        RAISE EXCEPTION 'webhook idempotency record was not persisted';
    END IF;
    BEGIN
        PERFORM 1 FROM public.stripe_webhook_events;
        RAISE EXCEPTION 'runtime role can read webhook events directly';
    EXCEPTION WHEN insufficient_privilege THEN
        NULL;
    END;
END $$;

ROLLBACK;
SELECT 'restricted runtime isolation and webhook smoke test passed' AS result;
