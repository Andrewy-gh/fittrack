-- RLS Audit Script
-- Step 5: Audit policy coverage & privilege escalation
-- 1. Cross-check pg_policies vs. information_schema.tables for any tables lacking RLS
-- 2. Check for tables with RLS enabled but no policies
-- 3. List all current RLS policies

\echo '====================================='
\echo 'RLS Policy Coverage Audit Report'
\echo '====================================='
\echo ''

-- Check 1: List all user tables and their RLS status
\echo '1. All Tables and RLS Status:'
\echo '-----------------------------'
SELECT 
    schemaname as schema,
    tablename as table_name,
    rowsecurity as rls_enabled,
    CASE 
        WHEN rowsecurity THEN 'ENABLED'
        ELSE 'DISABLED'
    END as rls_status
FROM pg_tables 
WHERE schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
ORDER BY schemaname, tablename;

\echo ''
\echo '2. Tables with RLS DISABLED (Security Risk):'
\echo '--------------------------------------------'
SELECT 
    schemaname as schema,
    tablename as table_name,
    'NO RLS - SECURITY RISK!' as issue
FROM pg_tables 
WHERE schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
AND tablename NOT IN (
    SELECT tablename 
    FROM pg_tables t
    JOIN pg_class c ON c.relname = t.tablename
    WHERE c.relrowsecurity = true
);

\echo ''
\echo '3. Current RLS Policies:'
\echo '------------------------'
SELECT 
    schemaname as schema,
    tablename as table_name,
    policyname as policy_name,
    permissive as permissive,
    roles as applicable_roles,
    cmd as command_type,
    CASE 
        WHEN qual IS NOT NULL THEN 'Has USING clause'
        ELSE 'No USING clause'
    END as using_clause,
    CASE 
        WHEN with_check IS NOT NULL THEN 'Has WITH CHECK clause'
        ELSE 'No WITH CHECK clause'
    END as with_check_clause
FROM pg_policies
ORDER BY schemaname, tablename, policyname;

\echo ''
\echo '4. Tables with RLS enabled but NO policies (Will block all access):'
\echo '-------------------------------------------------------------------'
SELECT DISTINCT
    t.schemaname as schema,
    t.tablename as table_name,
    'RLS enabled but NO policies - will block all access!' as issue
FROM pg_tables t
JOIN pg_class c ON c.relname = t.tablename AND c.relnamespace = (
    SELECT oid FROM pg_namespace WHERE nspname = t.schemaname
)
WHERE c.relrowsecurity = true
AND t.schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
AND NOT EXISTS (
    SELECT 1 FROM pg_policies p 
    WHERE p.schemaname = t.schemaname 
    AND p.tablename = t.tablename
);

\echo ''
\echo '5. Policy Coverage Summary:'
\echo '---------------------------'
WITH table_stats AS (
    SELECT 
        COUNT(*) as total_tables,
        COUNT(CASE WHEN c.relrowsecurity THEN 1 END) as tables_with_rls,
        COUNT(CASE WHEN NOT c.relrowsecurity THEN 1 END) as tables_without_rls
    FROM pg_tables t
    JOIN pg_class c ON c.relname = t.tablename AND c.relnamespace = (
        SELECT oid FROM pg_namespace WHERE nspname = t.schemaname
    )
    WHERE t.schemaname NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
),
policy_stats AS (
    SELECT COUNT(DISTINCT schemaname||'.'||tablename) as tables_with_policies
    FROM pg_policies
)
SELECT 
    ts.total_tables as "Total User Tables",
    ts.tables_with_rls as "Tables with RLS Enabled",
    ts.tables_without_rls as "Tables WITHOUT RLS (Risk)",
    ps.tables_with_policies as "Tables with Policies",
    (ts.tables_with_rls - ps.tables_with_policies) as "RLS enabled but No Policies"
FROM table_stats ts, policy_stats ps;

\echo ''
\echo '6. Detailed Policy Analysis:'
\echo '----------------------------'
SELECT 
    p.schemaname||'.'||p.tablename as table_name,
    COUNT(*) as policy_count,
    STRING_AGG(DISTINCT p.cmd, ', ') as covered_commands,
    CASE 
        WHEN COUNT(*) = 0 THEN 'NO POLICIES - BLOCKS ALL ACCESS'
        WHEN COUNT(CASE WHEN p.cmd = 'ALL' THEN 1 END) > 0 THEN 'Has ALL command policy'
        WHEN 'SELECT' = ANY(STRING_TO_ARRAY(STRING_AGG(DISTINCT p.cmd, ','), ',')) 
             AND 'INSERT' = ANY(STRING_TO_ARRAY(STRING_AGG(DISTINCT p.cmd, ','), ','))
             AND 'UPDATE' = ANY(STRING_TO_ARRAY(STRING_AGG(DISTINCT p.cmd, ','), ','))
             AND 'DELETE' = ANY(STRING_TO_ARRAY(STRING_AGG(DISTINCT p.cmd, ','), ','))
             THEN 'Full CRUD coverage'
        ELSE 'Partial coverage - check commands'
    END as coverage_assessment
FROM pg_policies p
GROUP BY p.schemaname, p.tablename
ORDER BY p.schemaname, p.tablename;

\echo ''
\echo '====================================='
\echo 'RLS Audit Complete'
\echo '====================================='
