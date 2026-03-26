## Development

### Quick Start

1. Create `setenv.sh` in this folder with:
```bash
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=fittrack
export DB_PORT=55432
export DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:55432/fittrack
```

2. Install migration CLI:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
export PATH="$HOME/go/bin:$PATH"
```

3. Run the complete development environment:
```bash
make dev
```

That's it! This single command will:
- ✅ Load environment variables from `setenv.sh`
- ✅ Check prerequisites (docker, docker compose, air)
- ✅ Start the PostgreSQL database container
- ✅ Wait for the database to be ready
- ✅ Start the hot-reload development server with air

### Manual Setup (Advanced)

If you prefer to run steps individually:

1. Load environment variables:
```bash
source ./setenv.sh
```

2. Start database:
```bash
make docker-up
```

3. Initialize the database (first time only):
```bash
cat schema.sql | docker exec -i db psql -U ${DB_USER} -d ${DB_NAME}
# OR run migrations:
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

The command respects existing shell env first, then loads `server/.env.local`, `server/.env`, and `server/setenv.sh` (or `server/.setenv.sh`) if present. It sends one real Genkit request to Gemini, times out after 20 seconds, and prints a short model response to stdout.

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
make dev               # Complete dev setup: env → docker up → wait DB → air
make build             # Compile the project
make run               # Run the compiled binary
make swagger           # Generate OpenAPI documentation
make sqlc              # Generate SQL code from query.sql
make test              # Run all tests
make vet               # Run go vet
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback last migration
make migrate-create    # Create a new migration file (NAME=<name>)
make docker-up         # Start PostgreSQL container
make docker-down       # Stop PostgreSQL container
make clean             # Clean build files and cache
```

## Documentation

- [Row Level Security (RLS)](docs/rls.md) - Multi-tenant data isolation and security
