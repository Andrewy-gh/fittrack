-- +goose Up
-- +goose StatementBegin
ALTER TABLE ai_chat_run
ADD COLUMN workout_draft JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ai_chat_run
DROP COLUMN workout_draft;
-- +goose StatementEnd
