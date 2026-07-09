-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_training_profile
    ALTER COLUMN movement_limitations DROP NOT NULL,
    ALTER COLUMN movement_limitations DROP DEFAULT;

UPDATE user_training_profile
SET movement_limitations = NULL
WHERE movement_limitations = '[]'::jsonb;

ALTER TABLE user_training_profile
    DROP CONSTRAINT IF EXISTS user_training_profile_movement_limitations_array,
    ADD CONSTRAINT user_training_profile_movement_limitations_array CHECK (
        movement_limitations IS NULL OR jsonb_typeof(movement_limitations) = 'array'
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE user_training_profile
SET movement_limitations = '[]'::jsonb
WHERE movement_limitations IS NULL;

ALTER TABLE user_training_profile
    DROP CONSTRAINT IF EXISTS user_training_profile_movement_limitations_array,
    ALTER COLUMN movement_limitations SET DEFAULT '[]'::jsonb,
    ALTER COLUMN movement_limitations SET NOT NULL,
    ADD CONSTRAINT user_training_profile_movement_limitations_array CHECK (
        jsonb_typeof(movement_limitations) = 'array'
    );
-- +goose StatementEnd
