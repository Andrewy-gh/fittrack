-- +goose Up
-- +goose StatementBegin
ALTER TABLE stripe_subscriptions
    ADD COLUMN cancel_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE stripe_subscriptions
    DROP COLUMN cancel_at;
-- +goose StatementEnd
