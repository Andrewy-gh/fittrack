# Middleware Error Handling Audit Results

**Date:** 2025-12-26
**Scope:** Auth handlers and all middleware files

---

## Executive Summary

During the audit of auth and middleware error handling, **5 critical inconsistencies** were discovered that prevent uniform error responses to the frontend. All regular API handlers (exercise, workout, user) correctly use `response.ErrorJSON`, but middleware files use manual error responses.

**Impact:** Frontend receives inconsistent error formats, making it difficult to display uniform error messages to users.

---

## Critical Issues Found

### 1. Rate Limit Middleware - Internal Error (CRITICAL)

**File:** `server/internal/middleware/ratelimit.go:44`

**Current Code:**
```go
context, err := instance.Get(r.Context(), userID)
if err != nil {
    logger.Error("failed to get rate limit context", "error", err, "userID", userID)
    http.Error(w, "internal server error", http.StatusInternalServerError)
    return
}
```

**Problems:**
- ❌ Uses `http.Error()` instead of `response.ErrorJSON()`
- ❌ Returns **plain text** instead of JSON
- ❌ Missing `request_id` in response
- ❌ Inconsistent with all other error responses in the app

**Frontend Impact:** Frontend receives plain text instead of JSON, breaking error parsing.

**Fix Required:**
```go
context, err := instance.Get(r.Context(), userID)
if err != nil {
    response.ErrorJSON(w, r, logger, http.StatusInternalServerError, "internal server error", err)
    return
}
```

---

### 2. Rate Limit Middleware - Rate Exceeded Response (CRITICAL)

**File:** `server/internal/middleware/ratelimit.go:62-72`

**Current Code:**
```go
w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusTooManyRequests)

response := map[string]string{
    "message": fmt.Sprintf("rate limit exceeded, retry after %d seconds", retryAfterSeconds),
}
if err := json.NewEncoder(w).Encode(response); err != nil {
    logger.Error("failed to encode rate limit response", "error", err)
}
```

**Problems:**
- ❌ Manual JSON construction using `map[string]string`
- ❌ Missing `request_id` in response body
- ❌ Inconsistent response structure (uses `message` instead of standard error format)
- ❌ No error handling for JSON encoding failures (just logs)
- ❌ Different from `response.Error` struct used everywhere else

**Frontend Impact:** Missing `request_id` makes it hard to correlate rate limit errors with logs.

**Fix Required:**
```go
w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
response.ErrorJSON(w, r, logger, http.StatusTooManyRequests,
    fmt.Sprintf("rate limit exceeded, retry after %d seconds", retryAfterSeconds), nil)
```

---

### 3. Basic Auth Middleware (HIGH PRIORITY)

**File:** `server/internal/middleware/basicauth.go:41-52`

**Current Code:**
```go
w.Header().Set("WWW-Authenticate", `Basic realm="Metrics"`)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusUnauthorized)

resp := map[string]string{
    "message": "unauthorized",
}
if requestID != "" {
    resp["request_id"] = requestID
}
json.NewEncoder(w).Encode(resp)
```

**Problems:**
- ❌ Manual JSON construction using `map[string]string`
- ❌ Manually adds `request_id` instead of using centralized approach
- ❌ Uses `message` field instead of standard error structure
- ❌ No error handling for `json.Encode()` failure
- ❌ Inconsistent with `response.ErrorJSON` pattern

**Frontend Impact:** Inconsistent error response structure.

**Fix Required:**
```go
w.Header().Set("WWW-Authenticate", `Basic realm="Metrics"`)
response.ErrorJSON(w, r, logger, http.StatusUnauthorized, "unauthorized", nil)
```

---

### 4. CORS Middleware - Empty Error Response (MEDIUM PRIORITY)

**File:** `server/internal/middleware/cors.go:31-39`

**Current Code:**
```go
if r.Method == http.MethodOptions {
    if allowed {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusForbidden)  // ❌ NO BODY
    }
    return
}
```

**Problems:**
- ❌ Returns HTTP 403 with **completely empty body**
- ❌ No error message explaining why CORS preflight was rejected
- ❌ No `request_id` for debugging
- ❌ Silent failure makes debugging difficult

**Frontend Impact:** Browser receives 403 with no explanation. Developers debugging CORS issues have no information.

**Consideration:** CORS preflight responses are typically minimal. Need to verify if adding a body breaks browser CORS handling.

**Suggested Fix (if safe):**
```go
if r.Method == http.MethodOptions {
    if allowed {
        w.WriteHeader(http.StatusOK)
    } else {
        response.ErrorJSON(w, r, logger, http.StatusForbidden, "CORS preflight not allowed", nil)
    }
    return
}
```

---

### 5. Health Handler - Raw Error Exposure (SECURITY CONCERN)

**File:** `server/internal/health/handler.go:83`

**Current Code:**
```go
if err := h.pool.Ping(r.Context()); err != nil {
    response := ReadyResponse{
        Status:    "unhealthy",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        Checks: map[string]string{
            "database": "failed: " + err.Error(),  // ❌ RAW DATABASE ERROR
        },
    }
    // ... send response
}
```

**Problems:**
- ❌ Exposes raw database error messages in response
- ❌ Could leak sensitive connection information (host, port, database name)
- ❌ Security risk if errors contain credentials or internal details

**Frontend Impact:** Potential information disclosure.

**Fix Required:**
```go
"database": "unhealthy",  // Generic message instead of raw error
```

And log the actual error:
```go
h.logger.Error("database health check failed", "error", err, "request_id", middleware.GetRequestID(r.Context()))
```

---

## Compliant Handlers

### ✅ Auth Middleware - CORRECT

**File:** `server/internal/auth/middleware.go`

All error responses correctly use `response.ErrorJSON`:
- Line 88: Missing access token
- Line 95: Invalid access token
- Line 102: Failed to ensure user
- Lines 114-117: Failed to set user context

**This is the pattern all middleware should follow.**

---

## Summary Table

| File | Line(s) | Issue | Priority | Frontend Impact |
|------|---------|-------|----------|-----------------|
| `middleware/ratelimit.go` | 44 | Plain text response | CRITICAL | Breaks JSON parsing |
| `middleware/ratelimit.go` | 62-72 | Manual JSON, missing request_id | CRITICAL | Inconsistent structure |
| `middleware/basicauth.go` | 41-52 | Manual JSON, manual request_id | HIGH | Inconsistent structure |
| `middleware/cors.go` | 37 | Empty 403 response | MEDIUM | No error info |
| `health/handler.go` | 83 | Raw error exposure | SECURITY | Info disclosure |

---

## Recommended Actions

### Immediate (Before Frontend Integration)
1. ✅ Fix `ratelimit.go:44` - blocking frontend JSON parsing
2. ✅ Fix `ratelimit.go:62-72` - missing request_id
3. ✅ Fix `basicauth.go:41-52` - inconsistent structure

### High Priority
4. ✅ Fix `health/handler.go:83` - security concern
5. ✅ Evaluate and fix `cors.go:37` - consider browser compatibility

### Documentation
6. ✅ Document standard error response format for frontend team
7. ✅ Document all HTTP status codes and when they're used
8. ✅ Document special headers (Retry-After, WWW-Authenticate)

---

## Standard Error Response Format

After fixes, all errors should return this consistent JSON structure:

```json
{
  "error": true,
  "message": "Human-readable error message",
  "request_id": "unique-request-id"
}
```

**Special Cases:**
- Rate limit errors include `Retry-After` header
- Basic auth errors include `WWW-Authenticate` header
- Health/ready endpoints may use custom format (for monitoring tools)

---

## Testing Requirements

After implementing fixes:

1. **Test each middleware error path returns JSON**
2. **Verify all errors include request_id**
3. **Verify special headers are preserved** (Retry-After, WWW-Authenticate)
4. **Verify no sensitive info in error messages**
5. **Test CORS preflight behavior** (if modified)
6. **Integration test: Frontend can parse all error responses**

---

## Next Steps

All findings have been incorporated into **Task 9** of the implementation plan:
- Subtask 9.3: Fix rate limit internal error
- Subtask 9.4: Fix rate limit exceeded response
- Subtask 9.5: Fix basic auth middleware
- Subtask 9.6: Fix CORS middleware
- Subtask 9.7: Fix health handler error exposure
- Subtask 9.9: Document error responses for frontend
- Subtask 9.10: Test all error responses

---

**End of Audit Report**
