-- +goose Up
-- +goose StatementBegin
ALTER TABLE workout
ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE exercise
ALTER COLUMN user_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE workout
ALTER COLUMN user_id DROP NOT NULL;

ALTER TABLE exercise
ALTER COLUMN user_id DROP NOT NULL;
-- +goose StatementEnd
