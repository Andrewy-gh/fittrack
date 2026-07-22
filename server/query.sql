-- Basic SELECT queries
-- name: GetWorkout :one
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: ListWorkoutFocusTemplates :many
WITH ranked_focus_workouts AS (
    SELECT
        id AS workout_id,
        date,
        BTRIM(workout_focus)::VARCHAR(256) AS workout_focus,
        ROW_NUMBER() OVER (
            PARTITION BY LOWER(BTRIM(workout_focus))
            ORDER BY date DESC, id DESC
        ) AS rank
    FROM workout
    WHERE user_id = $1
      AND workout_focus IS NOT NULL
      AND BTRIM(workout_focus) <> ''
)
SELECT workout_id, date, workout_focus
FROM ranked_focus_workouts
WHERE rank = 1
ORDER BY date DESC, workout_id DESC;

-- name: GetLatestWorkoutNote :one
SELECT id AS workout_id, date, notes
FROM workout
WHERE user_id = $1
  AND notes IS NOT NULL
  AND BTRIM(notes) <> ''
ORDER BY date DESC, id DESC
LIMIT 1;

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

-- name: GetExerciseMetricsHistoryRawAll :many
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
    FROM working_sets
    GROUP BY workout_id
)
SELECT workout_id, workout_day, session_best_e1rm, session_avg_e1rm, session_avg_intensity, session_best_intensity, total_volume_working
FROM workout_metrics
ORDER BY workout_day ASC, workout_id ASC;

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

-- name: GetUserTrainingProfile :one
SELECT
    user_id,
    primary_goal,
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
    NULLIF(sqlc.narg(experience_level)::text, ''),
    sqlc.narg(preferred_session_duration_minutes)::integer,
    NULLIF(sqlc.narg(usual_training_location)::text, ''),
    sqlc.arg(available_equipment)::jsonb,
    sqlc.arg(avoided_exercises)::jsonb,
    sqlc.narg(movement_limitations)::jsonb,
    -- Manual settings saves supersede AI-written profile provenance.
    NULL,
    NULL
)
ON CONFLICT (user_id) DO UPDATE SET
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

-- name: DeleteUser :execrows
DELETE FROM users
WHERE user_id = $1;

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

-- Billing queries
-- name: GetStripeCustomerByUserID :one
SELECT user_id, stripe_customer_id, created_at, updated_at
FROM stripe_customers
WHERE user_id = $1;

-- name: GetStripeCustomerByCustomerID :one
SELECT user_id, stripe_customer_id, created_at, updated_at
FROM stripe_customers
WHERE stripe_customer_id = $1;

-- name: GetBillingUserForUpdate :one
SELECT id, user_id, created_at
FROM users
WHERE user_id = $1
FOR UPDATE;

-- name: UpsertStripeCustomer :one
INSERT INTO stripe_customers (
    user_id,
    stripe_customer_id
)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET stripe_customer_id = EXCLUDED.stripe_customer_id,
    updated_at = CURRENT_TIMESTAMP
RETURNING user_id, stripe_customer_id, created_at, updated_at;

-- name: UpsertStripeSubscription :one
INSERT INTO stripe_subscriptions (
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end
)
VALUES (
    sqlc.arg(stripe_subscription_id),
    sqlc.arg(user_id),
    sqlc.arg(stripe_customer_id),
    NULLIF(sqlc.arg(stripe_price_id)::text, ''),
    sqlc.arg(stripe_event_created_at),
    sqlc.arg(status),
    sqlc.arg(cancel_at_period_end),
    sqlc.arg(cancel_at),
    sqlc.arg(current_period_start),
    sqlc.arg(current_period_end),
    sqlc.arg(trial_start),
    sqlc.arg(trial_end)
)
ON CONFLICT (stripe_subscription_id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    stripe_customer_id = EXCLUDED.stripe_customer_id,
    stripe_price_id = EXCLUDED.stripe_price_id,
    stripe_event_created_at = EXCLUDED.stripe_event_created_at,
    status = EXCLUDED.status,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    cancel_at = EXCLUDED.cancel_at,
    current_period_start = EXCLUDED.current_period_start,
    current_period_end = EXCLUDED.current_period_end,
    trial_start = EXCLUDED.trial_start,
    trial_end = EXCLUDED.trial_end,
    updated_at = CURRENT_TIMESTAMP
WHERE stripe_subscriptions.stripe_event_created_at < EXCLUDED.stripe_event_created_at
   OR (
       stripe_subscriptions.stripe_event_created_at = EXCLUDED.stripe_event_created_at
       AND NOT (
           stripe_subscriptions.status IN ('past_due', 'unpaid', 'canceled', 'incomplete', 'incomplete_expired', 'paused')
           AND EXCLUDED.status IN ('trialing', 'active')
       )
   )
RETURNING
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end,
    created_at,
    updated_at;

-- name: GetCurrentStripeSubscriptionByUserID :one
SELECT
    stripe_subscription_id,
    user_id,
    stripe_customer_id,
    stripe_price_id,
    stripe_event_created_at,
    status,
    cancel_at_period_end,
    cancel_at,
    current_period_start,
    current_period_end,
    trial_start,
    trial_end,
    created_at,
    updated_at
FROM stripe_subscriptions
WHERE user_id = $1
ORDER BY
    stripe_event_created_at DESC,
    updated_at DESC,
    created_at DESC
LIMIT 1;

-- name: HasProcessedStripeWebhookEvent :one
SELECT EXISTS (
    SELECT 1
    FROM stripe_webhook_events
    WHERE stripe_event_id = $1
);

-- name: MarkStripeWebhookEventProcessed :exec
INSERT INTO stripe_webhook_events (
    stripe_event_id,
    event_type
)
VALUES ($1, $2)
ON CONFLICT (stripe_event_id) DO NOTHING;

-- name: RevokeStripeFeatureAccess :exec
UPDATE user_feature_access
SET revoked_at = GREATEST(CURRENT_TIMESTAMP, starts_at)
WHERE user_id = $1
  AND feature_key = $2
  AND source = 'stripe'
  AND source_reference = $3
  AND revoked_at IS NULL;

-- name: GrantStripeFeatureAccess :exec
INSERT INTO user_feature_access (
    user_id,
    feature_key,
    source,
    source_reference,
    granted_by,
    note,
    starts_at,
    expires_at
)
VALUES ($1, $2, 'stripe', $3, 'stripe_webhook', $4, CURRENT_TIMESTAMP, $5);

-- name: GetAIChatTrialPromptUsage :one
SELECT user_id, stripe_subscription_id, prompt_count, created_at, updated_at
FROM ai_chat_trial_prompt_usage
WHERE user_id = $1
  AND stripe_subscription_id = $2;

-- name: ConsumeAIChatTrialPrompt :one
INSERT INTO ai_chat_trial_prompt_usage (
    user_id,
    stripe_subscription_id,
    prompt_count
)
VALUES ($1, $2, 1)
ON CONFLICT (user_id, stripe_subscription_id) DO UPDATE
SET prompt_count = ai_chat_trial_prompt_usage.prompt_count + 1,
    updated_at = CURRENT_TIMESTAMP
WHERE ai_chat_trial_prompt_usage.prompt_count < $3
RETURNING user_id, stripe_subscription_id, prompt_count, created_at, updated_at;

-- name: ConsumeAIChatTrialPromptForCurrentSubscription :one
WITH current_subscription AS (
    SELECT stripe_subscription_id, status
    FROM stripe_subscriptions
    WHERE stripe_subscriptions.user_id = $1
    ORDER BY
        stripe_event_created_at DESC,
        updated_at DESC,
        created_at DESC
    LIMIT 1
),
trial_subscription AS (
    SELECT stripe_subscription_id
    FROM current_subscription
    WHERE status = 'trialing'
),
consumed AS (
    INSERT INTO ai_chat_trial_prompt_usage (
        user_id,
        stripe_subscription_id,
        prompt_count
    )
    SELECT $1, stripe_subscription_id, 1
    FROM trial_subscription
    WHERE sqlc.arg(prompt_cap)::integer > 0
    ON CONFLICT (user_id, stripe_subscription_id) DO UPDATE
    SET prompt_count = ai_chat_trial_prompt_usage.prompt_count + 1,
        updated_at = CURRENT_TIMESTAMP
    WHERE ai_chat_trial_prompt_usage.prompt_count < sqlc.arg(prompt_cap)::integer
    RETURNING 1
)
SELECT
    NOT EXISTS (SELECT 1 FROM trial_subscription)
    OR EXISTS (SELECT 1 FROM consumed) AS allowed;

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
