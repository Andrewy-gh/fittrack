Auth Integration TODO list (Go backend)

### 0. Prep
- [X] Add new env vars
  - `VITE_PROJECT_ID`
  - `VITE_PUBLISHABLE_CLIENT_KEY`
  - `SECRET_SERVER_KEY`
- [X] Expose them to the server process ( `.env`, Docker, CI, etc.).

### 1. Database Migration Strategy
- [ ] **Create `app_user` Table:** The central table for users.
  - `id` (VARCHAR/UUID, PK from Stack Auth)
  - `email` (VARCHAR, nullable)
  - `name` (VARCHAR, nullable)
- [ ] **Add `user_id` to Root Tables:** Add a `user_id` foreign key to tables representing user-owned objects.
  - `workout` table: Add `user_id INT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE`.
  - `exercise` table: Add `user_id INT NOT NULL REFERENCES app_user(id) ON DELETE CASCADE`.
- [ ] **Update `exercise` Unique Constraint:** Change the unique constraint on the `exercise` table from `(name)` to `(user_id, name)` to allow different users to have exercises with the same name.
- [ ] **Do Not Add `user_id` to Child Tables:** The `set` table does not need a `user_id` column, as ownership is inferred from the parent `workout`.
- [ ] **Add Indexes for Performance:** Index the new foreign keys to ensure fast lookups.
  - `CREATE INDEX idx_workout_user_id ON workout(user_id);`
  - `CREATE INDEX idx_exercise_user_id ON exercise(user_id);`
  - *Note:* For queries filtering by user and sorting by date, a composite index like `ON workout(user_id, date)` could be even more performant.
- [ ] **Create and Run Migration Script:** Combine all schema changes into a single migration file and run it.

### 1.5. Update SQL Queries (`server/query.sql`)
- [ ] `GetWorkout`: Add `user_id` parameter to `WHERE` clause.
- [ ] `ListWorkouts`: Add `user_id` parameter to `WHERE` clause.
- [ ] `GetExercise`: Add `user_id` parameter to `WHERE` clause.
- [ ] `ListExercises`: Add `user_id` parameter to `WHERE` clause.
- [ ] `GetExerciseByName`: Add `user_id` parameter to `WHERE` clause.
- [ ] `GetWorkoutWithSets`: Add `user_id` parameter to `WHERE` clause. Resolve duplicate query name.
- [ ] `GetExerciseWithSets`: Add `user_id` parameter to `WHERE` clause.
- [ ] `CreateWorkout`: Add `user_id` parameter to `INSERT` statement.
- [ ] `GetOrCreateExercise`: Add `user_id` and update `ON CONFLICT` to use `(user_id, name)`.
- [ ] `GetSet`: Secure query by joining on `workout` and filtering by `user_id`.
- [ ] `ListSets`: Secure query by joining on `workout` and filtering by `user_id`.

### 2. Go auth middleware
- [ ] File server/middleware/auth.go
  1. Read x-stack-access-token header; abort 401 if missing.
  2. Build request to https://api.stack-auth.com/api/v1/users/me with headers:
    - `x-stack-access-type: server`
    - `x-stack-project-id: $STACK_PROJECT_ID`
    - `x-stack-secret-server-key: $STACK_SECRET_SERVER_KEY`
    - `x-stack-access-token: <token>`
  3. Parse JSON; if status ≠ 200 or missing "id", abort 401.
  4. Look up user ID in database; if not found, insert a new record (step 1).
  5. Attach user struct to context for downstream handlers.

### 3. Protected sample endpoint
- [ ] File `server/handlers/me.go`
  - Route `/api/users/me` (or similar)
  - Wrapped with `AuthMiddleware`
  - Returns `{ "id": "..", "message": "auth OK", "dummy": true }` for now.

### 4. Client update (already started)
- [ ] In `client/src/stack.ts` (or helper) add:
```ts
export async function fetchWithAuth(user: User) {
  const { accessToken } = await user.getAuthJson();
  return fetch('/api/users/me', {
    headers: { 'x-stack-access-token': accessToken },
  });
}
```
- [ ]Consume above helper where needed (e.g. handler.$.tsx).

### 5. Testing
- [ ] Unit-test `AuthMiddleware` (happy path & failure cases).
- [ ] Integration test hitting `/api/users/me` with mock Stack Auth using httptest server.

### 6. Docs / housekeeping
- [ ] Update README.md with:
  - New env vars
  - How to run migrations
  - Auth flow diagram

- [ ] Add TODO for future optimizations:
  - Cache token validation for N seconds to reduce external calls.
  - Role-/scope-based authorization helper.