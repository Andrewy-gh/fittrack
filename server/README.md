## Development

### Quick Start

1. Create the local env files in this folder:
```bash
cp setenv.example.sh setenv.sh
cp .env.example .env
```

2. Update `setenv.sh` and `.env` with your local values.

- `setenv.sh` powers `make dev`, `make migrate-up`, `make migrate-down`, and `make test-short`
- `.env` is handy for local integration tests, Playwright setup helpers, and AI smoke testing

3. Install goose and air:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/air-verse/air@latest
export PATH="$HOME/go/bin:$PATH"
```

4. Run the complete development environment:
```bash
make dev
```

That's it! This single command will:
- ✅ Load environment variables from `setenv.sh`
- ✅ Check prerequisites (docker, docker compose, goose, air)
- ✅ Start the PostgreSQL database container
- ✅ Wait for the database to be ready
- ✅ Apply pending database migrations
- ✅ Start the hot-reload development server with air

### Manual Setup (Advanced)

If you prefer to run steps individually:

1. Load environment variables:
```bash
source ./.setenv.sh
# or: source ./setenv.sh
```

2. Start database:
```bash
docker compose up -d postgres
```

3. Initialize the database (first time only):
```bash
make migrate-up
```

4. Start development server:
```bash
air
# OR build and run manually:
go run ./cmd/api
```

### Gemini Smoke Test

From `server/`, run:

```bash
go run ./cmd/gemini-smoke
```

Expected env var: `GEMINI_API_KEY` or `GOOGLE_API_KEY`

Optional env var: `GEMINI_MODEL` (defaults to `googleai/gemini-2.5-flash`)

The command respects existing shell env first, then loads `server/.env` and `server/setenv.sh` (or `server/.setenv.sh`) if present. It sends one real Genkit request to Gemini, times out after 20 seconds, and prints a short model response to stdout.

### AI Chat Runtime

The live API chat runtime uses the same model default as the smoke test:

- default model: `googleai/gemini-2.5-flash`
- supported API key env vars: `GEMINI_API_KEY` or `GOOGLE_API_KEY`
- optional model override: `GEMINI_MODEL`

For local development, keep using the existing server env workflow:

1. Put values in your shell, `server/.env`, `server/.setenv.sh`, or `server/setenv.sh`
2. Source `.setenv.sh` or `setenv.sh` (or otherwise export env) before `go run ./cmd/api` or `make dev`

The API server itself reads process env. It does not introduce a chat-only env file loader.

### Required vs Optional Variables

Required to boot the API:

- `DATABASE_URL`
- `PROJECT_ID`

Common optional variables:

- `PORT` (default `8080`)
- `LOG_LEVEL` (default `info`)
- `ENVIRONMENT` (default `development`)
- `RATE_LIMIT_RPM` (default `100`)
- `ALLOWED_ORIGINS`
- `DB_MAX_CONNS`, `DB_MIN_CONNS`, `DB_MAX_CONN_IDLE`, `DB_MAX_CONN_LIFE`, `DB_HEALTHCHECK`
- `METRICS_USERNAME`, `METRICS_PASSWORD`
- `GEMINI_API_KEY` or `GOOGLE_API_KEY`
- `GEMINI_MODEL`
- `SECRET_SERVER_KEY` for Playwright's server-side auth bootstrap

### AI Chat API

Phase 1 adds persisted chat endpoints under `/api/ai/*`:

- `POST /api/ai/conversations`
- `GET /api/ai/conversations/{id}`
- `POST /api/ai/conversations/{id}/messages/stream`
- `POST /api/ai/conversations/{id}/messages/recover`
- `POST /api/ai/chat/telemetry`

The stream endpoint is authenticated fetch-based SSE:

- preflight failures return normal JSON errors with non-2xx status
- successful requests switch to `text/event-stream`
- post-start failures emit SSE `error` events
- streaming assistant text is snapshotted into app-owned storage during long runs so a dropped client can reload persisted partial progress
- the client recovery path is still storage-backed inspection, not live SSE replay; interrupted sessions poll the persisted conversation until the run reaches a terminal state
- stale `streaming` runs older than the stream timeout grace window are auto-failed before a new send starts, which prevents a permanently blocked conversation after an interrupted server-side run
- interrupted runs can enqueue background recovery through Inngest when `INNGEST_EVENT_KEY` and `INNGEST_SIGNING_KEY` are configured

Proxy note:

- reverse proxies must not buffer the SSE response path
- the handler already sends `X-Accel-Buffering: no`
- keep `text/event-stream` passthrough enabled if you add nginx or another proxy in front

### Configuration

The `make dev` command supports these environment variables:
- `DB_SERVICE` (default: "postgres") - Docker service name for PostgreSQL
- `DB_READY_TIMEOUT` (default: 90) - Seconds to wait for database readiness  
- `AIR_CMD` (default: "air") - Command to run for hot-reload development

Example:
```bash
make DB_READY_TIMEOUT=60 dev
```

### Troubleshooting

**Database connection issues:**
```bash
# Check database logs
docker compose logs db

# Check database health
docker compose ps
```

**Missing goose:**
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export PATH="$HOME/go/bin:$PATH"
```

**Postgres auth failures after changing DB_USER/DB_PASSWORD:**
```bash
# reset local persisted postgres data
cd server
docker compose down
sudo rm -rf _db-data
mkdir -p _db-data
```

**Missing air:**
```bash
# Install via Go
go install github.com/air-verse/air@latest

# Or via Homebrew (macOS)
brew install air
```

### Available Commands

```bash
make help              # Show all available commands
make dev               # Complete dev setup: env → docker up → wait DB → migrate → air
make build             # Compile the project
make run               # Run the compiled binary
make swagger           # Generate OpenAPI documentation
make sqlc              # Generate SQL code from query.sql
make test              # Run all tests
make vet               # Run go vet
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback last migration
make migrate-create    # Create a new migration file (NAME=<name>)
make docker-down       # Stop PostgreSQL container
make clean             # Clean build files and cache
```

## Documentation

- [Row Level Security (RLS)](docs/rls.md) - Multi-tenant data isolation and security
