-- +goose Up
-- +goose StatementBegin
ALTER TABLE ai_chat_conversation
ADD CONSTRAINT ai_chat_conversation_user_id_id_unique UNIQUE (user_id, id);

ALTER TABLE ai_chat_message
ADD CONSTRAINT ai_chat_message_user_conversation_id_unique UNIQUE (user_id, conversation_id, id);

DROP INDEX IF EXISTS idx_ai_chat_message_user_conversation;

CREATE TABLE user_training_profile (
    user_id VARCHAR(256) PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
    primary_goal VARCHAR(64),
    experience_level VARCHAR(32),
    preferred_session_duration_minutes INTEGER,
    usual_training_location VARCHAR(64),
    available_equipment JSONB NOT NULL DEFAULT '[]'::jsonb,
    avoided_exercises JSONB NOT NULL DEFAULT '[]'::jsonb,
    movement_limitations JSONB NOT NULL DEFAULT '[]'::jsonb,
    source_conversation_id INTEGER,
    source_message_id INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT user_training_profile_primary_goal_valid CHECK (
        primary_goal IS NULL OR primary_goal IN (
            'strength',
            'hypertrophy',
            'endurance',
            'general_fitness',
            'weight_loss',
            'mobility'
        )
    ),
    CONSTRAINT user_training_profile_experience_level_valid CHECK (
        experience_level IS NULL OR experience_level IN (
            'beginner',
            'intermediate',
            'advanced'
        )
    ),
    CONSTRAINT user_training_profile_duration_bounds CHECK (
        preferred_session_duration_minutes IS NULL
        OR preferred_session_duration_minutes BETWEEN 10 AND 240
    ),
    CONSTRAINT user_training_profile_location_valid CHECK (
        usual_training_location IS NULL OR usual_training_location IN (
            'gym',
            'home',
            'outdoor',
            'travel'
        )
    ),
    CONSTRAINT user_training_profile_available_equipment_array CHECK (
        jsonb_typeof(available_equipment) = 'array'
    ),
    CONSTRAINT user_training_profile_avoided_exercises_array CHECK (
        jsonb_typeof(avoided_exercises) = 'array'
    ),
    CONSTRAINT user_training_profile_movement_limitations_array CHECK (
        jsonb_typeof(movement_limitations) = 'array'
    ),
    CONSTRAINT user_training_profile_source_message_requires_conversation CHECK (
        source_message_id IS NULL OR source_conversation_id IS NOT NULL
    ),
    CONSTRAINT user_training_profile_source_conversation_owner_fk FOREIGN KEY (
        user_id,
        source_conversation_id
    ) REFERENCES ai_chat_conversation (
        user_id,
        id
    ) ON DELETE SET NULL (source_conversation_id),
    CONSTRAINT user_training_profile_source_message_owner_conversation_fk FOREIGN KEY (
        user_id,
        source_conversation_id,
        source_message_id
    ) REFERENCES ai_chat_message (
        user_id,
        conversation_id,
        id
    ) ON DELETE SET NULL (source_message_id)
);

ALTER TABLE user_training_profile ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_training_profile_select_policy ON user_training_profile
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY user_training_profile_insert_policy ON user_training_profile
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY user_training_profile_update_policy ON user_training_profile
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY user_training_profile_delete_policy ON user_training_profile
    FOR DELETE TO PUBLIC
    USING (user_id = current_user_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON user_training_profile TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS user_training_profile_delete_policy ON user_training_profile;
DROP POLICY IF EXISTS user_training_profile_update_policy ON user_training_profile;
DROP POLICY IF EXISTS user_training_profile_insert_policy ON user_training_profile;
DROP POLICY IF EXISTS user_training_profile_select_policy ON user_training_profile;

ALTER TABLE user_training_profile DISABLE ROW LEVEL SECURITY;

REVOKE ALL ON user_training_profile FROM PUBLIC;

DROP TABLE IF EXISTS user_training_profile;

ALTER TABLE ai_chat_message
DROP CONSTRAINT IF EXISTS ai_chat_message_user_conversation_id_unique;

CREATE INDEX IF NOT EXISTS idx_ai_chat_message_user_conversation
    ON ai_chat_message (user_id, conversation_id, id ASC);

ALTER TABLE ai_chat_conversation
DROP CONSTRAINT IF EXISTS ai_chat_conversation_user_id_id_unique;
-- +goose StatementEnd
