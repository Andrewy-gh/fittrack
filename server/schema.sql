-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(256) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User feature access table
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

-- AI chat conversations table
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

-- AI chat messages table
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

-- AI chat runs table
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

-- Workouts table
CREATE TABLE workout (
    id SERIAL PRIMARY KEY,
    date TIMESTAMPTZ NOT NULL,
    notes VARCHAR(256),
    workout_focus VARCHAR(256),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE exercise_template (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(256) NOT NULL UNIQUE,
    name VARCHAR(256) NOT NULL,
    instructions TEXT,
    category VARCHAR(128),
    equipment VARCHAR(128),
    primary_muscle_group VARCHAR(128),
    secondary_muscle_groups TEXT[] NOT NULL DEFAULT '{}',
    source VARCHAR(64) NOT NULL,
    source_id VARCHAR(256) NOT NULL,
    CONSTRAINT exercise_template_slug_not_empty CHECK (btrim(slug) <> ''),
    CONSTRAINT exercise_template_name_not_empty CHECK (btrim(name) <> ''),
    CONSTRAINT exercise_template_source_not_empty CHECK (btrim(source) <> ''),
    CONSTRAINT exercise_template_source_id_not_empty CHECK (btrim(source_id) <> ''),
    CONSTRAINT exercise_template_source_source_id_key UNIQUE (source, source_id)
);

-- Exercises table  
CREATE TABLE exercise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    kind VARCHAR(32) NOT NULL DEFAULT 'custom',
    template_id INTEGER REFERENCES exercise_template(id),
    instructions TEXT,
    equipment VARCHAR(128),
    primary_muscle_group VARCHAR(128),
    secondary_muscle_groups TEXT[] NOT NULL DEFAULT '{}',
    historical_1rm NUMERIC(8,2),
    historical_1rm_updated_at TIMESTAMPTZ,
    historical_1rm_source_workout_id INTEGER REFERENCES workout(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT exercise_user_id_name_key UNIQUE (user_id, name),
    CONSTRAINT exercise_kind_valid CHECK (kind IN ('custom', 'template_based')),
    CONSTRAINT exercise_kind_template_state CHECK (
        (kind = 'custom' AND template_id IS NULL)
        OR (kind = 'template_based' AND template_id IS NOT NULL)
    )
);

-- Sets table
CREATE TABLE "set" (
    id SERIAL PRIMARY KEY,
    exercise_id INTEGER NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
    workout_id INTEGER NOT NULL REFERENCES workout(id) ON DELETE CASCADE,
    weight NUMERIC(10,1),
    reps INTEGER NOT NULL,
    set_type VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    exercise_order INTEGER NOT NULL,
    set_order INTEGER NOT NULL,
    CONSTRAINT weight_non_negative CHECK (weight IS NULL OR weight >= 0)
);

-- Indexes for foreign keys
CREATE INDEX idx_set_exercise_id ON "set"(exercise_id);
CREATE INDEX idx_set_workout_id ON "set"(workout_id);
CREATE INDEX idx_set_user_id ON "set"(user_id);
CREATE INDEX idx_set_user_exercise_id ON "set"(user_id, exercise_id);

-- Additional indexes for performance
CREATE INDEX idx_workout_user_id ON workout(user_id);
CREATE INDEX idx_exercise_user_id ON exercise(user_id);
CREATE INDEX idx_exercise_template_id ON exercise(template_id);
CREATE INDEX idx_workout_user_date ON workout(user_id, date);
CREATE INDEX idx_user_feature_access_user_feature ON user_feature_access(user_id, feature_key, starts_at DESC);
CREATE INDEX idx_user_feature_access_active_lookup ON user_feature_access(user_id, starts_at DESC) WHERE revoked_at IS NULL;
CREATE INDEX idx_ai_chat_conversation_user_updated ON ai_chat_conversation(user_id, updated_at DESC, id DESC);
CREATE INDEX idx_ai_chat_message_conversation ON ai_chat_message(conversation_id, id ASC);
CREATE INDEX idx_ai_chat_message_user_conversation ON ai_chat_message(user_id, conversation_id, id ASC);
CREATE INDEX idx_ai_chat_run_conversation_created ON ai_chat_run(conversation_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX idx_ai_chat_run_active_conversation ON ai_chat_run(conversation_id) WHERE status = 'streaming';
