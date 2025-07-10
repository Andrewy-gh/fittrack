Auth Integration TODO list (Go backend)

### 0. Prep
- [X] Add new env vars
  - `VITE_PROJECT_ID`
  - `VITE_PUBLISHABLE_CLIENT_KEY`
  - `SECRET_SERVER_KEY`
- [X] Expose them to the server process ( `.env`, Docker, CI, etc.).

### 1. Database
- [ ] Create migration `202507091_auth_user_table.sql`
- [ ] Table users
  - id (varchar / uuid, PK, from Stack Auth)
  - email (varchar, nullable)
  - name (varchar, nullable)
  - created_at, updated_at
- [ ] Run migration locally & in CI.

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