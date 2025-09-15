# SELECT * Query Removal Progress

## Analysis Summary

I've analyzed all frontend components and their usage of API responses to determine which columns are actually needed vs. which can be removed for security and performance.

## Frontend Component → Query Mapping

### 1. Workout Queries

**`ListWorkouts` query** (used by `/workouts/` page):
- **Fields USED**: `id`, `date`, `workout_focus`, `created_at`
- **Fields NOT USED**: `user_id`, `notes`, `updated_at`
- **Recommendation**: Remove `user_id` (security), keep `notes` and `updated_at` for potential future use

**`GetWorkout` query** (used for individual workout operations):
- **Fields POTENTIALLY USED**: All fields may be needed for backend operations
- **Recommendation**: Remove `user_id` from response (used for auth, not needed in response)

### 2. Exercise Queries

**`ListExercises` query** (used by `/exercises/` page):
- **Fields USED**: `id`, `name`
- **Fields NOT USED**: `user_id`, `created_at`, `updated_at`
- **Recommendation**: Remove `user_id`, `created_at`, `updated_at`

**`GetExercise` query** (used for individual exercise operations):
- **Fields POTENTIALLY USED**: `id`, `name`
- **Fields NOT USED**: `user_id`, `created_at`, `updated_at`
- **Recommendation**: Remove `user_id`, `created_at`, `updated_at`

**`GetExerciseByName` query** (used for exercise lookups):
- **Recommendation**: Same as `GetExercise` - remove `user_id`, `created_at`, `updated_at`

### 3. Set Queries

**`GetSet` and `ListSets`**:
- **Status**: Keep all fields for now (may be used in backend logic)
- **Recommendation**: Remove `user_id` from response

### 4. User Queries

**`GetUser` and `GetUserByUserID`**:
- **Frontend Usage**: None found - authentication handled by external service
- **Recommendation**: Evaluate if these endpoints are needed at all

## Required Changes

### 1. Update `server/query.sql`

Replace these SELECT * queries with specific columns:

```sql
-- name: GetWorkout :one
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE id = $1 AND user_id = $2;

-- name: ListWorkouts :many
SELECT id, date, notes, workout_focus, created_at, updated_at FROM workout WHERE user_id = $1 ORDER BY date DESC;

-- name: GetExercise :one
SELECT id, name FROM exercise WHERE id = $1 AND user_id = $2;

-- name: ListExercises :many
SELECT id, name FROM exercise WHERE user_id = $1 ORDER BY name;

-- name: GetSet :one
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set" WHERE id = $1 AND user_id = $2;

-- name: ListSets :many
SELECT id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order FROM "set" WHERE user_id = $1 ORDER BY exercise_order, set_order, id;

-- name: GetExerciseByName :one
SELECT id, name FROM exercise WHERE name = $1 AND user_id = $2;

-- name: GetUser :one
SELECT id, user_id, created_at FROM users WHERE id = $1;

-- name: GetUserByUserID :one
SELECT id, user_id, created_at FROM users WHERE user_id = $1 LIMIT 1;
```

Also update RETURNING clauses to match:
```sql
-- name: CreateWorkout :one
INSERT INTO workout (date, notes, workout_focus, user_id)
VALUES ($1, $2, $3, $4)
RETURNING id, date, notes, workout_focus, created_at, updated_at;

-- name: GetOrCreateExercise :one
INSERT INTO exercise (name, user_id)
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name;

-- name: CreateSet :one
INSERT INTO "set" (exercise_id, workout_id, weight, reps, set_type, user_id, exercise_order, set_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order;

-- name: UpdateWorkout :one
UPDATE workout
SET
    date = COALESCE($2, date),
    notes = COALESCE($3, notes),
    workout_focus = COALESCE($4, workout_focus),
    updated_at = NOW()
WHERE id = $1 AND user_id = $5
RETURNING id, date, notes, workout_focus, created_at, updated_at;

-- name: UpdateSet :one
UPDATE "set"
SET
    weight = COALESCE($2, weight),
    reps = COALESCE($3, reps),
    set_type = COALESCE($4, set_type),
    updated_at = NOW()
WHERE id = $1 AND user_id = $5
RETURNING id, exercise_id, workout_id, weight, reps, set_type, created_at, updated_at, exercise_order, set_order;

-- name: CreateUser :one
INSERT INTO users (user_id)
VALUES ($1)
RETURNING id, user_id, created_at;
```

### 2. Regenerate Database Code
```bash
cd server && make sqlc
```

### 3. Update Frontend Types
```bash
cd client && bun run openapi-ts
```

### 4. Additional Steps

1. **Run tests** to ensure no backend code depends on removed fields
2. **Check handlers** for any direct field access that might break
3. **Verify Swagger docs** still generate correctly
4. **Test frontend** to ensure all pages still work
5. **Run linting/type checking**:
   ```bash
   cd server && go vet ./...
   cd client && bun run typecheck
   ```

## Security Benefits

- Removes `user_id` from API responses (prevents potential data leakage)
- Reduces payload size by removing unnecessary timestamp fields
- Makes it explicit which fields are actually needed vs. accidentally exposed

## Questions for Review

1. Should we remove `notes` and `updated_at` from `ListWorkouts` since they're not displayed?
2. Are the User queries (`GetUser`, `GetUserByUserID`) actually needed by the backend?
3. Should we keep `created_at`/`updated_at` fields for any auditing purposes?

## Next Agent Instructions

1. Make the SQL changes in `server/query.sql` as specified above
2. Run `make sqlc` in the server directory
3. Run `bun run openapi-ts` in the client directory
4. Test the changes and fix any breaking issues
5. Run linting and type checking as specified