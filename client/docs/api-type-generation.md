# API type generation

The checked-in client is generated from Go Swagger annotations. Do not install a
second generator or hand-edit generated output. The repository uses
`@hey-api/openapi-ts` **0.99.0**, locked in `client/package.json` and `bun.lock`,
not `openapi-typescript-codegen` or an Axios service-class generator.

## Generate and verify

Use Node 24.22.1, Bun 1.4.2, and Go 1.26.11 (the documented/CI versions).
No running API, database, Docker, auth provider, or Stripe account is needed.
The first run may need network access for Go modules and client dependencies.

From the repository root:

```bash
bun install --cwd client --frozen-lockfile
bun run --cwd client api:generate
bun run --cwd client api:check
bun run --cwd client tsc
```

For concurrent linked worktrees, prepare and run in the credential-stripped
unit mode described in [worktree setup](../../docs/worktrees.md):

```bash
# Repository root, after installing the frozen client dependencies:
node scripts/worktree.ts prepare --mode unit --json
node scripts/worktree.ts run unit -- bun run --cwd client api:generate
node scripts/worktree.ts run unit -- bun run --cwd client api:check
node scripts/worktree.ts run unit -- bun run --cwd client tsc
```

`api:generate` runs `scripts/api-contract.mjs`, which:

1. Runs `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init` from `server/`,
   with entry point `cmd/api/main.go` and `--outputTypes json,yaml`.
2. Runs the installed `openapi-ts` with `client/openapi-ts.config.ts` against
   the freshly generated Swagger JSON, not a potentially stale checked-in spec.
3. Runs oxfmt and oxlint fixes on the temporary generated client.
4. Publishes Swagger JSON/YAML and the client only after generation succeeds.
   Publication failures restore backed-up outputs.

`api:check` runs that same chain into temporary output and compares every
expected generated path without publishing. It fails for changed, missing, or
unexpected output. Neither command stages files. `make swagger` from `server/`
is an alias for the complete generation command. `bun run openapi-ts` from
`client/` is only the lower-level client generator; it does **not** regenerate
Swagger. `bun run build` does not regenerate the API contract.

## Configuration and output

`client/openapi-ts.config.ts` normally reads `server/docs/swagger.json` and
writes `client/src/client/`. The orchestration script overrides those locations
with `FITTRACK_OPENAPI_INPUT` and `FITTRACK_OPENAPI_OUTPUT` for temporary output.
Its plugins generate schemas, TypeScript types/JavaScript enums, SDK functions,
and TanStack React Query query keys, query options, and mutation options.

Key paths:

- `server/docs/swagger.json` and `swagger.yaml`: generated Swagger 2 documents.
- `server/docs/docs.go`: custom Go embedding/registration wrapper, **not generated**.
- `client/src/client/types.gen.ts`: request/response types.
- `client/src/client/sdk.gen.ts`: endpoint functions, such as `getWorkouts`.
- `client/src/client/schemas.gen.ts`: generated schemas.
- `client/src/client/client.gen.ts`, `client/`, and `core/`: Fetch-based client support.
- `client/src/client/@tanstack/react-query.gen.ts`: React Query integration.
- `client/src/client/index.ts`: SDK/type exports; React Query imports use their own path.

For example, existing application code imports:

```typescript
import type { WorkoutWorkoutWithSetsResponse } from "@/client";
import { getWorkoutsQueryOptions } from "@/client/@tanstack/react-query.gen";
```

Use feature API modules such as `src/features/workouts/api/workouts.ts` for
application behavior and invalidation. Shared runtime configuration lives in
`src/lib/api/client-config.ts`: it sets `baseUrl` using `VITE_API_BASE_URL`
(default `/api`, or `http://localhost/api` in tests) and installs auth/error
interceptors. A custom API base must include `/api`; do not add a parallel
`OpenAPI.BASE` or `process.env` configuration in browser code.

## Review, staging, and CI

Review the generated diff, then stage only the intended contract output alongside
the source change (commands below are from the repository root):

```bash
git diff -- server/docs/swagger.json server/docs/swagger.yaml client/src/client
git add server/docs/swagger.json server/docs/swagger.yaml client/src/client
```

With Husky installed, the pre-commit hook generates for staged backend Go source,
Go module files, generator configuration/scripts, and generated contract inputs.
It refuses relevant unstaged or partially staged input/output before generating.
If output changes, the hook leaves it unstaged and fails with review/staging
instructions. Review, explicitly stage, and retry; there is no automatic broad
staging. Ordinary frontend, Markdown, migration, and backend test-only changes
do not trigger generation.

Client changes require Bun and run format, lint, both knip modes, and unit tests.
Missing Bun fails with installation/PATH guidance, not a pnpm fallback. Server
non-doc changes retain the fast Go check; server docs-only changes skip it.

The `api-contract-freshness` job in
[`.github/workflows/test.yml`](../../.github/workflows/test.yml) installs frozen
client dependencies, runs `api:check`, and tests the commit hook on every pull
request to `main`. It does not update generated files or create bot commits.

## Troubleshooting

- Missing Bun/Go/Node: install the documented tools and put them on PATH.
- Missing client generator/formatter/linter: install frozen client dependencies;
  do not install a substitute generator or modify the lockfile to work around it.
- Stale output: run `api:generate`, inspect the diff, and rerun `api:check`/`tsc`.
- Generation failure: fix the reported Go annotations, spec, or config error;
  failed generation does not publish temporary output.
- Hook refuses partial staging: resolve relevant unstaged input/output first.
  Do not bypass the guard by broadly staging unrelated work.
- Type errors: update the feature caller or backend contract as appropriate and
  regenerate; do not patch `*.gen.ts` manually.
