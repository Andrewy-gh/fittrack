# Environment Reference

This reference lists the variables used for local development, testing, preview deploys, and production deploys.

## Local App Runtime

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
- `STRIPE_SECRET_KEY`: enables Stripe customer and checkout session creation
- `STRIPE_WEBHOOK_SECRET`: verifies Stripe webhook signatures
- `STRIPE_PREMIUM_PRICE_ID`: Stripe recurring price that grants FitTrack premium AI chat access
- `APP_BASE_URL`: absolute frontend URL used for Stripe checkout success and cancel redirects, defaults to `http://localhost:5173`
- `AI_CHAT_TRIAL_PROMPT_CAP`: number of AI chat prompts allowed while a Stripe subscription is `trialing`, defaults to `30`
- `VITE_API_BASE_URL`: points the client at a non-default API base URL

See [Stripe Billing](stripe-billing.md) for the checkout and webhook integration details.

## Worktree-local commands

See [worktree setup](worktrees.md) for the initial helper and its provisioning
limits. These variables are read from the exported process environment:

- `FRONTEND_PORT`: Vite development port (default `5173`)
- `PREVIEW_PORT`: Vite preview port (default `4173`)
- `API_PROXY_TARGET`: API origin used by both dev and preview proxies (default `http://127.0.0.1:8080`)
- `E2E_PORT`: managed Playwright frontend port (default `5173`)
- `E2E_BASE_URL`: Playwright target URL; managed local launches must match `E2E_PORT`
- `E2E_REUSE_EXISTING_SERVER=true`: explicit opt-in to an existing test target; disabled by the worktree wrapper
- `E2E_LOCAL_AUTH_API_BASE_URL`: local auth bootstrap API origin, without `/api`
- `FITTRACK_LOCAL_DATABASE_URL`: existing owning-role URL for the shared local server at `127.0.0.1:55432`; the wrapper replaces its database name
- `FITTRACK_LOCAL_RUNTIME_DATABASE_URL`: existing `fittrack_app` URL, required only for the isolated `rls` mode

The wrapper generates mode-specific `DATABASE_URL`, API/metrics ports, frontend
origins, and proxy settings without saving credentials. It forces
`VITE_API_BASE_URL=/api` so local clients use the same-origin proxy. For direct
API access outside this wrapper, `VITE_API_BASE_URL` must include the `/api`
prefix, and `ALLOWED_ORIGINS` must allow the frontend origin.

## Local Testing

- `server/setenv.sh`: used by `make dev`, `make migrate-up`, `make migrate-down`, and `make test-short`
- `server/.env`: read by several Go integration tests, the Gemini smoke test, and Playwright setup helpers
- `client/.env`: read by the frontend and Playwright setup helpers
- `E2E_LOCAL_AUTH_ENABLED`: enables the local-only Playwright auth bootstrap on the server in `development`
- `E2E_LOCAL_AUTH_USER_ID`, `E2E_LOCAL_AUTH_EMAIL`, `E2E_LOCAL_AUTH_DISPLAY_NAME`: optional local test-user overrides for that bootstrap
- `VITE_E2E_LOCAL_AUTH_ENABLED`: lets the client trust the Playwright-seeded local auth session in dev
- `SECRET_SERVER_KEY`: optional for Playwright; lets tests create Stack Auth sessions directly instead of logging in through the UI
- `E2E_STACK_EMAIL` and `E2E_STACK_PASSWORD`: optional fallback for Playwright UI login

## Production Deploy

GitHub Actions currently expect these secrets:

- `DATABASE_URL`
- `VITE_PROJECT_ID`
- `VITE_PUBLISHABLE_CLIENT_KEY`
- `FLY_API_TOKEN`
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_PREMIUM_PRICE_ID`

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
- `STRIPE_SECRET_KEY`
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_PREMIUM_PRICE_ID`
- `APP_BASE_URL`
- `AI_CHAT_TRIAL_PROMPT_CAP`

Production note:

- `PROJECT_ID` should match the Stack project behind `VITE_PROJECT_ID`
- `DATABASE_URL` should use the production application database role, not a personal or admin connection string
