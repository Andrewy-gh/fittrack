-- +goose Up
-- +goose StatementBegin
ALTER TABLE workout ADD COLUMN workout_focus VARCHAR(256);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE workout DROP COLUMN workout_focus;
-- +goose StatementEnd