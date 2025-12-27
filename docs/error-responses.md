# API Error Response Documentation

**Last Updated:** 2025-12-27
**Audience:** Frontend developers, API consumers
**Purpose:** Complete reference for handling error responses from the FitTrack API

---

## Table of Contents

1. [Error Response Format](#error-response-format)
2. [HTTP Status Codes](#http-status-codes)
3. [Error Types and Examples](#error-types-and-examples)
4. [Middleware Error Responses](#middleware-error-responses)
5. [Error Sanitization](#error-sanitization)
6. [Best Practices for Frontend](#best-practices-for-frontend)

---

## Error Response Format

All error responses (except health endpoints) follow a standardized JSON structure:

```json
{
  "message": "Human-readable error description",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `message` | string | ✅ Yes | Human-readable error message, sanitized for security |
| `request_id` | string | ⚠️ Optional | UUID v4 request identifier for tracing/debugging (included when available) |

**Note:** The `request_id` field is included in most responses but may be absent in some edge cases (e.g., before request context is established).

---

## HTTP Status Codes

### 2xx - Success

| Code | Name | Usage |
|------|------|-------|
| 200 | OK | Successful GET requests with response body |
| 204 | No Content | Successful DELETE/UPDATE operations without response body |

### 4xx - Client Errors (User Fixable)

| Code | Name | When Used | Example Causes |
|------|------|-----------|----------------|
| 400 | Bad Request | Invalid input, validation failures | Missing required fields, invalid JSON, type mismatches, validation errors |
| 401 | Unauthorized | Authentication/authorization failures | Missing/invalid JWT token, user lacks permission for resource, token expired |
| 403 | Forbidden | Access forbidden by policy | CORS rejection (only used by CORS middleware) |
| 404 | Not Found | Resource doesn't exist | Exercise ID not found, workout not found for user |
| 409 | Conflict | Resource conflict | Duplicate exercise name (unique constraint violation) |
| 429 | Too Many Requests | Rate limit exceeded | Too many requests from IP address |

### 5xx - Server Errors (Not User Fixable)

| Code | Name | When Used | Example Causes |
|------|------|-----------|----------------|
| 500 | Internal Server Error | Unexpected server failures | Database errors, internal bugs, unexpected conditions |
| 503 | Service Unavailable | Service temporarily down | Database connection failures, service dependencies unavailable |

---

## Error Types and Examples

### 400 - Bad Request

**Validation Error:**
```json
{
  "message": "Name is required",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Invalid JSON:**
```json
{
  "message": "failed to decode request body",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Missing Parameter:**
```json
{
  "message": "Missing exercise ID",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Invalid Parameter Format:**
```json
{
  "message": "Invalid workout ID",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 401 - Unauthorized

**Missing Authentication:**
```json
{
  "message": "unauthorized",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Resource Access Denied:**
```json
{
  "message": "unauthorized access to exercise 123 by user 456",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Basic Auth Required (Metrics Endpoint):**
```json
{
  "message": "unauthorized"
}
```

**Response Headers:**
```
WWW-Authenticate: Basic realm="metrics"
```

### 403 - Forbidden

**CORS Rejection:**
```json
{
  "message": "origin not allowed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 404 - Not Found

**Resource Not Found:**
```json
{
  "message": "exercise 123 not found",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

```json
{
  "message": "workout 789 not found",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 409 - Conflict

**Unique Constraint Violation:**
```json
{
  "message": "Exercise name already exists",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 429 - Too Many Requests

**Rate Limit Exceeded:**
```json
{
  "message": "rate limit exceeded"
}
```

**Response Headers:**
```
Retry-After: 60
```

**Note:** The `Retry-After` header indicates how many seconds to wait before retrying.

### 500 - Internal Server Error

**Generic Server Error:**
```json
{
  "message": "internal error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Operation Failure:**
```json
{
  "message": "failed to list workouts",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 503 - Service Unavailable

**Database Unavailable:**
```json
{
  "message": "internal error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## Middleware Error Responses

Middleware components (rate limiter, CORS, basic auth) also follow the standard error response format:

### Rate Limit Middleware

**Endpoints:** All API endpoints
**Status Code:** 429 Too Many Requests
**Response:**
```json
{
  "message": "rate limit exceeded"
}
```
**Headers:**
```
Retry-After: 60
```

**Internal Rate Limiter Error (rare):**
**Status Code:** 500 Internal Server Error
**Response:**
```json
{
  "message": "internal error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### CORS Middleware

**Endpoints:** All preflight OPTIONS requests
**Status Code:** 403 Forbidden
**Response:**
```json
{
  "message": "origin not allowed",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Note:** This only occurs when an origin is explicitly rejected during preflight. Normal requests from disallowed origins are handled by the browser before reaching the server.

### Basic Auth Middleware

**Endpoints:** `/metrics` (Prometheus metrics)
**Status Code:** 401 Unauthorized
**Response:**
```json
{
  "message": "unauthorized"
}
```
**Headers:**
```
WWW-Authenticate: Basic realm="metrics"
```

---

## Error Sanitization

For security reasons, certain error details are **sanitized** before being returned to clients. The actual error details are logged server-side for debugging.

### Sanitized Errors (Generic Messages)

These errors return generic messages to prevent information leakage:

| Error Type | Returned Message |
|------------|------------------|
| Database errors | `"internal error"` or `"unauthorized"` |
| PostgreSQL error codes | `"internal error"` |
| Connection/pool errors | `"internal error"` |
| JWT/token parsing errors | `"unauthorized"` |
| Database connection details | `"internal error"` |

### Not Sanitized (Original Messages)

These errors are safe to return to clients:

- Validation errors (field requirements, format issues)
- "Missing parameter" errors
- "Invalid format" errors
- Custom application errors (NotFound, Unauthorized with context)
- Unique constraint violations (business logic errors)

### Example: Database Error Sanitization

**What the server logs:**
```
ERROR database query failed error="pq: connection refused at 127.0.0.1:5432"
```

**What the client receives:**
```json
{
  "message": "internal error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

## Health Endpoints

**Note:** Health and readiness endpoints (`/health`, `/ready`) use a **custom response format** designed for monitoring systems (Kubernetes, load balancers, etc.) and do **not** follow the standard error response format.

### Health Endpoint (`/health`)

**Success Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-27T10:30:00Z",
  "version": "1.0.0"
}
```

### Readiness Endpoint (`/ready`)

**Success Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-27T10:30:00Z",
  "checks": {
    "database": "ok"
  }
}
```

**Failure Response (503 Service Unavailable):**
```json
{
  "status": "unhealthy",
  "timestamp": "2025-12-27T10:30:00Z",
  "checks": {
    "database": "unavailable"
  }
}
```

**Note:** Database errors are sanitized - only generic status values (`"ok"`, `"unavailable"`) are returned. Actual errors are logged with `request_id` for debugging.

---

## Best Practices for Frontend

### 1. Always Check HTTP Status Code First

```typescript
if (response.status >= 200 && response.status < 300) {
  // Success
} else if (response.status >= 400 && response.status < 500) {
  // Client error - user can fix
} else if (response.status >= 500) {
  // Server error - not user's fault
}
```

### 2. Display Error Messages to Users

The `message` field is safe to display to users. It contains sanitized, human-readable error descriptions.

```typescript
const errorData = await response.json();
showErrorToUser(errorData.message);
```

### 3. Use Request ID for Support

Include the `request_id` in error reports or support tickets for easier debugging:

```typescript
const errorData = await response.json();
if (errorData.request_id) {
  console.error(`Error occurred. Request ID: ${errorData.request_id}`);
  // Show to user: "An error occurred. Please contact support with ID: ..."
}
```

### 4. Handle Rate Limiting

Check for 429 status and respect the `Retry-After` header:

```typescript
if (response.status === 429) {
  const retryAfter = response.headers.get('Retry-After');
  const waitSeconds = parseInt(retryAfter || '60', 10);
  // Wait before retrying
  await new Promise(resolve => setTimeout(resolve, waitSeconds * 1000));
}
```

### 5. Handle Common Error Patterns

**Validation Errors (400):**
```typescript
if (response.status === 400) {
  // Show validation error to user
  // Usually field-specific, user can fix
  showFieldError(errorData.message);
}
```

**Unauthorized (401):**
```typescript
if (response.status === 401) {
  // Token expired or invalid
  // Redirect to login
  redirectToLogin();
}
```

**Not Found (404):**
```typescript
if (response.status === 404) {
  // Resource doesn't exist
  // Show "not found" page or message
  showNotFoundMessage();
}
```

**Conflict (409):**
```typescript
if (response.status === 409) {
  // Duplicate resource
  // Show specific message (e.g., "Exercise name already exists")
  showConflictError(errorData.message);
}
```

**Server Error (500/503):**
```typescript
if (response.status >= 500) {
  // Server problem, not user's fault
  // Show generic error, include request_id for support
  showServerError(errorData.message, errorData.request_id);
}
```

### 6. TypeScript Type Definitions

```typescript
interface ErrorResponse {
  message: string;
  request_id?: string;
}

interface HealthResponse {
  status: 'healthy' | 'unhealthy';
  timestamp: string;
  version: string;
}

interface ReadyResponse {
  status: 'healthy' | 'unhealthy';
  timestamp: string;
  checks: {
    database: 'ok' | 'unavailable';
  };
}
```

### 7. Complete Error Handling Example

```typescript
async function fetchWorkouts(): Promise<Workout[]> {
  try {
    const response = await fetch('/api/v1/workouts', {
      headers: {
        'Authorization': `Bearer ${getToken()}`,
      },
    });

    if (!response.ok) {
      const errorData: ErrorResponse = await response.json();

      switch (response.status) {
        case 400:
          throw new ValidationError(errorData.message);
        case 401:
          // Token expired - redirect to login
          redirectToLogin();
          throw new AuthError(errorData.message);
        case 404:
          return []; // No workouts found
        case 429:
          const retryAfter = response.headers.get('Retry-After');
          throw new RateLimitError(
            errorData.message,
            parseInt(retryAfter || '60', 10)
          );
        case 500:
        case 503:
          // Server error - show support message with request_id
          throw new ServerError(
            errorData.message,
            errorData.request_id
          );
        default:
          throw new Error(errorData.message);
      }
    }

    return await response.json();
  } catch (error) {
    if (error instanceof ValidationError) {
      showFieldError(error.message);
    } else if (error instanceof RateLimitError) {
      showRateLimitMessage(error.retryAfter);
    } else if (error instanceof ServerError) {
      showSupportMessage(error.message, error.requestId);
    } else {
      showGenericError();
    }
    throw error;
  }
}
```

---

## Summary

- **Standard Format:** All errors use `{"message": "...", "request_id": "..."}` (except health endpoints)
- **Status Codes:** Semantically correct HTTP status codes (400s for client errors, 500s for server errors)
- **Sanitized:** Sensitive error details are removed; safe messages returned
- **Request ID:** Include in support requests for easier debugging
- **Middleware:** Rate limit, CORS, and basic auth follow the same standard format
- **Health:** Uses custom format for monitoring systems

For questions or issues, contact the backend team or file an issue with the `request_id` from the error response.
