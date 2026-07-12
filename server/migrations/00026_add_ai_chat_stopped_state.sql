-- +goose Up
ALTER TABLE ai_chat_message DROP CONSTRAINT ai_chat_message_status_valid;
ALTER TABLE ai_chat_message ADD CONSTRAINT ai_chat_message_status_valid
    CHECK (status IN ('streaming', 'completed', 'failed', 'stopped'));

ALTER TABLE ai_chat_run DROP CONSTRAINT ai_chat_run_status_valid;
ALTER TABLE ai_chat_run ADD CONSTRAINT ai_chat_run_status_valid
    CHECK (status IN ('streaming', 'completed', 'failed', 'stopped'));

ALTER TABLE ai_chat_run DROP CONSTRAINT ai_chat_run_generation_status_valid;
ALTER TABLE ai_chat_run ADD CONSTRAINT ai_chat_run_generation_status_valid
    CHECK (generation_status IN ('queued', 'generating', 'completed', 'failed', 'interrupted', 'stopped'));

-- +goose Down
UPDATE ai_chat_message
SET status = 'failed', error_message = 'generation stopped',
    completed_at = COALESCE(completed_at, CURRENT_TIMESTAMP)
WHERE status = 'stopped';
UPDATE ai_chat_run
SET status = 'failed', generation_status = 'failed',
    error_message = 'generation stopped',
    completed_at = COALESCE(completed_at, CURRENT_TIMESTAMP)
WHERE status = 'stopped' OR generation_status = 'stopped';
ALTER TABLE ai_chat_run DROP CONSTRAINT ai_chat_run_generation_status_valid;
ALTER TABLE ai_chat_run ADD CONSTRAINT ai_chat_run_generation_status_valid
    CHECK (generation_status IN ('queued', 'generating', 'completed', 'failed', 'interrupted'));
ALTER TABLE ai_chat_run DROP CONSTRAINT ai_chat_run_status_valid;
ALTER TABLE ai_chat_run ADD CONSTRAINT ai_chat_run_status_valid
    CHECK (status IN ('streaming', 'completed', 'failed'));
ALTER TABLE ai_chat_message DROP CONSTRAINT ai_chat_message_status_valid;
ALTER TABLE ai_chat_message ADD CONSTRAINT ai_chat_message_status_valid
    CHECK (status IN ('streaming', 'completed', 'failed'));
