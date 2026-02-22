-- +goose Up
ALTER TABLE exercise
ADD COLUMN historical_1rm NUMERIC(8,2),
ADD COLUMN historical_1rm_updated_at TIMESTAMPTZ,
ADD COLUMN historical_1rm_source_workout_id INTEGER REFERENCES workout(id);

-- Speed up metrics-history queries (filtering by user_id + exercise_id).
CREATE INDEX IF NOT EXISTS idx_set_user_exercise_id ON "set"(user_id, exercise_id);

-- +goose Down
DROP INDEX IF EXISTS idx_set_user_exercise_id;

ALTER TABLE exercise
DROP COLUMN IF EXISTS historical_1rm_source_workout_id,
DROP COLUMN IF EXISTS historical_1rm_updated_at,
DROP COLUMN IF EXISTS historical_1rm;

