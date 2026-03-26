-- Basic SELECT queries
-- name: GetWorkout :one
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: GetExercise :one
SELECT id, name FROM exercise WHERE id = $1 AND user_id = $2;

-- name: GetExerciseDetail :one
SELECT
    e.id,
    e.name,
    e.created_at,
    e.updated_at,
    e.user_id,
    e.historical_1rm,
    e.historical_1rm_updated_at,
    e.historical_1rm_source_workout_id,
    (
        SELECT MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)))::numeric(8,2)
        FROM "set" s
        WHERE s.exercise_id = e.id
          AND s.user_id = e.user_id
          AND s.set_type = 'working'
    ) AS best_e1rm
FROM exercise e
WHERE e.id = $1 AND e.user_id = $2;

-- name: ListExercises :many
SELECT id, name FROM exercise WHERE user_id = $1 ORDER BY name;

-- name: GetSet :one
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set"
WHERE id = $1 AND user_id = $2;

-- name: ListSets :many
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set"
WHERE user_id = $1
ORDER BY exercise_order, set_order, id;

-- name: GetExerciseWithSets :many
SELECT 
    s.workout_id,
    w.date as workout_date,
    w.notes as workout_notes,
    w.workout_focus as workout_focus,
    s.id as set_id,
    s.weight,
    s.reps,
    s.set_type,
    e.id as exercise_id,
    e.name as exercise_name,
    s.exercise_order,
    s.set_order,
    (COALESCE(s.weight, 0) * s.reps)::NUMERIC(10,1) as volume
FROM "set" s
JOIN exercise e ON e.id = s.exercise_id
JOIN workout w ON w.id = s.workout_id
WHERE s.exercise_id = $1 AND s.user_id = $2
ORDER BY w.date DESC, s.exercise_order, s.set_order, s.created_at, s.id;

-- name: GetExerciseMetricsHistoryRawWeek :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date::date AS workout_day,
        COALESCE(s.weight, 0)::numeric AS weight,
        s.reps AS reps,
        (COALESCE(s.weight, 0)::numeric * s.reps::numeric) AS volume,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)) AS e1rm,
        e.historical_1rm AS historical_1rm,
        MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))) OVER (PARTITION BY w.id) AS session_best_e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    JOIN exercise e ON e.id = s.exercise_id
    WHERE s.exercise_id = $1
      AND s.user_id = $2
      AND s.set_type = 'working'
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM working_sets
),
filtered AS (
    SELECT *
    FROM working_sets, end_day
    WHERE workout_day >= end_day - interval '7 days'
),
workout_metrics AS (
    SELECT
        workout_id,
        MIN(workout_day)::date AS workout_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_best_intensity,
        COALESCE(SUM(volume), 0)::float8 AS total_volume_working
    FROM filtered
    GROUP BY workout_id
)
SELECT workout_id, workout_day, session_best_e1rm, session_avg_e1rm, session_avg_intensity, session_best_intensity, total_volume_working
FROM workout_metrics
ORDER BY workout_day ASC;

-- name: GetExerciseMetricsHistoryRawMonth :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date::date AS workout_day,
        COALESCE(s.weight, 0)::numeric AS weight,
        s.reps AS reps,
        (COALESCE(s.weight, 0)::numeric * s.reps::numeric) AS volume,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)) AS e1rm,
        e.historical_1rm AS historical_1rm,
        MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))) OVER (PARTITION BY w.id) AS session_best_e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    JOIN exercise e ON e.id = s.exercise_id
    WHERE s.exercise_id = $1
      AND s.user_id = $2
      AND s.set_type = 'working'
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM working_sets
),
filtered AS (
    SELECT *
    FROM working_sets, end_day
    WHERE workout_day >= end_day - interval '30 days'
),
workout_metrics AS (
    SELECT
        workout_id,
        MIN(workout_day)::date AS workout_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_best_intensity,
        COALESCE(SUM(volume), 0)::float8 AS total_volume_working
    FROM filtered
    GROUP BY workout_id
)
SELECT workout_id, workout_day, session_best_e1rm, session_avg_e1rm, session_avg_intensity, session_best_intensity, total_volume_working
FROM workout_metrics
ORDER BY workout_day ASC;

-- name: GetExerciseMetricsHistoryRaw6M :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date::date AS workout_day,
        COALESCE(s.weight, 0)::numeric AS weight,
        s.reps AS reps,
        (COALESCE(s.weight, 0)::numeric * s.reps::numeric) AS volume,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)) AS e1rm,
        e.historical_1rm AS historical_1rm,
        MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))) OVER (PARTITION BY w.id) AS session_best_e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    JOIN exercise e ON e.id = s.exercise_id
    WHERE s.exercise_id = $1
      AND s.user_id = $2
      AND s.set_type = 'working'
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM working_sets
),
filtered AS (
    SELECT *
    FROM working_sets, end_day
    WHERE workout_day >= end_day - interval '6 months'
),
workout_metrics AS (
    SELECT
        workout_id,
        MIN(workout_day)::date AS workout_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_best_intensity,
        COALESCE(SUM(volume), 0)::float8 AS total_volume_working
    FROM filtered
    GROUP BY workout_id
)
SELECT workout_id, workout_day, session_best_e1rm, session_avg_e1rm, session_avg_intensity, session_best_intensity, total_volume_working
FROM workout_metrics
ORDER BY workout_day ASC, workout_id ASC;

-- name: GetExerciseMetricsHistoryRawYear :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date::date AS workout_day,
        COALESCE(s.weight, 0)::numeric AS weight,
        s.reps AS reps,
        (COALESCE(s.weight, 0)::numeric * s.reps::numeric) AS volume,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)) AS e1rm,
        e.historical_1rm AS historical_1rm,
        MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))) OVER (PARTITION BY w.id) AS session_best_e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    JOIN exercise e ON e.id = s.exercise_id
    WHERE s.exercise_id = $1
      AND s.user_id = $2
      AND s.set_type = 'working'
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM working_sets
),
filtered AS (
    SELECT *
    FROM working_sets, end_day
    WHERE workout_day >= end_day - interval '1 year'
),
workout_metrics AS (
    SELECT
        workout_id,
        MIN(workout_day)::date AS workout_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(
            CASE
                WHEN (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) > 0
                THEN (weight / (CASE WHEN historical_1rm > 0 THEN historical_1rm ELSE session_best_e1rm END) * 100)
            END
        ), 0)::float8 AS session_best_intensity,
        COALESCE(SUM(volume), 0)::float8 AS total_volume_working
    FROM filtered
    GROUP BY workout_id
)
SELECT workout_id, workout_day, session_best_e1rm, session_avg_e1rm, session_avg_intensity, session_best_intensity, total_volume_working
FROM workout_metrics
ORDER BY workout_day ASC, workout_id ASC;

-- INSERT queries for form submission
-- name: CreateWorkout :one
INSERT INTO workout (date, notes, workout_focus, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetOrCreateExercise :one
INSERT INTO exercise (name, user_id)
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: DeleteExercise :exec
DELETE FROM exercise WHERE id = $1 AND user_id = $2;

-- name: UpdateExerciseName :exec
UPDATE exercise
SET name = $2, updated_at = NOW()
WHERE id = $1 AND user_id = $3;

-- name: UpdateExerciseHistorical1RMManual :exec
UPDATE exercise
SET
    historical_1rm = $2,
    historical_1rm_updated_at = NOW(),
    historical_1rm_source_workout_id = NULL,
    updated_at = NOW()
WHERE id = $1 AND user_id = $3;

-- name: SetExerciseHistorical1RM :exec
UPDATE exercise
SET
    historical_1rm = $2,
    historical_1rm_updated_at = NOW(),
    historical_1rm_source_workout_id = $3,
    updated_at = NOW()
WHERE id = $1 AND user_id = $4;

-- name: UpdateExerciseHistorical1RMFromWorkoutIfBetter :exec
UPDATE exercise
SET
    historical_1rm = $2,
    historical_1rm_updated_at = NOW(),
    historical_1rm_source_workout_id = $3,
    updated_at = NOW()
WHERE id = $1
  AND user_id = $4
  AND (historical_1rm IS NULL OR historical_1rm < $2);

-- name: ListExercisesWithHistorical1RMSourceWorkout :many
SELECT id
FROM exercise
WHERE user_id = $1 AND historical_1rm_source_workout_id = $2;

-- name: GetWorkoutBestE1rmByExercise :many
SELECT
    s.exercise_id,
    MAX((COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30)))::numeric(8,2) AS best_e1rm
FROM "set" s
WHERE s.workout_id = $1
  AND s.user_id = $2
  AND s.set_type = 'working'
GROUP BY s.exercise_id;

-- name: GetExerciseBestE1rmWithWorkout :one
WITH working AS (
    SELECT
        s.workout_id,
        w.date AS workout_date,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))::numeric(8,2) AS e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    WHERE s.user_id = $1
      AND s.exercise_id = $2
      AND s.set_type = 'working'
)
SELECT workout_id, e1rm
FROM working
ORDER BY e1rm DESC, workout_date DESC, workout_id DESC
LIMIT 1;

-- name: GetExerciseBestE1rmWithWorkoutExcludingWorkout :one
WITH working AS (
    SELECT
        s.workout_id,
        w.date AS workout_date,
        (COALESCE(s.weight, 0)::numeric * (1 + s.reps::numeric / 30))::numeric(8,2) AS e1rm
    FROM "set" s
    JOIN workout w ON w.id = s.workout_id
    WHERE s.user_id = $1
      AND s.exercise_id = $2
      AND s.workout_id <> $3
      AND s.set_type = 'working'
)
SELECT workout_id, e1rm
FROM working
ORDER BY e1rm DESC, workout_date DESC, workout_id DESC
LIMIT 1;

-- name: CreateSet :one
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- Complex queries for joining data
-- name: GetWorkoutWithSets :many
SELECT 
    w.id as workout_id,
    w.date as workout_date,
    w.notes as workout_notes,
    w.workout_focus as workout_focus,
    s.id as set_id,
    s.weight,
    s.reps,
    s.set_type,
    e.id as exercise_id,
    e.name as exercise_name,
    s.exercise_order,
    s.set_order,
    (COALESCE(s.weight, 0) * s.reps)::NUMERIC(10,1) as volume
FROM workout w
JOIN "set" s ON w.id = s.workout_id
JOIN exercise e ON s.exercise_id = e.id
WHERE w.id = $1 AND w.user_id = $2
ORDER BY s.exercise_order, s.set_order, s.id;

-- name: GetExerciseByName :one
SELECT id, name FROM exercise WHERE name = $1 AND user_id = $2;

-- User queries
-- name: GetUser :one
SELECT id, user_id, created_at FROM users WHERE id = $1;

-- name: GetUserByUserID :one
SELECT id, user_id, created_at FROM users WHERE user_id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (user_id)
VALUES ($1)
RETURNING id;

-- Feature access queries
-- name: ListActiveFeatureAccess :many
SELECT
    id,
    user_id,
    feature_key,
    source,
    source_reference,
    granted_by,
    note,
    starts_at,
    expires_at,
    revoked_at,
    created_at
FROM user_feature_access
WHERE user_id = $1
  AND revoked_at IS NULL
  AND starts_at <= NOW()
  AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY feature_key, starts_at DESC, id DESC;

-- name: HasActiveFeatureAccess :one
SELECT EXISTS (
    SELECT 1
    FROM user_feature_access
    WHERE user_id = $1
      AND feature_key = $2
      AND revoked_at IS NULL
      AND starts_at <= NOW()
      AND (expires_at IS NULL OR expires_at > NOW())
);

-- UPDATE queries for PUT endpoint
-- name: UpdateWorkout :one
UPDATE workout
SET
    date = COALESCE($2, date),
    notes = COALESCE($3, notes),
    workout_focus = COALESCE($4, workout_focus),
    updated_at = NOW()
WHERE id = $1 AND user_id = $5
RETURNING id;

-- name: UpdateSet :one
UPDATE "set"
SET
    weight = COALESCE($2, weight),
    reps = COALESCE($3, reps),
    set_type = COALESCE($4, set_type),
    updated_at = NOW()
WHERE id = $1 AND user_id = $5
RETURNING id;

-- name: DeleteSetsByWorkout :exec
DELETE FROM "set" 
WHERE workout_id = $1 AND user_id = $2;

-- name: DeleteSetsByWorkoutAndExercise :exec
DELETE FROM "set" 
WHERE workout_id = $1 
  AND exercise_id = $2
  AND user_id = $3;

-- name: DeleteWorkout :exec
DELETE FROM workout 
WHERE id = $1 
  AND user_id = $2;

-- name: GetRecentSetsForExercise :many
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
JOIN workout w ON w.id = s.workout_id
WHERE s.exercise_id = $1 AND s.user_id = $2
ORDER BY w.date DESC, s.set_order DESC
LIMIT 3;

-- name: ListWorkoutFocusValues :many
SELECT DISTINCT workout_focus
FROM workout
WHERE user_id = $1
  AND workout_focus IS NOT NULL
ORDER BY workout_focus;

-- name: GetContributionData :many
-- Security: This query is protected by both application-level filtering and RLS policies.
-- The WHERE clause filters by user_id (parameter $1), ensuring only the authenticated user's
-- workouts are retrieved. RLS policies on the workout table provide defense-in-depth.
-- The GROUP BY on date and JSON_AGG of workout metadata ensures no cross-user data leakage.
WITH workout_totals AS (
    SELECT
        w.id,
        w.date,
        w.workout_focus,
        COUNT(s.id) FILTER (WHERE s.set_type = 'working')::INTEGER AS working_set_count,
        COALESCE(
            SUM(
                CASE
                    WHEN s.set_type = 'working' THEN COALESCE(s.weight, 0)::NUMERIC * s.reps::NUMERIC
                    ELSE 0
                END
            ),
            0
        )::FLOAT8 AS volume
    FROM workout w
    LEFT JOIN "set" s ON s.workout_id = w.id
    WHERE w.user_id = $1
      AND w.date >= CURRENT_DATE - INTERVAL '52 weeks'
    GROUP BY w.id, w.date, w.workout_focus
)
SELECT
    DATE_TRUNC('day', wt.date)::DATE as date,
    SUM(wt.working_set_count)::INTEGER as count,
    JSON_AGG(JSONB_BUILD_OBJECT(
        'id', wt.id,
        'time', wt.date,
        'focus', wt.workout_focus,
        'volume', wt.volume
    ) ORDER BY wt.date, wt.id) as workouts
FROM workout_totals wt
GROUP BY DATE_TRUNC('day', wt.date)
ORDER BY date;

-- name: CreateAIChatConversation :one
INSERT INTO ai_chat_conversation (
    user_id,
    title
)
VALUES ($1, $2)
RETURNING id, user_id, title, created_at, updated_at, last_message_at;

-- name: GetAIChatConversation :one
SELECT id, user_id, title, created_at, updated_at, last_message_at
FROM ai_chat_conversation
WHERE id = $1 AND user_id = $2;

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

-- name: HasActiveAIChatRunForConversation :one
SELECT EXISTS (
    SELECT 1
    FROM ai_chat_run
    WHERE conversation_id = $1
      AND user_id = $2
      AND status = 'streaming'
);

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
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: TouchAIChatConversation :exec
UPDATE ai_chat_conversation
SET updated_at = CURRENT_TIMESTAMP,
    last_message_at = $3
WHERE id = $1 AND user_id = $2;

-- name: SetAIChatConversationTitleIfEmpty :execrows
UPDATE ai_chat_conversation
SET title = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND user_id = $2
  AND title IS NULL
  AND $3 IS NOT NULL
  AND btrim($3) <> '';

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

-- name: UpdateAIChatRunCompleted :one
UPDATE ai_chat_run
SET status = 'completed',
    error_message = NULL,
    completed_at = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
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
    created_at,
    updated_at,
    started_at,
    completed_at;

-- name: UpdateAIChatRunFailed :one
UPDATE ai_chat_run
SET status = 'failed',
    error_message = $3,
    completed_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
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
    created_at,
    updated_at,
    started_at,
    completed_at;
