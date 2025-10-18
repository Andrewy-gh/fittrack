# Edit Exercise Name Feature Plan

## Overview
Add edit exercise name functionality via dialog (matches delete pattern). PATCH endpoint. Handle uniqueness constraint per user.

## Frontend

### 1. Create Dialog Component
**File**: `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`
- Single text input (name)
- Validation: required, trimmed, max 100 chars
- Handle 409 error → "You already have an exercise named 'X'"
- Pattern after `exercise-delete-dialog.tsx`

### 2. Wire to UI
**File**: `client/src/components/exercises/exercise-detail.tsx:85-88`
- Add dialog state + open handler
- Pass `exerciseId` + `exerciseName` to dialog
- Pattern after delete button (lines 89-98)

### 3. Add Mutation Hook
**File**: `client/src/lib/api/exercises.ts`
- `useUpdateExerciseMutation()` (~15 lines)
- Invalidate: exercises list + detail queries
- Pattern after `useDeleteExerciseMutation()` (lines 39-51)

### 4. Demo Mode - Storage
**File**: `client/src/lib/demo-data/storage.ts`
- Add `updateExercise(id: number, name: string): boolean` (~20 lines)
- Check uniqueness, throw error if duplicate
- Update exercise name + `updated_at`
- Pattern after `deleteExercise()` (lines 123-136)

### 5. Demo Mode - Mutation
**File**: `client/src/lib/demo-data/query-options.ts`
- Add `patchDemoExercisesByIdMutation()` (~20 lines)
- Invalidate exercises list + detail
- Pattern after `deleteDemoExercisesByIdMutation()` (lines 137-157)

## Backend

### 1. SQL Query
**File**: `server/query.sql`
```sql
-- name: UpdateExerciseName :exec
UPDATE exercise
SET name = $2, updated_at = NOW()
WHERE id = $1 AND user_id = $3;
```
Run: `make sqlc` to generate code

### 2. Request Validator
**File**: `server/internal/exercise/validator.go`
- Add `UpdateExerciseNameRequest` struct
- Validate: `Name string` (required, max 256)

### 3. Repository Interface + Impl
**File**: `server/internal/exercise/repository.go`
- Interface: Add `UpdateExerciseName(ctx, id int32, name, userID string) error`
- Impl: Call generated `queries.UpdateExerciseName()` (~15 lines)
- Pattern after `DeleteExercise()` (lines 241-266)

### 4. Service Layer
**File**: `server/internal/exercise/service.go`
- Add `UpdateExerciseName(ctx, id int32, name string) error` (~20 lines)
- Auth check via `user.Current(ctx)`
- Verify ownership: `GetExercise(ctx, id, userID)` → return `ErrNotFound` if missing
- Call `repo.UpdateExerciseName()`
- Pattern after `DeleteExercise()` (lines 117-137)

### 5. Handler
**File**: `server/internal/exercise/handler.go`
- Add `UpdateExerciseName(w, r)` handler (~35 lines)
- Parse path ID + JSON body
- Validate request
- Handle errors: 400/401/404/409/500
- **409 Conflict**: Check constraint violation → "Exercise name already exists"
- Return 204 No Content on success
- Swagger: `@Router /exercises/{id} [patch]`
- Pattern after `DeleteExercise()` (lines 229-258)

### 6. Route
**File**: `server/internal/server/routes.go` (or main router file)
- Add: `PATCH /exercises/{id}` → `handler.UpdateExerciseName`

## Constraints
- **Uniqueness**: `UNIQUE (user_id, name)` at DB (schema.sql:26)
- **Max length**: 256 chars (DB), 100 chars (frontend UX)
- **Trim whitespace**: Yes
- **Demo mode**: Yes (consistency with delete)

## Error Handling
- 400: Invalid ID, validation failed
- 401: Unauthorized
- 404: Exercise not found
- 409: Duplicate name (uniqueness violation)
- 500: Server error

## Testing Notes
- Test uniqueness constraint (rename to existing name)
- Test ownership (user can't edit another user's exercise)
- Test demo mode parity with auth mode
