-- +goose Up
-- +goose StatementBegin
-- Row Level Security (RLS) policies for multi-tenancy
-- Note: The application must set the session variable 'app.current_user_id' for each request
-- when using connection pooling to ensure proper user context isolation
-- Create function to get current user ID from session variable
CREATE OR REPLACE FUNCTION current_user_id() 
RETURNS TEXT AS $$
    SELECT current_setting('app.current_user_id', true);
$$ LANGUAGE SQL STABLE;

-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE workout ENABLE ROW LEVEL SECURITY;
ALTER TABLE exercise ENABLE ROW LEVEL SECURITY;
ALTER TABLE "set" ENABLE ROW LEVEL SECURITY;

-- Create policies for users table
CREATE POLICY users_policy ON users
    FOR ALL TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

-- Create policies for workout table
CREATE POLICY workout_select_policy ON workout
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY workout_insert_policy ON workout
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY workout_update_policy ON workout
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY workout_delete_policy ON workout
    FOR DELETE TO PUBLIC
    USING (user_id = current_user_id());

-- Create policies for exercise table
CREATE POLICY exercise_select_policy ON exercise
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY exercise_insert_policy ON exercise
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY exercise_update_policy ON exercise
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY exercise_delete_policy ON exercise
    FOR DELETE TO PUBLIC
    USING (user_id = current_user_id());

-- Create policies for set table
CREATE POLICY set_select_policy ON "set"
    FOR SELECT TO PUBLIC
    USING (
        EXISTS (
            SELECT 1 FROM workout w 
            WHERE w.id = "set".workout_id 
            AND w.user_id = current_user_id()
        )
    );

CREATE POLICY set_insert_policy ON "set"
    FOR INSERT TO PUBLIC
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM workout w 
            WHERE w.id = "set".workout_id 
            AND w.user_id = current_user_id()
        )
    );

CREATE POLICY set_update_policy ON "set"
    FOR UPDATE TO PUBLIC
    USING (
        EXISTS (
            SELECT 1 FROM workout w 
            WHERE w.id = "set".workout_id 
            AND w.user_id = current_user_id()
        )
    )
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM workout w 
            WHERE w.id = "set".workout_id 
            AND w.user_id = current_user_id()
        )
    );

CREATE POLICY set_delete_policy ON "set"
    FOR DELETE TO PUBLIC
    USING (
        EXISTS (
            SELECT 1 FROM workout w 
            WHERE w.id = "set".workout_id 
            AND w.user_id = current_user_id()
        )
    );

-- Grant necessary permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON workout TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON exercise TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON "set" TO PUBLIC;
GRANT USAGE ON SEQUENCE users_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE workout_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE exercise_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE set_id_seq TO PUBLIC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop policies
DROP POLICY IF EXISTS users_policy ON users;
DROP POLICY IF EXISTS workout_select_policy ON workout;
DROP POLICY IF EXISTS workout_insert_policy ON workout;
DROP POLICY IF EXISTS workout_update_policy ON workout;
DROP POLICY IF EXISTS workout_delete_policy ON workout;
DROP POLICY IF EXISTS exercise_select_policy ON exercise;
DROP POLICY IF EXISTS exercise_insert_policy ON exercise;
DROP POLICY IF EXISTS exercise_update_policy ON exercise;
DROP POLICY IF EXISTS exercise_delete_policy ON exercise;
DROP POLICY IF EXISTS set_select_policy ON "set";
DROP POLICY IF EXISTS set_insert_policy ON "set";
DROP POLICY IF EXISTS set_update_policy ON "set";
DROP POLICY IF EXISTS set_delete_policy ON "set";

-- Disable RLS
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE workout DISABLE ROW LEVEL SECURITY;
ALTER TABLE exercise DISABLE ROW LEVEL SECURITY;
ALTER TABLE "set" DISABLE ROW LEVEL SECURITY;

-- Drop function
DROP FUNCTION IF EXISTS current_user_id();

-- Revoke permissions
REVOKE ALL ON users FROM PUBLIC;
REVOKE ALL ON workout FROM PUBLIC;
REVOKE ALL ON exercise FROM PUBLIC;
REVOKE ALL ON "set" FROM PUBLIC;
REVOKE ALL ON SEQUENCE users_id_seq FROM PUBLIC;
REVOKE ALL ON SEQUENCE workout_id_seq FROM PUBLIC;
REVOKE ALL ON SEQUENCE exercise_id_seq FROM PUBLIC;
REVOKE ALL ON SEQUENCE set_id_seq FROM PUBLIC;
-- +goose StatementEnd