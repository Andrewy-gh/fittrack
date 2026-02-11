-- Basic SELECT queries
-- name: GetWorkout :one
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: GetExercise :one
SELECT id, name FROM exercise WHERE id = $1 AND user_id = $2;

-- name: GetExerciseDetail :one
SELECT
    id,
    name,
    created_at,
    updated_at,
    user_id,
    historical_1rm,
    historical_1rm_updated_at,
    historical_1rm_source_workout_id
FROM exercise
WHERE id = $1 AND user_id = $2;

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

-- name: GetExerciseMetricsHistoryWeekly6M :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date AS workout_date,
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
        MIN(workout_day) AS workout_day,
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
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM workout_metrics
),
filtered_workouts AS (
    SELECT *
    FROM workout_metrics, end_day
    WHERE workout_day >= end_day - interval '6 months'
),
bucketed AS (
    SELECT
        date_trunc('week', workout_day)::date AS bucket_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(session_avg_e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(session_avg_intensity), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(session_best_intensity), 0)::float8 AS session_best_intensity,
        COALESCE(AVG(total_volume_working), 0)::float8 AS total_volume_working
    FROM filtered_workouts
    GROUP BY bucket_day
)
SELECT *
FROM (
    SELECT *
    FROM bucketed
    ORDER BY bucket_day DESC
    LIMIT 26
) t
ORDER BY bucket_day ASC;

-- name: GetExerciseMetricsHistoryMonthlyYear :many
WITH working_sets AS (
    SELECT
        w.id AS workout_id,
        w.date AS workout_date,
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
        MIN(workout_day) AS workout_day,
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
),
end_day AS (
    SELECT MAX(workout_day) AS end_day
    FROM workout_metrics
),
filtered_workouts AS (
    SELECT *
    FROM workout_metrics, end_day
    WHERE workout_day >= end_day - interval '1 year'
),
bucketed AS (
    SELECT
        date_trunc('month', workout_day)::date AS bucket_day,
        COALESCE(MAX(session_best_e1rm), 0)::float8 AS session_best_e1rm,
        COALESCE(AVG(session_avg_e1rm), 0)::float8 AS session_avg_e1rm,
        COALESCE(AVG(session_avg_intensity), 0)::float8 AS session_avg_intensity,
        COALESCE(MAX(session_best_intensity), 0)::float8 AS session_best_intensity,
        COALESCE(AVG(total_volume_working), 0)::float8 AS total_volume_working
    FROM filtered_workouts
    GROUP BY bucket_day
)
SELECT *
FROM (
    SELECT *
    FROM bucketed
    ORDER BY bucket_day DESC
    LIMIT 12
) t
ORDER BY bucket_day ASC;

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
SELECT
    DATE_TRUNC('day', w.date)::DATE as date,
    COUNT(s.id)::INTEGER as count,
    JSON_AGG(DISTINCT JSONB_BUILD_OBJECT(
        'id', w.id,
        'time', w.date,
        'focus', w.workout_focus
    )) as workouts
FROM workout w
LEFT JOIN "set" s ON s.workout_id = w.id AND s.set_type = 'working'
WHERE w.user_id = $1
  AND w.date >= CURRENT_DATE - INTERVAL '52 weeks'
GROUP BY DATE_TRUNC('day', w.date)
ORDER BY date;
