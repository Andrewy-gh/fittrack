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

-- name: GetExerciseByName :one
SELECT id, name FROM exercise WHERE name = $1 AND user_id = $2;

-- User queries

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
