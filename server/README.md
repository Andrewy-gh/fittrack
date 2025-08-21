## Development

### Quick Start

1. Copy `setenv.example.sh` to `setenv.sh` and modify it to your liking.

2. Run the complete development environment:
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

**Missing air:**
```bash
# Install via Go
go install github.com/air-verse/air@latest

# Or via Homebrew (macOS)
brew install air
```

## Documentation

- [Row Level Security (RLS)](docs/rls.md) - Multi-tenant data isolation and security
