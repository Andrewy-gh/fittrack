-- +goose Up
-- +goose StatementBegin
ALTER TABLE ai_chat_run
ADD COLUMN generation_status VARCHAR(32),
ADD COLUMN generation_owner VARCHAR(128),
ADD COLUMN generation_lease_expires_at TIMESTAMPTZ,
ADD COLUMN generation_heartbeat_at TIMESTAMPTZ,
ADD COLUMN generation_attempt INTEGER NOT NULL DEFAULT 0,
ADD COLUMN interrupted_at TIMESTAMPTZ,
ADD COLUMN interruption_reason VARCHAR(128);

UPDATE ai_chat_run
SET status = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN 'failed'
    ELSE status
END,
error_message = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN 'ai chat stream was interrupted before completion'
    ELSE error_message
END,
completed_at = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN updated_at
    ELSE completed_at
END,
generation_status = CASE
    WHEN status = 'completed' THEN 'completed'
    WHEN status = 'failed' THEN 'failed'
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN 'interrupted'
    WHEN status = 'streaming'
      AND error_message = 'ai chat recovery claimed and in progress'
      THEN 'generating'
    ELSE 'queued'
END,
generation_owner = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat recovery claimed and in progress'
      THEN 'legacy:recovery'
    ELSE generation_owner
END,
generation_lease_expires_at = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat recovery claimed and in progress'
      THEN updated_at + INTERVAL '60 seconds'
    ELSE generation_lease_expires_at
END,
generation_heartbeat_at = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat recovery claimed and in progress'
      THEN updated_at
    ELSE generation_heartbeat_at
END,
interrupted_at = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN updated_at
    ELSE interrupted_at
END,
interruption_reason = CASE
    WHEN status = 'streaming'
      AND error_message = 'ai chat stream interrupted and awaiting recovery handoff'
      THEN 'client_disconnect'
    ELSE interruption_reason
END;

UPDATE ai_chat_message message
SET status = 'failed',
    error_message = 'ai chat stream was interrupted before completion',
    completed_at = run.updated_at,
    updated_at = CURRENT_TIMESTAMP
FROM ai_chat_run run
WHERE message.id = run.assistant_message_id
  AND message.user_id = run.user_id
  AND run.generation_status = 'interrupted'
  AND message.status = 'streaming';

ALTER TABLE ai_chat_run
ALTER COLUMN generation_status SET NOT NULL,
ALTER COLUMN generation_status SET DEFAULT 'queued';

ALTER TABLE ai_chat_run
ADD CONSTRAINT ai_chat_run_generation_status_valid CHECK (
    generation_status IN ('queued', 'generating', 'completed', 'failed', 'interrupted')
),
ADD CONSTRAINT ai_chat_run_generation_owner_not_empty CHECK (
    generation_owner IS NULL OR btrim(generation_owner) <> ''
),
ADD CONSTRAINT ai_chat_run_generation_attempt_non_negative CHECK (
    generation_attempt >= 0
),
ADD CONSTRAINT ai_chat_run_interruption_reason_not_empty CHECK (
    interruption_reason IS NULL OR btrim(interruption_reason) <> ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ai_chat_run
DROP CONSTRAINT ai_chat_run_interruption_reason_not_empty,
DROP CONSTRAINT ai_chat_run_generation_attempt_non_negative,
DROP CONSTRAINT ai_chat_run_generation_owner_not_empty,
DROP CONSTRAINT ai_chat_run_generation_status_valid;

ALTER TABLE ai_chat_run
DROP COLUMN interruption_reason,
DROP COLUMN interrupted_at,
DROP COLUMN generation_attempt,
DROP COLUMN generation_heartbeat_at,
DROP COLUMN generation_lease_expires_at,
DROP COLUMN generation_owner,
DROP COLUMN generation_status;
-- +goose StatementEnd
