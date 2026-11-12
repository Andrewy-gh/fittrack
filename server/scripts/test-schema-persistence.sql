-- Run with psql -X -v ON_ERROR_STOP=1 against a DISPOSABLE database only.
-- Exercises both the migrated database and a database built from schema.sql.
-- Fixtures roll back; SERIAL sequences can still advance. No provider calls.
BEGIN;

CREATE FUNCTION pg_temp.expect_constraint_failure(statement TEXT, expected_state TEXT, expected_constraint TEXT)
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE actual_state TEXT; actual_constraint TEXT;
BEGIN
    BEGIN
        EXECUTE statement;
    EXCEPTION WHEN OTHERS THEN
        GET STACKED DIAGNOSTICS actual_state = RETURNED_SQLSTATE, actual_constraint = CONSTRAINT_NAME;
        IF actual_state = expected_state AND actual_constraint = expected_constraint THEN
            RETURN;
        END IF;
        RAISE EXCEPTION 'Expected % / %, got % / %', expected_state, expected_constraint, actual_state, actual_constraint;
    END;
    RAISE EXCEPTION 'Expected % / %, but statement succeeded', expected_state, expected_constraint;
END $$;

DO $$
DECLARE
    test_user TEXT := 'schema-persistence-' || txid_current();
    conversation INTEGER; user_message INTEGER; assistant_message INTEGER;
    run INTEGER; saved_workout INTEGER; other_run INTEGER;
    current_status TEXT;
BEGIN
    INSERT INTO users(user_id) VALUES (test_user);
    INSERT INTO ai_chat_conversation(user_id) VALUES (test_user) RETURNING id INTO conversation;
    INSERT INTO ai_chat_message(conversation_id, user_id, role, status)
        VALUES (conversation, test_user, 'user', 'completed') RETURNING id INTO user_message;
    INSERT INTO ai_chat_message(conversation_id, user_id, role, status)
        VALUES (conversation, test_user, 'assistant', 'stopped') RETURNING id INTO assistant_message;
    INSERT INTO ai_chat_run(conversation_id, user_id, user_message_id, assistant_message_id, model, status, generation_status)
        VALUES (conversation, test_user, user_message, assistant_message, 'local-test', 'stopped', 'stopped')
        RETURNING id INTO run;

    -- All existing statuses remain legal, including interrupted generation.
    FOREACH current_status IN ARRAY ARRAY['streaming', 'completed', 'failed', 'stopped'] LOOP
        UPDATE ai_chat_message SET status = current_status WHERE id = assistant_message;
        UPDATE ai_chat_run SET status = current_status WHERE id = run;
    END LOOP;
    FOREACH current_status IN ARRAY ARRAY['queued', 'generating', 'completed', 'failed', 'interrupted', 'stopped'] LOOP
        UPDATE ai_chat_run SET generation_status = current_status WHERE id = run;
    END LOOP;
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE ai_chat_message SET status = ''bogus'' WHERE id = %s', assistant_message), '23514', 'ai_chat_message_status_valid');
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE ai_chat_run SET status = ''bogus'' WHERE id = %s', run), '23514', 'ai_chat_run_status_valid');
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE ai_chat_run SET generation_status = ''bogus'' WHERE id = %s', run), '23514', 'ai_chat_run_generation_status_valid');
    IF NOT EXISTS (SELECT 1 FROM ai_chat_run WHERE id = run AND status = 'stopped' AND generation_status = 'stopped')
       OR NOT EXISTS (SELECT 1 FROM ai_chat_message WHERE id = assistant_message AND status = 'stopped') THEN
        RAISE EXCEPTION 'Stopped status did not persist';
    END IF;

    INSERT INTO workout(user_id, date) VALUES (test_user, CURRENT_TIMESTAMP) RETURNING id INTO saved_workout;
    UPDATE workout SET recommendation_context = '[{"reason":"local fixture"}]'::jsonb WHERE id = saved_workout;
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE workout SET recommendation_context = ''{}''::jsonb WHERE id = %s', saved_workout), '23514', 'workout_recommendation_context_check');
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE workout SET recommendation_context = ''null''::jsonb WHERE id = %s', saved_workout), '23514', 'workout_recommendation_context_check');

    UPDATE ai_chat_conversation SET latest_workout_draft_source_run_id = run,
        latest_workout_draft_saved_workout_id = saved_workout WHERE id = conversation;
    -- Reserve then delete a real ID: missing-parent probes cannot accidentally hit a fixture.
    INSERT INTO workout(user_id, date) VALUES (test_user, CURRENT_TIMESTAMP) RETURNING id INTO other_run;
    DELETE FROM workout WHERE id = other_run;
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE ai_chat_conversation SET latest_workout_draft_saved_workout_id = %s WHERE id = %s', other_run, conversation), '23503', 'ai_chat_conversation_latest_workout_draft_saved_workout_id_fkey');
    DELETE FROM ai_chat_run WHERE id = run;
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE ai_chat_conversation SET latest_workout_draft_source_run_id = %s WHERE id = %s', run, conversation), '23503', 'ai_chat_conversation_latest_workout_draft_source_run_id_fkey');
    IF NOT EXISTS (SELECT 1 FROM ai_chat_conversation WHERE id = conversation AND latest_workout_draft_source_run_id IS NULL AND latest_workout_draft_saved_workout_id = saved_workout) THEN
        RAISE EXCEPTION 'Run delete must clear only its saved draft reference';
    END IF;
    DELETE FROM workout WHERE id = saved_workout;
    IF NOT EXISTS (SELECT 1 FROM ai_chat_conversation WHERE id = conversation AND latest_workout_draft_source_run_id IS NULL AND latest_workout_draft_saved_workout_id IS NULL) THEN
        RAISE EXCEPTION 'Workout delete must clear saved workout reference without deleting conversation';
    END IF;

    -- Preserve A010 timestamp defaults and NOT NULL, and existing exercise uniqueness.
    INSERT INTO exercise(user_id, name) VALUES (test_user, 'schema timestamp fixture');
    IF NOT EXISTS (SELECT 1 FROM exercise WHERE user_id = test_user AND created_at = updated_at AND updated_at IS NOT NULL) THEN
        RAISE EXCEPTION 'Exercise timestamps must default to the same creation time';
    END IF;
    PERFORM pg_temp.expect_constraint_failure(format('INSERT INTO exercise(user_id, name) VALUES (%L, ''schema timestamp fixture'')', test_user), '23505', 'exercise_user_id_name_key');
    PERFORM pg_temp.expect_constraint_failure(format('UPDATE exercise SET updated_at = NULL WHERE user_id = %L', test_user), '23502', '');
    RAISE NOTICE 'PASS: statuses, rejection SQLSTATE/constraint identity, draft FK SET NULL, JSON arrays, timestamps and uniqueness';
END $$;

ROLLBACK;
