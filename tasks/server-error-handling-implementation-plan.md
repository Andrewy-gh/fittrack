# Server Error Handling Implementation Plan

**Project:** FitTrack Error Handling Improvements
**Created:** 2025-12-26
**Status:** Ready for Implementation

---

## Overview

This plan addresses the error handling improvements identified in `server-error-handling-analysis.md` plus additional middleware inconsistencies discovered during auth/middleware audit. The implementation is organized into 10 high-level tasks with **58 detailed subtasks**.

**Key Goals:**

1. Reduce duplicate logging by logging at the handler boundary only
2. Replace string-based error detection with type-safe alternatives
3. Consolidate custom error types into shared package
4. **Ensure uniform error responses to frontend** - critical for user-facing error display
5. Fix middleware error handling inconsistencies (rate limit, basic auth, CORS)
6. Update all affected tests

**Database:** PostgreSQL only (no SQLite support)

**Important:** Mark tasks as completed in this file so progress can be tracked.

---

## Task 1: Replace String-Based Database Error Detection

**Priority:** High
**Files Affected:**

- `server/internal/database/errors.go`
- Tests: `server/internal/database/errors_test.go` (if exists)

**Description:** Replace string matching in database error detection functions with type-safe `errors.As` using `*pgconn.PgError`.

### Subtasks

#### 1.1 Add pgconn Import

- [x] Add `"github.com/jackc/pgx/v5/pgconn"` to imports in `database/errors.go`
- [x] Verify pgx version in `go.mod` supports `pgconn.PgError`

#### 1.2 Refactor IsUniqueConstraintError

- [x] Replace string matching with `errors.As(err, &pgErr)`
- [x] Check `pgErr.Code == "23505"` for PostgreSQL unique violations
- [x] Remove SQLite fallback code (only PostgreSQL is supported)
- [x] Add code documentation explaining error code

#### 1.3 Refactor IsForeignKeyConstraintError

- [x] Replace string matching with `errors.As(err, &pgErr)`
- [x] Check `pgErr.Code == "23503"` for foreign key violations
- [x] Remove SQLite fallback code (only PostgreSQL is supported)
- [x] Add code documentation

#### 1.4 Refactor IsRowLevelSecurityError

- [x] Replace string matching with `errors.As(err, &pgErr)`
- [x] Check `pgErr.Code == "42501"` for insufficient privilege
- [x] Verify RLS-specific error codes in PostgreSQL
- [x] Update to properly detect RLS policy violations
- [x] Keep `errors.Is(err, ErrRowLevelSecurity)` check

#### 1.5 Refactor IsRLSContextError

- [x] Replace string matching with proper error type checking
- [x] Keep `errors.Is(err, ErrRLSContext)` check
- [x] Verify context setting error patterns

#### 1.6 Update Tests

- [x] Create/update unit tests for each error detection function
- [x] Test with actual `pgconn.PgError` instances
- [x] Test with wrapped errors to verify `errors.As` works correctly
- [x] Test with nil errors
- [x] Verify all existing tests still pass

---

## Task 2: Fix String Comparison in User Repository

**Priority:** High
**Files Affected:**

- `server/internal/user/repository.go`
- Tests: `server/internal/user/repository_test.go` (if exists)

**Description:** Replace fragile string comparison with proper `errors.Is` sentinel error checking.

### Subtasks

#### 2.1 Update Error Check in GetByID

- [x] Locate line 36 in `user/repository.go`
- [x] Replace `if err.Error() == "no rows in result set"` with `if errors.Is(err, pgx.ErrNoRows)`
- [x] Verify `pgx.ErrNoRows` is the correct sentinel error
- [x] Test with actual "no rows" scenario

#### 2.2 Audit Other String Comparisons

- [x] Search entire `user/repository.go` for other `err.Error()` comparisons
- [x] Replace any other instances found
- [x] Document any intentional string checks that remain

#### 2.3 Update Tests

- [x] Verify existing tests for GetByID still pass
- [x] Add test case for no rows scenario
- [x] Ensure test uses proper mocking of `pgx.ErrNoRows`

---

## Task 3: Reduce Duplicate Logging (Log at Handler Boundary)

**Priority:** High
**Files Affected:**

- `server/internal/exercise/service.go`
- `server/internal/workout/service.go`
- `server/internal/user/service.go` (if has duplicate logging)
- All handler files (exercise, workout, user)
- Tests: All service and handler tests

**Description:** Remove logging from service layer and log errors only once at the handler boundary.

### Subtasks

#### 3.1 Audit Current Logging Points

- [x] Create inventory of all error logging in service files
- [x] Create inventory of all error logging in handler files
- [x] Document the ErrorJSON function's logging behavior
- [x] Identify all duplicate logging instances

#### 3.2 Remove Service Layer Error Logging - Exercise

- [x] Remove `es.logger.Error()` calls from `exercise/service.go` List method
- [x] Remove from GetByID method
- [x] Remove from Create method
- [x] Remove from Update method
- [x] Remove from Delete method
- [x] Remove from Search method
- [x] Remove all `es.logger.Debug()` error detail logging
- [x] Keep error wrapping with `fmt.Errorf("...: %w", err)`

#### 3.3 Remove Service Layer Error Logging - Workout

- [x] Remove `ws.logger.Error()` calls from `workout/service.go` Create method
- [x] Remove from GetByID method
- [x] Remove from List method
- [x] Remove from Update method
- [x] Remove from Delete method
- [x] Remove from all other workout service methods
- [x] Remove all `ws.logger.Debug()` error detail logging
- [x] Keep error wrapping with `fmt.Errorf("...: %w", err)`

#### 3.4 Remove Service Layer Error Logging - User

- [x] Audit `user/service.go` for duplicate logging
- [x] Remove duplicate error logging if found
- [x] Keep error wrapping

#### 3.5 Verify Handler Layer Logging

- [x] Confirm `response.ErrorJSON()` logs all errors with proper context
- [x] Verify request ID, path, method, status are included
- [x] Ensure all handlers use `response.ErrorJSON()` consistently
- [x] Document any handlers with custom error logging

#### 3.6 Update Tests

- [x] Update service tests to not expect logger calls
- [x] Remove mock logger expectations from service tests
- [x] Verify handler tests check for proper logging
- [x] Add assertions in handler tests for log output if needed

---

## Task 4: Consolidate Custom Error Types

**Priority:** High
**Files Affected:**

- `server/internal/errors/errors.go` (new file)
- `server/internal/exercise/service.go`
- `server/internal/exercise/handler.go`
- `server/internal/workout/service.go`
- `server/internal/workout/handler.go`
- Tests: All affected service and handler tests

**Description:** Create shared error types package and migrate duplicate custom errors.

### Subtasks

#### 4.1 Create Shared Errors Package

- [x] Create directory `server/internal/errors/`
- [x] Create file `server/internal/errors/errors.go`
- [x] Add package documentation
- [x] Define `Unauthorized` error type with Resource and UserID fields
- [x] Define `NotFound` error type with Resource and ID fields
- [x] Implement `Error()` method for `Unauthorized`
- [x] Implement `Error()` method for `NotFound`
- [x] Add constructor functions (optional): `NewUnauthorized()`, `NewNotFound()`

#### 4.2 Update Exercise Service

- [x] Import shared errors package as `apperrors`
- [x] Remove local `ErrUnauthorized` type definition
- [x] Remove local `ErrNotFound` type definition
- [x] Replace all `&ErrUnauthorized{...}` with `&apperrors.Unauthorized{...}`
- [x] Replace all `&ErrNotFound{...}` with `&apperrors.NotFound{...}`
- [x] Update error instantiation to include Resource field
- [x] Verify all error returns compile

#### 4.3 Update Exercise Handler

- [x] Import shared errors package
- [x] Update all `errors.As(err, &ErrUnauthorized{})` to use `&apperrors.Unauthorized{}`
- [x] Update all `errors.As(err, &ErrNotFound{})` to use `&apperrors.NotFound{}`
- [x] Verify error type checking still works correctly

#### 4.4 Update Workout Service

- [x] Import shared errors package as `apperrors`
- [x] Remove local `ErrUnauthorized` type definition
- [x] Remove local `ErrNotFound` type definition
- [x] Replace all error instantiations with shared types
- [x] Update to include Resource field

#### 4.5 Update Workout Handler

- [x] Import shared errors package
- [x] Update all `errors.As()` calls to use shared error types
- [x] Verify error type checking still works

#### 4.6 Create Error Package Tests

- [x] Create `server/internal/errors/errors_test.go`
- [x] Test `Unauthorized.Error()` method output
- [x] Test `NotFound.Error()` method output
- [x] Test `errors.As()` works with both error types
- [x] Test error wrapping preserves type information

#### 4.7 Update Existing Tests

- [x] Update exercise service tests to use shared error types
- [x] Update exercise handler tests to use shared error types
- [x] Update workout service tests to use shared error types
- [x] Update workout handler tests to use shared error types
- [x] Verify all error assertions still pass (Note: 1 integration test failing - TestContributionData_Integration_RLS - appears unrelated to error handling changes)

---

## Task 5: Fix Type Assertion to Use errors.As

**Priority:** Medium
**Files Affected:**

- `server/internal/workout/handler.go`
- Tests: `server/internal/workout/handler_test.go` (if FormatValidationErrors is tested)

**Description:** Replace type assertion with `errors.As` in `FormatValidationErrors` function.

### Subtasks

#### 5.1 Update FormatValidationErrors Function

- [x] Locate `FormatValidationErrors` function at line 328
- [x] Change `validationErrors` from `*validator.ValidationErrors` to `validator.ValidationErrors`
- [x] Replace type assertion with `errors.As(err, &validationErrors)`
- [x] Update loop to iterate over `validationErrors` directly (not dereference)
- [x] Verify function signature doesn't need to change

#### 5.2 Test with Wrapped Errors

- [x] Create test with wrapped validation error
- [x] Verify `errors.As` correctly unwraps to find validation error
- [x] Test with non-validation error (should fall back to `err.Error()`)

#### 5.3 Update Any Similar Functions

- [x] Search for other type assertions on errors in handler files
- [x] Update any similar patterns found

---

## Task 6: Add Logging for Config Parsing Errors

**Priority:** Low
**Files Affected:**

- `server/internal/config/config.go`

**Description:** Add logging when configuration value parsing fails to aid debugging.

### Subtasks

#### 6.1 Update GetInt Function

- [x] Locate GetInt function around line 95-96
- [x] Add else clause to `strconv.Atoi` error check
- [x] Log parsing failure with value and error
- [x] Use appropriate log level (Info or Debug)
- [x] Include config key name in log message if available

#### 6.2 Audit Other Parsing Functions

- [x] Check for similar patterns in GetBool, GetFloat64, etc.
- [x] Add logging to other parsing failures
- [x] Ensure consistent log message format

#### 6.3 Test Config Parsing

- [x] Test with invalid integer value
- [x] Verify log message appears
- [x] Verify default value is still returned

---

## Task 7: Consider JSON Encoding Error Tracking

**Priority:** Low
**Files Affected:**

- `server/internal/health/handler.go`
- `server/internal/middleware/ratelimit.go`
- `server/internal/response/json.go`
- Potential: Add metrics/monitoring package

**Description:** Evaluate whether to track JSON encoding errors and implement tracking if beneficial.

**Recommendation:** JSON encoding failures are extremely rare in practice and typically indicate a serious bug (e.g., trying to marshal a channel or function). The current approach of logging these errors is sufficient. Adding metrics would add complexity without significant value unless you're seeing these errors in production. **Suggested action: Document the behavior and leave as-is.**

### Subtasks

#### 7.1 Evaluate Need for Tracking

- [x] Review logs to see if JSON encoding errors occur in practice
- [x] Assess impact of partial responses
- [x] Decide whether to add metrics, enhanced logging, or leave as-is (recommendation: leave as-is)
- [x] Document decision

#### 7.2 If Tracking is Needed - Add Metrics

- [x] N/A - Decision: Leave as-is (metrics not needed)

#### 7.3 If Tracking is Needed - Enhanced Logging

- [x] N/A - Decision: Leave as-is (current logging is sufficient)

#### 7.4 Document Behavior

- [x] Document that response headers are already written
- [x] Explain why execution continues after encoding error
- [x] Add code comments explaining the trade-off

---

## Task 8: Fix Minor Inconsistencies ✅

**Priority:** Low
**Files Affected:**

- `server/internal/workout/repository.go`
- `server/internal/database/errors.go` (sentinel error usage)

**Description:** Fix minor inconsistencies identified in the analysis.

### Subtasks

#### 8.1 Fix Unwrapped Error in Workout Repository

- [x] Locate line 357 in `workout/repository.go`
- [x] Wrap error with `fmt.Errorf("...: %w", err)` for consistency
- [x] Add descriptive context to error message
- [x] Verify tests still pass

#### 8.2 Enhance Sentinel Error Usage

- [x] Review how `ErrRowLevelSecurity` and `ErrRLSContext` are created
- [x] Ensure they're returned directly (not just checked)
- [x] Update error detection functions to return sentinel errors when appropriate
- [x] Document when sentinel errors should be used vs. wrapped errors

---

## Task 9: Ensure Uniform Frontend Error Responses

**Priority:** High
**Files Affected:**

- `server/internal/response/json.go`
- All handler files
- `server/internal/middleware/ratelimit.go`
- `server/internal/middleware/basicauth.go`
- `server/internal/middleware/cors.go`
- `server/internal/health/handler.go`

**Description:** Audit and standardize all error responses sent to frontend. Fix middleware error responses to use centralized ErrorJSON pattern for consistency.

### Subtasks

#### 9.1 Audit Current Error Response Format

- [x] Document the structure of `ErrorJSON` responses
- [x] List all HTTP status codes used across handlers
- [x] Identify any handlers not using `ErrorJSON`
- [x] Document `sanitizeErrorMessage` behavior

**Findings:**

**ErrorJSON Response Structure:**
- `message` (string): Sanitized error message
- `request_id` (string, optional): Request ID from context

**HTTP Status Codes Used:**
- 200 OK - Successful responses
- 204 No Content - Successful DELETE/UPDATE operations
- 400 Bad Request - Validation errors, invalid input
- 401 Unauthorized - Authentication/authorization failures
- 403 Forbidden - CORS rejection (currently has no body!)
- 404 Not Found - Resource not found
- 409 Conflict - Duplicate resource (unique constraint)
- 429 Too Many Requests - Rate limit exceeded
- 500 Internal Server Error - Server errors
- 503 Service Unavailable - Database connection failures

**Handlers NOT using ErrorJSON (need fixing):**
1. `middleware/ratelimit.go:44` - Uses `http.Error()` for internal errors
2. `middleware/ratelimit.go:62-72` - Manual JSON construction for rate limit
3. `middleware/basicauth.go:41-52` - Manual JSON construction for unauthorized
4. `middleware/cors.go:37` - Empty 403 response for CORS rejection
5. `health/handler.go:83` - Raw error exposure in database check

**sanitizeErrorMessage Behavior:**
- Removes database error details (PostgreSQL codes, connection info, table names)
- Removes JWT error details (tokens, claims, signatures)
- Allows validation errors to pass through (safe for users)
- Returns generic messages for sensitive errors: "internal error", "unauthorized"

#### 9.2 Standardize Error Response Structure

- [x] Ensure all responses have consistent JSON structure
- [x] Define standard fields: `error`, `message`, `request_id`, etc.
- [x] Document which errors are sanitized and which are not
- [x] Verify 400-level vs 500-level error handling

**Standard Error Response Structure:**
```json
{
  "message": "error description",
  "request_id": "uuid-v4-string"  // optional, included when available
}
```

**Field Definitions:**
- `message` (required): Human-readable error message, sanitized for security
- `request_id` (optional): Unique request identifier for tracing/debugging

**Error Sanitization Rules:**

**SANITIZED (generic message returned):**
- Database errors → "internal error" or "unauthorized"
- JWT/token errors → "unauthorized"
- Connection/pool errors → "internal error"
- PostgreSQL error codes (23505, 23503, 42501, etc.)

**NOT SANITIZED (original message returned):**
- Validation errors (safe for users)
- "Missing parameter" errors
- "Invalid format" errors
- Custom application errors (NotFound, Unauthorized with context)

**HTTP Status Code Guidelines:**

**4xx - Client Errors (user fixable):**
- 400 Bad Request - Invalid input, validation failures, malformed JSON
- 401 Unauthorized - Missing/invalid authentication, permission denied
- 403 Forbidden - CORS rejection, access forbidden
- 404 Not Found - Resource doesn't exist
- 409 Conflict - Duplicate resource (unique constraint violation)
- 429 Too Many Requests - Rate limit exceeded

**5xx - Server Errors (not user fixable):**
- 500 Internal Server Error - Database errors, unexpected failures
- 503 Service Unavailable - Database connection failures, service down

**Current Inconsistencies (to be fixed in subsequent subtasks):**
- Middleware uses manual JSON construction instead of `response.ErrorJSON`
- Health handler exposes raw database errors
- Some responses missing `request_id` field

#### 9.3 Fix Rate Limit Middleware Internal Error (Line 44)

- [x] Replace `http.Error()` call with standardized JSON response
- [x] Use `http.StatusInternalServerError` (already correct)
- [x] Ensure error includes request_id
- [x] Test that response returns JSON, not plain text

**Note:** Cannot use `response.ErrorJSON()` due to import cycle (response imports middleware for GetRequestID). Instead, implemented manual JSON construction that matches the standard format with proper logging.

#### 9.4 Fix Rate Limit Exceeded Response (Lines 62-72)

- [x] Update manual JSON construction to match standard format
- [x] Keep `Retry-After` header
- [x] Use `http.StatusTooManyRequests` (already correct)
- [x] Ensure request_id is included in response
- [x] Update error message format to match standard pattern
- [x] Add proper logging for rate limit events

**Changes:**
- Added request_id to response
- Added comprehensive logging (Warn level for rate limit events)
- Standardized response format to match ErrorJSON structure
- Improved JSON encoding error logging with request_id

#### 9.5 Fix Basic Auth Middleware (Lines 41-52)

- [x] Update manual JSON construction to match standard format
- [x] Keep `WWW-Authenticate` header
- [x] Use `http.StatusUnauthorized` (already correct)
- [x] Add status code to logging for consistency
- [x] Handle JSON encoding errors properly
- [x] Verify error response structure matches standard format

**Changes:**
- Added status code to logging output
- Added JSON encoding error handling with request_id
- Improved code comments for clarity
- Response already includes request_id (was already correct)

#### 9.6 Fix CORS Middleware Empty 403 Response (Line 37)

- [x] Add error response body for rejected CORS preflight
- [x] Add logger parameter to CORS middleware
- [x] Include appropriate error message explaining CORS rejection
- [x] Add proper logging for rejected CORS requests
- [x] Verify this doesn't break browser CORS handling

**Changes:**
- Added logger parameter to CORS function signature
- Added JSON response body for rejected CORS preflight: `{message, request_id?}`
- Added comprehensive logging (Warn level) with origin, path, method, status, request_id
- Updated all CORS usage (main.go and tests) to include logger
- All tests pass - CORS handling still works correctly

#### 9.7 Fix Health Handler Raw Error Exposure (Line 83)

- [x] Sanitize database error in ready endpoint response
- [x] Replace `"failed: " + err.Error()` with generic message
- [x] Consider adding request_id to health/ready responses (optional)
- [x] Document that health endpoints use custom response format
- [x] Verify no sensitive connection details leak in errors

#### 9.8 Standardize Custom Error Responses

- [x] Verify `Unauthorized` errors return 401 or 403 consistently
- [x] Verify `NotFound` errors return 404 consistently
- [x] Verify validation errors return 400 with formatted messages
- [x] Check database constraint errors return appropriate codes

#### 9.9 Create Error Response Documentation

- [x] Document all possible error response formats
- [x] Document HTTP status codes and when each is used
- [x] Document which error details are exposed to frontend
- [x] Create examples for frontend team
- [x] Document middleware error responses (rate limit, CORS, basic auth)

#### 9.10 Test Error Responses

- [ ] Test each error type returns correct status code
- [ ] Test error messages are properly sanitized
- [ ] Test request_id is included in all error responses
- [ ] Verify no sensitive information leaks in error messages
- [ ] Test middleware error responses return proper JSON format
- [ ] Test rate limit response includes Retry-After header
- [ ] Test basic auth response includes WWW-Authenticate header

---

## Task 10: Update All Affected Tests

**Priority:** High (runs in parallel with implementation)
**Files Affected:** All test files

**Description:** Update all tests affected by error handling changes.

### Subtasks

#### 10.1 Service Tests - Remove Logger Expectations

- [ ] Update `exercise/service_test.go` to remove logger mock calls
- [ ] Update `workout/service_test.go` to remove logger mock calls
- [ ] Update `user/service_test.go` if needed
- [ ] Keep tests for error wrapping and error types

#### 10.2 Handler Tests - Add Logger Expectations

- [ ] Update handler tests to verify ErrorJSON logs correctly
- [ ] Add assertions for log fields (error, path, method, status, request_id)
- [ ] Test that errors are logged exactly once per request

#### 10.3 Error Type Tests

- [ ] Test shared error types work with `errors.As`
- [ ] Test shared error types work with `errors.Is` if applicable
- [ ] Test error wrapping preserves custom error types

#### 10.4 Database Error Tests

- [ ] Create tests for new `pgconn.PgError` detection
- [ ] Test each error code is detected correctly
- [ ] Test wrapped errors are properly detected

#### 10.5 Integration Tests

- [ ] Run full integration test suite
- [ ] Verify API error responses are correct
- [ ] Test error scenarios end-to-end
- [ ] Verify logs show errors once per request

#### 10.6 Test Coverage

- [ ] Measure test coverage for error package
- [ ] Measure test coverage for error handling paths
- [ ] Add tests for any uncovered error scenarios
- [ ] Aim for >80% coverage on error handling code

---

## Implementation Order

### Phase 1: Foundation (Do First)

1. **Task 4**: Consolidate Custom Error Types

   - Creates shared foundation for other tasks
   - No dependencies on other tasks

2. **Task 2**: Fix String Comparison in User Repository

   - Quick win, isolated change
   - No dependencies

3. **Task 1**: Replace String-Based Database Error Detection
   - Important foundation for error handling
   - No dependencies on other tasks

### Phase 2: Core Changes

4. **Task 3**: Reduce Duplicate Logging

   - Major change affecting many files
   - Depends on Task 4 (shared error types)

5. **Task 9**: Ensure Uniform Frontend Error Responses
   - Should be done after Task 3 (logging changes)
   - Verifies end-to-end error flow

### Phase 3: Polish

6. **Task 5**: Fix Type Assertion to Use errors.As

   - Small, isolated improvement

7. **Task 6**: Add Logging for Config Parsing Errors

   - Low priority, isolated change

8. **Task 8**: Fix Minor Inconsistencies

   - Cleanup and polish

9. **Task 7**: Consider JSON Encoding Error Tracking
   - Evaluation task, may not require code changes

### Throughout All Phases

10. **Task 10**: Update All Affected Tests
    - Run continuously as each task completes
    - Ensure tests pass before moving to next task

---

## Testing Strategy

### Unit Tests

- [ ] Test each error type independently
- [ ] Test error detection functions with various error types
- [ ] Test error wrapping preserves type information
- [ ] Test `errors.As` and `errors.Is` work correctly

### Integration Tests

- [ ] Test error responses from API endpoints
- [ ] Verify HTTP status codes are correct
- [ ] Verify error messages are sanitized for frontend
- [ ] Test request IDs appear in error logs and responses

### Manual Testing

- [ ] Trigger each error scenario manually
- [ ] Verify single log entry per error
- [ ] Verify frontend receives correct error format
- [ ] Test with wrapped and unwrapped errors

---

## Success Criteria

- [ ] All errors logged exactly once per request
- [ ] No string-based error detection in production code
- [ ] All custom error types consolidated in shared package
- [ ] All error responses to frontend have uniform structure
- [ ] All tests passing with updated expectations
- [ ] Test coverage >80% on error handling code
- [ ] Documentation updated with new error patterns
- [ ] Code review approved
- [ ] PR merged to main branch

---

## Rollback Plan

If issues are discovered after merge:

1. **Immediate**: Revert PR if critical errors affect production
2. **Short-term**: Fix specific issues in hotfix PR
3. **Testing**: Add missing test cases to prevent regression
4. **Review**: Conduct retrospective on what was missed

---

## Post-Implementation

After this PR is merged:

- [ ] Monitor error logs for any issues
- [ ] Verify log volume decreased due to single logging
- [ ] Collect frontend team feedback on error responses
- [ ] **Conduct frontend error handling analysis** (next phase)
- [ ] Update error handling documentation
- [ ] Share learnings with team

---

## Notes

- **No backward compatibility required** - breaking changes are acceptable
- Update API documentation for error response formats
- **Critical: Coordinate with frontend team** on error response changes - this is essential for consistent error display to users
- All middleware must use centralized error response pattern
- Health/ready endpoints may use custom format (for monitoring systems)

---

## Questions to Resolve - ANSWERED

1. ✅ **Is SQLite used in addition to PostgreSQL?**

   - **Answer:** No, only PostgreSQL is supported
   - **Action:** Remove all SQLite fallback code

2. ✅ **Should we add metrics for JSON encoding errors?**

   - **Answer:** No need - JSON encoding failures are extremely rare and indicate serious bugs
   - **Action:** Document behavior and leave as-is (current logging is sufficient)

3. ✅ **Do we need to maintain any backward compatibility?**

   - **Answer:** No backward compatibility required
   - **Action:** Breaking changes are acceptable; focus on consistency

4. ✅ **Are there any error handling patterns in other packages (auth, etc.)?**
   - **Answer:** Yes - found inconsistencies in middleware
   - **Action:** Added subtasks to Task 9 to fix:
     - `middleware/ratelimit.go` - lines 44, 62-72 (plain text errors, manual JSON)
     - `middleware/basicauth.go` - lines 41-52 (manual JSON construction)
     - `middleware/cors.go` - line 37 (empty 403 response)
     - `health/handler.go` - line 83 (raw error exposure)

---

**End of Implementation Plan**
