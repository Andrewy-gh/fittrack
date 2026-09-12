-- +goose Up
CREATE TABLE exercise_prescription (
    exercise_id INTEGER PRIMARY KEY REFERENCES exercise(id) ON DELETE CASCADE,
    user_id VARCHAR(256) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    min_sets INTEGER NOT NULL CHECK (min_sets BETWEEN 1 AND 20),
    max_sets INTEGER NOT NULL CHECK (max_sets BETWEEN min_sets AND 20),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE exercise_prescription ENABLE ROW LEVEL SECURITY;
CREATE POLICY exercise_prescription_owner ON exercise_prescription
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id() AND EXISTS (
        SELECT 1 FROM exercise WHERE exercise.id = exercise_id AND exercise.user_id = current_user_id()
    ));
ALTER TABLE workout ADD COLUMN recommendation_context JSONB NOT NULL DEFAULT '[]'::jsonb
    CHECK (jsonb_typeof(recommendation_context) = 'array');
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'fittrack_app') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON exercise_prescription TO fittrack_app;
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE workout DROP COLUMN recommendation_context;
DROP TABLE exercise_prescription;
