# FitTrack

A full-stack fitness tracking application with Go backend and React frontend.

## Architecture

- **Backend**: Go API with PostgreSQL database
- **Frontend**: React with TypeScript, TanStack Router, and TailwindCSS
- **API**: OpenAPI/Swagger specification with auto-generated client code

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
npm run generate:api
```

This generates TypeScript client code and types from the updated swagger specification.

#### 3. Commit Generated Code

**Important**: Always commit the generated code to prevent drift:

```bash
git add server/docs/ client/src/generated/
git commit -m "feat: update API types for [your change]"
```

### Why This Workflow Matters

- **Type Safety**: Ensures frontend and backend stay in sync
- **Explicit Regeneration**: Makes API changes visible in code reviews
- **Prevents Runtime Errors**: Catches type mismatches at compile time
- **Documentation**: Keeps API documentation current

### Quick Start

#### Backend Setup

1. Copy `server/setenv.example.sh` to `server/setenv.sh` and configure
2. Start services:
   ```bash
   cd server
   . setenv.sh
   make docker-up
   ```
3. Initialize database:
   ```bash
   cat schema.sql | docker exec -i db psql -U ${DB_USER} -d ${DB_NAME}
   ```
4. Run migrations:
   ```bash
   make migrate-up
   ```
5. Start API server:
   ```bash
   go run ./cmd/api
   ```

#### Frontend Setup

```bash
cd client
bun install
bun run start
```

### Available Commands

#### Backend (server/)
```bash
make help          # Show all available commands
make build         # Compile the project
make swagger       # Generate OpenAPI documentation
make sqlc          # Generate SQL code from queries
make test          # Run tests
make migrate-up    # Run database migrations
make docker-up     # Start PostgreSQL container
```

#### Frontend (client/)
```bash
npm run dev           # Development server
npm run build         # Production build
npm run generate:api  # Generate API client from OpenAPI spec
npm run test          # Run tests
```

## Project Structure

```
├── server/           # Go backend
│   ├── cmd/api/     # API entry point
│   ├── docs/        # Generated OpenAPI docs
│   ├── internal/    # Internal packages
│   └── migrations/  # Database migrations
├── client/          # React frontend
│   ├── src/
│   │   └── generated/  # Generated API client
│   └── package.json
└── config/          # Shared configuration
```

## Contributing

1. Create a feature branch
2. Make your changes following the type-safe workflow above
3. Ensure tests pass: `make test` (backend) and `npm run test` (frontend)
4. Submit a pull request

## Documentation

- [Backend Setup](server/README.md) - Detailed backend development guide
- [Frontend Setup](client/README.md) - React/TanStack development guide
- [Database RLS](server/docs/rls.md) - Row Level Security documentation
