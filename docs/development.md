# Development Guide

This guide covers the repo-level workflow for running FitTrack locally and keeping backend and frontend contracts in sync.

## Worktree-local development

For concurrent linked worktrees sharing a local PostgreSQL server, see the
[worktree setup guide](worktrees.md). Run its `init`, `db setup --major`, and
`run` commands in each worktree; creating a worktree alone is not setup. Local
PostgreSQL 15.19 is pinned for local development and CI; the guide includes the
shared Docker startup, version checks, and safe fresh-storage transition. The Docker quick start below is
the existing single-checkout workflow.

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

### Generate and verify the API client

After changing API structs, annotations, or generator configuration, run the combined command from the repository root:

```bash
bun install --cwd client --frozen-lockfile
bun run --cwd client api:generate
```

The command runs the pinned `swag` CLI version `v1.16.6`, then the locked
`@hey-api/openapi-ts` package and its formatter/linter post-processing. It writes
Swagger JSON/YAML and the generated TypeScript client. `server/docs/docs.go` is
a custom embedding wrapper and is intentionally not generated or replaced.

Review and stage the output with the API change. The pre-commit hook runs this
same generator when staged backend Go source, `server/go.mod`/`go.sum`,
generation configuration/scripts, or generated contract files change. It stops
before writing if relevant files have unstaged edits, including partial staging.
If generated files change, it leaves them for review and fails the commit with
the exact paths and `git add` guidance; it never stages files automatically.
Stage reviewed outputs and retry. Ordinary frontend, Markdown, migration, and
backend test-only changes do not run generation.

CI runs the same generation chain into temporary output and compares every
generated path with the checkout. It reports changed, missing, and unexpected
files and does not write generated files or create bot commits. The job runs on
every pull request so path filters cannot omit the freshness check.

The generator version is pinned in `scripts/api-contract.mjs`; update it
deliberately with regenerated Swagger files and a freshness-check run. Client
generator dependencies remain pinned by `client/package.json` and `bun.lock`.
The first run needs Go module download access and frozen Bun dependencies. A
missing tool or failed generator exits nonzero before publishing any output.

### Commit Generated Code

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

Authenticated Playwright setup now verifies the app's own browser request to
`GET /api/workouts` before saving auth state or running tests. Local auth needs
the Vite **development** server, both public Stack client settings
(`VITE_PROJECT_ID` and `VITE_PUBLISHABLE_CLIENT_KEY`), and
`VITE_E2E_LOCAL_AUTH_ENABLED=true`. Restart Vite after changing those settings.
Do not use `CI=true`/production preview with local auth; preview requires real
Stack authentication. A successful bootstrap or an existing auth file alone
does not establish browser authentication.

With no auth configured, a clean checkout still runs guest checks and reports
that authenticated coverage is unverified. If a previous
`client/tests/e2e/.auth/stack.json` exists, setup stops: configure fresh auth or
archive the old file before a guest-only run. Failed readiness leaves existing
state untouched. For concurrent worktrees, use the isolated E2E setup in
[worktrees.md](worktrees.md), rather than the single-checkout commands above.

## Landing and README Media

The landing page's "How it works" clips and the README screenshots/GIF are recorded from the real app in guest demo mode, so they match the current UI. Re-record them after UI changes to the workout form, workout detail, or analytics pages.

Requires `ffmpeg` (with `libx264` and `libwebp`) on your PATH.

```bash
cd client
bun run build && bun run serve:test   # terminal 1
bun run media:capture                  # terminal 2
```

The script seeds 12 weeks of sample training into demo storage (browser-only; the app's built-in demo data is untouched), records the log, repeat, and review flows in light and dark, and writes:

- `client/public/media/how-it-works/{log,repeat,review}-{light,dark}.{mp4,webp}`: landing clips and poster frames
- `docs/media/app-screens-{light,dark}.png` and `docs/media/how-it-works.gif`: README artifacts

Pass clip names to re-record a subset, for example `bun run media:capture review readme`.

## Available Commands

### Backend

Run these from `server/`.

```bash
make help             # Show all available commands
make dev              # Complete dev setup: docker, migrations, and hot reload via air
make build            # Compile the project
make run              # Run the compiled binary
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
bun run api:generate  # Regenerate Swagger and the TypeScript API client
bun run test          # Run unit tests with Vitest
bun run test:e2e      # Run end-to-end tests with Playwright
bun run lint          # Lint with oxlint
bun run knip          # Check for unused files and dependency drift
bun run knip:production # Check for code unused by the production app
bun run tsc           # Type-check without emitting
```

## Checks Before Handoff

Run the full repo-defined gate when practical:

- backend: `make test`
- frontend: `bun run lint && bun run knip && bun run knip:production && bun run test && bun run tsc`
- docs or build checks when the change touches deploy, docs generation, or production output
