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

-- Exercises table  
CREATE TABLE exercise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    historical_1rm NUMERIC(8,2),
    historical_1rm_updated_at TIMESTAMPTZ,
    historical_1rm_source_workout_id INTEGER REFERENCES workout(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT exercise_user_id_name_key UNIQUE (user_id, name)
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
CREATE INDEX idx_workout_user_date ON workout(user_id, date);
CREATE INDEX idx_user_feature_access_user_feature ON user_feature_access(user_id, feature_key, starts_at DESC);
CREATE INDEX idx_user_feature_access_active_lookup ON user_feature_access(user_id, starts_at DESC) WHERE revoked_at IS NULL;
