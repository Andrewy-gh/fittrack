-- +goose Up
-- +goose StatementBegin
ALTER TABLE workout
ADD COLUMN user_id VARCHAR(256) REFERENCES users(user_id) ON DELETE CASCADE;

ALTER TABLE exercise
ADD COLUMN user_id VARCHAR(256) REFERENCES users(user_id) ON DELETE CASCADE;

ALTER TABLE exercise
ADD CONSTRAINT exercise_user_id_name_key UNIQUE NULLS NOT DISTINCT (user_id, name);

CREATE INDEX idx_workout_user_id ON workout(user_id);
CREATE INDEX idx_exercise_user_id ON exercise(user_id);
CREATE INDEX idx_workout_user_date ON workout(user_id, date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_workout_user_date;
DROP INDEX IF EXISTS idx_exercise_user_id;
DROP INDEX IF EXISTS idx_workout_user_id;

ALTER TABLE exercise
DROP CONSTRAINT IF EXISTS exercise_user_id_name_key;

ALTER TABLE exercise
DROP COLUMN IF EXISTS user_id;

ALTER TABLE workout
DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd