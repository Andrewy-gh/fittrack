-- +goose Up
-- +goose StatementBegin
ALTER TABLE ai_chat_conversation
ADD COLUMN latest_workout_draft JSONB;

ALTER TABLE ai_chat_run
ADD COLUMN workout_draft JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ai_chat_run
DROP COLUMN workout_draft;

ALTER TABLE ai_chat_conversation
DROP COLUMN latest_workout_draft;
-- +goose StatementEnd
