# FitTrack Client

React frontend for FitTrack, built with Vite, TypeScript, TanStack Router, and TailwindCSS v4.

## Getting Started

```bash
cp .env.example .env
# set VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY first
bun install
bun run prepare  # optional: enable local Husky git hooks
bun run dev       # http://localhost:5173 (proxies API calls to :8080)
# or
bun run start     # http://localhost:3000 unless FRONTEND_PORT is set
```

`bunfig.toml` skips install scripts and delays newly published package versions by default. Husky is still installed through dev dependencies; run `bun run prepare` when you want local git hooks enabled.

## Available Commands

```bash
bun run dev           # Development server (FRONTEND_PORT, default 5173)
bun run start         # Development server (FRONTEND_PORT, default 3000)
bun run build         # Production build
bun run serve         # Preview production build (PREVIEW_PORT, default 4173)
bun run serve:test    # Preview on E2E_PORT (default 5173)
bun run openapi-ts    # Generate API client from OpenAPI spec
bun run fmt           # Format client files with oxfmt
bun run fmt:check     # Check client formatting with oxfmt
bun run test          # Run unit tests (Vitest)
bun run test:e2e      # Run end-to-end tests (Playwright)
bun run test:e2e:ci   # Build then run E2E tests (CI mode)
bun run lint          # Lint with oxlint
bun run knip          # Check for unused files and dependency drift
bun run tsc           # Type-check without emitting
```

## Authentication

This project uses [Stack Auth](https://stack-auth.com/) for authentication. Configure the following in `.env`:

```env
VITE_PROJECT_ID=<your-stack-project-id>
VITE_PUBLISHABLE_CLIENT_KEY=<your-stack-publishable-key>
```

Optional:

```env
VITE_API_BASE_URL=http://localhost:8080
```

You usually do not need `VITE_API_BASE_URL` during local dev because Vite proxies `/api` to the backend by default.

### Local ports and worktrees

Export `FRONTEND_PORT` (default `5173`), `PREVIEW_PORT` (default `4173`),
`E2E_PORT` (default `5173`), and `API_PROXY_TARGET` (default
`http://127.0.0.1:8080`) in the command environment. These settings are not
browser-visible `VITE_` variables, and Vite's `.env` files are not loaded by the
server configuration. The repo's worktree helper supplies them automatically;
see [the worktree setup guide](../docs/worktrees.md).

`bun run start` keeps its legacy default port `3000`, but honors an exported
`FRONTEND_PORT` just like `bun run dev`.

```bash
FRONTEND_PORT=24000 API_PROXY_TARGET=http://127.0.0.1:24001 bun run dev
PREVIEW_PORT=24006 API_PROXY_TARGET=http://127.0.0.1:24001 bun run serve
E2E_PORT=24003 API_PROXY_TARGET=http://127.0.0.1:24004 bun run test:e2e
```

Development and preview servers bind to `127.0.0.1` by default and fail when their requested port is occupied;
they never silently move to the next port. Playwright launches its server on
`E2E_PORT`, even when the development and preview port settings differ, and
refuses to reuse an existing server by default. Use
`E2E_REUSE_EXISTING_SERVER=true` only when you deliberately want that server.

An explicit `E2E_BASE_URL` must use HTTP `localhost` or `127.0.0.1` on
`E2E_PORT` for a managed server. A different origin requires
`E2E_REUSE_EXISTING_SERVER=true`; Playwright then tests that existing target
without launching a local server. Local auth bootstrap also uses
`API_PROXY_TARGET` unless `E2E_LOCAL_AUTH_API_BASE_URL` is set explicitly.

For local Playwright auth bootstrap without manual social login, also set:

```env
VITE_E2E_LOCAL_AUTH_ENABLED=true
```

## API Client Generation

TypeScript client code is auto-generated from the backend's OpenAPI spec. After the backend swagger docs are updated, regenerate with:

```bash
bun run openapi-ts
```

This reads `../server/docs/swagger.json` and outputs to `src/client/`. Always commit the generated files.

## Adding shadcn Components

```bash
bunx shadcn@latest add button
```

## Routing

File-based routing via TanStack Router. Routes live in `src/routes/`. The router Vite plugin auto-generates the route tree on file changes.

## Testing

- **Unit tests**: Vitest with jsdom - `bun run test`
- **E2E tests**: Playwright (Chrome) - `bun run test:e2e`
- **Project hygiene**: oxlint and knip - `bun run lint` and `bun run knip`
- E2E tests live in `tests/e2e/`

### Local Structured Chat E2E

When the server also has `E2E_LOCAL_AUTH_ENABLED=true`, Playwright global setup can create a deterministic local test session instead of using Google, GitHub, or email/password login.

Run the structured chat reopen/import E2E with:

```bash
bun run test:e2e -- tests/e2e/auth/structured-workout-chat-import.test.ts
```

## PWA

The app ships as a Progressive Web App. The service worker caches API responses and serves the app offline.
