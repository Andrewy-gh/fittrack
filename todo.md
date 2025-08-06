# Updated RLS Implementation Plan

## 1. Database Migration (Goose)

- [x] Create a new migration file for RLS setup
- [x] Add RLS and policies to the migration file
- [x] Test the migration locally **NOTE** tests are passing but double check to see they are properly implemented
- [x] Document the migration for other developers

### Additional Migration Tasks:

- [x] **Policy Validation:** Verify RLS policies handle edge cases (null user context, missing session variables)
- [x] **Migration Testing:** Test both up and down migrations work correctly in development
- [x] **Admin Access:** Ensure admin users (if any) have appropriate bypass or elevated access

## 2. SQLC Configuration

- [x] **Schema Review:** Verify no sqlc configuration changes are needed (likely none required)
- [x] **Code Regeneration:** Regenerate SQLC code if any schema changes were made:
  ```bash
  sqlc generate
  ```
- [-] **Query Validation:** Ensure existing queries work correctly with RLS enabled

## 3. Database Connection Layer

- [x] **Middleware:** The `server/internal/auth/middleware.go` file already contains the necessary logic to set the `app.current_user_id` for each request
- [ ] **Connection Pool Validation:** Verify `pgxpool` correctly isolates user context between concurrent requests
- [x] **Context Propagation:** User ID extraction from JWT and context propagation is implemented
- [ ] **Session Variable Validation:** Add optional logging/validation to verify `app.current_user_id` is set correctly
- [ ] **Error Handling:** Enhance error handling for RLS-related failures

## 4. Testing

### Unit Tests:

- [ ] **Middleware Tests (`server/internal/auth/middleware_test.go`):**
  - Verify `app.current_user_id` is set correctly in database session
  - Test error scenarios (invalid tokens, missing headers, database failures)
  - Mock database connection to verify SQL queries

### Integration Tests:

- [ ] **Workout Handler Tests (`server/internal/workout/handler_test.go`):**

  - **Scenario 1:** User A creates a workout → User A can retrieve it
  - **Scenario 2:** User B attempts to retrieve User A's workout → Request denied/empty result
  - **Scenario 3:** Anonymous user attempts to access workout data → Request denied
  - **Scenario 4:** Concurrent requests from different users maintain proper isolation

- [ ] **Exercise Handler Tests (`server/internal/exercise/handler_test.go`):**
  - Similar multi-user scenarios for exercise data
  - Verify RLS policies work across all protected resources

### Performance & Security Tests:

- [ ] **Connection Pool Testing:** Test that concurrent requests from different users maintain proper context isolation
- [ ] **Performance Impact:** Benchmark the performance impact of `set_config()` on each request
- [ ] **Policy Bypass Testing:** Verify RLS policies can't be bypassed through direct database access
- [ ] **Session Variable Edge Cases:** Test behavior when session variables are missing or malformed

### End-to-end API Tests:

- [ ] Manual or automated E2E tests using different user credentials
- [ ] Test complete request lifecycle with authentication and authorization

## 5. Error Handling & Monitoring

- [ ] **RLS-Specific Logging:** Add specific logging for RLS policy violations to aid debugging
- [ ] **Error Messages:** Ensure error messages don't leak sensitive information
- [ ] **Monitoring:** Consider adding metrics for RLS-related operations

## 6. Documentation

- [ ] **Implementation Documentation:** Document the RLS implementation in `README.md` or create `docs/rls.md`
- [ ] **API Documentation:** Update API docs to reflect authentication and authorization requirements
- [ ] **Developer Guide:** Document how RLS affects local development and testing
- [ ] **Rollback Plan:** Create rollback procedures including `goose down` commands
- [ ] **Troubleshooting Guide:** Document common RLS-related issues and solutions

## 7. Deployment Considerations

- [ ] **Environment Parity:** Verify RLS works identically in development and production
- [ ] **Migration Coordination:** Plan for zero-downtime deployment if needed
- [ ] **Database Permissions:** Verify application database user has necessary permissions for RLS operations
- [ ] **Connection Pool Configuration:** Review pool settings for RLS compatibility

## 8. Security Review

- [ ] **Policy Coverage:** Ensure all sensitive tables have appropriate RLS policies
- [ ] **Privilege Escalation:** Test that users cannot escalate privileges through RLS bypasses
- [ ] **Token Validation:** Verify JWT validation is robust and secure
- [ ] **Session Security:** Ensure session variables can't be manipulated by malicious requests

## Notes

- The existing middleware implementation in `server/internal/auth/middleware.go` correctly sets user context per request
- Connection pooling is properly handled with per-request session variable setting
- Consider performance implications of `set_config()` on every request - monitor in production
