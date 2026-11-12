# Database security tools

Row-level security (RLS) limits which rows a database connection can access for
its current user context. These tools have different purposes and credentials;
do not run every SQL file as one audit.

## Maintained tools

| Script | Purpose and callers | Safety assumptions |
| --- | --- | --- |
| [provision-runtime-role.sql](provision-runtime-role.sql) | Establishes the restricted `fittrack_app` role and deny-by-default grants. Used by CI and `server/cmd/worktree-db`. | Writes role attributes, grants and default privileges. Requires the migration owner/admin. Shared local setup uses preserve-role mode; do not run ordinary provisioning or reset the shared role's password while worktrees are using it. |
| [verify-runtime-role.sql](verify-runtime-role.sql) | Rejects unsafe runtime privileges, checks webhook access restrictions and checks an unknown user's workout visibility. Used by CI, the isolated database test and the production-triage runbook. | Connect as the candidate runtime role, not the owner. Reads catalogs/data and changes only the current session's user context. It does not prove all policies or deployed HTTP authentication. |
| [test-runtime-role.sql](test-runtime-role.sql) | Asserts two-user workout isolation, cross-user write rejection, missing-context denial, billing isolation and narrow webhook-function access. Used by CI and the isolated database test. | Disposable migrated/provisioned database only, connected as `fittrack_app`. Fixture writes roll back, but sequences can advance. Never use production or a shared development database. |
| [audit-rls-coverage.sql](audit-rls-coverage.sql) | Manual, read-only catalog report of table RLS flags, policies and command coverage. No automated caller remains; retained because this inventory is useful when reviewing schema changes. | Uses `psql` meta-commands and SELECTs. Report only: it does not fail when a policy is missing and is not an isolation test. See interpretation limits below. |

### Current verification path

[`.github/workflows/test.yml`](../../.github/workflows/test.yml), in the
**Verify restricted runtime role** step, creates a separate disposable
`fittrack_rls_test` database, migrates it, provisions the role, then runs role
verification and the rollback-only SQL smoke test as `fittrack_app`. It also runs:

- `TestOpenDatabaseSetsRLSContextForEveryAcquisition` in
  [`internal/app/rls_connection_integration_test.go`](../internal/app/rls_connection_integration_test.go):
  checks that the real application pool sets each user's context and clears it
  for an unauthenticated acquisition.
- `TestRecommendationPersistenceWithRLS` in
  [`internal/recommendation/repository_integration_test.go`](../internal/recommendation/repository_integration_test.go):
  exercises workout/exercise/set persistence and cross-user denial through
  repositories under the restricted role.

For local execution, first follow [linked-worktree setup](../../docs/worktrees.md).
Its `rls` mode supplies the isolated runtime and admin URLs; use the dedicated
restricted checks, not the full Go suite or migrations in that mode. The opt-in
`scripts/worktree.database.test.ts` creates its own temporary cluster and runs
the retained SQL checks; its prerequisites are documented in the same guide.
The general database tests include owner/superuser-sensitive checks that can
skip, so a green general suite alone is not restricted-role proof.

These checks do **not** establish live Stack authentication, deployed API
isolation or session-pooler parity. Those need a separately authorized scenario;
there is no `rls-smoke-test.yml` workflow or remote smoke-test secret requirement
in the current Test workflow.

### Manual catalog inspection

After selecting an approved **local** database explicitly, from `server/`:

```bash
# Set LOCAL_INSPECTION_DATABASE_URL privately to the intended local database.
psql "$LOCAL_INSPECTION_DATABASE_URL" -X -v ON_ERROR_STOP=1 -f scripts/audit-rls-coverage.sql
```

Do not substitute a production credential or print connection strings. This
report performs no fixture writes or role provisioning. Interpret it manually:

- No RLS is not automatically a vulnerability: migration metadata and deliberately
  restricted tables need their grants and intended access reviewed.
- RLS without policies denies access only to roles subject to RLS; owners and
  bypass roles may behave differently.
- The disabled-table section matches table names without schema qualification;
  duplicate table names across schemas can hide a disabled table. Check the
  schema-qualified first section instead when multiple schemas are present.
- Command counts and clause-presence labels do not validate policy expressions,
  ownership, grants or cross-user behavior. Use the restricted-role checks above.

## Storage-schema checks (not RLS proof)

The [canonical schema verification guide](../schema-verification.md) describes
these separate checks against an owned disposable PostgreSQL cluster:

- [schema-storage-catalog.sql](schema-storage-catalog.sql) reads application
  storage definitions for comparison between migrated and schema-built databases.
  It deliberately does not compare runtime security configuration.
- [test-schema-persistence.sql](test-schema-persistence.sql) verifies status,
  foreign-key, JSON, timestamp and uniqueness behavior in both databases.
  Fixtures roll back, but sequences can advance: never run it against production
  or shared development data.

`schema.sql` is the SQLC storage model, not a deployment bootstrap. These checks
complement—not replace—the restricted-role verification above.

## Retired historical tooling

The following files originated in the 2025 RLS investigation (`f598297`, #36).
Tracked-file searches found no executable callers outside that legacy group;
the old README was the only documented entry point for the HTTP scripts.
External/manual usage outside the repository cannot be established by that search. They were removed rather
than reviving an unsafe second verification path:

| Removed script | Historical purpose | Why retired / remaining coverage |
| --- | --- | --- |
| `quick-rls-test.sh` | Two-token HTTP workout creation/list isolation and anonymous rejection. | Defaulted to the live app, left writes behind and treated successful curl transport as successful HTTP creation. Error objects could satisfy its list-length checks. The restricted SQL/pool checks are the maintained isolation path, not a claim of deployed HTTP parity. |
| `rls-smoke-test.sh` | Broader manual HTTP workout/exercise isolation check. | Also defaulted to the live app without cleanup. Expected 200/empty data for cross-user detail instead of the current 404 contract; one list check required User B's own notes count to be zero, contradicting its fixture. Not a trustworthy live acceptance test. |
| `run-complete-rls-audit.sh` | Chained the catalog report, persistent fixtures and privilege probes. | Used PowerShell `Get-Content` inside Bash and the retired `db` container/credentials, lacked reliable SQL failure propagation, and printed hard-coded findings rather than generating a report. Dedicated CI/local restricted checks replace the runner; the catalog report remains independently usable. |
| `setup-test-data.sql` | Persisted three users' workouts/exercises for the old probes. | Only called by the retired runner, accumulated workout fixtures and had no rollback/cleanup. Maintained tests own their fixtures. |
| `test-privilege-escalation.sql` | Exploratory WHERE TRUE, OR/UNION, session and role probes. | No caller; unbounded UPDATE/DELETE without an outer rollback, an undeclared loop record and notice-only error handling. Restricted-role assertions and the existing database bypass tests retain meaningful isolation checks without treating notices as acceptance. |
| `test-privilege-escalation-fixed.sql` | Alternate exploratory probe used by the runner. | Not a title-only copy of the other SQL file: scenarios differ. Still used persistent updates, invalid aggregate LIMIT syntax, notice-only failures and obsolete claims that the set table has no policies. Current migrations include set policies; maintained runtime/repository checks exercise actual access. |

Removing the runner supersedes its repair finding (A029/F16); it does not assert
that the historical complete audit now passes. No production verification or
database modification is part of this cleanup.
