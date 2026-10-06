# Concurrent linked worktrees

**Creating a Git worktree is not setup.** In each worktree use
`node scripts/worktree.ts init`, then `db setup`, then `run` below.
Linked worktrees from one clone share a locked allocation registry. Separate
clones do not coordinate application ports. Never copy `.worktree/` or source
another checkout's database environment.

## Prerequisites and version gate

- Node **24.14.0**, Go **1.26.6**, Bun **1.4.2**, Goose **v3.24.3**.
- Windows 10/11 with Windows PowerShell 5.1 and available `Add-Type`, or POSIX.
  PowerShell and Git Bash invoke the same Node entry point on Windows.
- A running **local PostgreSQL cluster** at `127.0.0.1:55432`. Setup does not
  install/start/upgrade clusters or change their configuration.
- A local owning login with `CREATEDB` (and `CREATEROLE` for first setup), plus
  stable local credentials for the shared restricted `fittrack_app` login.

Production reported PostgreSQL **15.8 / 150008** through an authorized read-only
query on October 5, 2026. Local development and CI pin **15.19 / 150019**:

```text
postgres:15.19-bookworm@sha256:539ceaaae49b3a7c8a04467cf00cc6788d8e3f1675df41860d86eebc4c40524f
```

Review release notes and update Compose, CI, the Docker setup guard, and its test
together for deliberate patch updates. Local Docker does not reproduce Supabase
extensions, managed services, or session-pooler behavior. The application's
production session-pooler guard is unchanged; local checks do not prove that parity.

### Shell and TypeScript entry point

Use the same `node scripts/worktree.ts` commands from Git Bash, PowerShell,
or a POSIX shell. The pinned Node 24.14.0 executes the TypeScript directly;
no transpiler, loader, or Bash wrapper is needed. Check `node --version` first.
If your nvm-managed shell selects another version, run `nvm use 24.14.0`.

The only PowerShell script is an internal compile step for the Windows native
supervisor. Git Bash still launches Windows Node and therefore uses that
supervisor too. Windows Job Objects and exclusive sockets provide process-tree
cleanup and port ownership; switching shells cannot replace those OS guarantees.
The shared CLI, allocation logic, environment wiring, and tests stay in TypeScript.

### One shared Docker server

Start this once from one checkout's `server/` directory, then reuse the same
container from all linked worktrees. Set the local owner password in your session:

```bash
export DB_USER='postgres'
export DB_PASSWORD='YOUR_LOCAL_OWNER_PASSWORD'
export DB_NAME='postgres'
export DB_PORT='55432'
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) docker_host='npipe:////./pipe/docker_engine' ;;
  *) docker_host='unix:///var/run/docker.sock' ;;
esac
MSYS_NO_PATHCONV=1 docker --host "$docker_host" compose -p fittrack-local-pg15 up -d postgres
export FITTRACK_LOCAL_POSTGRES_CONTAINER='fittrack-postgres15'
```

The Bash example works in Git Bash and POSIX shells; `MSYS_NO_PATHCONV` keeps
Git Bash from rewriting the Docker endpoint. In PowerShell, set variables with
`$env:NAME='value'` and use `--host npipe:////./pipe/docker_engine`. The helper
supports these local daemons only. It ignores Docker environment/context overrides
when verifying the named container, its exact pinned image, ownership label,
running state, and sole `5432/tcp` publication at `127.0.0.1:55432`. It then
matches the connected server's bridge address/internal port and exact version
`150019`. Without this opt-in, native PostgreSQL must report the loopback
address/port directly. Setup never starts or replaces a container itself.

Compose and all Make lifecycle commands select the separate
`fittrack-local-pg15` project; Make passes it explicitly even when
`COMPOSE_PROJECT_NAME` is inherited. Compose uses a separately named container and fresh `server/_db-data-pg15`.
Leave the old `db` PostgreSQL 16 container and `server/_db-data` untouched.
Never run PostgreSQL 15 against PostgreSQL 16 storage. Rollback means returning
to the original PostgreSQL 16 container and its original data; keep both data
directories separate. `pg_upgrade` cannot downgrade 16 to 15. If old data is
needed, separately back it up and rehearse a compatible logical/data-only export
and import into a disposable PostgreSQL 15 database; review unsupported features
and validate the restored data before planning any migration. No data transfer
is performed by this workflow.

If global Goose differs, set `FITTRACK_GOOSE` to a separately installed pinned
executable. On Windows, bound compilation and omit unused database drivers:

```bash
export GOBIN="$(pwd -W)/.worktree/tools"
go install -p=1 -tags='no_mysql,no_sqlite3,no_mssql,no_ydb,no_vertica,no_clickhouse,no_turso' github.com/pressly/goose/v3/cmd/goose@v3.24.3
export FITTRACK_GOOSE="$GOBIN/goose.exe"
```

This installation example uses Git Bash on Windows. On POSIX, use `pwd`
instead of `pwd -W` and omit the `.exe` suffix.

## Setup in each worktree

```bash
node scripts/worktree.ts init
node scripts/worktree.ts status

# Set this in every worktree terminal when using the shared Docker server.
export FITTRACK_LOCAL_POSTGRES_CONTAINER='fittrack-postgres15'
# Supply these through a local secret manager/session, never a committed file.
export FITTRACK_LOCAL_DATABASE_URL='postgresql://LOCAL_OWNER:LOCAL_PASSWORD@127.0.0.1:55432/postgres?sslmode=disable'
export FITTRACK_LOCAL_RUNTIME_DATABASE_URL='postgresql://fittrack_app:STABLE_LOCAL_PASSWORD@127.0.0.1:55432/postgres?sslmode=disable'
node scripts/worktree.ts db setup --major 15
```

In PowerShell, use `$env:NAME='value'` for the same session variables; the
`node scripts/worktree.ts` commands are identical.

Replace placeholders locally; URL-encode special characters in credentials.
Never paste credentials into review artifacts. This workflow needs no production
credential or write.

`init` writes ignored `.worktree/config.json` with names and ports only. Its
registry is under the clone's Git common directory at
`fittrack-worktrees/registry.json`. Concurrent initializers briefly wait for
the lock without stealing it. The canonical checkout path determines the ID;
branch changes keep the allocation, moving the directory does not.

Setup verifies the native or explicitly inspected Docker endpoint, authenticated role, database, and
explicit major before mutations. A PostgreSQL advisory lock in `postgres`
coordinates role/catalog operations across clones throughout setup. Competing
setup fails with a retryable lock message. Existing databases must match both
the owning login and the exact `fittrack-worktree:<id>` comment. Unmarked or
differently owned databases are never adopted. A crash between creation and
marking requires manual ownership verification. Otherwise reruns create only
missing databases and run Goose `up` idempotently. There is no drop/reset command.

| Suffix | Purpose |
| --- | --- |
| `_dev` | Persistent development data; never used by test modes |
| `_test` | Go integration fixtures and destructive RLS-state tests |
| `_rls` | Restricted-role checks, with separate owner/runtime connections |
| `_e2e` | Browser test API data |

Names are `ft_<worktree-id>_<suffix>`. Setup applies this checkout's migrations
and grants to all four databases. Stop wrapped commands before setup; their
run records prevent migrations underneath them. Unwrapped commands cannot
participate in this coordination.

The first setup creates `fittrack_app` only if absent, using the supplied local
runtime password. Later setup authenticates its existing password and checks
restricted attributes/memberships; it **never resets passwords, alters unsafe
existing attributes, or drops the role**. Unsafe or NOLOGIN existing roles need
operator attention. Per-database grants use the provisioning script's
preserve-role mode. Do not run CI's password-reset or ordinary role provisioning
commands against this shared cluster while agents are using it.

## Commands and cancellation

Each terminal uses its own wrapper, which preserves the current directory:

```bash
# Terminal 1, from server/
node ../scripts/worktree.ts run dev -- go run ./cmd/api
# Terminal 2, from client/ after bun install --frozen-lockfile
node ../scripts/worktree.ts run dev -- bun run dev
# From server/: serialized packages, isolated disposable test database
node ../scripts/worktree.ts run test -- go test -p 1 ./...
# From the worktree root: cancel all its dev commands
node scripts/worktree.ts stop dev
```

`stop` requests cancellation; wait for the original commands to exit before
restarting. Ctrl+C also cancels. `test` and `rls` each hold an exclusive mode
lock until the owned process tree has stopped. Dev API/frontend and E2E
API/Playwright share their respective mode because they use separate ports.
Do not run concurrent destructive E2E suites in one worktree.

PowerShell compiles a small native supervisor once per source hash into
`%TEMP%/fittrack-worktree-native/`; long-running commands do not retain a
PowerShell process. Windows atomically places commands in non-breakaway Job Objects with
kill-on-close. Normal root exit, failure, cancellation, and wrapper death also
terminate grandchildren. The supervisor verifies the job is empty before
releasing locks. Windows cancellation is forced, not a POSIX graceful signal.
Commands are noninteractive native executables (`go`, `bun`, `node`); stdin is
closed. `.cmd`/`.bat` and shell expressions are not implicitly interpreted.
Invoke an explicit shell when needed. Arguments use Windows argv quoting.

POSIX retains separate process groups and TERM/KILL cleanup. Deliberately
escaping a POSIX group is outside this guardrail. This is local coordination,
not a security sandbox: arbitrary commands can override variables, and shared
owning roles are not a tenant security boundary.

Do not wrap `make dev`, `make test-*`, or `make migrate-*`; those source
`setenv.sh` and manage Docker. Use the direct commands above. Supply normal app
settings such as `PROJECT_ID` per terminal; setup does not copy `.env` files or
create Stack accounts.

## Ports, frontend, and auth

Ten-port blocks use 21000â€“29999: dev frontend `+0`, API `+1`, metrics `+2`,
E2E frontend `+3`, API `+4`, metrics `+5`, preview `+6`, three reserved ports.
`init`/`check` probe both address families. Windows uses exclusive native sockets:
Node's `exclusive: true` was observed accepting a conflicting loopback listener
even on pinned Node 24.14.0.

```bash
node scripts/worktree.ts check
```

Stop services before `init`/`check`; use `status` while running. New allocations
skip occupied blocks; existing allocations fail without moving. The registry
does not reserve OS ports after a check. Another program can race startup;
Vite/Playwright use strict ports and the API reports bind failure.

For E2E, set `export E2E_LOCAL_AUTH_ENABLED='true'` and start the API from `server/`
using `run e2e -- go run ./cmd/api`. From `client/`, set
`export VITE_E2E_LOCAL_AUTH_ENABLED='true'` and use `run e2e -- bun run test:e2e`.
PowerShell uses `$env:NAME='value'` for these variables. Playwright owns the E2E frontend
with existing-server reuse disabled. Build preview/CI assets under the same
`e2e` environment. The wrapper aligns `APP_BASE_URL`, `API_PROXY_TARGET`,
`E2E_BASE_URL`, and `E2E_LOCAL_AUTH_API_BASE_URL`, and forces browser requests to
same-origin `/api`.

`rls` uses the runtime URL for `_rls`, retains the owning `ADMIN_DATABASE_URL`,
and enables RLS enforcement. Use CI's dedicated restricted checks, not the full
Go suite or migrations. Separate browser profiles/Playwright contexts are still
needed for independent localhost cookies. External Stack authentication may
require allowlisting the actual allocated origins.

## Recovery and verification

Locks/run records are never automatically stolen. After a hard kill, inspect
`owner.json` in the reported registry directory, verify the owned supervisor
and tree are gone, then remove only the exact stale mode lock/run record. PID
reuse means a live PID alone does not prove ownership. If uncertain, keep the
record. Never kill by process-name/port wildcard. Old allocations survive
worktree deletion/moves until explicitly reviewed.

After installing the pinned client dependencies, type-check the CLI and both
test files from the repository root:

```bash
node client/node_modules/typescript/bin/tsc -p scripts/tsconfig.json
node --test scripts/worktree.test.ts
```

Node strips types at runtime; the separate strict TypeScript check runs in CI.
The lifecycle tests exercise native process/socket and real linked-worktree coverage. For the disposable PostgreSQL test, ensure 55432 is
unused, set `FITTRACK_TEST_DOCKER=1` for the pinned Docker image (or
`FITTRACK_TEST_POSTGRES_BIN` for a native PostgreSQL bin directory), and
`FITTRACK_GOOSE` to v3.24.3, then run:

```bash
node --test scripts/worktree.database.test.ts
```

This opt-in test creates a fresh temporary cluster and eight databases in two
real linked worktrees, exercises migration/write/read/reuse and refusal paths,
then stops/removes its own resources. A skipped database test or green unit
tests do not prove database isolation. For live API/proxy verification too,
build this checkout's API (`go build -p=1 -o ../.worktree/api.exe ./cmd/api`
from `server/`), install the client's locked dependencies, and set
`FITTRACK_TEST_API` to that executable. Set `FITTRACK_TEST_PROJECT_ID` to your
existing local Stack project ID: API initialization requires public JWKS access
even with local E2E authentication. The test uses two temporary frontend
checkouts and links their dependencies; it does not regenerate files in your
checkout. It checks real auth/API/metrics/proxy endpoints and loads Playwright's
actual configuration, but does not execute the full browser E2E suite.

References: [atomic Job Object assignment](https://devblogs.microsoft.com/oldnewthing/20230209-00/?p=107812),
[Job Object lifecycle](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects),
[PostgreSQL version policy](https://www.postgresql.org/support/versioning/).
