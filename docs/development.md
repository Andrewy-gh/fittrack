# Development Guide

This guide covers the repo-level workflow for running FitTrack locally and keeping backend and frontend contracts in sync.

## Quick Start

### Backend Setup

1. Create your local server env files from the examples:

```bash
cd server
cp setenv.example.sh setenv.sh
cp .env.example .env
```

2. Update the copied values:

```bash
# server/setenv.sh powers make dev, make migrate-*, and make test-short
export PROJECT_ID=your-stack-project-id
export DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:55432/fittrack?sslmode=disable

# server/.env is useful for local integration tests, Playwright setup, and AI smoke tests
PROJECT_ID=your-stack-project-id
DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:55432/fittrack?sslmode=disable
```

3. Install goose and air, and ensure PATH includes `$HOME/go/bin`:

```bash
go install github.com/pressly/goose/v3/cmd/goose@v3.24.3
go install github.com/air-verse/air@v1.65.1
export PATH="$HOME/go/bin:$PATH"
```

4. Start services:

```bash
cd server
make dev
```

5. If you prefer to run the database step separately, you can still apply migrations manually:

```bash
make migrate-up
```

### Frontend Setup

```bash
cd client
cp .env.example .env
# then set VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY
bun install
bun run prepare   # optional: enable local Husky git hooks
bun run dev        # starts on http://localhost:5173 (proxies API to :8080)
# or: bun run start  # starts on http://localhost:3000
```

The client uses `client/bunfig.toml` to skip install scripts and delay newly published package versions by default. Husky is still installed as a dev dependency; `bun run prepare` enables the local git hooks when you want them.

After signing in, the minimal phase-1 chat proof is available at `/chat`.

## API Type Workflow

To maintain type safety between backend and frontend, follow this workflow when modifying API endpoints.

### 1. Modify Go Structs

After changing any Go structs used in API endpoints, including request and response types:

```bash
cd server
make swagger
```

This regenerates the OpenAPI documentation from your Go code annotations.

### 2. Update API Client

After swagger documentation is updated:

```bash
cd client
bun run openapi-ts
```

This generates TypeScript client code and types from the updated swagger specification.

### 3. Commit Generated Code

Always commit the generated code to prevent drift:

```bash
git add server/docs/ client/src/client/
git commit -m "feat: update API types for your-change"
```

This workflow keeps frontend and backend types in sync, makes API changes visible in code reviews, and catches type mismatches before runtime.

## Local Structured Chat E2E

For local browser E2E runs, you can skip manual Google or GitHub login by enabling the repo's local-only Playwright bootstrap path.

Add these server vars in `server/.env` or `server/setenv.sh`:

```env
E2E_LOCAL_AUTH_ENABLED=true
E2E_LOCAL_AUTH_USER_ID=local-e2e-user
E2E_LOCAL_AUTH_EMAIL=local-e2e-user@example.test
E2E_LOCAL_AUTH_DISPLAY_NAME=Local E2E User
```

Add this client var in `client/.env`:

```env
VITE_E2E_LOCAL_AUTH_ENABLED=true
```

Then run the normal local servers and Playwright:

```bash
cd server && make dev
cd client && bun run test:e2e -- tests/e2e/auth/structured-workout-chat-import.test.ts
```

That flow:

- bootstraps one deterministic local test user
- grants that user `ai_chatbot` feature access
- seeds a persisted chat conversation with a structured workout draft
- reopens the chat in the browser and imports the draft into `/workouts/new`

## Available Commands

### Backend

Run these from `server/`.

```bash
make help             # Show all available commands
make dev              # Complete dev setup: docker, migrations, and hot reload via air
make build            # Compile the project
make run              # Run the compiled binary
make swagger          # Generate OpenAPI documentation
make sqlc             # Generate SQL code from queries
make test             # Run tests
make vet              # Run go vet
make migrate-up       # Run database migrations
make migrate-down     # Rollback last migration
make migrate-create   # Create a new migration
make docker-up        # Start PostgreSQL container
make docker-down      # Stop PostgreSQL container
make clean            # Clean build files and cache
```

### Frontend

Run these from `client/`.

```bash
bun run dev           # Development server on http://localhost:5173
bun run start         # Development server on port 3000
bun run build         # Production build
bun run serve         # Preview production build
bun run openapi-ts    # Generate API client from OpenAPI spec
bun run test          # Run unit tests with Vitest
bun run test:e2e      # Run end-to-end tests with Playwright
bun run lint          # Lint with oxlint
bun run knip          # Check for unused files and dependency drift
bun run tsc           # Type-check without emitting
```

## Checks Before Handoff

Run the full repo-defined gate when practical:

- backend: `make test`
- frontend: `bun run lint && bun run knip && bun run test && bun run tsc`
- docs or build checks when the change touches deploy, docs generation, or production output
