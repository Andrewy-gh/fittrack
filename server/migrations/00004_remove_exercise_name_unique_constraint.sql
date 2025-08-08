-- +goose Up
-- +goose StatementBegin
ALTER TABLE exercise DROP CONSTRAINT IF EXISTS exercise_name_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Note: We're not adding the constraint back in the down migration since we want to allow duplicate exercise names
-- +goose StatementEnd