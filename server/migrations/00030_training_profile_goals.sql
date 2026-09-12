-- +goose Up
ALTER TABLE user_training_profile ADD COLUMN goals TEXT[] NOT NULL DEFAULT '{}';
UPDATE user_training_profile SET goals = ARRAY[primary_goal] WHERE primary_goal IS NOT NULL;
ALTER TABLE user_training_profile ADD CONSTRAINT user_training_profile_goals_valid
CHECK (goals <@ ARRAY['strength', 'hypertrophy', 'endurance', 'general_fitness', 'weight_loss', 'mobility']::text[] AND array_position(goals, NULL) IS NULL AND cardinality(goals) <= 6);

-- +goose Down
ALTER TABLE user_training_profile DROP COLUMN goals;
