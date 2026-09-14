-- name: GetRecommendationTrainingHistory :many
WITH recent_sessions AS (
    SELECT DISTINCT w.id AS workout_id, w.date
    FROM workout w
    JOIN "set" observed ON observed.workout_id = w.id AND observed.user_id = w.user_id
    WHERE observed.exercise_id = sqlc.arg(exercise_id)
      AND w.user_id = sqlc.arg(user_id)
      AND w.date <= sqlc.arg(as_of)
    ORDER BY w.date DESC, w.id DESC
    LIMIT 2
)
SELECT
    recent_sessions.workout_id,
    recent_sessions.date,
    working.id AS working_set_id,
    working.weight,
    working.reps
FROM recent_sessions
LEFT JOIN "set" working
    ON working.workout_id = recent_sessions.workout_id
    AND working.exercise_id = sqlc.arg(exercise_id)
    AND working.user_id = sqlc.arg(user_id)
    AND working.set_type = 'working'
ORDER BY recent_sessions.date DESC, recent_sessions.workout_id DESC, working.set_order, working.id;

-- name: SaveWorkoutRecommendationContext :exec
UPDATE workout SET recommendation_context = sqlc.arg(recommendation_context)::text::jsonb
WHERE id = sqlc.arg(workout_id) AND user_id = sqlc.arg(user_id);

-- name: GetWorkoutRecommendationContext :one
SELECT recommendation_context FROM workout WHERE id = $1 AND user_id = $2;
