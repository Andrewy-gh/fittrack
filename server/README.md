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
source ./setenv.sh
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

The command respects existing shell env first, then loads `server/.env` and `server/setenv.sh` if present. It sends one real Genkit request to Gemini, times out after 20 seconds, and prints a short model response to stdout.

### AI Chat Scenario Sweep

From `server/`, run real-provider AI chat scenario evals with:

```bash
# Full single-turn default pack
go run ./cmd/ai-chat-scenario-sweep -mode single_turn

# Full two-turn default pack
go run ./cmd/ai-chat-scenario-sweep -mode two_turn

# Selected two-turn single scenario
go run ./cmd/ai-chat-scenario-sweep -mode two_turn -scenario prompt-03

# Selected two-turn batch
go run ./cmd/ai-chat-scenario-sweep -mode two_turn -scenarios prompt-03,prompt-04,prompt-12
```

Expected env var: `GEMINI_API_KEY` or `GOOGLE_API_KEY`

Optional env vars:

- `GEMINI_MODEL` to override the default model
- `FITTRACK_AI_CHAT_SWEEP_OUT` to override the JSON report path

Optional flags:

- `-mode single_turn` runs one assistant turn for each default scenario
- `-mode two_turn` answers configured follow-up questions after the assistant asks one
- `-scenario prompt-03` runs one scenario id from the default pack
- `-scenarios prompt-03,prompt-04,prompt-12` runs selected scenario ids while preserving default-pack order
- `-from prompt-03 -to prompt-08` runs an inclusive contiguous id range in default-pack order
- `-timeout 15m` limits the full sweep wall-clock runtime; lower it for quick checks or raise it for slower provider runs

The command writes a JSON report to `FITTRACK_AI_CHAT_SWEEP_OUT` when set, otherwise to `~/.codex/diagrams/fittrack-ai-chat-scenario-sweep.json`. The summary includes structured draft count, follow-up count, text-only count, error count, and conversion rates so model or prompt changes can be compared across runs. When scenario selection flags are used, `scenario_count`, summary totals, results, and stderr progress include only the selected scenarios.

### AI Chat Runtime

The live API chat runtime uses the same model default as the smoke test:

- default model: `googleai/gemini-2.5-flash`
- supported API key env vars: `GEMINI_API_KEY` or `GOOGLE_API_KEY`
- optional model override: `GEMINI_MODEL`

For local development, keep using the existing server env workflow:

1. Put values in your shell, `server/.env`, or `server/setenv.sh`
2. Source `setenv.sh` (or otherwise export env) before `go run ./cmd/api` or `make dev`

The API server itself reads process env. It does not introduce a chat-only env file loader.

### AI Chat Recovery In Dev

Background AI chat recovery uses Inngest. In local development, recovery stays disabled until both `INNGEST_EVENT_KEY` and `INNGEST_SIGNING_KEY` are set.

To run recovery locally:

1. Export the Inngest env vars in `setenv.sh` or your shell.
2. Set `INNGEST_DEV=1` to use Inngest's default local dev server address (`http://127.0.0.1:8288`).
   You can also set `INNGEST_DEV` to an explicit URL if your dev server is running elsewhere.
3. Start the Inngest dev server in another terminal:

```bash
inngest dev
```

4. Start the API:

```bash
make dev
```

Notes:

- If `INNGEST_EVENT_KEY` and `INNGEST_SIGNING_KEY` are not both set, `POST /api/ai/conversations/{id}/messages/recover` returns `503`.
- `GET /inngest` should respond once recovery is configured and the handler is mounted.

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
- `AI_CHAT_TRACE_LOGS=true` to temporarily enable verbose AI chat stream timing logs while debugging
- `E2E_LOCAL_AUTH_ENABLED` to enable the local-only Playwright auth bootstrap in `development`
- `E2E_LOCAL_AUTH_USER_ID`, `E2E_LOCAL_AUTH_EMAIL`, `E2E_LOCAL_AUTH_DISPLAY_NAME` to override the deterministic local test user
- `SECRET_SERVER_KEY` for Playwright's server-side auth bootstrap

### Local Playwright Auth Bootstrap

Local structured-chat E2E runs can avoid manual Google or GitHub login.

Set these values in `server/.env` or `setenv.sh`:

```env
E2E_LOCAL_AUTH_ENABLED=true
E2E_LOCAL_AUTH_USER_ID=local-e2e-user
E2E_LOCAL_AUTH_EMAIL=local-e2e-user@example.test
E2E_LOCAL_AUTH_DISPLAY_NAME=Local E2E User
```

This local-only path is tightly gated:

- it only turns on when `ENVIRONMENT=development`
- it requires `E2E_LOCAL_AUTH_ENABLED=true`
- it only accepts the one configured local E2E user id

When enabled, the server exposes two dev-only helpers:

- `POST /dev/e2e/auth/bootstrap`
- `POST /dev/e2e/ai-chat/conversations`

Playwright uses them to:

- ensure the local test user exists
- grant `ai_chatbot` feature access
- seed a persisted AI chat conversation with `latest_workout_draft`

### AI Chat API

The structured workout chat slice exposes persisted chat endpoints under `/api/ai/*`:

- `POST /api/ai/conversations`
- `GET /api/ai/conversations/{id}`
- `POST /api/ai/conversations/{id}/messages/stream`
- `POST /api/ai/conversations/{id}/messages/recover`
- `POST /api/ai/chat/telemetry`

`GET /api/ai/conversations/{id}` also returns the conversation's `latest_workout_draft` when the assistant has produced a structured workout draft for that thread.

The stream endpoint is authenticated fetch-based SSE:

- preflight failures return normal JSON errors with non-2xx status
- successful requests switch to `text/event-stream`
- post-start failures emit SSE `error` events
- streaming assistant text is snapshotted into app-owned storage during long runs so a dropped client can reload persisted partial progress
- the client recovery path is still storage-backed inspection, not live SSE replay; interrupted sessions poll the persisted conversation until the run reaches a terminal state
- stale `streaming` runs older than the stream timeout grace window are auto-failed before a new send starts, which prevents a permanently blocked conversation after an interrupted server-side run
- `POST /api/ai/conversations/{id}/messages/recover` returns `503` until background recovery is configured
- interrupted runs can enqueue background recovery through Inngest when both `INNGEST_EVENT_KEY` and `INNGEST_SIGNING_KEY` are configured

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
docker compose logs postgres

# Check database health
docker compose ps
```

**Missing goose:**

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export PATH="$HOME/go/bin:$PATH"
```

**Postgres auth failures after changing DB_USER/DB_PASSWORD or switching branches:**

This repo bind-mounts PostgreSQL data into `server/_db-data`. If that folder was initialized with different `DB_USER`, `DB_PASSWORD`, or `DB_NAME` values, Postgres may still start but `make dev` and `goose` will fail authentication.

PowerShell:

```powershell
cd server
docker compose down
Remove-Item -Recurse -Force _db-data
New-Item -ItemType Directory _db-data | Out-Null
```

Bash:

```bash
cd server
docker compose down
rm -rf _db-data
mkdir -p _db-data
```

After that, rerun `make dev` so Postgres initializes with the values from `setenv.sh`.

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
