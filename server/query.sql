-- Basic SELECT queries
-- name: GetWorkout :one
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: GetExercise :one
SELECT id, name FROM exercise WHERE id = $1 AND user_id = $2;

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
ORDER BY w.date DESC, s.exercise_order, s.set_order, s.created_at DESC
LIMIT 3;

-- name: ListWorkoutFocusValues :many
SELECT DISTINCT workout_focus 
FROM workout 
WHERE user_id = $1 
  AND workout_focus IS NOT NULL
ORDER BY workout_focus;
