1. Create a feature branch and confirm migration numbering
- Branch name suggestion: feat/add-set-order-columns
- Confirm the next Goose migration number is 00008; if your repo already has higher numbers, increment accordingly.
2. Add Goose migration 00008_add_order_columns_to_set_table.sql
Create server/migrations/00008_add_order_columns_to_set_table.sql with:
```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE "set"
  ADD COLUMN exercise_order INTEGER,
  ADD COLUMN set_order INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "set"
  DROP COLUMN IF EXISTS set_order,
  DROP COLUMN IF EXISTS exercise_order;
-- +goose StatementEnd
```
Notes:
- Columns are intentionally NULLable; they will be 1-indexed by convention and made NOT NULL in a future migration after backfilling.
- Table name remains quoted as "set" because it is a reserved word.
3. Apply the migration locally and verify
- Run: goose -dir server/migrations postgres "$POSTGRES_DSN" up
- Verify columns exist:
```sql
SELECT column_name, is_nullable
FROM information_schema.columns
WHERE table_name = 'set' AND column_name IN ('exercise_order', 'set_order');
```
Both should be present and nullable.
4. Update SQL queries to use exercise_order and set_order
Edit server/query.sql:
1) GetWorkoutWithSets: replace ORDER BY with new ordering
From:
```sql
ORDER BY e.name, s.id;
```
To:
```sql
ORDER BY s.exercise_order NULLS LAST, s.set_order NULLS LAST, s.id;
```
2) ListSets: replace ORDER BY id
From:
```sql
ORDER BY id;
```
To:
```sql
ORDER BY exercise_order NULLS LAST, set_order NULLS LAST, id;
```
3) GetExerciseWithSets: add exercise_order/set_order as secondary ordering
From:
```sql
ORDER BY w.date DESC, s.created_at;
```
To:
```sql
ORDER BY w.date DESC, s.exercise_order NULLS LAST, s.set_order NULLS LAST, s.created_at, s.id;
```
Rationale:
- NULLS LAST keeps existing data (which will initially be NULL) grouped at the end while still yielding deterministic results thanks to the id tie-breaker.
- The id tie-breaker preserves stable ordering during the transition.
5. Regenerate sqlc code
- Run: sqlc generate
- Verify server/internal/database/query.sql.go is updated and compiles.
- No new fields are required in SELECT lists for this task; we are only changing ORDER BY clauses.
6. Build and run unit/integration tests to capture ordering-related failures
- Build: go build ./...
- Test: go test ./...
- Expect tests that previously assumed ordering by e.name or set id to fail; note all occurrences.
7. Update tests to set and assert ordering via exercise_order and set_order
- Wherever tests relied on implicit ordering, explicitly set exercise_order and set_order in fixtures before querying, for deterministic expectations. For example:
```sql
-- Example: set orders for a single workout deterministically based on previous behavior
WITH ranked AS (
  SELECT
    s.id,
    DENSE_RANK() OVER (PARTITION BY s.workout_id ORDER BY e.name, e.id) AS ex_ord,
    ROW_NUMBER() OVER (PARTITION BY s.workout_id, s.exercise_id ORDER BY s.created_at, s.id) AS set_ord
  FROM "set" s
  JOIN exercise e ON e.id = s.exercise_id
  WHERE s.workout_id = $1 AND s.user_id = $2
)
UPDATE "set" s
SET exercise_order = r.ex_ord,
    set_order = r.set_ord
FROM ranked r
WHERE r.id = s.id;
```
- If tests create sets across multiple workouts, run the above per workout_id as part of test setup.
- Update assertions to reflect new ordering semantics:
  - ListSets returns ordered by exercise_order, then set_order.
  - GetWorkoutWithSets returns sets ordered by exercise_order, then set_order.
  - GetExerciseWithSets returns by workout date DESC, then exercise_order, set_order, then created_at and id.
- Ensure tests with RLS continue to set the session via testutils.SetTestUserContext or set app.current_user_id explicitly.
8. Optional: add a local backfill script to ease manual QA
Create a SQL script (not a migration) to populate existing rows during manual testing:
```sql
-- Backfill per workout to emulate prior grouping (by exercise name) and per-exercise set order (by created_at then id)
WITH ranked AS (
  SELECT
    s.id,
    s.workout_id,
    DENSE_RANK() OVER (PARTITION BY s.workout_id ORDER BY e.name, e.id) AS ex_ord,
    ROW_NUMBER() OVER (PARTITION BY s.workout_id, s.exercise_id ORDER BY s.created_at, s.id) AS set_ord
  FROM "set" s
  JOIN exercise e ON e.id = s.exercise_id
  WHERE s.user_id = $1
)
UPDATE "set" s
SET exercise_order = r.ex_ord,
    set_order = r.set_ord
FROM ranked r
WHERE r.id = s.id;
```
Run this locally to confirm UI/API ordering behaves as expected before production backfill.
9. Smoke test API endpoints that consume updated queries
- Verify endpoints that call:
  - GetWorkoutWithSets
  - ListSets
  - GetExerciseWithSets
Return data in the expected order after setting exercise_order and set_order.
- Check pagination or client-side sorting assumptions remain valid.
10. Update migration README to document 00008
- Add an entry describing 00008_add_order_columns_to_set_table.sql:
  - Adds nullable INTEGER columns exercise_order and set_order to "set"
  - Columns are 1-indexed by convention; they will be made NOT NULL in a future migration after backfill
  - No indexes added at this time (pure ordering; no filtering)
11. Create PR with rollout/backfill guidance
Include in the PR description:
- Summary of schema changes and updated ORDER BY clauses
- Test strategy and how ordering is set in tests
- Backfill plan for staging/production:
  - Run the provided backfill script per tenant/user or globally, ensuring RLS session variable is set or superuser is used appropriately
  - Validate ordering in the UI/API
- Follow-up migration plan to enforce NOT NULL and (optionally) a CHECK constraint ensuring values are ≥ 1
12. Post-merge deployment checklist
- Run goose up in each environment
- Backfill exercise_order and set_order in each environment
- Monitor logs for queries touching "set" to ensure no regressions
- Later: add a new migration to set DEFAULTs, enforce NOT NULL, and possibly add a CHECK (exercise_order &gt;= 1, set_order &gt;= 1)

**Summary of the Plan**

Goal: Add two 1-indexed nullable INTEGER columns (exercise_order and set_order) to the "set" table to control ordering of exercises within workouts and sets within exercises.

Key Features:
•  ✅ Nullable columns initially - Safe for production deployment since you have existing data
•  ✅ 1-indexed ordering - Natural for users (1, 2, 3, etc.)
•  ✅ Preserves existing behavior - Uses NULLS LAST with id tie-breaker during transition
•  ✅ Updates 3 key queries that currently order by e.name or s.id
•  ✅ Comprehensive test updates - Ensures deterministic ordering in tests
•  ✅ Includes backfill strategy - Script to populate existing data based on current ordering logic

Migration Strategy:
1. Phase 1 (this plan): Add nullable columns, update queries, test
2. Phase 2 (future): Backfill production data, make columns NOT NULL

The plan includes 12 detailed tasks covering migration creation, query updates, code generation, testing, and deployment guidance.