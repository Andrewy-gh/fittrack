# Tasks: Production Hardening - Phase 1

## Relevant Files

### New Files to Create

- `server/internal/config/config.go` - Centralized configuration validation with struct-based config and environment variable parsing
- `server/internal/config/config_test.go` - Unit tests for configuration validation
- `server/internal/health/handler.go` - Health and readiness endpoint handlers with database connectivity checks
- `server/internal/health/handler_test.go` - Unit tests for health endpoints
- `server/internal/middleware/security.go` - Security headers middleware (X-Content-Type-Options, X-Frame-Options, etc.)
- `server/internal/middleware/security_test.go` - Unit tests for security headers
- `server/internal/middleware/cors.go` - CORS middleware extracted from auth with configurable origins
- `server/internal/middleware/cors_test.go` - Unit tests for CORS middleware
- `server/internal/middleware/requestid.go` - Request ID tracking middleware using UUID v4
- `server/internal/middleware/requestid_test.go` - Unit tests for request ID middleware
- `server/internal/middleware/ratelimit.go` - Rate limiting middleware using ulule/limiter
- `server/internal/middleware/ratelimit_test.go` - Unit tests for rate limiting
- `server/internal/middleware/metrics.go` - Prometheus metrics collection middleware
- `server/internal/middleware/metrics_test.go` - Unit tests for metrics middleware

### Existing Files to Modify

- `server/cmd/api/main.go` - Add graceful shutdown, config validation, HTTP timeouts, middleware initialization
- `server/cmd/api/routes.go` - Add `/health`, `/ready`, and `/metrics` endpoints
- `server/internal/auth/middleware.go` - Remove hardcoded CORS configuration (line 80-84)
- `server/internal/response/json.go` - Add optional `request_id` field to error responses
- `server/internal/response/types.go` - Define error response type with request_id field
- `server/internal/user/repository.go` - Add 30-second context timeouts to all database queries
- `server/internal/workout/repository.go` - Add 30-second context timeouts to all database queries
- `server/internal/exercise/repository.go` - Add 30-second context timeouts to all database queries
- `fly.toml` - Add health check configuration and update min_machines_running
- `go.mod` - Add new dependencies (ulule/limiter/v3, prometheus/client_golang, google/uuid)

### Notes

- Unit tests should be placed alongside the code files they test (same directory)
- Use `go test ./...` to run all tests
- Follow existing project patterns for error handling and logging from `server/CLAUDE.md`
- Middleware ordering is critical: Security Headers → CORS → Request ID → Metrics → Rate Limiting → Authentication → Handlers

## Tasks

- [ ] 1.0 Set up configuration management and validation system
  - [x] 1.1 Create `server/internal/config/config.go` with Config struct containing all required and optional environment variables (DATABASE_URL, PROJECT_ID, PORT, LOG_LEVEL, ENVIRONMENT, RATE_LIMIT_RPM, ALLOWED_ORIGINS)
  - [x] 1.2 Implement `Load()` function that reads environment variables and validates using go-playground/validator
  - [x] 1.3 Add validation rules: DATABASE_URL must be valid PostgreSQL URL, PROJECT_ID required, LOG_LEVEL enum (debug/info/warn/error), ENVIRONMENT enum (development/staging/production)
  - [x] 1.4 Create `server/internal/config/config_test.go` with tests for valid config, missing required vars, invalid formats, and default values
  - [x] 1.5 Add helper methods for parsing comma-separated ALLOWED_ORIGINS into string slice
  - [x] 1.6 Ensure config exits with code 1 and clear error message if validation fails

- [ ] 2.0 Implement graceful shutdown and server lifecycle management
  - [ ] 2.1 Update `server/cmd/api/main.go` to create http.Server with ReadTimeout: 10s, WriteTimeout: 10s, IdleTimeout: 120s (FR-4.1 to FR-4.3)
  - [ ] 2.2 Set up signal handling for SIGTERM and SIGINT using os.Signal channel (FR-1.1)
  - [ ] 2.3 Implement shutdown goroutine that stops accepting new requests and waits up to 30 seconds for in-flight requests (FR-1.2, FR-1.3)
  - [ ] 2.4 Add pool.Close() in shutdown handler to cleanly close database connections (FR-1.4)
  - [ ] 2.5 Add structured logging for shutdown events: "shutdown signal received", "draining connections", "shutdown complete", "forced shutdown" on timeout (FR-1.5)
  - [ ] 2.6 Handle http.ErrServerClosed separately from actual server errors
  - [ ] 2.7 Add startup log with configuration summary showing environment, port, log_level, db_max_conns, rate_limit_rpm (FR-10.5)

- [ ] 3.0 Create middleware infrastructure (security headers, CORS, request ID)
  - [ ] 3.1 Create `server/internal/middleware/security.go` with SecurityHeaders() middleware that adds X-Content-Type-Options: nosniff, X-Frame-Options: DENY, X-XSS-Protection: 1; mode=block (FR-6.1)
  - [ ] 3.2 Add logic in SecurityHeaders to allow X-Frame-Options: SAMEORIGIN for /swagger/* paths (FR-6.2)
  - [ ] 3.3 Add Strict-Transport-Security header only when request is HTTPS (check r.TLS != nil) (FR-6.1)
  - [ ] 3.4 Create `server/internal/middleware/cors.go` with CORS() middleware that accepts allowed origins slice as parameter
  - [ ] 3.5 Implement origin validation in CORS middleware to reject unknown origins, handle preflight OPTIONS requests (FR-5.5)
  - [ ] 3.6 Remove hardcoded CORS logic from `server/internal/auth/middleware.go` lines 80-84 (FR-5.4)
  - [ ] 3.7 Create `server/internal/middleware/requestid.go` with RequestID() middleware using github.com/google/uuid for UUID v4 generation (FR-8.1)
  - [ ] 3.8 Add X-Request-ID to response headers, support client-provided X-Request-ID for tracing (FR-8.2, FR-8.4)
  - [ ] 3.9 Store request ID in request context for access by handlers and logging (FR-8.3)
  - [ ] 3.10 Update `server/cmd/api/main.go` to apply middleware in order: SecurityHeaders → CORS → RequestID → Authentication
  - [ ] 3.11 Create unit tests for all three middleware files testing headers, context values, and edge cases

- [ ] 4.0 Add health and readiness endpoints with database validation
  - [ ] 4.1 Create `server/internal/health/handler.go` with Handler struct containing logger and database pool
  - [ ] 4.2 Implement GET /health endpoint that returns 200 with JSON: {"status": "healthy", "timestamp": "...", "version": "1.0.0"} (FR-2.1, FR-2.4)
  - [ ] 4.3 Implement GET /ready endpoint that calls pool.Ping(ctx) to validate database connection (FR-2.2)
  - [ ] 4.4 Return 200 with {"status": "healthy", "timestamp": "...", "checks": {"database": "ok"}} if ready (FR-2.2, FR-2.4)
  - [ ] 4.5 Return 503 with {"status": "unhealthy", "timestamp": "...", "checks": {"database": "failed: connection refused"}} if database check fails (FR-2.2)
  - [ ] 4.6 Ensure health endpoints do NOT require authentication (FR-2.3)
  - [ ] 4.7 Add health endpoint routes to `server/cmd/api/routes.go` at /health and /ready
  - [ ] 4.8 Create `server/internal/health/handler_test.go` with tests for healthy state, database connection failure, and response format

- [ ] 5.0 Implement rate limiting middleware
  - [ ] 5.1 Add `github.com/ulule/limiter/v3` dependency to go.mod (FR-3.7)
  - [ ] 5.2 Create `server/internal/middleware/ratelimit.go` with RateLimit() middleware using in-memory store (FR-3.6)
  - [ ] 5.3 Implement per-user rate limiting at 100 requests per minute, extracting user ID from context (FR-3.1, FR-3.2)
  - [ ] 5.4 Make rate limit configurable via RATE_LIMIT_RPM config value (FR-3.3)
  - [ ] 5.5 When rate limit exceeded, return HTTP 429 with Retry-After header and JSON {"message": "rate limit exceeded, retry after 45 seconds"} (FR-3.4, FR-3.5)
  - [ ] 5.6 Apply rate limiting middleware to all /api/* endpoints after RequestID middleware but before Authentication
  - [ ] 5.7 Create `server/internal/middleware/ratelimit_test.go` with tests for rate limit enforcement, header presence, and configurable limits

- [ ] 6.0 Add Prometheus metrics collection and endpoint
  - [ ] 6.1 Add `github.com/prometheus/client_golang` dependency to go.mod (FR-7.4)
  - [ ] 6.2 Create `server/internal/middleware/metrics.go` with Metrics() middleware
  - [ ] 6.3 Initialize Prometheus metrics: http_requests_total (counter), http_request_duration_seconds (histogram), db_connections_active (gauge), db_connections_idle (gauge) (FR-7.3)
  - [ ] 6.4 Implement middleware to track request count, duration, method, path, and status code labels (FR-7.3)
  - [ ] 6.5 Add function to update database connection metrics from pool.Stat() (FR-7.3)
  - [ ] 6.6 Apply metrics middleware to all /api/* handlers after CORS but before rate limiting (FR-7.5)
  - [ ] 6.7 Add GET /metrics endpoint to `server/cmd/api/routes.go` that serves promhttp.Handler() (FR-7.1)
  - [ ] 6.8 Ensure /metrics endpoint does NOT require authentication (FR-7.2)
  - [ ] 6.9 Create `server/internal/middleware/metrics_test.go` with tests for metric collection, labels, and endpoint response format

- [ ] 7.0 Update fly.toml and add database context timeouts
  - [ ] 7.1 Update `fly.toml` to add [http_service.checks.health] section with grace_period: 10s, interval: 30s, method: GET, path: /ready, timeout: 5s (FR-2.5)
  - [ ] 7.2 Change min_machines_running from 0 to 1 in fly.toml to avoid cold starts
  - [ ] 7.3 Add 30-second context timeout pattern to all methods in `server/internal/user/repository.go` (FR-4.4)
  - [ ] 7.4 Add 30-second context timeout pattern to all methods in `server/internal/workout/repository.go` (FR-4.4)
  - [ ] 7.5 Add 30-second context timeout pattern to all methods in `server/internal/exercise/repository.go` (FR-4.4)
  - [ ] 7.6 Update `server/internal/response/json.go` to include request_id in error responses when available in context (FR-8.3, FR-8.5)
  - [ ] 7.7 Update all handler logging to include request_id from context (FR-10.4)
  - [ ] 7.8 Update `server/cmd/api/main.go` to use LOG_LEVEL from config and set slog.HandlerOptions.Level dynamically (FR-10.1, FR-10.2, FR-10.3)
