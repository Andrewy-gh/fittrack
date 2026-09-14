-- name: GetUserTrainingProfile :one
SELECT
    user_id,
    primary_goal,
    goals,
    experience_level,
    preferred_session_duration_minutes,
    usual_training_location,
    available_equipment,
    avoided_exercises,
    movement_limitations,
    source_conversation_id,
    source_message_id,
    created_at,
    updated_at
FROM user_training_profile
WHERE user_id = $1;

-- name: UpsertUserTrainingProfileForChat :one
INSERT INTO user_training_profile (
    user_id,
    primary_goal,
    goals,
    experience_level,
    preferred_session_duration_minutes,
    usual_training_location,
    available_equipment,
    avoided_exercises,
    movement_limitations,
    source_conversation_id,
    source_message_id
)
VALUES (
    sqlc.arg(user_id),
    NULLIF(sqlc.narg(primary_goal)::text, ''),
    COALESCE(sqlc.narg(goals)::text[], CASE WHEN NULLIF(sqlc.narg(primary_goal)::text, '') IS NULL THEN '{}'::text[] ELSE ARRAY[sqlc.narg(primary_goal)::text] END),
    NULLIF(sqlc.narg(experience_level)::text, ''),
    CASE
        WHEN sqlc.narg(preferred_session_duration_minutes)::integer IS NULL THEN NULL
        WHEN sqlc.narg(preferred_session_duration_minutes)::integer <= 0 THEN NULL
        ELSE sqlc.narg(preferred_session_duration_minutes)::integer
    END,
    NULLIF(sqlc.narg(usual_training_location)::text, ''),
    COALESCE(sqlc.narg(available_equipment)::jsonb, '[]'::jsonb),
    COALESCE(sqlc.narg(avoided_exercises)::jsonb, '[]'::jsonb),
    sqlc.narg(movement_limitations)::jsonb,
    sqlc.narg(source_conversation_id),
    sqlc.narg(source_message_id)
)
ON CONFLICT (user_id) DO UPDATE SET
    -- Omitted or unchanged legacy goals preserve a newer multi-goal selection.
    goals = CASE
        WHEN sqlc.narg(goals)::text[] IS NOT NULL THEN sqlc.narg(goals)::text[]
        WHEN sqlc.narg(primary_goal)::text IS NULL THEN user_training_profile.goals
        WHEN sqlc.narg(primary_goal)::text = user_training_profile.primary_goal THEN user_training_profile.goals
        WHEN sqlc.narg(primary_goal)::text = '' THEN '{}'::text[]
        ELSE ARRAY[sqlc.narg(primary_goal)::text]
    END,
    primary_goal = CASE
        WHEN sqlc.narg(primary_goal)::text IS NULL THEN user_training_profile.primary_goal
        ELSE NULLIF(sqlc.narg(primary_goal)::text, '')
    END,
    experience_level = CASE
        WHEN sqlc.narg(experience_level)::text IS NULL THEN user_training_profile.experience_level
        ELSE NULLIF(sqlc.narg(experience_level)::text, '')
    END,
    preferred_session_duration_minutes = CASE
        WHEN sqlc.narg(preferred_session_duration_minutes)::integer IS NULL THEN user_training_profile.preferred_session_duration_minutes
        WHEN sqlc.narg(preferred_session_duration_minutes)::integer <= 0 THEN NULL
        ELSE sqlc.narg(preferred_session_duration_minutes)::integer
    END,
    usual_training_location = CASE
        WHEN sqlc.narg(usual_training_location)::text IS NULL THEN user_training_profile.usual_training_location
        ELSE NULLIF(sqlc.narg(usual_training_location)::text, '')
    END,
    available_equipment = COALESCE(sqlc.narg(available_equipment)::jsonb, user_training_profile.available_equipment),
    avoided_exercises = COALESCE(sqlc.narg(avoided_exercises)::jsonb, user_training_profile.avoided_exercises),
    movement_limitations = COALESCE(sqlc.narg(movement_limitations)::jsonb, user_training_profile.movement_limitations),
    source_conversation_id = COALESCE(sqlc.narg(source_conversation_id), user_training_profile.source_conversation_id),
    source_message_id = COALESCE(sqlc.narg(source_message_id), user_training_profile.source_message_id),
    updated_at = CURRENT_TIMESTAMP
RETURNING
    user_id,
    primary_goal,
    goals,
    experience_level,
    preferred_session_duration_minutes,
    usual_training_location,
    available_equipment,
    avoided_exercises,
    movement_limitations,
    source_conversation_id,
    source_message_id,
    created_at,
    updated_at;

-- name: UpsertUserTrainingProfileForSettings :one
INSERT INTO user_training_profile (
    user_id,
    primary_goal,
    goals,
    experience_level,
    preferred_session_duration_minutes,
    usual_training_location,
    available_equipment,
    avoided_exercises,
    movement_limitations,
    source_conversation_id,
    source_message_id
)
VALUES (
    sqlc.arg(user_id),
    NULLIF(sqlc.narg(primary_goal)::text, ''),
    COALESCE(sqlc.narg(goals)::text[], CASE WHEN NULLIF(sqlc.narg(primary_goal)::text, '') IS NULL THEN '{}'::text[] ELSE ARRAY[sqlc.narg(primary_goal)::text] END),
    NULLIF(sqlc.narg(experience_level)::text, ''),
    sqlc.narg(preferred_session_duration_minutes)::integer,
    NULLIF(sqlc.narg(usual_training_location)::text, ''),
    sqlc.arg(available_equipment)::text::jsonb,
    sqlc.arg(avoided_exercises)::text::jsonb,
    sqlc.narg(movement_limitations)::text::jsonb,
    -- Manual settings saves supersede AI-written profile provenance.
    NULL,
    NULL
)
ON CONFLICT (user_id) DO UPDATE SET
    goals = CASE
        WHEN sqlc.narg(goals)::text[] IS NOT NULL THEN sqlc.narg(goals)::text[]
        WHEN sqlc.narg(primary_goal)::text = user_training_profile.primary_goal THEN user_training_profile.goals
        WHEN NULLIF(sqlc.narg(primary_goal)::text, '') IS NULL THEN '{}'::text[]
        ELSE ARRAY[sqlc.narg(primary_goal)::text]
    END,
    primary_goal = NULLIF(sqlc.narg(primary_goal)::text, ''),
    experience_level = NULLIF(sqlc.narg(experience_level)::text, ''),
    preferred_session_duration_minutes = sqlc.narg(preferred_session_duration_minutes)::integer,
    usual_training_location = NULLIF(sqlc.narg(usual_training_location)::text, ''),
    available_equipment = sqlc.arg(available_equipment)::jsonb,
    avoided_exercises = sqlc.arg(avoided_exercises)::jsonb,
    movement_limitations = sqlc.narg(movement_limitations)::jsonb,
    -- These columns only reference the chat message that last wrote the profile via AI.
    source_conversation_id = NULL,
    source_message_id = NULL,
    updated_at = CURRENT_TIMESTAMP
RETURNING
    user_id,
    primary_goal,
    goals,
    experience_level,
    preferred_session_duration_minutes,
    usual_training_location,
    available_equipment,
    avoided_exercises,
    movement_limitations,
    source_conversation_id,
    source_message_id,
    created_at,
    updated_at;

-- INSERT queries for form submission
