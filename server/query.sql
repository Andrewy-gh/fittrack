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
SELECT s.* FROM "set" s
JOIN workout w ON w.id = s.workout_id
WHERE s.id = $1 AND w.user_id = $2;

-- name: ListSets :many
SELECT s.* FROM "set" s
JOIN workout w ON w.id = s.workout_id
WHERE w.user_id = $1
ORDER BY s.id;

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
WHERE s.exercise_id = $1 AND e.user_id = $2
ORDER BY w.date DESC, s.created_at;

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
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type)
SELECT $1, $2, $3, $4, $5
FROM workout w
JOIN exercise e ON e.id = $1
WHERE w.id = $2 
  AND w.user_id = $6
  AND e.user_id = $6
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
ORDER BY e.name, s.id;

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