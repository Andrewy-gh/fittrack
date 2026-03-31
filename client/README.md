# FitTrack Client

React frontend for FitTrack, built with Vite, TypeScript, TanStack Router, and TailwindCSS v4.

## Getting Started

```bash
cp .env.example .env
# set VITE_PROJECT_ID and VITE_PUBLISHABLE_CLIENT_KEY first
bun install
bun run dev       # http://localhost:5173 (proxies API calls to :8080)
# or
bun run start     # http://localhost:3000
```

## Available Commands

```bash
bun run dev           # Development server (Vite default port)
bun run start         # Development server on port 3000
bun run build         # Production build
bun run serve         # Preview production build
bun run openapi-ts    # Generate API client from OpenAPI spec
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

## PWA

The app ships as a Progressive Web App. The service worker caches API responses and serves the app offline.
