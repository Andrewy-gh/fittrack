-- +goose Up
-- +goose StatementBegin
ALTER TABLE ai_chat_conversation
ADD COLUMN latest_workout_draft_source_run_id INTEGER REFERENCES ai_chat_run(id) ON DELETE SET NULL,
ADD COLUMN latest_workout_draft_saved_workout_id INTEGER REFERENCES workout(id) ON DELETE SET NULL,
ADD COLUMN latest_workout_draft_saved_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ai_chat_conversation
DROP COLUMN latest_workout_draft_saved_at,
DROP COLUMN latest_workout_draft_saved_workout_id,
DROP COLUMN latest_workout_draft_source_run_id;
-- +goose StatementEnd
