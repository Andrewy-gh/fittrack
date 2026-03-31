# FitTrack

A full-stack fitness tracking application with Go backend and React frontend.

## Architecture

- **Backend**: Go API with PostgreSQL database
- **Frontend**: React with TypeScript, TanStack Router, and TailwindCSS
- **Auth**: Stack authentication with JWT verification and Row Level Security
- **API**: OpenAPI/Swagger specification with auto-generated client code
- **AI chat**: In-process Go API slice under `/api/ai/*` with Postgres-owned conversation state and fetch-based SSE

## Developer Workflow

### Preventing Type Drift

To maintain type safety between backend and frontend, follow this workflow when modifying API endpoints:

#### 1. Modify Go Structs (Backend)

After changing any Go structs used in API endpoints (request/response types):

```bash
cd server
make swagger
```

This regenerates the OpenAPI documentation from your Go code annotations.

#### 2. Update API Client (Frontend)

After swagger documentation is updated:

```bash
cd client
bun run openapi-ts
```

This generates TypeScript client code and types from the updated swagger specification.

#### 3. Commit Generated Code

**Important**: Always commit the generated code to prevent drift:

```bash
git add server/docs/ client/src/client/
git commit -m "feat: update API types for [your change]"
```

### Why This Workflow Matters

- **Type Safety**: Ensures frontend and backend stay in sync
- **Explicit Regeneration**: Makes API changes visible in code reviews
- **Prevents Runtime Errors**: Catches type mismatches at compile time
- **Documentation**: Keeps API documentation current

### Quick Start

#### Backend Setup

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
   go install github.com/pressly/goose/v3/cmd/goose@latest
   go install github.com/air-verse/air@latest
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

#### Frontend Setup

```bash
cd client
cp .env.example .env
# then set VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY
bun install
bun run dev        # starts on http://localhost:5173 (proxies API to :8080)
# or: bun run start  # starts on http://localhost:3000
```

After signing in, the minimal phase-1 chat proof is available at `/chat`.

## Environment Cheat Sheet

### Local app runtime

Minimum server vars:

- `DATABASE_URL`: Postgres connection string for the API
- `PROJECT_ID`: Stack Auth project ID used by the backend to verify tokens

Minimum client vars:

- `VITE_PROJECT_ID`: same Stack Auth project ID, but exposed to the browser
- `VITE_PUBLISHABLE_CLIENT_KEY`: Stack Auth publishable browser key

Useful optional local vars:

- `ALLOWED_ORIGINS`: needed when the frontend is not using the default Vite proxy or same-origin `/api`
- `GEMINI_API_KEY` or `GOOGLE_API_KEY`: enables AI chat and `go run ./cmd/gemini-smoke`
- `GEMINI_MODEL`: overrides the default Gemini model
- `VITE_API_BASE_URL`: points the client at a non-default API base URL

### Local testing

- `server/setenv.sh`: used by `make dev`, `make migrate-up`, `make migrate-down`, and `make test-short`
- `server/.env`: read by several Go integration tests, the Gemini smoke test, and Playwright setup helpers
- `client/.env`: read by the frontend and Playwright setup helpers
- `SECRET_SERVER_KEY`: optional for Playwright; lets tests create Stack Auth sessions directly instead of logging in through the UI
- `E2E_STACK_EMAIL` and `E2E_STACK_PASSWORD`: optional fallback for Playwright UI login

### Production deploy

GitHub Actions currently expect these secrets:

- `DATABASE_URL`
- `VITE_PROJECT_ID`
- `VITE_PUBLISHABLE_CLIENT_KEY`
- `FLY_API_TOKEN`

Preview deploys also use:

- `PREVIEW_DATABASE_URL`
- `PREVIEW_METRICS_USERNAME`
- `PREVIEW_METRICS_PASSWORD`

Fly preview apps set these runtime vars on the server:

- `DATABASE_URL`
- `PROJECT_ID`
- `ENVIRONMENT=staging`
- `METRICS_USERNAME`
- `METRICS_PASSWORD`

Production note:

- `PROJECT_ID` should match the Stack project behind `VITE_PROJECT_ID`
- `DATABASE_URL` should use a non-superuser app role in production so Postgres row-level security keeps working

### Available Commands

#### Backend (server/)
```bash
make help             # Show all available commands
make dev              # Complete dev setup (docker + migrations + hot reload via air)
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

#### Frontend (client/)
```bash
bun run dev           # Development server (http://localhost:5173)
bun run start         # Development server on port 3000
bun run build         # Production build
bun run serve         # Preview production build
bun run openapi-ts    # Generate API client from OpenAPI spec
bun run test          # Run unit tests (Vitest)
bun run test:e2e      # Run end-to-end tests (Playwright)
bun run lint          # Lint with oxlint
bun run knip          # Check for unused files and dependency drift
bun run tsc           # Type-check without emitting
```

## Project Structure

```
.
+-- server/           # Go backend
|   +-- cmd/api/      # API entry point
|   +-- docs/         # Generated OpenAPI docs
|   +-- internal/     # Internal packages
|   +-- migrations/   # Database migrations
+-- client/           # React frontend
|   +-- src/
|   |   +-- client/     # Generated API client (from OpenAPI spec)
|   +-- package.json
+-- docs/             # Design and API notes
```

## Contributing

1. Create a feature branch
2. Make your changes following the type-safe workflow above
3. Ensure checks pass: `make test` (backend) and `bun run lint && bun run knip && bun run test` (frontend)
4. Submit a pull request

## Documentation

- [Backend Setup](server/README.md) - Detailed backend development guide
- [Frontend Setup](client/README.md) - React/TanStack development guide
