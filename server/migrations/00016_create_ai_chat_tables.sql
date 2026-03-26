-- +goose Up
-- +goose StatementBegin
CREATE TABLE ai_chat_conversation (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    title VARCHAR(256),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMPTZ,
    CONSTRAINT ai_chat_conversation_title_not_empty CHECK (
        title IS NULL OR btrim(title) <> ''
    )
);

CREATE TABLE ai_chat_message (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER NOT NULL REFERENCES ai_chat_conversation(id) ON DELETE CASCADE,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL,
    error_message VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    CONSTRAINT ai_chat_message_role_valid CHECK (
        role IN ('user', 'assistant')
    ),
    CONSTRAINT ai_chat_message_status_valid CHECK (
        status IN ('streaming', 'completed', 'failed')
    ),
    CONSTRAINT ai_chat_message_error_not_empty CHECK (
        error_message IS NULL OR btrim(error_message) <> ''
    )
);

CREATE TABLE ai_chat_run (
    id SERIAL PRIMARY KEY,
    conversation_id INTEGER NOT NULL REFERENCES ai_chat_conversation(id) ON DELETE CASCADE,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    user_message_id INTEGER NOT NULL REFERENCES ai_chat_message(id) ON DELETE CASCADE,
    assistant_message_id INTEGER NOT NULL REFERENCES ai_chat_message(id) ON DELETE CASCADE,
    model VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    request_id VARCHAR(128),
    error_message VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    CONSTRAINT ai_chat_run_status_valid CHECK (
        status IN ('streaming', 'completed', 'failed')
    ),
    CONSTRAINT ai_chat_run_model_not_empty CHECK (btrim(model) <> ''),
    CONSTRAINT ai_chat_run_request_id_not_empty CHECK (
        request_id IS NULL OR btrim(request_id) <> ''
    ),
    CONSTRAINT ai_chat_run_error_not_empty CHECK (
        error_message IS NULL OR btrim(error_message) <> ''
    ),
    CONSTRAINT ai_chat_run_user_message_unique UNIQUE (user_message_id),
    CONSTRAINT ai_chat_run_assistant_message_unique UNIQUE (assistant_message_id)
);

CREATE INDEX idx_ai_chat_conversation_user_updated
    ON ai_chat_conversation (user_id, updated_at DESC, id DESC);

CREATE INDEX idx_ai_chat_message_conversation
    ON ai_chat_message (conversation_id, id ASC);

CREATE INDEX idx_ai_chat_message_user_conversation
    ON ai_chat_message (user_id, conversation_id, id ASC);

CREATE INDEX idx_ai_chat_run_conversation_created
    ON ai_chat_run (conversation_id, created_at DESC, id DESC);

CREATE UNIQUE INDEX idx_ai_chat_run_active_conversation
    ON ai_chat_run (conversation_id)
    WHERE status = 'streaming';

ALTER TABLE ai_chat_conversation ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_chat_message ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_chat_run ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_conversation'
          AND policyname = 'ai_chat_conversation_select_policy'
    ) THEN
        CREATE POLICY ai_chat_conversation_select_policy ON ai_chat_conversation
            FOR SELECT TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_conversation'
          AND policyname = 'ai_chat_conversation_insert_policy'
    ) THEN
        CREATE POLICY ai_chat_conversation_insert_policy ON ai_chat_conversation
            FOR INSERT TO PUBLIC
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_conversation'
          AND policyname = 'ai_chat_conversation_update_policy'
    ) THEN
        CREATE POLICY ai_chat_conversation_update_policy ON ai_chat_conversation
            FOR UPDATE TO PUBLIC
            USING (user_id = current_user_id())
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_conversation'
          AND policyname = 'ai_chat_conversation_delete_policy'
    ) THEN
        CREATE POLICY ai_chat_conversation_delete_policy ON ai_chat_conversation
            FOR DELETE TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_message'
          AND policyname = 'ai_chat_message_select_policy'
    ) THEN
        CREATE POLICY ai_chat_message_select_policy ON ai_chat_message
            FOR SELECT TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_message'
          AND policyname = 'ai_chat_message_insert_policy'
    ) THEN
        CREATE POLICY ai_chat_message_insert_policy ON ai_chat_message
            FOR INSERT TO PUBLIC
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_message'
          AND policyname = 'ai_chat_message_update_policy'
    ) THEN
        CREATE POLICY ai_chat_message_update_policy ON ai_chat_message
            FOR UPDATE TO PUBLIC
            USING (user_id = current_user_id())
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_message'
          AND policyname = 'ai_chat_message_delete_policy'
    ) THEN
        CREATE POLICY ai_chat_message_delete_policy ON ai_chat_message
            FOR DELETE TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_run'
          AND policyname = 'ai_chat_run_select_policy'
    ) THEN
        CREATE POLICY ai_chat_run_select_policy ON ai_chat_run
            FOR SELECT TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_run'
          AND policyname = 'ai_chat_run_insert_policy'
    ) THEN
        CREATE POLICY ai_chat_run_insert_policy ON ai_chat_run
            FOR INSERT TO PUBLIC
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_run'
          AND policyname = 'ai_chat_run_update_policy'
    ) THEN
        CREATE POLICY ai_chat_run_update_policy ON ai_chat_run
            FOR UPDATE TO PUBLIC
            USING (user_id = current_user_id())
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_run'
          AND policyname = 'ai_chat_run_delete_policy'
    ) THEN
        CREATE POLICY ai_chat_run_delete_policy ON ai_chat_run
            FOR DELETE TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

GRANT SELECT, INSERT, UPDATE, DELETE ON ai_chat_conversation TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON ai_chat_message TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON ai_chat_run TO PUBLIC;
GRANT USAGE ON SEQUENCE ai_chat_conversation_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE ai_chat_message_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE ai_chat_run_id_seq TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS ai_chat_run_select_policy ON ai_chat_run;
DROP POLICY IF EXISTS ai_chat_run_insert_policy ON ai_chat_run;
DROP POLICY IF EXISTS ai_chat_run_update_policy ON ai_chat_run;
DROP POLICY IF EXISTS ai_chat_run_delete_policy ON ai_chat_run;

DROP POLICY IF EXISTS ai_chat_message_select_policy ON ai_chat_message;
DROP POLICY IF EXISTS ai_chat_message_insert_policy ON ai_chat_message;
DROP POLICY IF EXISTS ai_chat_message_update_policy ON ai_chat_message;
DROP POLICY IF EXISTS ai_chat_message_delete_policy ON ai_chat_message;

DROP POLICY IF EXISTS ai_chat_conversation_select_policy ON ai_chat_conversation;
DROP POLICY IF EXISTS ai_chat_conversation_insert_policy ON ai_chat_conversation;
DROP POLICY IF EXISTS ai_chat_conversation_update_policy ON ai_chat_conversation;
DROP POLICY IF EXISTS ai_chat_conversation_delete_policy ON ai_chat_conversation;

ALTER TABLE ai_chat_run DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_chat_message DISABLE ROW LEVEL SECURITY;
ALTER TABLE ai_chat_conversation DISABLE ROW LEVEL SECURITY;

REVOKE ALL ON ai_chat_run FROM PUBLIC;
REVOKE ALL ON ai_chat_message FROM PUBLIC;
REVOKE ALL ON ai_chat_conversation FROM PUBLIC;
REVOKE ALL ON SEQUENCE ai_chat_run_id_seq FROM PUBLIC;
REVOKE ALL ON SEQUENCE ai_chat_message_id_seq FROM PUBLIC;
REVOKE ALL ON SEQUENCE ai_chat_conversation_id_seq FROM PUBLIC;

DROP TABLE IF EXISTS ai_chat_run;
DROP TABLE IF EXISTS ai_chat_message;
DROP TABLE IF EXISTS ai_chat_conversation;
-- +goose StatementEnd
