-- Basic SELECT queries
-- name: GetWorkout :one
SELECT * FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many  
SELECT * FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: GetExercise :one
SELECT * FROM exercise WHERE id = $1 AND user_id = $2;

-- name: ListExercises :many
SELECT * FROM exercise WHERE user_id = $1 ORDER BY name;

-- name: GetSet :one
SELECT * FROM "set" 
WHERE id = $1 AND user_id = $2;

-- name: ListSets :many
SELECT * FROM "set" 
WHERE user_id = $1
ORDER BY exercise_order NULLS LAST, set_order NULLS LAST, id;

-- name: GetExerciseWithSets :many
SELECT 
    s.workout_id,
    w.date as workout_date,
    w.notes as workout_notes,
    s.id as set_id,
    s.weight,
    s.reps,
    s.set_type,
    s.exercise_id,
    e.name as exercise_name,
    (COALESCE(s.weight, 0) * s.reps) as volume
FROM "set" s
JOIN exercise e ON e.id = s.exercise_id
JOIN workout w ON w.id = s.workout_id
WHERE s.exercise_id = $1 AND s.user_id = $2
ORDER BY w.date DESC, s.set_order NULLS LAST, s.created_at, s.id;

-- INSERT queries for form submission
-- name: CreateWorkout :one
INSERT INTO workout (date, notes, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrCreateExercise :one
INSERT INTO exercise (name, user_id) 
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: CreateSet :one
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- Complex queries for joining data
-- name: GetWorkoutWithSets :many
SELECT 
    w.id as workout_id,
    w.date as workout_date,
    w.notes as workout_notes,
    s.id as set_id,
    s.weight,
    s.reps,
    s.set_type,
    e.id as exercise_id,
    e.name as exercise_name,
    (COALESCE(s.weight, 0) * s.reps) as volume
FROM workout w
JOIN "set" s ON w.id = s.workout_id
JOIN exercise e ON s.exercise_id = e.id
WHERE w.id = $1 AND w.user_id = $2
ORDER BY s.exercise_order NULLS LAST, s.set_order NULLS LAST, s.id;

-- name: GetExerciseByName :one
SELECT * FROM exercise WHERE name = $1 AND user_id = $2;

-- User queries
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUserID :one
SELECT * FROM users WHERE user_id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (user_id)
VALUES ($1)
RETURNING *;

-- UPDATE queries for PUT endpoint
-- name: UpdateWorkout :one
UPDATE workout 
SET 
    date = COALESCE($2, date),
    notes = COALESCE($3, notes),
    updated_at = NOW()
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: UpdateSet :one
UPDATE "set" 
SET 
    weight = COALESCE($2, weight),
    reps = COALESCE($3, reps),
    set_type = COALESCE($4, set_type),
    updated_at = NOW()
WHERE id = $1 AND user_id = $5
RETURNING *;

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
    w.date AS workout_date,
    s.weight,
    s.reps,
    s.created_at
FROM "set" s
JOIN workout w ON w.id = s.workout_id
WHERE s.exercise_id = $1 AND s.user_id = $2
ORDER BY s.created_at DESC
LIMIT 3;
