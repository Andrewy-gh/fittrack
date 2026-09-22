-- name: ListExerciseAssessmentSets :many
SELECT w.id AS workout_id, w.date AS workout_date, s.id AS set_id,
       s.set_type, s.reps, s.weight
FROM "set" s
JOIN workout w ON w.id = s.workout_id AND w.user_id = s.user_id
JOIN exercise e ON e.id = s.exercise_id AND e.user_id = s.user_id
WHERE s.user_id = sqlc.arg(user_id)
  AND s.exercise_id = sqlc.arg(exercise_id)
  AND w.date >= sqlc.arg(start_at)::timestamptz
  AND w.date < sqlc.arg(end_at)::timestamptz
  AND w.date <= sqlc.arg(observed_at)::timestamptz
ORDER BY w.date, w.id, s.id;

-- name: ListMuscleContributionSets :many
SELECT w.id AS workout_id, w.date AS workout_date, e.id AS exercise_id,
       e.name AS exercise_name, e.catalog_id, s.id AS set_id, s.set_type, s.reps, s.weight
FROM "set" s
JOIN workout w ON w.id = s.workout_id AND w.user_id = s.user_id
JOIN exercise e ON e.id = s.exercise_id AND e.user_id = s.user_id
WHERE s.user_id = sqlc.arg(user_id)
  AND w.date >= sqlc.arg(start_at)::timestamptz
  AND w.date < sqlc.arg(end_at)::timestamptz
  AND w.date <= sqlc.arg(observed_at)::timestamptz
ORDER BY w.date, w.id, s.id;
