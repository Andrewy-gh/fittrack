-- Privilege Escalation Test Script
-- Step 5.2: Attempt UPDATE with tampered WHERE TRUE clause as normal user to confirm denial

\echo '================================================='
\echo 'RLS Privilege Escalation Test'
\echo '================================================='
\echo 'Testing if normal users can bypass RLS with tampered WHERE clauses'
\echo ''

-- First, let's set up the test environment
-- Create a test user context (simulating what the application would do)
SET app.current_user_id = 'test-user-1';

\echo 'Current test user context: test-user-1'
\echo ''

-- Test 1: Check current user can see their own data
\echo '1. Testing legitimate access (should work):'
\echo '------------------------------------------'
SELECT 
    'users' as table_name,
    user_id,
    created_at
FROM users 
WHERE user_id = current_user_id() 
LIMIT 3;

\echo ''
\echo '2. Testing tampered SELECT with WHERE TRUE (should be blocked by RLS):'
\echo '-----------------------------------------------------------------------'
-- This should only return the current user's data, not all data
-- RLS should prevent seeing other users' data even with WHERE TRUE
SELECT 
    'users' as table_name,
    user_id,
    'SHOULD ONLY SEE CURRENT USER' as expectation
FROM users 
WHERE TRUE
LIMIT 5;

\echo ''
\echo '3. Testing tampered UPDATE with WHERE TRUE (should be blocked by RLS):'
\echo '------------------------------------------------------------------------'
-- This should fail or only affect the current user's records
-- First, let's see what we're trying to update
\echo 'Before tampered UPDATE attempt:'
SELECT COUNT(*) as total_workouts_visible FROM workout;

-- Now attempt a tampered UPDATE that tries to modify all records
-- RLS should prevent this from affecting other users' data
DO $$
DECLARE
    affected_rows INTEGER;
BEGIN
    -- Attempt to update all workouts (tampered WHERE TRUE)
    UPDATE workout 
    SET notes = 'TAMPERED BY TEST USER - ' || COALESCE(notes, 'no notes')
    WHERE TRUE;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RAISE NOTICE 'Tampered UPDATE affected % rows (should only be current user''s rows)', affected_rows;
    
    -- Check what we can see after the update
    RAISE NOTICE 'Workouts visible after tampered UPDATE: %', (SELECT COUNT(*) FROM workout);
    
    -- Show a sample of what we can see
    FOR rec IN SELECT id, notes, user_id FROM workout LIMIT 3 LOOP
        RAISE NOTICE 'Workout ID %, User ID %, Notes: %', rec.id, rec.user_id, rec.notes;
    END LOOP;
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Tampered UPDATE failed with error: %', SQLERRM;
END $$;

\echo ''
\echo '4. Testing tampered DELETE with WHERE TRUE (should be blocked by RLS):'
\echo '-----------------------------------------------------------------------'
DO $$
DECLARE
    affected_rows INTEGER;
    initial_count INTEGER;
BEGIN
    -- Check initial count
    SELECT COUNT(*) INTO initial_count FROM exercise;
    RAISE NOTICE 'Exercises visible before tampered DELETE: %', initial_count;
    
    -- Attempt to delete all exercises (tampered WHERE TRUE)
    DELETE FROM exercise 
    WHERE TRUE;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RAISE NOTICE 'Tampered DELETE affected % rows (should only be current user''s rows)', affected_rows;
    
    -- Check what remains
    RAISE NOTICE 'Exercises remaining after tampered DELETE: %', (SELECT COUNT(*) FROM exercise);
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Tampered DELETE failed with error: %', SQLERRM;
END $$;

\echo ''
\echo '5. Testing SQL injection style attempts:'
\echo '------------------------------------------'
-- Test various SQL injection patterns that might try to bypass RLS
DO $$
BEGIN
    -- Test 1: Try to bypass with OR condition
    PERFORM COUNT(*) FROM users WHERE user_id = current_user_id() OR 1=1;
    RAISE NOTICE 'OR 1=1 injection attempt - returned % rows (should only be 1 for current user)', 
        (SELECT COUNT(*) FROM users WHERE user_id = current_user_id() OR 1=1);
        
    -- Test 2: Try UNION injection (should fail due to RLS)
    BEGIN
        PERFORM * FROM (
            SELECT user_id FROM users WHERE user_id = current_user_id()
            UNION 
            SELECT user_id FROM users WHERE TRUE
        ) AS combined;
        RAISE NOTICE 'UNION injection - RLS should still apply to the unioned query';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'UNION injection blocked: %', SQLERRM;
    END;
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'SQL injection test failed: %', SQLERRM;
END $$;

\echo ''
\echo '6. Testing function-based bypass attempts:'
\echo '-------------------------------------------'
DO $$
BEGIN
    -- Try to bypass RLS by manipulating the current_user_id function
    -- This should fail because we don't have privileges to modify the function
    
    RAISE NOTICE 'Current user function returns: %', current_user_id();
    
    -- Try to change session variable directly (should work, but RLS should still apply)
    SET app.current_user_id = 'different-user-id';
    RAISE NOTICE 'After changing session var, function returns: %', current_user_id();
    
    -- Check if we can see different data now
    RAISE NOTICE 'Workouts visible with different user ID: %', (SELECT COUNT(*) FROM workout);
    
    -- Reset back
    SET app.current_user_id = 'test-user-1';
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Function bypass attempt failed: %', SQLERRM;
END $$;

\echo ''
\echo '7. Testing role-based bypass attempts:'
\echo '---------------------------------------'
DO $$
BEGIN
    -- Check current role
    RAISE NOTICE 'Current role: %', current_user;
    
    -- Try to set role (should fail for normal users)
    BEGIN
        SET ROLE postgres;
        RAISE NOTICE 'Successfully changed to postgres role - THIS IS A SECURITY ISSUE!';
        RESET ROLE;
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'Role change blocked (good): %', SQLERRM;
    END;
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Role test failed: %', SQLERRM;
END $$;

\echo ''
\echo '============================================='
\echo 'Privilege Escalation Test Results Summary:'
\echo '============================================='
\echo ''
\echo 'Key Tests Performed:'
\echo '• Tampered WHERE TRUE in SELECT, UPDATE, DELETE'
\echo '• SQL injection patterns (OR 1=1, UNION)'
\echo '• Session variable manipulation'  
\echo '• Role escalation attempts'
\echo ''
\echo 'Expected Results (for properly configured RLS):'
\echo '• All tampered queries should only affect current user data'
\echo '• No unauthorized data should be visible or modifiable'
\echo '• Role changes should be blocked for normal users'
\echo '• RLS policies should remain effective despite bypass attempts'
\echo ''
\echo 'If any test shows unauthorized access, RLS may have gaps!'
\echo '============================================='
