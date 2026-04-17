-- +goose Up
-- +goose StatementBegin
CREATE TABLE ai_chat_stream_chunk (
    run_id INTEGER NOT NULL REFERENCES ai_chat_run(id) ON DELETE CASCADE,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    delta_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (run_id, sequence),
    CONSTRAINT ai_chat_stream_chunk_sequence_positive CHECK (sequence > 0),
    CONSTRAINT ai_chat_stream_chunk_delta_not_empty CHECK (btrim(delta_text) <> '')
);

CREATE INDEX idx_ai_chat_stream_chunk_user_run_sequence
    ON ai_chat_stream_chunk (user_id, run_id, sequence ASC);

ALTER TABLE ai_chat_stream_chunk ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_stream_chunk'
          AND policyname = 'ai_chat_stream_chunk_select_policy'
    ) THEN
        CREATE POLICY ai_chat_stream_chunk_select_policy ON ai_chat_stream_chunk
            FOR SELECT TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_stream_chunk'
          AND policyname = 'ai_chat_stream_chunk_insert_policy'
    ) THEN
        CREATE POLICY ai_chat_stream_chunk_insert_policy ON ai_chat_stream_chunk
            FOR INSERT TO PUBLIC
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'ai_chat_stream_chunk'
          AND policyname = 'ai_chat_stream_chunk_delete_policy'
    ) THEN
        CREATE POLICY ai_chat_stream_chunk_delete_policy ON ai_chat_stream_chunk
            FOR DELETE TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

GRANT SELECT, INSERT, DELETE ON ai_chat_stream_chunk TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS ai_chat_stream_chunk_select_policy ON ai_chat_stream_chunk;
DROP POLICY IF EXISTS ai_chat_stream_chunk_insert_policy ON ai_chat_stream_chunk;
DROP POLICY IF EXISTS ai_chat_stream_chunk_delete_policy ON ai_chat_stream_chunk;

ALTER TABLE ai_chat_stream_chunk DISABLE ROW LEVEL SECURITY;

REVOKE ALL ON ai_chat_stream_chunk FROM PUBLIC;

DROP TABLE IF EXISTS ai_chat_stream_chunk;
-- +goose StatementEnd
