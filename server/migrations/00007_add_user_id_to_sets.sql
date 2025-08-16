-- +goose Up
-- +goose StatementBegin

-- Migration: Add user_id to set table and enable CASCADE delete
-- This improves security and simplifies queries by removing the need for complex JOINs

-- Step 1: Add user_id column to set table (nullable initially)
ALTER TABLE "set" ADD COLUMN user_id VARCHAR(256);

-- Step 2: Populate user_id from workout table for existing data
UPDATE "set" 
SET user_id = w.user_id 
FROM workout w 
WHERE "set".workout_id = w.id;

-- Step 3: Make user_id NOT NULL and add foreign key constraint
ALTER TABLE "set" ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE "set" ADD CONSTRAINT set_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- Step 4: Drop and recreate workout_id foreign key with CASCADE
-- Note: Need to find the actual constraint name first
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Find the existing foreign key constraint name
    SELECT conname INTO constraint_name
    FROM pg_constraint 
    WHERE conrelid = '"set"'::regclass 
    AND confrelid = 'workout'::regclass;
    
    -- Drop the existing constraint if it exists
    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE "set" DROP CONSTRAINT ' || constraint_name;
    END IF;
    
    -- Add the new constraint with CASCADE
    ALTER TABLE "set" ADD CONSTRAINT set_workout_id_fkey 
        FOREIGN KEY (workout_id) REFERENCES workout(id) ON DELETE CASCADE;
END $$;

-- Step 5: Add index for performance
CREATE INDEX idx_set_user_id ON "set"(user_id);

-- Step 6: Drop existing complex RLS policies for set table
DROP POLICY IF EXISTS set_select_policy ON "set";
DROP POLICY IF EXISTS set_insert_policy ON "set";
DROP POLICY IF EXISTS set_update_policy ON "set";
DROP POLICY IF EXISTS set_delete_policy ON "set";

-- Step 7: Create simplified RLS policies using direct user_id column
CREATE POLICY set_select_policy ON "set"
    FOR SELECT TO PUBLIC
    USING (user_id = current_user_id());

CREATE POLICY set_insert_policy ON "set"
    FOR INSERT TO PUBLIC
    WITH CHECK (user_id = current_user_id());

CREATE POLICY set_update_policy ON "set"
    FOR UPDATE TO PUBLIC
    USING (user_id = current_user_id())
    WITH CHECK (user_id = current_user_id());

CREATE POLICY set_delete_policy ON "set"
    FOR DELETE TO PUBLIC
    USING (user_id = current_user_id());

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Step 1: Drop simplified RLS policies
DROP POLICY IF EXISTS set_select_policy ON "set";
DROP POLICY IF EXISTS set_insert_policy ON "set";
DROP POLICY IF EXISTS set_update_policy ON "set";
DROP POLICY IF EXISTS set_delete_policy ON "set";

-- Step 2: Recreate complex RLS policies using EXISTS subqueries
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

-- Step 3: Drop index
DROP INDEX IF EXISTS idx_set_user_id;

-- Step 4: Drop and recreate workout_id foreign key without CASCADE
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    -- Find the existing foreign key constraint name
    SELECT conname INTO constraint_name
    FROM pg_constraint 
    WHERE conrelid = '"set"'::regclass 
    AND confrelid = 'workout'::regclass;
    
    -- Drop the existing constraint if it exists
    IF constraint_name IS NOT NULL THEN
        EXECUTE 'ALTER TABLE "set" DROP CONSTRAINT ' || constraint_name;
    END IF;
    
    -- Add the constraint back without CASCADE (original behavior)
    ALTER TABLE "set" ADD CONSTRAINT set_workout_id_fkey 
        FOREIGN KEY (workout_id) REFERENCES workout(id);
END $$;

-- Step 5: Drop user_id constraint and column
ALTER TABLE "set" DROP CONSTRAINT IF EXISTS set_user_id_fkey;
ALTER TABLE "set" DROP COLUMN IF EXISTS user_id;

-- +goose StatementEnd
