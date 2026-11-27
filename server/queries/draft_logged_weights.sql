-- name: ListDraftLoggedWeights :many
SELECT e.name, s.reps, s.set_type, MIN(s.weight)::numeric AS weight
FROM exercise e
CROSS JOIN LATERAL (
    SELECT w.id
    FROM "set" prior
    JOIN workout w ON w.id = prior.workout_id AND w.user_id = prior.user_id
    WHERE prior.exercise_id = e.id AND prior.user_id = e.user_id
      AND w.date BETWEEN sqlc.arg(history_start)::timestamptz AND sqlc.arg(history_end)::timestamptz
    ORDER BY w.date DESC, w.id DESC
    LIMIT 1
) latest
JOIN "set" s ON s.workout_id = latest.id AND s.exercise_id = e.id AND s.user_id = e.user_id
WHERE e.user_id = sqlc.arg(user_id)
  AND e.name = ANY(sqlc.arg(exercise_names)::text[])
GROUP BY e.name, s.reps, s.set_type
HAVING COUNT(*) = COUNT(s.weight) AND MIN(s.weight) = MAX(s.weight) AND MIN(s.weight) > 0;
