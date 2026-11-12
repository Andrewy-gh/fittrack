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
- `VITE_ENABLE_TRAINING_GUIDANCE`: shows the Training assessment (including muscle sets), Today's suggestion, and saved-workout suggestion panels only when set to `true`; defaults to `false`. This is a client build-time flag, so changing it requires rebuilding the frontend.

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

The production workflow, [fly-deploy.yml](../.github/workflows/fly-deploy.yml),
currently reads these GitHub Actions secrets:

- `DATABASE_URL`
- `VITE_PROJECT_ID`
- `VITE_PUBLISHABLE_CLIENT_KEY`
- `FLY_API_TOKEN`

Stripe settings are **runtime configuration**, not secrets consumed or synced by
this workflow. Billing still requires `STRIPE_SECRET_KEY`,
`STRIPE_WEBHOOK_SECRET`, and `STRIPE_PREMIUM_PRICE_ID` on the deployed API;
configure them separately in the Fly runtime secret store. Merely adding them
to GitHub Actions does not configure billing. Preserve existing runtime settings;
no checked-in workflow currently supplies Stripe credentials. See
[Stripe Billing](stripe-billing.md) for runtime setup and verification.

Preview deploys use these secrets in the GitHub `review` environment, copied from
Doppler project `fittrack`, config `stg` (storage in Doppler alone does not sync them):

- `PREVIEW_MIGRATION_DATABASE_URL`: owning `postgres.<project-ref>` connection; only Goose receives it
- `PREVIEW_DATABASE_URL`: restricted `fittrack_app.<project-ref>` runtime connection
- `PREVIEW_METRICS_USERNAME`
- `PREVIEW_METRICS_PASSWORD`

Set the `review` environment variable `PREVIEW_SUPABASE_PROJECT_REF` to the
nonproduction Supabase project reference. Both database URLs must use that project,
the same shared session-pooler host on port `5432`, database `postgres`, and
`sslmode=require`. The workflow validates these before migrations. Bootstrap a new
database with migrations and `server/scripts/provision-runtime-role.sql`, then
configure the runtime login/password separately; do not use its credential for Goose.

All preview apps reuse this one database; deploying or closing a PR never resets it.
Deployments are serialized across PRs, but existing apps continue serving while a
migration runs. Keep migrations compatible with those apps and coordinate schema
changes across PRs: serialization does not reconcile divergent migration histories
or isolate test data. Use isolated local databases for destructive tests.

The workflow fails on migration errors and requires deployed `/ready` to report
`healthy` with database `ok`. Closing a PR only removes its Fly app and skips
migrations/readiness. It does not remove the shared database.

Fly preview apps set these runtime vars on the server:

- `DATABASE_URL`
- `PROJECT_ID`
- `ENVIRONMENT=staging`
- `RLS_ENFORCEMENT_REQUIRED=true`
- `DB_MAX_CONNS=3` and `DB_MIN_CONNS=0` to limit aggregate preview connections
- `METRICS_USERNAME`
- `METRICS_PASSWORD`

AI and Stripe credentials are not supplied by the preview workflow. Database
readiness does not verify provider login, AI access, or billing scenarios.

Production note:

- `PROJECT_ID` should match the Stack project behind `VITE_PROJECT_ID`
- `DATABASE_URL` should use the production application database role, not a personal or admin connection string
