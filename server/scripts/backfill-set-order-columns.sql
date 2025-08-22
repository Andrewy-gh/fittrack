-- Backfill script for exercise_order and set_order columns
-- This script populates the new columns based on the previous ordering logic:
-- - exercise_order: ordered by exercise name within each workout
-- - set_order: ordered by created_at and id within each exercise

-- NOTE: This is NOT a migration - it's a standalone script for backfilling data
-- Run this after applying migration 00008 but before making the columns NOT NULL

-- For RLS-enabled environments, ensure you have proper permissions or run as superuser
-- Alternatively, set the session variable: SELECT set_config('app.current_user_id', 'your-user-id', false);

BEGIN;

-- Backfill exercise_order and set_order for all existing sets
-- This query maintains the previous ordering behavior:
-- 1. Within each workout, exercises are ordered by exercise name (as in GetWorkoutWithSets)
-- 2. Within each exercise, sets are ordered by created_at then id (as in previous ordering)
WITH ranked AS (
  SELECT
    s.id,
    s.workout_id,
    s.exercise_id,
    s.user_id,
    -- Exercise order within the workout (by exercise name, then exercise id for determinism)
    DENSE_RANK() OVER (
      PARTITION BY s.workout_id 
      ORDER BY e.name, e.id
    ) AS new_exercise_order,
    -- Set order within the exercise (by created_at, then id for determinism)  
    ROW_NUMBER() OVER (
      PARTITION BY s.workout_id, s.exercise_id 
      ORDER BY s.created_at, s.id
    ) AS new_set_order
  FROM "set" s
  JOIN exercise e ON e.id = s.exercise_id
  WHERE s.exercise_order IS NULL OR s.set_order IS NULL  -- Only update rows that haven't been set
)
UPDATE "set" s
SET 
  exercise_order = r.new_exercise_order,
  set_order = r.new_set_order
FROM ranked r
WHERE r.id = s.id;

-- Verify the results
SELECT 
  COUNT(*) as total_sets,
  COUNT(exercise_order) as sets_with_exercise_order,
  COUNT(set_order) as sets_with_set_order,
  MIN(exercise_order) as min_exercise_order,
  MAX(exercise_order) as max_exercise_order,
  MIN(set_order) as min_set_order,
  MAX(set_order) as max_set_order
FROM "set";

-- Show a sample of the results
SELECT 
  w.id as workout_id,
  w.date::date as workout_date,
  e.name as exercise_name,
  s.exercise_order,
  s.set_order,
  s.weight,
  s.reps,
  s.created_at::timestamp as set_created_at
FROM "set" s
JOIN workout w ON w.id = s.workout_id
JOIN exercise e ON e.id = s.exercise_id
ORDER BY w.id, s.exercise_order, s.set_order
LIMIT 20;

COMMIT;

-- Instructions:
-- 1. Apply this script after running migration 00008
-- 2. For production, consider running this per user or in batches:
--    WHERE s.user_id = 'specific-user-id'
-- 3. Monitor performance on large datasets
-- 4. Verify the results before proceeding with making columns NOT NULL
