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

1. Create `server/setenv.sh` and configure DB vars (example):
   ```bash
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=fittrack
   export DB_PORT=55432
   export DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:55432/fittrack
   ```
2. Install goose and ensure PATH includes `$HOME/go/bin`:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   export PATH="$HOME/go/bin:$PATH"
   ```
3. Start services:
   ```bash
   cd server
   . setenv.sh
   make docker-up
   ```
4. Initialize database:
   ```bash
   cat schema.sql | docker exec -i db psql -U ${DB_USER} -d ${DB_NAME}
   ```
5. Run migrations:
   ```bash
   make migrate-up
   ```
6. Start API server:
   ```bash
   go run ./cmd/api
   ```

#### Frontend Setup

```bash
cd client
bun install
bun run dev        # starts on http://localhost:5173 (proxies API to :8080)
# or: bun run start  # starts on http://localhost:3000
```

After signing in, the minimal phase-1 chat proof is available at `/chat`.

### Available Commands

#### Backend (server/)
```bash
make help             # Show all available commands
make dev              # Complete dev setup (docker + hot reload via air)
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
