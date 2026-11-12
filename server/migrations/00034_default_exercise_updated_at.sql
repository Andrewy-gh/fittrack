-- +goose Up
-- New exercises have not been edited yet, so both timestamps start at creation.
ALTER TABLE exercise ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE exercise ALTER COLUMN updated_at DROP DEFAULT;
