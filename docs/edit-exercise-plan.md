# Edit Exercise Name Feature Plan

## Overview
PATCH endpoint for updating exercise name. Handle uniqueness constraint per user. TDD approach: tests → backend → frontend.

## Backend Tests

### 1. Service Tests
**File**: `server/internal/exercise/service_test.go`
- Success: valid name update
- Error: exercise not found
- Error: unauthorized (wrong user)
- Error: duplicate name (uniqueness)

### 2. Handler Tests
**File**: `server/internal/exercise/handler_test.go`
- 204: successful update
- 400: invalid ID, validation fails
- 401: unauthorized
- 404: not found
- 409: duplicate name

## Backend Implementation

### 1. SQL Query
**File**: `server/query.sql`
```sql
-- name: UpdateExerciseName :exec
UPDATE exercise
SET name = $2, updated_at = NOW()
WHERE id = $1 AND user_id = $3;
```
Run: `make sqlc`

### 2. Validator
**File**: `server/internal/exercise/validator.go`
- `UpdateExerciseNameRequest` struct
- Validate: required, max 256

### 3. Repository
**File**: `server/internal/exercise/repository.go`
- Interface: `UpdateExerciseName(ctx, id int32, name, userID string) error`
- Impl: call `queries.UpdateExerciseName()`
- Pattern: `DeleteExercise()` (lines 241-266)

### 4. Service
**File**: `server/internal/exercise/service.go`
- `UpdateExerciseName(ctx, id int32, name string) error`
- Auth check, verify ownership, call repo
- Pattern: `DeleteExercise()` (lines 117-137)

### 5. Handler
**File**: `server/internal/exercise/handler.go`
- `UpdateExerciseName(w, r)`
- Parse ID + body, validate, handle 400/401/404/409/500
- 409: check constraint violation → "Exercise name already exists"
- Return 204
- Pattern: `DeleteExercise()` (lines 229-258)

### 6. Route
**File**: `server/internal/server/routes.go`
- `PATCH /exercises/{id}` → `handler.UpdateExerciseName`

## Frontend Tests (Minimal)

### 1. Demo Storage Tests
**File**: `client/src/lib/demo-data/storage.test.ts`
- Success: update name
- Error: duplicate name
- Updates `updated_at`

## Frontend Implementation

### 1. Mutation Hook
**File**: `client/src/lib/api/exercises.ts`
- `useUpdateExerciseMutation()`
- Invalidate: exercises list + detail
- Pattern: `useDeleteExerciseMutation()` (lines 39-51)

### 2. Demo Storage
**File**: `client/src/lib/demo-data/storage.ts`
- `updateExercise(id, name): boolean`
- Check uniqueness, throw if duplicate
- Update name + `updated_at`
- Pattern: `deleteExercise()` (lines 123-136)

### 3. Demo Mutation
**File**: `client/src/lib/demo-data/query-options.ts`
- `patchDemoExercisesByIdMutation()`
- Invalidate exercises list + detail
- Pattern: `deleteDemoExercisesByIdMutation()` (lines 137-157)

### 4. Dialog Component
**File**: `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`
- Single text input
- Validation: required, trimmed, max 100
- Handle 409 → "You already have an exercise named 'X'"
- Pattern: `exercise-delete-dialog.tsx`

### 5. Wire to UI
**File**: `client/src/components/exercises/exercise-detail.tsx:85-88`
- Dialog state + open handler
- Pass `exerciseId` + `exerciseName`
- Pattern: delete button (lines 89-98)

## Constraints
- Uniqueness: `UNIQUE (user_id, name)` at DB
- Max: 256 chars (DB), 100 chars (frontend)
- Trim whitespace
