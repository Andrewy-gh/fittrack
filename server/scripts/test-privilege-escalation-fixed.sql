-- Fixed Privilege Escalation Test Script
-- Step 5.2: Attempt UPDATE with tampered WHERE TRUE clause as normal user to confirm denial

\echo '================================================='
\echo 'RLS Privilege Escalation Test - FIXED VERSION'
\echo '================================================='
\echo ''

-- Set up test environment
SET app.current_user_id = 'test-user-1';

\echo 'Current test user context: test-user-1'
\echo ''

-- Test 1: Baseline - what should we see normally?
\echo '1. BASELINE: Normal user access (what we should see):'
\echo '-----------------------------------------------------'
SELECT 'users' as table_name, COUNT(*) as count_visible FROM users;
SELECT 'workout' as table_name, COUNT(*) as count_visible FROM workout; 
SELECT 'exercise' as table_name, COUNT(*) as count_visible FROM exercise;

\echo ''
\echo '2. CRITICAL TEST: Tampered SELECT with WHERE TRUE:'
\echo '---------------------------------------------------'
-- This should only show current user data due to RLS, not ALL data
SELECT 
    'CRITICAL: users with WHERE TRUE' as test_name,
    COUNT(*) as rows_returned,
    CASE 
        WHEN COUNT(*) = 1 THEN 'PASS - RLS working'
        ELSE 'FAIL - RLS BYPASS DETECTED!'
    END as result
FROM users WHERE TRUE;

\echo ''
\echo '3. CRITICAL TEST: Tampered UPDATE with WHERE TRUE:'
\echo '----------------------------------------------------'
DO $$
DECLARE
    before_count INTEGER;
    affected_rows INTEGER;
    after_count INTEGER;
    visible_notes TEXT[];
BEGIN
    -- Count before
    SELECT COUNT(*) INTO before_count FROM workout;
    
    -- Attempt tampered UPDATE
    UPDATE workout 
    SET notes = 'TAMPERED-' || id || '-' || COALESCE(notes, 'none')
    WHERE TRUE;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    
    -- Count after
    SELECT COUNT(*) INTO after_count FROM workout;
    
    -- Check what we can see (sample)
    SELECT array_agg(notes ORDER BY id LIMIT 3) INTO visible_notes FROM workout;
    
    RAISE NOTICE 'UPDATE Test Results:';
    RAISE NOTICE '  Workouts visible before: %', before_count;
    RAISE NOTICE '  Rows affected by UPDATE: %', affected_rows; 
    RAISE NOTICE '  Workouts visible after: %', after_count;
    RAISE NOTICE '  Sample notes visible: %', visible_notes;
    RAISE NOTICE '  Assessment: % rows affected (should only be test-user-1 rows)', affected_rows;
    
    IF affected_rows = before_count AND before_count > 2 THEN
        RAISE NOTICE '  RESULT: FAIL - May have updated other users data!';
    ELSIF affected_rows <= 2 THEN
        RAISE NOTICE '  RESULT: PASS - Only affected expected user rows';
    ELSE
        RAISE NOTICE '  RESULT: UNCLEAR - Need manual verification';
    END IF;
    
END $$;

\echo ''
\echo '4. CRITICAL TEST: Cross-user data visibility:'
\echo '-----------------------------------------------'
DO $$
DECLARE
    user1_count INTEGER;
    user2_count INTEGER;
    user3_count INTEGER;
BEGIN
    -- Check what each user context can see
    SET app.current_user_id = 'test-user-1';
    SELECT COUNT(*) INTO user1_count FROM workout;
    
    SET app.current_user_id = 'test-user-2'; 
    SELECT COUNT(*) INTO user2_count FROM workout;
    
    SET app.current_user_id = 'test-user-3';
    SELECT COUNT(*) INTO user3_count FROM workout;
    
    RAISE NOTICE 'Cross-user visibility test:';
    RAISE NOTICE '  User 1 sees % workouts', user1_count;
    RAISE NOTICE '  User 2 sees % workouts', user2_count; 
    RAISE NOTICE '  User 3 sees % workouts', user3_count;
    
    -- Each user should only see their own data (2 workouts each)
    IF user1_count = 2 AND user2_count = 2 AND user3_count = 2 THEN
        RAISE NOTICE '  RESULT: PASS - Each user only sees own data';
    ELSE
        RAISE NOTICE '  RESULT: FAIL - Users seeing other users data!';
    END IF;
    
    -- Reset context
    SET app.current_user_id = 'test-user-1';
END $$;

\echo ''
\echo '5. CRITICAL TEST: SQL Injection patterns:'
\echo '-------------------------------------------'
DO $$
DECLARE
    injection_count INTEGER;
    normal_count INTEGER;
BEGIN
    -- Normal query
    SELECT COUNT(*) INTO normal_count FROM users WHERE user_id = current_user_id();
    
    -- Injection attempt: OR 1=1
    SELECT COUNT(*) INTO injection_count FROM users WHERE user_id = current_user_id() OR 1=1;
    
    RAISE NOTICE 'SQL Injection Test:';
    RAISE NOTICE '  Normal query returns: % rows', normal_count;
    RAISE NOTICE '  OR 1=1 injection returns: % rows', injection_count;
    
    IF injection_count = normal_count AND injection_count = 1 THEN
        RAISE NOTICE '  RESULT: PASS - Injection blocked by RLS';
    ELSE
        RAISE NOTICE '  RESULT: FAIL - SQL injection successful!';
    END IF;
END $$;

\echo ''
\echo '6. Test RLS on "set" table (found to have no policies):'
\echo '---------------------------------------------------------'
DO $$
DECLARE
    set_count INTEGER;
BEGIN
    -- The audit found that "set" table has RLS enabled but no policies
    -- This should block ALL access
    BEGIN
        SELECT COUNT(*) INTO set_count FROM "set";
        RAISE NOTICE 'Set table access: % rows visible', set_count;
        
        IF set_count = 0 THEN
            RAISE NOTICE '  RESULT: Expected - No policies means no access';
        ELSE
            RAISE NOTICE '  RESULT: UNEXPECTED - Should be blocked due to no policies';
        END IF;
        
    EXCEPTION
        WHEN insufficient_privilege THEN
            RAISE NOTICE '  RESULT: PASS - Access properly blocked (insufficient privilege)';
        WHEN OTHERS THEN
            RAISE NOTICE '  RESULT: Access blocked with error: %', SQLERRM;
    END;
END $$;

\echo ''
\echo '============================================='
\echo 'FINAL ASSESSMENT:'
\echo '============================================='
\echo ''
\echo 'Key findings from privilege escalation tests:'
\echo ''
\echo 'Issues found in audit:'
\echo '• goose_db_version table has NO RLS (low risk - migration table)'  
\echo '• "set" table has RLS enabled but NO policies (blocks all access)'
\echo '• Core tables (users, workout, exercise) have proper RLS policies'
\echo ''
\echo 'Security test results:'
\echo '• WHERE TRUE bypass attempts'
\echo '• SQL injection pattern tests'  
\echo '• Cross-user data visibility tests'
\echo '• Policy gap verification'
\echo ''
\echo 'Review the NOTICE messages above for specific test results.'
\echo '============================================='
