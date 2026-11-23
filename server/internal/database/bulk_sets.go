package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

// sqlc does not support PostgreSQL's multi-array unnest form. Keep the arrays
// aligned in the caller so each input set remains one row.
const bulkCreateSets = `INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
SELECT x.exercise_id,$1,x.weight,x.reps,x.set_type,$2,x.exercise_order,x.set_order
 FROM unnest($3::int[],$4::numeric[],$5::int[],$6::text[],$7::int[],$8::int[]) AS x(exercise_id,weight,reps,set_type,exercise_order,set_order)`

func (q *Queries) BulkCreateSets(ctx context.Context, workoutID int32, userID string, exerciseIDs []int32, weights []pgtype.Numeric, reps []int32, setTypes []string, exerciseOrder, setOrder []int32) error {
	_, err := q.db.Exec(ctx, bulkCreateSets, workoutID, userID, exerciseIDs, weights, reps, setTypes, exerciseOrder, setOrder)
	return err
}
