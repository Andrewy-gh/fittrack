# Error Handling Analysis Report - FitTrack Server

**Analysis Date:** 2025-12-26
**Directory Analyzed:** `server/`
**Total Files Analyzed:** 33 Go files (excluding tests)

---

## Executive Summary

The FitTrack server demonstrates **strong foundational error handling practices**, including:
- Consistent use of `%w` for error wrapping
- Proper use of `errors.As` for custom error type checking (17 instances)
- Structured logging throughout
- Good transaction error handling with deferred rollbacks

**Key Areas for Improvement:**
1. Reduce duplicate logging (same error logged 2-3 times)
2. Replace string-based error detection with type-safe alternatives
3. Consolidate duplicate custom error types across packages

---

## Detailed Findings

### 1. Unchecked or Silently Ignored Errors

#### 1.1 Silent Configuration Parsing Failures

**Location:** `server/internal/config/config.go:95-96`

```go
if intValue, err := strconv.Atoi(value); err == nil {
    return intValue
}
// Falls through to return default value
```

**Issue:** When `strconv.Atoi` fails, the error is silently ignored and the function returns the default value without any logging.

**Impact:** Configuration issues are difficult to debug because parsing failures are invisible.

**Recommendation:**
```go
if intValue, err := strconv.Atoi(value); err == nil {
    return intValue
} else {
    log.Printf("failed to parse config value %q as int: %v, using default", value, err)
}
```

---

#### 1.2 JSON Encoding Errors Not Propagated

**Location:** `server/internal/health/handler.go:56-58`

```go
if err := json.NewEncoder(w).Encode(response); err != nil {
    h.logger.Error("failed to encode health response", "error", err, "request_id", middleware.GetRequestID(r.Context()))
}
// Execution continues
```

**Also found in:**
- `health/handler.go:90` (Ready endpoint)
- `health/handler.go:106` (Ready endpoint)
- `middleware/ratelimit.go:69-71`
- `response/json.go:153-155`

**Issue:** Error is logged but execution continues, potentially leaving response in partial state.

**Note:** This may be intentional since response headers are already written, but it's worth considering whether to track these failures.

---

### 2. String-Based Error Detection (Should Use errors.Is/errors.As)

#### 2.1 String Comparison Instead of Sentinel Error

**Location:** `server/internal/user/repository.go:36`

```go
if err.Error() == "no rows in result set" {
    return db.Users{}, sql.ErrNoRows
}
```

**Issue:** Fragile string comparison on error message instead of using proper error checking.

**Recommendation:**
```go
if errors.Is(err, pgx.ErrNoRows) {
    return db.Users{}, sql.ErrNoRows
}
```

---

#### 2.2 String Matching on Database Errors

**Location:** `server/internal/database/errors.go:10-97`

All error detection functions use string matching:

```go
func IsUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }
    msg := err.Error()
    // PostgreSQL unique constraint violation
    if strings.Contains(msg, "SQLSTATE 23505") {
        return true
    }
    // SQLite unique constraint violation
    if strings.Contains(msg, "UNIQUE constraint failed") {
        return true
    }
    return false
}
```

**Functions affected:**
- `IsUniqueConstraintError` (lines 10-33)
- `IsForeignKeyConstraintError` (lines 36-51)
- `IsRowLevelSecurityError` (lines 57-76)
- `IsRLSContextError` (lines 79-97)

**Issue:** String matching is fragile and doesn't work well with wrapped errors.

**Recommendation:** Use `errors.As` with pgx-specific error types:

```go
func IsUniqueConstraintError(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" // unique_violation
    }
    // Keep SQLite fallback if needed
    return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func IsForeignKeyConstraintError(err error) bool {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23503" // foreign_key_violation
    }
    return strings.Contains(err.Error(), "FOREIGN KEY constraint failed")
}
```

---

### 3. Duplicate Logging of the Same Error

**Pattern found throughout the codebase:**

#### Error Flow Example (Exercise Service)

**Step 1 - Service Layer:** `server/internal/exercise/service.go:58-61`
```go
if err != nil {
    es.logger.Error("failed to list exercises", "error", err)
    es.logger.Debug("raw database error details", "error", err.Error(), "error_type", fmt.Sprintf("%T", err), "user_id", userID)
    return nil, fmt.Errorf("failed to list exercises: %w", err)
}
```

**Step 2 - Handler Layer:** `server/internal/exercise/handler.go:49`
```go
response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to list exercises", err)
```

**Step 3 - Response Helper:** `server/internal/response/json.go:140`
```go
logger.Error(message, "error", err, "path", r.URL.Path, "method", r.Method, "status", status, "request_id", requestID)
```

**Result:** The same error is logged **3 times** (plus a debug line).

**Affected areas:**
- All Exercise service methods (List, GetByID, Create, Update, Delete, Search)
- All Workout service methods (Create, GetByID, List, Update, Delete, etc.)
- Repository layers

**Recommendation:**

Choose one layer for error logging:

**Option A: Log at the boundary (handler layer)**
```go
// Service - just return wrapped errors
if err != nil {
    return nil, fmt.Errorf("failed to list exercises: %w", err)
}

// Handler - log once with full context
if err := h.exerciseService.List(ctx, userID); err != nil {
    response.ErrorJSON(w, r, h.logger, http.StatusInternalServerError, "Failed to list exercises", err)
    return
}
```

**Option B: Log at the service layer, skip in response helper**
```go
// Modify ErrorJSON to have a flag to skip logging when already logged
response.ErrorJSONNoLog(w, r, http.StatusInternalServerError, "Failed to list exercises")
```

---

### 4. Custom Error Types - Good Practices with Duplication

#### Custom Errors Defined

**Location 1:** `server/internal/exercise/service.go:23-37`
```go
type ErrUnauthorized struct {
    Message string
}

func (e *ErrUnauthorized) Error() string {
    return e.Message
}

type ErrNotFound struct {
    Message string
}

func (e *ErrNotFound) Error() string {
    return e.Message
}
```

**Location 2:** `server/internal/workout/service.go:31-45`
```go
// Identical definitions
type ErrUnauthorized struct { Message string }
type ErrNotFound struct { Message string }
```

**Good Practice:** These errors are properly checked using `errors.As`:
- Exercise handler: lines 46, 103, 105, 148, 201, 261, 263, 309, 311
- Workout handler: lines 45, 75, 110, 154, 196, 260, 262, 311, 313

**Issue:** Duplicate definitions across packages mean they're not interchangeable.

**Recommendation:** Create a shared errors package:

```go
// server/internal/errors/errors.go
package errors

type Unauthorized struct {
    Resource string
    UserID   string
}

func (e *Unauthorized) Error() string {
    return fmt.Sprintf("unauthorized access to %s by user %s", e.Resource, e.UserID)
}

type NotFound struct {
    Resource string
    ID       string
}

func (e *NotFound) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}
```

Then use across both packages:
```go
import apperrors "yourapp/internal/errors"

return nil, &apperrors.NotFound{Resource: "exercise", ID: id}
```

---

### 5. Sentinel Errors - Good Usage

**Location:** `server/internal/database/errors.go:110-113`

```go
var ErrRowLevelSecurity = errors.New("access denied by row level security policy")
var ErrRLSContext = errors.New("failed to set RLS context")
```

**Proper usage with errors.Is:**
- `database/errors.go:75` - `errors.Is(err, ErrRowLevelSecurity)`
- `database/errors.go:96` - `errors.Is(err, ErrRLSContext)`
- `user/service.go:38` - `errors.Is(err, sql.ErrNoRows)`

**Good Practice:** Correctly using `errors.Is` for sentinel error checking.

**Minor Issue:** The sentinel errors are only checked within wrapper functions that primarily use string matching, limiting their utility.

---

### 6. Type Assertion Instead of errors.As

**Location:** `server/internal/workout/handler.go:328`

```go
func FormatValidationErrors(err error) string {
    if validationErrors, ok := err.(*validator.ValidationErrors); ok {
        var messages []string
        for _, err := range *validationErrors {
            messages = append(messages, fmt.Sprintf("Field '%s' %s", err.Field(), err.Tag()))
        }
        return strings.Join(messages, "; ")
    }
    return err.Error()
}
```

**Issue:** Type assertion doesn't work with wrapped errors.

**Recommendation:**
```go
func FormatValidationErrors(err error) string {
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        var messages []string
        for _, err := range validationErrors {
            messages = append(messages, fmt.Sprintf("Field '%s' %s", err.Field(), err.Tag()))
        }
        return strings.Join(messages, "; ")
    }
    return err.Error()
}
```

---

### 7. Error Wrapping - Excellent Consistency

**Good Practice Found Throughout:**

The codebase consistently uses `%w` for error wrapping:

```go
// Examples from exercise/service.go
return nil, fmt.Errorf("failed to list exercises: %w", err)
return nil, fmt.Errorf("failed to get exercise: %w", err)
return fmt.Errorf("failed to create exercise: %w", err)

// Examples from workout/service.go
return nil, fmt.Errorf("failed to create workout: %w", err)
return nil, fmt.Errorf("failed to get workout: %w", err)
return fmt.Errorf("failed to update workout: %w", err)

// Repository methods all properly wrap errors
```

**Files with consistent wrapping:**
- All service files
- All repository files
- All handler files

**Minor Exception:** `workout/repository.go:357` returns unwrapped error in internal helper, but this is acceptable for internal functions.

---

### 8. Error Context and Messages

#### Public vs Internal Error Messages

**Location:** `server/internal/response/json.go:30-59`

The `sanitizeErrorMessage` function intentionally makes error messages generic for security:

```go
if containsDatabaseError(errMsg) {
    if strings.Contains(strings.ToLower(message), "unauthorized") ||
       strings.Contains(strings.ToLower(message), "auth") {
        return "unauthorized"
    }
    return "internal error"
}
```

**Good Practice:** Prevents information leakage while maintaining detailed internal logs.

---

## Statistics

### Error Handling Metrics

| Metric | Count |
|--------|-------|
| Sentinel errors defined | 2 |
| Custom error types | 4 (2 duplicates) |
| Proper `errors.As` usage | 17 |
| Proper `errors.Is` usage | 3 |
| Type assertions on errors | 1 |
| String-based error checking | 5 functions + 1 instance |
| Files with consistent `%w` wrapping | 33/33 |

---

## Priority Recommendations

### High Priority

1. **Replace string-based database error detection** in `database/errors.go`
   - Use `errors.As` with `*pgconn.PgError` and check error codes
   - Estimated impact: 5 functions, more reliable error handling

2. **Fix string comparison in user repository** (`user/repository.go:36`)
   - Replace `err.Error() == "no rows in result set"` with `errors.Is(err, pgx.ErrNoRows)`
   - Estimated impact: More reliable, works with wrapped errors

3. **Reduce duplicate logging**
   - Choose single layer for error logging (recommend handler layer)
   - Estimated impact: Cleaner logs, easier debugging, 2-3x reduction in log volume

### Medium Priority

4. **Consolidate custom error types**
   - Create shared `internal/errors` package
   - Move `ErrUnauthorized` and `ErrNotFound` to shared location
   - Estimated impact: Better code reuse, type safety across packages

5. **Replace type assertion with errors.As** in `FormatValidationErrors`
   - Estimated impact: Works correctly with wrapped validation errors

### Low Priority

6. **Add logging for config parsing errors**
   - Help debug configuration issues
   - Estimated impact: Better observability of config problems

7. **Consider tracking JSON encoding errors**
   - Add metrics or structured tracking for encoding failures
   - Estimated impact: Better visibility into partial response issues

---

## Conclusion

The FitTrack server codebase demonstrates **strong error handling fundamentals** with consistent use of modern Go error handling patterns. The main opportunities for improvement are:

1. **Reducing noise** through single-layer logging
2. **Improving reliability** by replacing string matching with type-safe error checking
3. **Reducing duplication** by consolidating common error types

These changes would elevate the error handling from good to excellent while maintaining the existing strong practices around error wrapping, structured logging, and custom error types.

---

## References

- [Go Error Handling Best Practices](https://go.dev/blog/go1.13-errors)
- [Effective Error Handling in Go](https://earthly.dev/blog/golang-errors/)
- [PostgreSQL Error Codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
- Project: `error-handling.md` (guidelines for `errors.As` vs `errors.Is`)
