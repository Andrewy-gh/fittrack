# Extract Request Context Package - Implementation Plan

**Project:** FitTrack - Fix Middleware Import Cycle
**Created:** 2025-12-27
**Status:** 🔄 In Progress

---

## Overview

This plan addresses the circular dependency between the `middleware` and `response` packages that prevents middleware from using `response.ErrorJSON()`.

**Current Problem:**
- `middleware` → `response` (wants to call ErrorJSON)
- `response` → `middleware` (calls GetRequestID)
- Result: Import cycle error

**Solution:**
Extract request context utilities (`GetRequestID`, `SetRequestID`) into a new shared package `server/internal/request` that both packages can depend on.

**Benefits:**
1. Eliminates import cycle
2. Allows middleware to use `response.ErrorJSON()` for consistent error handling
3. Centralizes request context utilities
4. Follows existing pattern established by `user/context.go`

---

## Research Findings

### Files Using GetRequestID
1. `server/internal/middleware/requestid.go` - Defines GetRequestID (38 LOC)
2. `server/internal/response/json.go` - Uses GetRequestID (1 usage)
3. `server/internal/health/handler.go` - Uses GetRequestID (4 usages)
4. `server/internal/auth/middleware.go` - Uses GetRequestID (4 usages)
5. `server/internal/middleware/requestid_test.go` - Tests GetRequestID (3 usages)
6. `server/internal/response/error_response_test.go` - Uses in test setup (2 usages)

**Total:** 6 files, 14 usages

### Existing Pattern (user/context.go)
The codebase already has a similar pattern in `user/context.go`:
```go
type contextKey string
const UserIDKey contextKey = "user_id"

func WithContext(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, UserIDKey, userID)
}

func Current(ctx context.Context) (string, bool) {
    userID, ok := ctx.Value(UserIDKey).(string)
    return userID, ok
}
```

We'll follow this pattern for consistency.

---

## Implementation Plan

### Task 1: Create New Request Package ⬜

**Priority:** High
**Estimated Complexity:** Low

#### Subtask 1.1: Create request package directory and file
- [ ] Create directory `server/internal/request/`
- [ ] Create file `server/internal/request/context.go`

#### Subtask 1.2: Implement request context utilities
- [ ] Define `contextKey` type
- [ ] Define `RequestIDKey` constant
- [ ] Implement `WithRequestID(ctx context.Context, requestID string) context.Context`
  - Sets request ID in context
  - Returns new context with value
- [ ] Implement `GetRequestID(ctx context.Context) string`
  - Retrieves request ID from context
  - Returns empty string if not found (matches current behavior)
- [ ] Add package documentation

**Expected File Structure:**
```go
package request

import "context"

type contextKey string

const RequestIDKey contextKey = "request_id"

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from the context.
// Returns empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
    if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
        return requestID
    }
    return ""
}
```

#### Subtask 1.3: Create tests for request package
- [ ] Create file `server/internal/request/context_test.go`
- [ ] Test `WithRequestID` - sets value correctly
- [ ] Test `GetRequestID` - retrieves value correctly
- [ ] Test `GetRequestID` - returns empty string when not set
- [ ] Test round-trip (set then get)

---

### Task 2: Update Middleware Package ⬜

**Priority:** High
**Estimated Complexity:** Medium

#### Subtask 2.1: Update requestid.go to use request package
- [ ] Add import: `"github.com/Andrewy-gh/fittrack/server/internal/request"`
- [ ] Remove `type contextKey string` definition
- [ ] Remove `const requestIDKey contextKey = "request_id"`
- [ ] Remove `GetRequestID()` function (now in request package)
- [ ] Update `RequestID()` middleware to use `request.WithRequestID()`
  - Replace `context.WithValue(r.Context(), requestIDKey, requestID)`
  - With `request.WithRequestID(r.Context(), requestID)`

#### Subtask 2.2: Update requestid_test.go
- [ ] Add import: `"github.com/Andrewy-gh/fittrack/server/internal/request"`
- [ ] Replace `GetRequestID(ctx)` with `request.GetRequestID(ctx)` (1 usage, line 44)
- [ ] Update test function names if needed (keep `TestGetRequestID_NoRequestID` for clarity)

---

### Task 3: Update Response Package ⬜

**Priority:** High
**Estimated Complexity:** Low

#### Subtask 3.1: Update json.go to use request package
- [ ] Add import: `"github.com/Andrewy-gh/fittrack/server/internal/request"`
- [ ] Remove import: `"github.com/Andrewy-gh/fittrack/server/internal/middleware"`
- [ ] Replace `middleware.GetRequestID(r.Context())` with `request.GetRequestID(r.Context())` (1 usage, line 145)

#### Subtask 3.2: Update error_response_test.go
- [ ] Update imports to use `request` instead of `middleware` for GetRequestID
- [ ] Note: May still need middleware import for `middleware.RequestID()` middleware function
- [ ] Update test setup code (2 usages on lines 86, 138)

---

### Task 4: Update Other Packages Using GetRequestID ⬜

**Priority:** High
**Estimated Complexity:** Low

#### Subtask 4.1: Update health/handler.go
- [ ] Add import: `"github.com/Andrewy-gh/fittrack/server/internal/request"`
- [ ] Replace `middleware.GetRequestID(r.Context())` with `request.GetRequestID(r.Context())` (4 usages)
  - Line 62: Health endpoint error logging
  - Line 92: Ready endpoint error logging
  - Line 106: Ready endpoint error logging
  - Line 122: Ready endpoint error logging

#### Subtask 4.2: Update auth/middleware.go
- [ ] Add import: `"github.com/Andrewy-gh/fittrack/server/internal/request"`
- [ ] Replace `middleware.GetRequestID(r.Context())` with `request.GetRequestID(r.Context())` (4 usages)
  - Line 87: Missing token warning
  - Line 94: Invalid token error
  - Line 101: Failed to ensure user error
  - Line 113: Failed to set user context error

---

### Task 5: Refactor Middleware Error Responses ⬜

**Priority:** High (Original Goal)
**Estimated Complexity:** Low

**Note:** This can now be done since the import cycle is resolved.

#### Subtask 5.1: Update ratelimit.go
- [ ] Already done in previous changes
- [ ] Verify import: `"github.com/Andrewy-gh/fittrack/server/internal/response"`
- [ ] Verify line 43: Uses `response.ErrorJSON()` for internal error
- [ ] Verify line 65-66: Uses `response.ErrorJSON()` for rate limit exceeded

#### Subtask 5.2: Update basicauth.go
- [ ] Already done in previous changes
- [ ] Verify import: `"github.com/Andrewy-gh/fittrack/server/internal/response"`
- [ ] Verify line 35: Uses `response.ErrorJSON()` for unauthorized

#### Subtask 5.3: Update cors.go
- [ ] Already done in previous changes
- [ ] Verify import: `"github.com/Andrewy-gh/fittrack/server/internal/response"`
- [ ] Verify line 42: Uses `response.ErrorJSON()` for CORS rejection

---

### Task 6: Update Documentation ⬜

**Priority:** Medium
**Estimated Complexity:** Low

#### Subtask 6.1: Update code comments
- [ ] Add package-level doc comment to `request/context.go` explaining purpose
- [ ] Update any references in existing comments

#### Subtask 6.2: Update error response documentation (if needed)
- [ ] Verify `docs/error-responses.md` reflects new middleware error responses
- [ ] Already updated in previous work - just verify

---

### Task 7: Testing ⬜

**Priority:** CRITICAL
**Estimated Complexity:** Medium

#### Subtask 7.1: Run unit tests for new request package
- [ ] Run `go test ./server/internal/request/...`
- [ ] Verify all tests pass
- [ ] Check test coverage: `go test -cover ./server/internal/request/...`
- [ ] Target: 100% coverage (small package, achievable)

#### Subtask 7.2: Run middleware tests
- [ ] Run `go test ./server/internal/middleware/...`
- [ ] Verify all existing tests still pass
- [ ] Verify no import cycle errors

#### Subtask 7.3: Run response tests
- [ ] Run `go test ./server/internal/response/...`
- [ ] Verify all existing tests still pass
- [ ] Verify error response tests work correctly

#### Subtask 7.4: Run auth tests
- [ ] Run `go test ./server/internal/auth/...`
- [ ] Verify middleware tests pass

#### Subtask 7.5: Run health tests
- [ ] Run `go test ./server/internal/health/...`
- [ ] Verify handler tests pass

#### Subtask 7.6: Run full test suite
- [ ] Run `go test ./server/...`
- [ ] Verify ALL tests pass across entire server
- [ ] Check for any build errors
- [ ] Verify no import cycle warnings

#### Subtask 7.7: Integration testing (manual)
- [ ] Start server: `cd server && go run cmd/api/main.go`
- [ ] Test middleware error responses:
  - [ ] Trigger rate limit error (if possible in dev)
  - [ ] Test CORS preflight rejection
  - [ ] Test basic auth failure
- [ ] Verify request IDs appear in responses
- [ ] Verify error format is consistent

---

### Task 8: Code Review Checklist ⬜

**Priority:** Medium
**Estimated Complexity:** Low

#### Subtask 8.1: Verify consistency
- [ ] Check that `request` package follows same pattern as `user` package
- [ ] Verify all imports are correct
- [ ] Check for any missed usages of `middleware.GetRequestID`

#### Subtask 8.2: Verify no regressions
- [ ] Request IDs still appear in logs
- [ ] Request IDs still appear in error responses
- [ ] Middleware still functions correctly

#### Subtask 8.3: Check for performance impact
- [ ] Verify no additional allocations introduced
- [ ] Context value access is still O(1)

---

## Summary of Changes

### Files to Create (1)
1. `server/internal/request/context.go` - New request context utilities
2. `server/internal/request/context_test.go` - Tests for new package

### Files to Modify (8)
1. `server/internal/middleware/requestid.go` - Use request.WithRequestID
2. `server/internal/middleware/requestid_test.go` - Use request.GetRequestID
3. `server/internal/response/json.go` - Use request.GetRequestID
4. `server/internal/response/error_response_test.go` - Update imports
5. `server/internal/health/handler.go` - Use request.GetRequestID
6. `server/internal/auth/middleware.go` - Use request.GetRequestID
7. `server/internal/middleware/ratelimit.go` - Already updated
8. `server/internal/middleware/basicauth.go` - Already updated
9. `server/internal/middleware/cors.go` - Already updated

### Imports Changed
- **Add** `request` package import to: response, health, auth, middleware (requestid.go, requestid_test.go)
- **Remove** `middleware` import from: response/json.go
- **Update** test imports in: response/error_response_test.go

---

## Risk Assessment

### Low Risk ✅
- Small, focused refactor
- Follows existing pattern (user/context.go)
- No behavioral changes (just code organization)
- All existing tests will verify behavior unchanged

### Potential Issues
- **Import updates** - Must update all files consistently
- **Test setup** - Tests may need import changes
- **Forgotten usages** - Must search thoroughly for all GetRequestID calls

### Mitigation
- Comprehensive grep search completed
- All files identified and documented
- Full test suite will catch any missed updates
- Small, incremental commits for easy rollback

---

## Success Criteria

### Must Have ✅
1. [ ] Import cycle eliminated - `go build` succeeds
2. [ ] All tests pass - `go test ./server/...` returns 0 exit code
3. [ ] Middleware uses `response.ErrorJSON()` consistently
4. [ ] Request IDs still appear in all error responses
5. [ ] No behavioral changes (request ID functionality identical)

### Should Have ✅
1. [ ] New `request` package has 100% test coverage
2. [ ] Code follows established patterns (user/context.go)
3. [ ] Documentation updated

### Nice to Have
1. [ ] Manual integration testing confirms everything works
2. [ ] Performance benchmarks show no regression

---

## Rollback Plan

If issues arise:
1. Revert middleware changes (ratelimit.go, basicauth.go, cors.go) to manual JSON
2. Delete new `request` package
3. Restore all imports to use `middleware.GetRequestID`
4. All tests should pass again

**Estimated Rollback Time:** < 5 minutes (just revert commits)

---

## Timeline Estimate

- **Task 1** (Create request package): 15 minutes
- **Task 2** (Update middleware): 10 minutes
- **Task 3** (Update response): 10 minutes
- **Task 4** (Update other packages): 15 minutes
- **Task 5** (Verify middleware errors): 5 minutes
- **Task 6** (Documentation): 5 minutes
- **Task 7** (Testing): 20 minutes
- **Task 8** (Code review): 10 minutes

**Total Estimated Time:** ~90 minutes

---

## Next Steps

1. Create request package (Task 1)
2. Run tests to verify package works
3. Update middleware package (Task 2)
4. Update response package (Task 3)
5. Update remaining packages (Task 4)
6. Run full test suite (Task 7)
7. Manual integration test
8. Commit changes

---

**Status:** Ready to begin implementation
**Confidence:** High (straightforward refactor, well-researched)
