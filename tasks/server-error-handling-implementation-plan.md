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

---

## Task 1: Replace String-Based Database Error Detection

**Priority:** High
**Files Affected:**
- `server/internal/database/errors.go`
- Tests: `server/internal/database/errors_test.go` (if exists)

**Description:** Replace string matching in database error detection functions with type-safe `errors.As` using `*pgconn.PgError`.

### Subtasks

#### 1.1 Add pgconn Import
- [ ] Add `"github.com/jackc/pgx/v5/pgconn"` to imports in `database/errors.go`
- [ ] Verify pgx version in `go.mod` supports `pgconn.PgError`

#### 1.2 Refactor IsUniqueConstraintError
- [ ] Replace string matching with `errors.As(err, &pgErr)`
- [ ] Check `pgErr.Code == "23505"` for PostgreSQL unique violations
- [ ] Remove SQLite fallback code (only PostgreSQL is supported)
- [ ] Add code documentation explaining error code

#### 1.3 Refactor IsForeignKeyConstraintError
- [ ] Replace string matching with `errors.As(err, &pgErr)`
- [ ] Check `pgErr.Code == "23503"` for foreign key violations
- [ ] Remove SQLite fallback code (only PostgreSQL is supported)
- [ ] Add code documentation

#### 1.4 Refactor IsRowLevelSecurityError
- [ ] Replace string matching with `errors.As(err, &pgErr)`
- [ ] Check `pgErr.Code == "42501"` for insufficient privilege
- [ ] Verify RLS-specific error codes in PostgreSQL
- [ ] Update to properly detect RLS policy violations
- [ ] Keep `errors.Is(err, ErrRowLevelSecurity)` check

#### 1.5 Refactor IsRLSContextError
- [ ] Replace string matching with proper error type checking
- [ ] Keep `errors.Is(err, ErrRLSContext)` check
- [ ] Verify context setting error patterns

#### 1.6 Update Tests
- [ ] Create/update unit tests for each error detection function
- [ ] Test with actual `pgconn.PgError` instances
- [ ] Test with wrapped errors to verify `errors.As` works correctly
- [ ] Test with nil errors
- [ ] Verify all existing tests still pass

---

## Task 2: Fix String Comparison in User Repository

**Priority:** High
**Files Affected:**
- `server/internal/user/repository.go`
- Tests: `server/internal/user/repository_test.go` (if exists)

**Description:** Replace fragile string comparison with proper `errors.Is` sentinel error checking.

### Subtasks

#### 2.1 Update Error Check in GetByID
- [ ] Locate line 36 in `user/repository.go`
- [ ] Replace `if err.Error() == "no rows in result set"` with `if errors.Is(err, pgx.ErrNoRows)`
- [ ] Verify `pgx.ErrNoRows` is the correct sentinel error
- [ ] Test with actual "no rows" scenario

#### 2.2 Audit Other String Comparisons
- [ ] Search entire `user/repository.go` for other `err.Error()` comparisons
- [ ] Replace any other instances found
- [ ] Document any intentional string checks that remain

#### 2.3 Update Tests
- [ ] Verify existing tests for GetByID still pass
- [ ] Add test case for no rows scenario
- [ ] Ensure test uses proper mocking of `pgx.ErrNoRows`

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
- [ ] Create inventory of all error logging in service files
- [ ] Create inventory of all error logging in handler files
- [ ] Document the ErrorJSON function's logging behavior
- [ ] Identify all duplicate logging instances

#### 3.2 Remove Service Layer Error Logging - Exercise
- [ ] Remove `es.logger.Error()` calls from `exercise/service.go` List method
- [ ] Remove from GetByID method
- [ ] Remove from Create method
- [ ] Remove from Update method
- [ ] Remove from Delete method
- [ ] Remove from Search method
- [ ] Remove all `es.logger.Debug()` error detail logging
- [ ] Keep error wrapping with `fmt.Errorf("...: %w", err)`

#### 3.3 Remove Service Layer Error Logging - Workout
- [ ] Remove `ws.logger.Error()` calls from `workout/service.go` Create method
- [ ] Remove from GetByID method
- [ ] Remove from List method
- [ ] Remove from Update method
- [ ] Remove from Delete method
- [ ] Remove from all other workout service methods
- [ ] Remove all `ws.logger.Debug()` error detail logging
- [ ] Keep error wrapping with `fmt.Errorf("...: %w", err)`

#### 3.4 Remove Service Layer Error Logging - User
- [ ] Audit `user/service.go` for duplicate logging
- [ ] Remove duplicate error logging if found
- [ ] Keep error wrapping

#### 3.5 Verify Handler Layer Logging
- [ ] Confirm `response.ErrorJSON()` logs all errors with proper context
- [ ] Verify request ID, path, method, status are included
- [ ] Ensure all handlers use `response.ErrorJSON()` consistently
- [ ] Document any handlers with custom error logging

#### 3.6 Update Tests
- [ ] Update service tests to not expect logger calls
- [ ] Remove mock logger expectations from service tests
- [ ] Verify handler tests check for proper logging
- [ ] Add assertions in handler tests for log output if needed

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
- [ ] Create directory `server/internal/errors/`
- [ ] Create file `server/internal/errors/errors.go`
- [ ] Add package documentation
- [ ] Define `Unauthorized` error type with Resource and UserID fields
- [ ] Define `NotFound` error type with Resource and ID fields
- [ ] Implement `Error()` method for `Unauthorized`
- [ ] Implement `Error()` method for `NotFound`
- [ ] Add constructor functions (optional): `NewUnauthorized()`, `NewNotFound()`

#### 4.2 Update Exercise Service
- [ ] Import shared errors package as `apperrors`
- [ ] Remove local `ErrUnauthorized` type definition
- [ ] Remove local `ErrNotFound` type definition
- [ ] Replace all `&ErrUnauthorized{...}` with `&apperrors.Unauthorized{...}`
- [ ] Replace all `&ErrNotFound{...}` with `&apperrors.NotFound{...}`
- [ ] Update error instantiation to include Resource field
- [ ] Verify all error returns compile

#### 4.3 Update Exercise Handler
- [ ] Import shared errors package
- [ ] Update all `errors.As(err, &ErrUnauthorized{})` to use `&apperrors.Unauthorized{}`
- [ ] Update all `errors.As(err, &ErrNotFound{})` to use `&apperrors.NotFound{}`
- [ ] Verify error type checking still works correctly

#### 4.4 Update Workout Service
- [ ] Import shared errors package as `apperrors`
- [ ] Remove local `ErrUnauthorized` type definition
- [ ] Remove local `ErrNotFound` type definition
- [ ] Replace all error instantiations with shared types
- [ ] Update to include Resource field

#### 4.5 Update Workout Handler
- [ ] Import shared errors package
- [ ] Update all `errors.As()` calls to use shared error types
- [ ] Verify error type checking still works

#### 4.6 Create Error Package Tests
- [ ] Create `server/internal/errors/errors_test.go`
- [ ] Test `Unauthorized.Error()` method output
- [ ] Test `NotFound.Error()` method output
- [ ] Test `errors.As()` works with both error types
- [ ] Test error wrapping preserves type information

#### 4.7 Update Existing Tests
- [ ] Update exercise service tests to use shared error types
- [ ] Update exercise handler tests to use shared error types
- [ ] Update workout service tests to use shared error types
- [ ] Update workout handler tests to use shared error types
- [ ] Verify all error assertions still pass

---

## Task 5: Fix Type Assertion to Use errors.As

**Priority:** Medium
**Files Affected:**
- `server/internal/workout/handler.go`
- Tests: `server/internal/workout/handler_test.go` (if FormatValidationErrors is tested)

**Description:** Replace type assertion with `errors.As` in `FormatValidationErrors` function.

### Subtasks

#### 5.1 Update FormatValidationErrors Function
- [ ] Locate `FormatValidationErrors` function at line 328
- [ ] Change `validationErrors` from `*validator.ValidationErrors` to `validator.ValidationErrors`
- [ ] Replace type assertion with `errors.As(err, &validationErrors)`
- [ ] Update loop to iterate over `validationErrors` directly (not dereference)
- [ ] Verify function signature doesn't need to change

#### 5.2 Test with Wrapped Errors
- [ ] Create test with wrapped validation error
- [ ] Verify `errors.As` correctly unwraps to find validation error
- [ ] Test with non-validation error (should fall back to `err.Error()`)

#### 5.3 Update Any Similar Functions
- [ ] Search for other type assertions on errors in handler files
- [ ] Update any similar patterns found

---

## Task 6: Add Logging for Config Parsing Errors

**Priority:** Low
**Files Affected:**
- `server/internal/config/config.go`

**Description:** Add logging when configuration value parsing fails to aid debugging.

### Subtasks

#### 6.1 Update GetInt Function
- [ ] Locate GetInt function around line 95-96
- [ ] Add else clause to `strconv.Atoi` error check
- [ ] Log parsing failure with value and error
- [ ] Use appropriate log level (Info or Debug)
- [ ] Include config key name in log message if available

#### 6.2 Audit Other Parsing Functions
- [ ] Check for similar patterns in GetBool, GetFloat64, etc.
- [ ] Add logging to other parsing failures
- [ ] Ensure consistent log message format

#### 6.3 Test Config Parsing
- [ ] Test with invalid integer value
- [ ] Verify log message appears
- [ ] Verify default value is still returned

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
- [ ] Review logs to see if JSON encoding errors occur in practice
- [ ] Assess impact of partial responses
- [ ] Decide whether to add metrics, enhanced logging, or leave as-is (recommendation: leave as-is)
- [ ] Document decision

#### 7.2 If Tracking is Needed - Add Metrics
- [ ] Define metric for JSON encoding failures
- [ ] Increment metric in error paths
- [ ] Add metric to monitoring dashboard

#### 7.3 If Tracking is Needed - Enhanced Logging
- [ ] Add structured fields to encoding error logs
- [ ] Include response size, endpoint, error type
- [ ] Create alerts for encoding error spikes

#### 7.4 Document Behavior
- [ ] Document that response headers are already written
- [ ] Explain why execution continues after encoding error
- [ ] Add code comments explaining the trade-off

---

## Task 8: Fix Minor Inconsistencies

**Priority:** Low
**Files Affected:**
- `server/internal/workout/repository.go`
- `server/internal/database/errors.go` (sentinel error usage)

**Description:** Fix minor inconsistencies identified in the analysis.

### Subtasks

#### 8.1 Fix Unwrapped Error in Workout Repository
- [ ] Locate line 357 in `workout/repository.go`
- [ ] Wrap error with `fmt.Errorf("...: %w", err)` for consistency
- [ ] Add descriptive context to error message
- [ ] Verify tests still pass

#### 8.2 Enhance Sentinel Error Usage
- [ ] Review how `ErrRowLevelSecurity` and `ErrRLSContext` are created
- [ ] Ensure they're returned directly (not just checked)
- [ ] Update error detection functions to return sentinel errors when appropriate
- [ ] Document when sentinel errors should be used vs. wrapped errors

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
- [ ] Document the structure of `ErrorJSON` responses
- [ ] List all HTTP status codes used across handlers
- [ ] Identify any handlers not using `ErrorJSON`
- [ ] Document `sanitizeErrorMessage` behavior

#### 9.2 Standardize Error Response Structure
- [ ] Ensure all responses have consistent JSON structure
- [ ] Define standard fields: `error`, `message`, `request_id`, etc.
- [ ] Document which errors are sanitized and which are not
- [ ] Verify 400-level vs 500-level error handling

#### 9.3 Fix Rate Limit Middleware Internal Error (Line 44)
- [ ] Replace `http.Error()` call with `response.ErrorJSON()`
- [ ] Use `http.StatusInternalServerError` (already correct)
- [ ] Ensure error includes request_id
- [ ] Test that response returns JSON, not plain text

#### 9.4 Fix Rate Limit Exceeded Response (Lines 62-72)
- [ ] Replace manual JSON construction with `response.ErrorJSON()`
- [ ] Keep `Retry-After` header
- [ ] Use `http.StatusTooManyRequests` (already correct)
- [ ] Ensure request_id is included in response
- [ ] Update error message format to match standard pattern
- [ ] Handle JSON encoding errors properly (ErrorJSON already does this)

#### 9.5 Fix Basic Auth Middleware (Lines 41-52)
- [ ] Replace manual JSON construction with `response.ErrorJSON()`
- [ ] Keep `WWW-Authenticate` header
- [ ] Use `http.StatusUnauthorized` (already correct)
- [ ] Remove manual request_id handling (ErrorJSON handles this)
- [ ] Verify error response structure matches standard format

#### 9.6 Fix CORS Middleware Empty 403 Response (Line 37)
- [ ] Add error response body for rejected CORS preflight
- [ ] Use `response.ErrorJSON()` or document why empty response is intentional
- [ ] Include appropriate error message explaining CORS rejection
- [ ] Verify this doesn't break browser CORS handling

#### 9.7 Fix Health Handler Raw Error Exposure (Line 83)
- [ ] Sanitize database error in ready endpoint response
- [ ] Replace `"failed: " + err.Error()` with generic message
- [ ] Consider adding request_id to health/ready responses (optional)
- [ ] Document that health endpoints use custom response format
- [ ] Verify no sensitive connection details leak in errors

#### 9.8 Standardize Custom Error Responses
- [ ] Verify `Unauthorized` errors return 401 or 403 consistently
- [ ] Verify `NotFound` errors return 404 consistently
- [ ] Verify validation errors return 400 with formatted messages
- [ ] Check database constraint errors return appropriate codes

#### 9.9 Create Error Response Documentation
- [ ] Document all possible error response formats
- [ ] Document HTTP status codes and when each is used
- [ ] Document which error details are exposed to frontend
- [ ] Create examples for frontend team
- [ ] Document middleware error responses (rate limit, CORS, basic auth)

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
