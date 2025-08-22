-- +goose Up
-- +goose StatementBegin

-- Add exercise_order and set_order columns to the "set" table
-- These columns will be used to control the order of exercises within a workout
-- and the order of sets within each exercise
-- Columns are nullable initially to allow for safe deployment to production
-- They will be made NOT NULL in a future migration after backfilling existing data
ALTER TABLE "set"
  ADD COLUMN exercise_order INTEGER,
  ADD COLUMN set_order INTEGER;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the order columns if rolling back
ALTER TABLE "set"
  DROP COLUMN IF EXISTS set_order,
  DROP COLUMN IF EXISTS exercise_order;

-- +goose StatementEnd
