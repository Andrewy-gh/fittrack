-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(256) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Workouts table
CREATE TABLE workout (
    id SERIAL PRIMARY KEY,
    date TIMESTAMPTZ NOT NULL,
    notes VARCHAR(256),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE
);

-- Exercises table  
CREATE TABLE exercise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT exercise_user_id_name_key UNIQUE (user_id, name)
);

-- Sets table
CREATE TABLE "set" (
    id SERIAL PRIMARY KEY,
    exercise_id INTEGER NOT NULL REFERENCES exercise(id),
    workout_id INTEGER NOT NULL REFERENCES workout(id) ON DELETE CASCADE,
    weight INTEGER,
    reps INTEGER NOT NULL,
    set_type VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    exercise_order INTEGER,
    set_order INTEGER
);

-- Indexes for foreign keys
CREATE INDEX idx_set_exercise_id ON "set"(exercise_id);
CREATE INDEX idx_set_workout_id ON "set"(workout_id);
CREATE INDEX idx_set_user_id ON "set"(user_id);

-- Additional indexes for performance
CREATE INDEX idx_workout_user_id ON workout(user_id);
CREATE INDEX idx_exercise_user_id ON exercise(user_id);
CREATE INDEX idx_workout_user_date ON workout(user_id, date);