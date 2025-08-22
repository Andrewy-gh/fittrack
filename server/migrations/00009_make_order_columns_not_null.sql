-- +goose Up
-- +goose StatementBegin

-- Make exercise_order and set_order columns NOT NULL in the "set" table
-- First, update any existing NULL values to have default values
UPDATE "set" 
SET exercise_order = 0 
WHERE exercise_order IS NULL;

UPDATE "set" 
SET set_order = 0 
WHERE set_order IS NULL;

-- Then modify the columns to be NOT NULL
ALTER TABLE "set"
  ALTER COLUMN exercise_order SET NOT NULL,
  ALTER COLUMN set_order SET NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert exercise_order and set_order columns to nullable
ALTER TABLE "set"
  ALTER COLUMN exercise_order DROP NOT NULL,
  ALTER COLUMN set_order DROP NOT NULL;

-- +goose StatementEnd