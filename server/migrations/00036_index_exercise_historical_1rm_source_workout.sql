-- +goose Up
SET LOCAL lock_timeout = '100ms';
SET LOCAL statement_timeout = '250ms';
SET LOCAL idle_in_transaction_session_timeout = '2s';
LOCK TABLE exercise IN SHARE MODE NOWAIT;

CREATE INDEX idx_exercise_historical_1rm_source_workout_id
    ON exercise (historical_1rm_source_workout_id)
    WHERE historical_1rm_source_workout_id IS NOT NULL;

-- +goose Down
SET LOCAL lock_timeout = '100ms';
SET LOCAL statement_timeout = '250ms';
SET LOCAL idle_in_transaction_session_timeout = '2s';
LOCK TABLE exercise IN ACCESS EXCLUSIVE MODE NOWAIT;

DROP INDEX idx_exercise_historical_1rm_source_workout_id;
