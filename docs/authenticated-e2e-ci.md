# Required authenticated browser coverage

`Run Required Authenticated E2E` in `.github/workflows/authenticated-e2e.yml` runs every
Playwright test discovered under `client/tests/e2e/auth/`. New authenticated tests
belong there. The ordinary frontend job keeps production-preview guest coverage,
including its separate known guest fixme.

The authenticated job uses a fresh PostgreSQL 15.19 service container, applies all
Goose migrations, and starts the real Go API. It uses the documented local E2E
owner connection: existing dev fixture bootstrap writes without a request user
context and is not compatible with the restricted runtime role. This job does
not prove runtime-role RLS isolation; the backend job retains that separate proof.
The production enforcement flag also requires Supabase pooler transport and stays
off for local PostgreSQL. Ports 25432 (database), 25080 (API),
25090 (metrics) and 25173 (frontend) are local to that runner. Credentials are
disposable literals for this container. The job needs no GitHub secrets and runs
on `pull_request`, including forks and dependency branches, with read-only repository permission.
The separate workflow lets a stacked PR exercise this job before its prerequisite
is merged; the existing guest workflow remains unchanged.

`CI=true` retains Playwright's worker, timeout, retry and forbid-only policies.
`E2E_FRONTEND_MODE=development` explicitly selects Vite development mode instead
of CI's default production preview. The application's `import.meta.env.DEV` and
backend development-only authentication guards remain in force. Never enable
these local fixture routes on a shared preview or production API.

The public Stack project/client key are copied from the deployed public frontend
on November 4, 2026. They initialize StackProvider and the API's public JWKS cache;
they are not server secrets or real-user credentials. Startup still depends on
Stack's public JWKS endpoint being available. If those public settings change,
update them deliberately; an initialization failure must fail this job. The job
does not create Stack users or sessions, contact Stripe, or generate AI responses.

Global setup bootstraps only `ci-synthetic-user` in the disposable database, grants
its fixture entitlement, then observes the application's own authenticated
`GET /api/workouts` before saving browser state. Tests exercise real workout CRUD,
exercise catalog classification at desktop/mobile sizes, structured conversation
import, and history deletion after access expires. One worker serializes shared
fixture mutation; fixture endpoints prepare data rather than stand in for the
browser/API assertions.

## Exit policy and local verification

From `client/`, run `bun run test:e2e:auth` with the job's environment, a migrated
isolated API/database, and installed Chromium. Playwright owns and stops Vite;
the caller owns and stops the API/database. A linked local worktree must use the
managed setup in [worktrees.md](worktrees.md) or an equivalent fresh cluster and
exclusive ports. Never point this suite at shared fixture-user history.

`playwright.auth.config.ts` selects the whole auth directory without a copied test
inventory. Playwright 1.63 has no built-in fail-on-skip option in its [CLI](https://playwright.dev/docs/test-cli).
The required-suite reporter uses [Reporter.onEnd](https://playwright.dev/docs/api/class-reporter#reporter-on-end)
to check final framework results: zero cases, skips/fixmes, unexecuted
cases, expected failures, flaky outcomes and ordinary failures all fail the job.
It never converts a framework failure into success. The console records selected,
passed and skipped/unexecuted totals; HTML, traces, screenshots and API logs are
uploaded even on failure. Auth storage state is deliberately excluded from uploads.

`node --test scripts/required-auth-gate.test.mjs` starts real Playwright processes
against temporary test files to verify success, missing-auth/runtime skip, static
skip, assertion failure, retry-to-skip, flaky retry-to-pass, expected failure and empty selection
(including `--pass-with-no-tests`). These checks require locked client dependencies
but no browser, API or database. The CI job runs them before the product suite.

This job proves local authenticated product behavior. External Stack login,
Stripe checkout, live AI providers and hosted deployment remain separate proof.
