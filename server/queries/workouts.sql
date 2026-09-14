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

-- name: GetSet :one
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set"
WHERE id = $1 AND user_id = $2;

-- name: ListSets :many
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set"
WHERE user_id = $1
ORDER BY exercise_order, set_order, id;

-- name: CreateWorkout :one
INSERT INTO workout (date, notes, workout_focus, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

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
