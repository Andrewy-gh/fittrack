-- name: ListWorkoutsWithSetsForChat :many
WITH matching_workouts AS (
    SELECT w.id
    FROM workout w
    WHERE w.user_id = sqlc.arg(user_id)
      AND (sqlc.narg(start_date)::timestamptz IS NULL OR w.date >= sqlc.narg(start_date)::timestamptz)
      AND (sqlc.narg(end_date)::timestamptz IS NULL OR w.date <= sqlc.narg(end_date)::timestamptz)
      AND (
          NULLIF(sqlc.narg(exercise_name)::text, '') IS NULL
          OR EXISTS (
              SELECT 1
              FROM "set" filter_set
              JOIN exercise filter_exercise ON filter_exercise.id = filter_set.exercise_id
              WHERE filter_set.workout_id = w.id
                AND filter_set.user_id = w.user_id
                AND filter_exercise.user_id = w.user_id
                AND filter_exercise.name = sqlc.narg(exercise_name)::text
          )
      )
      AND (
          NULLIF(sqlc.narg(workout_focus)::text, '') IS NULL
          OR w.workout_focus ILIKE '%' || sqlc.narg(workout_focus)::text || '%'
      )
    ORDER BY w.date DESC, w.id DESC
    LIMIT sqlc.arg(row_limit)
)
SELECT
    w.id AS workout_id,
    w.date,
    w.notes,
    w.workout_focus,
    e.name AS exercise_name,
    s.exercise_order,
    s.set_order,
    s.weight,
    s.reps,
    s.set_type
FROM matching_workouts mw
JOIN workout w ON w.id = mw.id
LEFT JOIN "set" s ON s.workout_id = w.id
    AND s.user_id = w.user_id
    AND (
        NULLIF(sqlc.narg(exercise_name)::text, '') IS NULL
        OR EXISTS (
            SELECT 1
            FROM exercise selected_exercise
            WHERE selected_exercise.id = s.exercise_id
              AND selected_exercise.user_id = s.user_id
              AND selected_exercise.name = sqlc.narg(exercise_name)::text
        )
    )
LEFT JOIN exercise e ON e.id = s.exercise_id AND e.user_id = w.user_id
ORDER BY w.date DESC, w.id DESC, s.exercise_order, s.set_order, s.id;

-- name: ListExerciseNameMatches :many
SELECT id, name
FROM exercise
WHERE user_id = $1
  AND name ILIKE '%' || sqlc.arg(name_query)::text || '%'
ORDER BY name
LIMIT 8;

-- name: GetChatWorkoutSnapshotStats :one
SELECT
    MAX(date)::timestamptz AS last_workout_date,
    COUNT(*) FILTER (WHERE date >= now() - interval '30 days') AS workouts_last_30d
FROM workout
WHERE user_id = $1;

-- name: ListTopExercisesByFrequency :many
SELECT
    e.name,
    COUNT(DISTINCT s.workout_id)::integer AS workout_count
FROM "set" s
JOIN exercise e ON e.id = s.exercise_id AND e.user_id = s.user_id
JOIN workout w ON w.id = s.workout_id AND w.user_id = s.user_id
WHERE s.user_id = $1
  AND w.date >= now() - interval '90 days'
GROUP BY e.name
ORDER BY workout_count DESC, e.name
LIMIT 5;

-- name: GetLastSessionSetsForExerciseChat :many
WITH latest_workout AS (
    SELECT s.workout_id
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id AND w.user_id = s.user_id
    WHERE s.exercise_id = $1
      AND s.user_id = $2
      AND s.set_type = 'working'
    ORDER BY w.date DESC, w.id DESC
    LIMIT 1
)
SELECT
    s.id AS set_id,
    w.id AS workout_id,
    w.date AS workout_date,
    w.workout_focus AS workout_focus,
    s.weight,
    s.reps,
    s.exercise_order,
    s.set_order,
    s.created_at
FROM "set" s
JOIN latest_workout lw ON lw.workout_id = s.workout_id
JOIN workout w ON w.id = s.workout_id AND w.user_id = s.user_id
WHERE s.exercise_id = $1
  AND s.user_id = $2
  AND s.set_type = 'working'
ORDER BY s.set_order ASC;

-- name: LockAIChatUserMutation :exec
-- Serializes conversation creation, stream start, and deletion for one owner.
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(user_id)::text, 250));

-- name: CreateAIChatConversation :one
INSERT INTO ai_chat_conversation (
    user_id,
    title
)
VALUES ($1, $2)
RETURNING
    id,
    user_id,
    title,
    latest_workout_draft,
    latest_workout_draft_source_run_id,
    latest_workout_draft_saved_workout_id,
    latest_workout_draft_saved_at,
    created_at,
    updated_at,
    last_message_at;

-- name: GetAIChatConversation :one
SELECT
    id,
    user_id,
    title,
    latest_workout_draft,
    latest_workout_draft_source_run_id,
    latest_workout_draft_saved_workout_id,
    latest_workout_draft_saved_at,
    created_at,
    updated_at,
    last_message_at
FROM ai_chat_conversation
WHERE id = $1 AND user_id = $2;

-- name: GetAIChatConversationForUpdate :one
SELECT
    id,
    user_id,
    title,
    latest_workout_draft,
    latest_workout_draft_source_run_id,
    latest_workout_draft_saved_workout_id,
    latest_workout_draft_saved_at,
    created_at,
    updated_at,
    last_message_at
FROM ai_chat_conversation
WHERE id = $1 AND user_id = $2
FOR UPDATE;

-- name: DeleteAIChatConversation :execrows
DELETE FROM ai_chat_conversation
WHERE id = $1 AND user_id = $2;

-- name: LockAIChatConversationsByUser :many
SELECT id
FROM ai_chat_conversation
WHERE user_id = $1
ORDER BY id
FOR UPDATE;

-- name: LockAIChatRunsByUser :many
SELECT id, conversation_id, assistant_message_id, status
FROM ai_chat_run
WHERE user_id = $1
ORDER BY conversation_id, id
FOR UPDATE;

-- name: ClearUserTrainingProfileSourcesByUser :exec
UPDATE user_training_profile
SET
    source_conversation_id = NULL,
    source_message_id = NULL
WHERE user_id = $1
  AND (source_conversation_id IS NOT NULL OR source_message_id IS NOT NULL);

-- name: DeleteAIChatConversationsByUser :execrows
DELETE FROM ai_chat_conversation
WHERE user_id = $1;

-- name: ClearUserTrainingProfileConversationSource :exec
UPDATE user_training_profile
SET
    source_conversation_id = NULL,
    source_message_id = NULL
WHERE user_id = $1
  AND source_conversation_id = $2;

-- name: ListAIChatConversationsByUser :many
SELECT
    id,
    user_id,
    title,
    created_at,
    updated_at,
    last_message_at
FROM ai_chat_conversation
WHERE user_id = $1
ORDER BY COALESCE(last_message_at, updated_at) DESC, updated_at DESC, id DESC
LIMIT $2;

-- name: ListAIChatMessagesByConversation :many
SELECT
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at
FROM ai_chat_message
WHERE conversation_id = $1 AND user_id = $2
ORDER BY id ASC;

-- name: GetAIChatMessage :one
SELECT
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at
FROM ai_chat_message
WHERE id = $1 AND user_id = $2;

-- name: GetActiveAIChatRunForConversation :one
SELECT
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at
FROM ai_chat_run
WHERE conversation_id = $1
  AND user_id = $2
  AND status = 'streaming'
ORDER BY id DESC
LIMIT 1;

-- name: GetAIChatRun :one
SELECT
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at
FROM ai_chat_run
WHERE id = $1
  AND user_id = $2;

-- name: GetAIChatRunForUpdate :one
SELECT id, conversation_id, user_id, user_message_id, assistant_message_id,
       model, status, request_id, error_message, workout_draft,
       generation_status, generation_owner, generation_lease_expires_at,
       generation_heartbeat_at, generation_attempt, interrupted_at,
       interruption_reason, created_at, updated_at, started_at, completed_at
FROM ai_chat_run
WHERE id = $1 AND conversation_id = $2 AND user_id = $3
FOR UPDATE;

-- name: GetLatestAIChatStreamChunkSequence :one
SELECT COALESCE(MAX(sequence), 0)::INTEGER
FROM ai_chat_stream_chunk
WHERE run_id = $1
  AND user_id = $2;

-- name: ListAIChatStreamChunksAfter :many
SELECT
    run_id,
    user_id,
    sequence,
    delta_text,
    created_at
FROM ai_chat_stream_chunk
WHERE run_id = $1
  AND user_id = $2
  AND sequence > $3
ORDER BY sequence ASC;

-- name: CreateAIChatStreamChunk :one
INSERT INTO ai_chat_stream_chunk (
    run_id,
    user_id,
    sequence,
    delta_text
)
SELECT
    sqlc.arg(run_id),
    sqlc.arg(user_id)::varchar,
    sqlc.arg(sequence),
    sqlc.arg(delta_text)
WHERE (
    NULLIF(sqlc.arg(generation_owner)::text, '') IS NULL
    OR EXISTS (
        SELECT 1
        FROM ai_chat_run
        WHERE id = sqlc.arg(run_id)
          AND user_id = sqlc.arg(user_id)::varchar
          AND status = 'streaming'
          AND generation_status = 'generating'
          AND generation_owner = sqlc.arg(generation_owner)::varchar
    )
)
RETURNING
    run_id,
    user_id,
    sequence,
    delta_text,
    created_at;

-- name: CreateAIChatMessage :one
INSERT INTO ai_chat_message (
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    completed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at;

-- name: CreateAIChatRun :one
INSERT INTO ai_chat_run (
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    completed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: TouchAIChatConversation :exec
UPDATE ai_chat_conversation
SET updated_at = CURRENT_TIMESTAMP,
    last_message_at = $3
WHERE id = $1 AND user_id = $2;

-- name: SetAIChatConversationLatestWorkoutDraft :exec
UPDATE ai_chat_conversation
SET latest_workout_draft = NULLIF($3::text, '')::jsonb,
    latest_workout_draft_source_run_id = $4,
    latest_workout_draft_saved_workout_id = NULL,
    latest_workout_draft_saved_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: MarkAIChatConversationLatestWorkoutDraftSaved :one
UPDATE ai_chat_conversation
SET latest_workout_draft_saved_workout_id = $4,
    latest_workout_draft_saved_at = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND latest_workout_draft IS NOT NULL
  AND (
    latest_workout_draft_source_run_id = $3
    OR ($3::integer IS NULL AND latest_workout_draft_source_run_id IS NULL)
  )
RETURNING
    id,
    user_id,
    title,
    latest_workout_draft,
    latest_workout_draft_source_run_id,
    latest_workout_draft_saved_workout_id,
    latest_workout_draft_saved_at,
    created_at,
    updated_at,
    last_message_at;

-- name: SetAIChatConversationTitleIfEmpty :execrows
UPDATE ai_chat_conversation
SET title = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND title IS NULL
  AND $3::text IS NOT NULL
  AND btrim($3::text) <> '';

-- name: UpdateAIChatMessageCompleted :one
UPDATE ai_chat_message
SET content = $3,
    status = 'completed',
    error_message = NULL,
    completed_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at;

-- name: UpdateAIChatMessageStreaming :one
UPDATE ai_chat_message
SET content = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
RETURNING
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at;

-- name: UpdateAIChatMessageFailed :one
UPDATE ai_chat_message
SET content = $3,
    status = 'failed',
    error_message = $4,
    completed_at = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING
    id,
    conversation_id,
    user_id,
    role,
    content,
    status,
    error_message,
    created_at,
    updated_at,
    completed_at;

-- name: UpdateAIChatMessageStopped :one
UPDATE ai_chat_message
SET status = 'stopped', error_message = NULL, completed_at = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND status = 'streaming'
RETURNING id, conversation_id, user_id, role, content, status, error_message,
          created_at, updated_at, completed_at;

-- name: UpdateAIChatRunStopped :one
UPDATE ai_chat_run
SET status = 'stopped', error_message = NULL, completed_at = $3,
    workout_draft = NULL, generation_status = 'stopped',
    generation_owner = NULL, generation_lease_expires_at = NULL,
    generation_heartbeat_at = NULL, interrupted_at = NULL,
    interruption_reason = NULL, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND status = 'streaming'
RETURNING id, conversation_id, user_id, user_message_id, assistant_message_id,
          model, status, request_id, error_message, workout_draft,
          generation_status, generation_owner, generation_lease_expires_at,
          generation_heartbeat_at, generation_attempt, interrupted_at,
          interruption_reason, created_at, updated_at, started_at, completed_at;

-- name: UpdateAIChatRunCompleted :one
UPDATE ai_chat_run
SET status = 'completed',
    error_message = NULL,
    completed_at = $3,
    workout_draft = NULLIF($4::text, '')::jsonb,
    generation_status = 'completed',
    generation_owner = NULL,
    generation_lease_expires_at = NULL,
    generation_heartbeat_at = NULL,
    interrupted_at = NULL,
    interruption_reason = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
  AND (NULLIF($5::text, '') IS NULL OR generation_owner = $5)
RETURNING
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: TouchAIChatRun :exec
UPDATE ai_chat_run
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming';

-- name: ClaimAIChatRunGeneration :one
UPDATE ai_chat_run
SET generation_status = 'generating',
    generation_owner = $3,
    generation_lease_expires_at = $4,
    generation_heartbeat_at = $5,
    generation_attempt = generation_attempt + 1,
    error_message = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
  AND generation_attempt < $6
  AND (
    generation_status = 'queued'
    OR (
      generation_status = 'generating'
      AND generation_lease_expires_at IS NOT NULL
      AND generation_lease_expires_at < $5
    )
  )
RETURNING
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: HeartbeatAIChatRunGeneration :execrows
UPDATE ai_chat_run
SET generation_lease_expires_at = $3,
    generation_heartbeat_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
  AND generation_status = 'generating'
  AND generation_owner = $5;

-- name: OwnsAIChatRunGeneration :one
SELECT EXISTS (
    SELECT 1
    FROM ai_chat_run
    WHERE id = $1
      AND user_id = $2
      AND status = 'streaming'
      AND generation_status = 'generating'
      AND generation_owner = $3
);

-- name: UpdateAIChatRunFailed :one
UPDATE ai_chat_run
SET status = 'failed',
    error_message = $3,
    completed_at = $4,
    workout_draft = NULL,
    generation_status = 'failed',
    generation_owner = NULL,
    generation_lease_expires_at = NULL,
    generation_heartbeat_at = NULL,
    interrupted_at = NULL,
    interruption_reason = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
  AND (NULLIF($5::text, '') IS NULL OR generation_owner = $5)
RETURNING
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: UpdateAIChatRunInterrupted :one
UPDATE ai_chat_run
SET status = 'failed',
    error_message = $3,
    completed_at = $4,
    workout_draft = NULL,
    generation_status = 'interrupted',
    generation_owner = NULL,
    generation_lease_expires_at = NULL,
    generation_heartbeat_at = NULL,
    interrupted_at = $4,
    interruption_reason = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND status = 'streaming'
  AND generation_status = sqlc.arg(expected_generation_status)
  AND generation_owner IS NOT DISTINCT FROM sqlc.narg(expected_generation_owner)
  AND generation_lease_expires_at IS NOT DISTINCT FROM sqlc.narg(expected_generation_lease_expires_at)
  AND generation_attempt = sqlc.arg(expected_generation_attempt)
RETURNING
    id,
    conversation_id,
    user_id,
    user_message_id,
    assistant_message_id,
    model,
    status,
    request_id,
    error_message,
    workout_draft,
    generation_status,
    generation_owner,
    generation_lease_expires_at,
    generation_heartbeat_at,
    generation_attempt,
    interrupted_at,
    interruption_reason,
    created_at,
    updated_at,
    started_at,
    completed_at;
