-- Basic SELECT queries
-- name: GetWorkout :one
SELECT * FROM workout WHERE id = $1;

-- name: ListWorkouts :many  
SELECT * FROM workout ORDER BY date DESC;

-- name: GetExercise :one
SELECT * FROM exercise WHERE id = $1;

-- name: ListExercises :many
SELECT * FROM exercise ORDER BY name;

-- name: GetSet :one
SELECT * FROM "set" WHERE id = $1;

-- name: ListSets :many
SELECT * FROM "set" ORDER BY id;

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
WHERE s.exercise_id = $1
ORDER BY w.date DESC, s.created_at;

-- INSERT queries for form submission
-- name: CreateWorkout :one
INSERT INTO workout (date, notes)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrCreateExercise :one
INSERT INTO exercise (name) 
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: CreateSet :one
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- Batch insert for better performance (optional)
-- name: CreateSets :copyfrom
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type)
VALUES ($1, $2, $3, $4, $5);

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
    e.name as exercise_name
FROM workout w
JOIN "set" s ON w.id = s.workout_id
JOIN exercise e ON s.exercise_id = e.id
WHERE w.id = $1
ORDER BY e.name, s.id;

-- name: GetExerciseByName :one
SELECT * FROM exercise WHERE name = $1;