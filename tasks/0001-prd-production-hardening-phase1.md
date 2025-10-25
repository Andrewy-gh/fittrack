# PRD: Production Hardening - Phase 1 (Critical Foundation)

## Introduction/Overview

The FitTrack server is currently deployed to production on Fly.io but lacks critical production-grade infrastructure for reliability, observability, and security. This PRD addresses the **must-have** foundational improvements needed before scaling to multiple users.

**Problem Statement:** Without proper monitoring, rate limiting, graceful shutdown, and health checks, the server is vulnerable to:
- API abuse and security attacks
- Poor debugging experience when issues occur
- Dropped requests during deployments
- Uncontrolled resource exhaustion
- Lack of visibility into system health

**Goal:** Implement production-grade infrastructure primitives within 1-2 weeks using free/open-source tools, enabling safe scaling from single-user to multi-user without requiring paid services.

---

## Goals

1. **Reliability:** Achieve zero dropped requests during deployments via graceful shutdown
2. **Observability:** Gain visibility into request rates, latency, errors, and system health
3. **Security:** Prevent API abuse through rate limiting and security headers
4. **Operational Readiness:** Enable quick incident response with health checks and structured logging
5. **Cost Efficiency:** Implement all improvements using free/open-source tools

---

## User Stories

**As a solo developer:**
- I want graceful shutdown so that deployments don't drop active user requests
- I want health/readiness endpoints so that Fly.io can properly health-check my app
- I want basic metrics so that I can see if my API is slow or erroring
- I want rate limiting so that a single abusive user can't take down my service
- I want production CORS configured so that my deployed frontend can call the API
- I want to see request IDs in logs so that I can trace user issues end-to-end

**As a future user:**
- I want the API to stay responsive even if other users are making many requests
- I want my data operations to complete successfully without timing out

---

## Functional Requirements

### 1. Graceful Shutdown

**FR-1.1:** The server MUST handle SIGTERM and SIGINT signals
**FR-1.2:** On receiving shutdown signal, the server MUST stop accepting new requests
**FR-1.3:** The server MUST wait up to 30 seconds for in-flight requests to complete
**FR-1.4:** The server MUST close database connections cleanly before exit
**FR-1.5:** The server MUST log shutdown events (initiated, completed, timeout)

**Implementation Location:** `server/cmd/api/main.go`

---

### 2. Health & Readiness Endpoints

**FR-2.1:** The server MUST expose `GET /health` endpoint (HTTP 200 = alive)
**FR-2.2:** The server MUST expose `GET /ready` endpoint that validates:
- Database connection is alive (via `pool.Ping()`)
- Returns HTTP 200 if ready, HTTP 503 if not ready

**FR-2.3:** Health endpoints MUST NOT require authentication
**FR-2.4:** Health endpoints MUST return JSON with status and timestamp:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-25T12:34:56Z",
  "checks": {
    "database": "ok"
  }
}
```

**FR-2.5:** Update `fly.toml` to use `/ready` for health checks

**Implementation Location:**
- `server/internal/health/handler.go` (new)
- `server/cmd/api/routes.go` (add routes)
- `fly.toml` (configure health check)

---

### 3. Rate Limiting

**FR-3.1:** The server MUST implement per-user rate limiting on all `/api/*` endpoints
**FR-3.2:** Default rate limit: **100 requests per minute per user**
**FR-3.3:** Rate limit MUST be configurable via environment variable `RATE_LIMIT_RPM`
**FR-3.4:** When rate limit exceeded, return HTTP 429 with header `Retry-After`
**FR-3.5:** Rate limit responses MUST include:
```json
{
  "message": "rate limit exceeded, retry after 45 seconds"
}
```

**FR-3.6:** Use in-memory rate limiting (no Redis required for Phase 1)
**FR-3.7:** Library recommendation: `github.com/ulule/limiter/v3` (free, simple, no dependencies)

**Implementation Location:**
- `server/internal/middleware/ratelimit.go` (new)
- `server/cmd/api/main.go` (add middleware)

---

### 4. Request Timeouts

**FR-4.1:** HTTP server MUST enforce `ReadTimeout: 10 seconds`
**FR-4.2:** HTTP server MUST enforce `WriteTimeout: 10 seconds`
**FR-4.3:** HTTP server MUST enforce `IdleTimeout: 120 seconds`
**FR-4.4:** Database queries MUST use contexts with 30-second timeout
**FR-4.5:** Timeout errors MUST return HTTP 408 Request Timeout

**Implementation Location:** `server/cmd/api/main.go`

---

### 5. Production CORS Configuration

**FR-5.1:** CORS allowed origins MUST be configurable via `ALLOWED_ORIGINS` environment variable
**FR-5.2:** Default development origin: `http://localhost:5173`
**FR-5.3:** Production MUST support multiple origins (e.g., `https://fittrack.fly.dev,https://www.fittrack.com`)
**FR-5.4:** Remove hardcoded CORS from `server/internal/auth/middleware.go:81`
**FR-5.5:** CORS middleware MUST validate origins against allowlist (reject unknown origins)

**Implementation Location:**
- `server/internal/middleware/cors.go` (new, extract from auth middleware)
- `server/cmd/api/main.go` (configure from env)

---

### 6. Security Headers

**FR-6.1:** The server MUST add security headers to all responses:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000; includeSubDomains` (HTTPS only)

**FR-6.2:** Swagger UI (`/swagger/*`) MUST allow `X-Frame-Options: SAMEORIGIN`
**FR-6.3:** Security headers MUST be applied via middleware

**Implementation Location:** `server/internal/middleware/security.go` (new)

---

### 7. Basic Prometheus Metrics

**FR-7.1:** The server MUST expose `GET /metrics` endpoint (Prometheus format)
**FR-7.2:** Metrics endpoint MUST NOT require authentication
**FR-7.3:** The server MUST track:
- `http_requests_total` (counter) - labels: method, path, status
- `http_request_duration_seconds` (histogram) - labels: method, path
- `db_connections_active` (gauge)
- `db_connections_idle` (gauge)

**FR-7.4:** Use `github.com/prometheus/client_golang` (official, free)
**FR-7.5:** Metrics middleware MUST wrap all `/api/*` handlers

**Implementation Location:**
- `server/internal/middleware/metrics.go` (new)
- `server/cmd/api/routes.go` (add /metrics endpoint)

---

### 8. Request ID Tracking

**FR-8.1:** Every request MUST have a unique request ID (UUID v4)
**FR-8.2:** Request ID MUST be added to response header: `X-Request-ID`
**FR-8.3:** Request ID MUST be included in all structured logs for that request
**FR-8.4:** If client provides `X-Request-ID` header, use it (for tracing)
**FR-8.5:** Log format example:
```json
{
  "level": "error",
  "msg": "failed to create workout",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user_123",
  "path": "/api/workouts",
  "error": "database connection lost"
}
```

**Implementation Location:**
- `server/internal/middleware/requestid.go` (new)
- Update all handlers to include request ID in logs

---

### 9. Configuration Validation

**FR-9.1:** Server MUST validate all required environment variables at startup
**FR-9.2:** Required variables:
- `DATABASE_URL` (must be valid PostgreSQL URL)
- `PROJECT_ID` (Stack Auth project ID)

**FR-9.3:** Optional variables with defaults:
- `PORT` (default: 8080)
- `LOG_LEVEL` (default: info, allowed: debug, info, warn, error)
- `ENVIRONMENT` (default: development, allowed: development, staging, production)
- `RATE_LIMIT_RPM` (default: 100)
- `ALLOWED_ORIGINS` (default: http://localhost:5173)

**FR-9.4:** If validation fails, server MUST exit with code 1 and log clear error
**FR-9.5:** Use struct-based config with `github.com/go-playground/validator` (already in project)

**Implementation Location:**
- `server/internal/config/config.go` (new)
- `server/cmd/api/main.go` (validate config first)

---

### 10. Structured Logging Improvements

**FR-10.1:** Log level MUST be configurable via `LOG_LEVEL` environment variable
**FR-10.2:** In production (`ENVIRONMENT=production`), default to `info` level
**FR-10.3:** In development (`ENVIRONMENT=development`), default to `debug` level
**FR-10.4:** All error logs MUST include:
- `request_id`
- `user_id` (if authenticated)
- `path`
- `method`
- `error` (sanitized for client responses)

**FR-10.5:** Add startup log with configuration summary:
```json
{
  "level": "info",
  "msg": "server starting",
  "environment": "production",
  "port": 8080,
  "log_level": "info",
  "db_max_conns": 15,
  "rate_limit_rpm": 100
}
```

**Implementation Location:** `server/cmd/api/main.go`

---

## Non-Goals (Out of Scope for Phase 1)

**Explicitly NOT included in Phase 1:**
- ❌ Distributed tracing (OpenTelemetry) - deferred to Phase 2
- ❌ Redis caching - deferred to Phase 2
- ❌ Circuit breakers - deferred to Phase 2
- ❌ APM tools (DataDog, New Relic) - too expensive, stick with Prometheus
- ❌ Advanced alerting (PagerDuty) - can use free Fly.io monitoring alerts
- ❌ Feature flags - not critical yet
- ❌ Admin/debug endpoints - deferred
- ❌ API versioning - not needed for single-version API yet
- ❌ GDPR compliance features - deferred until needed

---

## Design Considerations

### Middleware Ordering

Apply middleware in this order (outside-in):
1. Security headers (outermost)
2. CORS
3. Request ID
4. Metrics collection
5. Rate limiting
6. Authentication (existing)
7. Request timeout
8. Handlers (innermost)

### Error Response Format

Maintain existing error response format from `server/internal/response/json.go`:
```json
{
  "message": "sanitized error message"
}
```

Add optional `request_id` field for debugging:
```json
{
  "message": "internal error",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Health Check Response Format

```json
{
  "status": "healthy",
  "timestamp": "2025-10-25T12:34:56Z",
  "version": "1.0.0",
  "checks": {
    "database": "ok"
  }
}
```

On failure:
```json
{
  "status": "unhealthy",
  "timestamp": "2025-10-25T12:34:56Z",
  "version": "1.0.0",
  "checks": {
    "database": "failed: connection refused"
  }
}
```

---

## Technical Considerations

### Dependencies to Add

All dependencies are free and open-source:
```go
// go.mod additions
github.com/ulule/limiter/v3         // Rate limiting (MIT license)
github.com/prometheus/client_golang // Metrics (Apache 2.0)
github.com/google/uuid              // Request IDs (BSD 3-Clause)
```

### Fly.io Configuration Changes

**fly.toml:**
```toml
[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = 'stop'
  auto_start_machines = true
  min_machines_running = 1  # Changed from 0 to avoid cold starts
  processes = ['app']

  # Add health check configuration
  [http_service.checks]
    [http_service.checks.health]
      grace_period = "10s"
      interval = "30s"
      method = "GET"
      path = "/ready"
      timeout = "5s"
```

### Database Context Timeout Pattern

```go
// Example pattern for all repository methods
func (r *Repository) GetWorkout(ctx context.Context, id int64) (*Workout, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // ... existing query logic
}
```

### Graceful Shutdown Implementation Pattern

```go
// Create server with timeouts
server := &http.Server{
    Addr:         fmt.Sprintf(":%d", config.Port),
    Handler:      handler,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

// Handle shutdown signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    logger.Info("shutdown signal received, draining connections")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        logger.Error("forced shutdown", "error", err)
    }

    pool.Close()
    logger.Info("shutdown complete")
}()

// Start server
if err := server.ListenAndServe(); err != http.ErrServerClosed {
    logger.Error("server error", "error", err)
}
```

---

## Success Metrics

### Quantitative Metrics

1. **Deployment Safety:** Zero 5xx errors during deployments (measured via Fly.io logs)
2. **Observability:** 100% of requests tagged with request IDs in logs
3. **Rate Limiting:** Successful 429 responses when exceeding 100 req/min
4. **Response Times:** p95 latency visible in Prometheus metrics
5. **Health Checks:** `/ready` endpoint responds within 100ms

### Qualitative Metrics

1. **Developer Experience:** Can debug production issues using request IDs in under 5 minutes
2. **Operational Confidence:** Can deploy to production without fear of dropped requests
3. **Security Posture:** Pass basic security header scans (securityheaders.com)

---

## Implementation Checklist

### Week 1: Core Infrastructure
- [ ] Create configuration validation (`internal/config/config.go`)
- [ ] Implement graceful shutdown in `main.go`
- [ ] Add HTTP server timeouts
- [ ] Create health/readiness endpoints
- [ ] Update `fly.toml` with health checks
- [ ] Add request ID middleware
- [ ] Update logging to include request IDs

### Week 2: Security & Observability
- [ ] Implement rate limiting middleware
- [ ] Add security headers middleware
- [ ] Extract and fix CORS configuration
- [ ] Add Prometheus metrics middleware
- [ ] Add `/metrics` endpoint
- [ ] Add database context timeouts to all repository methods
- [ ] Update documentation (README.md)

### Testing & Validation
- [ ] Test graceful shutdown (deploy and check for dropped requests)
- [ ] Test health endpoints (curl /health, /ready)
- [ ] Test rate limiting (script to send 101 requests in 1 minute)
- [ ] Verify metrics endpoint (curl /metrics)
- [ ] Verify security headers (curl -I https://fittrack.fly.dev/api/workouts)
- [ ] Test production CORS (deploy with real frontend origin)
- [ ] Load test with 50 concurrent users (use `hey` or `ab`)

---

## Open Questions

1. **Monitoring Dashboard:** Do you want a local Grafana dashboard for development, or just rely on Fly.io's built-in monitoring for now?
2. **Rate Limit Customization:** Should different endpoints have different rate limits? (e.g., login = 10/min, workouts = 100/min)
3. **Metrics Retention:** Should we set up a free Prometheus instance on Fly.io to scrape metrics, or just expose `/metrics` for future use?
4. **Alerting:** Should we set up basic email alerts via Fly.io (free tier), or defer alerting entirely?
5. **Request ID Format:** UUID v4 or shorter format (e.g., nanoid, cuid)?

---

## References

- [Fly.io Health Checks](https://fly.io/docs/reference/configuration/#http_service-checks)
- [Go Graceful Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Ulule Limiter](https://github.com/ulule/limiter)
- [OWASP Security Headers](https://owasp.org/www-project-secure-headers/)

---

## Estimated Effort

**Total:** 12-16 hours for solo developer

**Breakdown:**
- Configuration & graceful shutdown: 2-3 hours
- Health endpoints: 1-2 hours
- Rate limiting: 2-3 hours
- Security headers & CORS: 1-2 hours
- Prometheus metrics: 3-4 hours
- Request ID tracking: 1 hour
- Testing & validation: 2-3 hours

**Timeline:** Completable in 1-2 weeks working part-time (5-8 hours/week)
