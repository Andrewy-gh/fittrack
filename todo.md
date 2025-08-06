## Updated RLS Implementation Plan

### 1. Database Migration (Goose)

- [x] Create a new migration file for RLS setup
- [x] Add RLS and policies to the migration file
- [x] Test the migration locally **NOTE** tests are passing but double check to see they are properly implemented
- [x] Document the migration for other developers

### 2. SQLC Configuration

- [x] Update `sqlc.yaml` to ignore the new `app.current_user_id` setting.
- [ ] Regenerate SQLC code to reflect any schema changes.
    ```bash
    sqlc generate
    ```

### 3. Database Connection Layer

- [x] **Middleware:** The `server/internal/auth/middleware.go` file already contains the necessary logic to set the `app.current_user_id` for each request. This is being handled in the `Authenticator.Middleware` function.
- [ ] **Connection Pool:** The `pgxpool` is created in `server/cmd/api/main.go` and passed to the `Authenticator`. The existing implementation correctly handles setting the user ID on a per-request basis, which is appropriate for a connection pool.
- [ ] **Context Propagation:** The user ID is extracted from the JWT and is available in the middleware. The `user.WithContext` function is already being used to add the user ID to the request context.

### 4. Testing

- [ ] **Unit Tests:**
    - Create a new test file `server/internal/auth/middleware_test.go` if it doesn't exist.
    - Add a test case to `middleware_test.go` to verify that the `app.current_user_id` is set correctly in the database session. This can be done by using a mock database connection that records the SQL queries being executed.
- [ ] **Integration Tests:**
    - Create a new test file `server/internal/workout/handler_test.go` if it doesn't exist.
    - Add integration tests to `handler_test.go` to cover multi-user scenarios:
        - **Scenario 1:** User A creates a workout. Verify that User A can retrieve it.
        - **Scenario 2:** User B attempts to retrieve User A's workout. Verify that the request is denied or returns an empty result set.
        - **Scenario 3:** An anonymous user (no token) attempts to access any workout data. Verify that the request is denied.
- [ ] **End-to-end API tests:**
    - These tests will be performed manually or with a separate E2E testing suite. The focus will be on verifying the RLS policies through actual API calls using different user credentials.

### 5. Documentation

- [ ] Document the RLS implementation in the project's `README.md` or a new `docs/rls.md` file.
- [ ] Update the API documentation to reflect the new authentication and authorization requirements.
- [ ] Create a rollback plan that includes the `goose down` command to revert the RLS migration.
