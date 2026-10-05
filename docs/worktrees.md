# Worktree-local development (initial increment)

One local PostgreSQL server can serve several FitTrack worktrees. Each worktree
needs its own databases and application ports. Database isolation is necessary:
the Go tests toggle RLS and delete fixtures, so a test suite must not use your
persistent development database.

This first increment allocates names and ports, supplies isolated command
environments, and makes Vite/Playwright honor them. It **does not install or
start PostgreSQL, create databases or roles, set passwords, run migrations, or
start the app for you**. Database provisioning is a separate next step, after the
production PostgreSQL version is verified.

## What works now

From the worktree root, with the repository's Node version available (the
command runner currently supports Linux, macOS, and WSL):

```bash
node scripts/worktree.mjs init
node scripts/worktree.mjs status
node --test scripts/worktree.test.mjs
```

`init` writes an ignored `.worktree/config.json` containing no credentials. It
reserves a ten-port block in a locked registry under the repository's shared Git
common directory (`fittrack-worktrees/registry.json`). Linked Git worktrees share
that registry. Branch changes keep the same allocation. The worktree ID combines
a short directory label with a hash of the canonical absolute path; moving a
worktree changes its identity and leaves the old reservation intact.

The block starts in the range 21000–29999:

| Offset | Service |
| --- | --- |
| +0 | Dev frontend |
| +1 | Dev API |
| +2 | Dev metrics |
| +3 | E2E frontend |
| +4 | E2E API |
| +5 | E2E metrics |
| +6 | Standalone production preview |
| +7–9 | Reserved for later needs |

All ten ports are checked against actual IPv4/IPv6 listeners before allocation.
A new worktree skips occupied blocks; an existing worktree keeps its assignment
and fails clearly if a port is occupied. Stop its services before rerunning
`init` or `check`. `status` can be used while services are running.

```bash
node scripts/worktree.mjs check
```

A registry reservation coordinates cooperating worktrees, not the operating
system. Another program can still take a port after the check; Vite, Playwright,
and the API fail on a bind conflict instead of silently selecting another port.
Separate clones have separate registries: keep linked worktrees in one clone
for coordinated reservations. The initial helper has no automatic pruning or
cleanup. It never assumes that an old-looking reservation is safe to reuse.

## Database layout and connection guardrails

For an ID such as `fittrack_123456789abc`, the allocated databases are:

- `ft_fittrack_123456789abc_dev`: persistent local development data
- `ft_fittrack_123456789abc_test`: disposable Go integration test data
- `ft_fittrack_123456789abc_rls`: restricted-runtime/RLS checks only
- `ft_fittrack_123456789abc_e2e`: browser test API data, separate from Go tests

All connect to a single local server at `127.0.0.1:55432`. This is a local endpoint
convention, not a PostgreSQL version selection. The databases must be created
and migrated independently before running their respective commands.

Once that server and its owning role exist, export an existing local connection
URL in your shell as `FITTRACK_LOCAL_DATABASE_URL`. Its database component is
replaced by the selected worktree/mode name; credentials are passed to the child
process in memory and are never written into the registry or config. The helper
refuses remote hosts, other ports, and query parameters that could override the
selected database or host. Only `sslmode` and `connect_timeout` query options are
accepted. Do not put production credentials in this variable.

For example, with `YOUR_LOCAL_OWNER` replaced by an already-provisioned local
role and its existing authentication configured:

```bash
export FITTRACK_LOCAL_DATABASE_URL='postgresql://YOUR_LOCAL_OWNER@127.0.0.1:55432/postgres?sslmode=disable'
```

The helper preserves other environment settings, including application keys.
It does not copy `.env` files, create auth accounts, or change Stack configuration.
Each worktree still needs its own normal client/server application settings.
Never commit local credentials. The command wrapper is a development guardrail,
not a security sandbox: arbitrary commands can override variables or connect
elsewhere, and owning roles on a shared cluster are not a security boundary.

## Running commands after database provisioning

The wrapper resolves its worktree from the current directory and leaves that
directory unchanged. Use it for each process rather than sourcing another
worktree's settings. These examples require the selected database to exist and
have that checkout's migrations applied.

```bash
# Terminal 1: development API
cd server
node ../scripts/worktree.mjs run dev -- go run ./cmd/api

# Terminal 2: development frontend
cd client
node ../scripts/worktree.mjs run dev -- bun run dev

# Disposable integration database, never the development database
cd server
node ../scripts/worktree.mjs run test -- go test -p 1 ./...
```

`test` holds a per-worktree test lock and sets `GOFLAGS` to serialize Go packages.
Keep `-p 1` for this suite; do not override it with higher package parallelism.
Different worktrees can run concurrently because their databases differ.

The original `make dev`, `make test-short`, `make test-integration`, and `make migrate-*` still own
a Docker Compose lifecycle and source `setenv.sh`; **do not use them through this
wrapper**. Sourcing an env file afterward can overwrite the isolation variables.
Use the direct Go commands above for this shared-server workflow. Existing
single-checkout Docker workflows remain unchanged.

For E2E, run the API in `e2e` mode in one terminal and Playwright in `e2e` mode in
another. Enable the existing local-only auth bootstrap explicitly if needed:

```bash
# API terminal (from server/)
E2E_LOCAL_AUTH_ENABLED=true node ../scripts/worktree.mjs run e2e -- go run ./cmd/api

# Browser test terminal (from client/)
VITE_E2E_LOCAL_AUTH_ENABLED=true node ../scripts/worktree.mjs run e2e -- bun run test:e2e
```

Playwright starts its own frontend on the E2E port. In CI/preview mode, build the
client with the same `e2e` environment and local-auth setting before serving it.
The wrapper supplies matching `APP_BASE_URL`, `API_PROXY_TARGET`, `E2E_BASE_URL`,
and `E2E_LOCAL_AUTH_API_BASE_URL`. It forces same-origin `/api` client requests
and disables reuse of an existing frontend so one worktree cannot silently test
another. E2E API startup and migrations are still manual in this increment.

`rls` mode additionally requires an existing
`FITTRACK_LOCAL_RUNTIME_DATABASE_URL` for `fittrack_app` at the same local
endpoint. It sets `DATABASE_URL` and `RECOMMENDATION_TEST_DATABASE_URL` to the
restricted `_rls` database, retains its owning-role URL as `ADMIN_DATABASE_URL`,
and enables RLS enforcement. Use only the dedicated restricted-role checks from
CI in this mode, not the full Go suite or owner migrations.

The fixed `fittrack_app` role is cluster-wide. Its provisioning alters role
attributes, and current CI changes its password. A future shared-cluster setup
must provision it once under a cluster-wide lock with stable existing credentials,
then apply grants separately per database. Worktree setup must never reset or
drop that shared role. Supporting distinct per-worktree runtime roles would
require changing the existing hardcoded migrations and tests.

Separate ports do not isolate localhost cookies. Use separate browser profiles
or Playwright contexts for independent Stack Auth sessions. Auth allowlists and
redirects must include the actual local origins if the selected auth flow needs
them.

## Failure recovery

- Port conflict: stop the listener or choose a different worktree. An existing
  assignment is deliberately not reassigned automatically
- Registry lock: wait for the other command. After a crash, inspect `owner.json`
  in the reported lock directory and verify that process is gone before removing
  the lock directory. Locks are never automatically stolen
- Missing/different config: stop services and rerun `init` to regenerate it from
  the registry. Do not copy `.worktree/` between checkouts
- Test failure: the child's exit status is preserved and its test lock is
  released once its process group has stopped. Cancellation reaches subprocesses
  too. If a process group cannot be verified as gone, the lock is retained; this
  includes unreaped zombie processes. A killed helper can also leave a lock; use
  the crash procedure above
- Worktree deletion/move: keep the old reservation until a future explicit
  cleanup flow can verify process state and exact owned database names. Never
  use wildcard database deletion

## Version policy and next increments

Production Supabase's actual PostgreSQL version has not yet been verified.
Existing definitions disagree: `server/docker-compose.yaml` pins `postgres:16.2`,
while `.github/workflows/test.yml` uses `postgres:15`. Do not infer production's
major from a Supabase default or select a new major merely because it is newer.

1. Obtain `SHOW server_version;` from the actual production database using an
   authorized read-only connection, or have its owner provide the result
2. Select a currently maintained patch release in that verified major, and
   align local/native installation and CI to that explicit version. If using
   containers, pin the artifact digest too and plan deliberate update reviews
3. Create a fresh local cluster for a different major. Never start a new major
   against the existing Compose `server/_db-data` directory. Preserve needed
   data through an explicit backup/migration plan
4. Add idempotent local-only creation of the four registered databases, verify
   server version, migrate each checkout separately, and implement centrally
   locked restricted-role provisioning. Keep dev data persistent
5. Verify two real linked worktrees simultaneously: migrations, full Go tests,
   restricted-role checks, distinct API/frontend/metrics listeners, and browser
   E2E. Then add narrowly scoped cleanup with ownership checks and confirmation

Tool versions already declared in this repository are Node `24.14.0`
(`.node-version`), Go `1.26.6` (`server/go.mod` and CI), Bun `1.4.2` (CI), and
Goose `v3.24.3` (CI/development guide). Use the lockfile for client dependencies.
A later DX pass can consolidate duplicated tool pins without conflating that
work with a database major upgrade.

Relevant upstream references:

- [PostgreSQL versioning and upgrade policy](https://www.postgresql.org/support/versioning/)
- [PostgreSQL roles are cluster-wide](https://www.postgresql.org/docs/current/database-roles.html)
- [Supabase upgrade guidance](https://supabase.com/docs/guides/platform/upgrading)
- [Git worktrees and shared repository state](https://git-scm.com/docs/git-worktree)
- [Cookie isolation does not follow port boundaries](https://www.rfc-editor.org/rfc/rfc6265#section-8.5)
