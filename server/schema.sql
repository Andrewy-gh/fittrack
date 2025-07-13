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
    updated_at TIMESTAMPTZ
);

-- Exercises table  
CREATE TABLE exercise (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

-- Sets table
CREATE TABLE "set" (
    id SERIAL PRIMARY KEY,
    exercise_id INTEGER NOT NULL REFERENCES exercise(id),
    workout_id INTEGER NOT NULL REFERENCES workout(id),
    weight INTEGER,
    reps INTEGER NOT NULL,
    set_type VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ
);

-- Indexes for foreign keys (optional but recommended for performance)
CREATE INDEX idx_set_exercise_id ON "set"(exercise_id);
CREATE INDEX idx_set_workout_id ON "set"(workout_id);