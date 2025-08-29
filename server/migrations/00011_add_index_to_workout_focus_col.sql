-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_workout_user_focus ON workout(user_id, workout_focus);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_workout_user_focus;
-- +goose StatementEnd