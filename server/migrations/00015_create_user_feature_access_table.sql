-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_feature_access (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    feature_key VARCHAR(128) NOT NULL,
    source VARCHAR(64) NOT NULL,
    source_reference VARCHAR(256),
    granted_by VARCHAR(256),
    note VARCHAR(512),
    starts_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_feature_access_feature_key_not_empty CHECK (btrim(feature_key) <> ''),
    CONSTRAINT user_feature_access_source_not_empty CHECK (btrim(source) <> ''),
    CONSTRAINT user_feature_access_expiry_after_start CHECK (
        expires_at IS NULL OR expires_at > starts_at
    ),
    CONSTRAINT user_feature_access_revoked_after_start CHECK (
        revoked_at IS NULL OR revoked_at >= starts_at
    )
);

CREATE INDEX idx_user_feature_access_user_feature
    ON user_feature_access (user_id, feature_key, starts_at DESC);

CREATE INDEX idx_user_feature_access_active_lookup
    ON user_feature_access (user_id, starts_at DESC)
    WHERE revoked_at IS NULL;

ALTER TABLE user_feature_access ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'user_feature_access'
          AND policyname = 'user_feature_access_select_policy'
    ) THEN
        CREATE POLICY user_feature_access_select_policy ON user_feature_access
            FOR SELECT TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'user_feature_access'
          AND policyname = 'user_feature_access_insert_policy'
    ) THEN
        CREATE POLICY user_feature_access_insert_policy ON user_feature_access
            FOR INSERT TO PUBLIC
            WITH CHECK (user_id = current_user_id());
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE tablename = 'user_feature_access'
          AND policyname = 'user_feature_access_update_policy'
    ) THEN
        CREATE POLICY user_feature_access_update_policy ON user_feature_access
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
        WHERE tablename = 'user_feature_access'
          AND policyname = 'user_feature_access_delete_policy'
    ) THEN
        CREATE POLICY user_feature_access_delete_policy ON user_feature_access
            FOR DELETE TO PUBLIC
            USING (user_id = current_user_id());
    END IF;
END $$;

GRANT SELECT, INSERT, UPDATE, DELETE ON user_feature_access TO PUBLIC;
GRANT USAGE ON SEQUENCE user_feature_access_id_seq TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS user_feature_access_select_policy ON user_feature_access;
DROP POLICY IF EXISTS user_feature_access_insert_policy ON user_feature_access;
DROP POLICY IF EXISTS user_feature_access_update_policy ON user_feature_access;
DROP POLICY IF EXISTS user_feature_access_delete_policy ON user_feature_access;

ALTER TABLE user_feature_access DISABLE ROW LEVEL SECURITY;

REVOKE ALL ON user_feature_access FROM PUBLIC;
REVOKE ALL ON SEQUENCE user_feature_access_id_seq FROM PUBLIC;

DROP TABLE IF EXISTS user_feature_access;
-- +goose StatementEnd
