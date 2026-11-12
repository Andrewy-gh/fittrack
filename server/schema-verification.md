# Canonical storage schema verification

`schema.sql` is SQLC's storage model, not a deployment/bootstrap script.
Migrations remain authoritative for runtime security and schema history. Never
apply this file over an existing database. Do not renumber historical migrations,
fill numbering gaps, or remove indexes as part of schema reconciliation.

## Disposable comparison procedure

Follow [worktree isolation](../docs/worktrees.md). Use a fresh, uniquely owned,
tmpfs PostgreSQL **15.19** container pinned to the documented image digest, only
when `127.0.0.1:55432` is free. Keep credentials in memory. Do not use an inherited
`DATABASE_URL`, production credentials, or an old PostgreSQL data directory.
Clean up only that container in `finally`, including on failure.

1. From the target checkout, with explicit local owner/runtime URLs and pinned
   Goose v3.24.3, run `node scripts/worktree.ts prepare --mode test --json`.
   Preparation must finish successfully before managed tests start: migration
   admission forbids setup beneath active runs. This applies every migration
   through 00035 and provisions separate dev/test/rls/e2e databases.
2. Inside that **owned disposable cluster only**, create an additional empty
   database and load `server/schema.sql` using `psql -X -v ON_ERROR_STOP=1`.
3. Run `server/scripts/schema-storage-catalog.sql` through `psql -X -qAt
   -v ON_ERROR_STOP=1` in both the migrated test database and the schema-built
   database. Compare the parsed JSON objects by key/value, not output order.
   Any difference needs an explanation or repair; do not filter out a failing
   constraint/index. PostgreSQL's catalog, not source-text similarity, is the
   comparison boundary.
4. Run `server/scripts/test-schema-persistence.sql` through `psql -X
   -v ON_ERROR_STOP=1` in **both** databases. It checks accepted/rejected statuses,
   exact SQLSTATE and constraint identities, valid/missing draft parents,
   `ON DELETE SET NULL`, workout JSON arrays, exercise timestamp defaults and
   non-nullability, and duplicate-name rejection. Fixtures roll back; sequences
   can advance, so disposable databases are mandatory.
5. Run `make -C server sqlc` (SQLC v1.29.0), inspect generated diffs, and run managed
   Go build/vet and relevant repository tests. Run retained runtime-role SQL
   checks separately in the provisioned RLS database. Neither a skipped Go test
   nor a static grep substitutes for the two real-database checks above.

The catalog query compares table presence/kind/persistence; named column types,
nullability, defaults, collation and identity/generated flags; all checks,
foreign keys and unique/primary constraints (including validation/deferral);
all index definitions and readiness/validity/null-distinct semantics; and
sequence configuration and ownership. It does not compare OIDs, physical column
order, dropped-column slots, sequence current values, table data, or statistics:
these depend on creation order/history, not the intended storage contract.

### Deliberate boundaries, not silent parity claims

- Goose's `goose_db_version` and its sequence are bookkeeping, not application
  storage. They are the only named storage objects omitted by the catalog query.
- The SQLC file intentionally has no roles/owners/ACLs/default privileges, RLS
  flags/policies, or SQL functions. Those are **not equal** between these two
  databases. Deployments must use migrations plus
  `scripts/provision-runtime-role.sql`, never `schema.sql`.
- The migrated/provisioned database has RLS enabled on all 14 application tables,
  45 policies, and four functions: `current_user_id`,
  `lookup_stripe_customer_user_id`, `has_processed_stripe_webhook_event`, and
  `record_stripe_webhook_event`. The schema-built database has no such security
  setup. Verify the real security boundary with the retained
  `verify-runtime-role.sql` and `test-runtime-role.sql` as `fittrack_app` in the
  migrated RLS database; do not add duplicate security SQL to the SQLC model.
- This local cluster does not model Supabase extensions/services or pooler
  behavior, and the comparison makes no claim about deployed production state.

## A044 reconciliation evidence (2026-11-12)

A fresh PostgreSQL `server_version_num = 150019` comparison found nine catalog
entries differing, representing eight storage definitions (one unique
constraint also owns an index):

| Object | Migrated definition / repair to canonical file |
| --- | --- |
| `idx_workout_user_focus` | `CREATE INDEX ... ON workout USING btree (user_id, workout_focus)` (00011) |
| Conversation source-run FK | `FOREIGN KEY (latest_workout_draft_source_run_id) REFERENCES ai_chat_run(id) ON DELETE SET NULL` (00020) |
| Conversation saved-workout FK | `FOREIGN KEY (latest_workout_draft_saved_workout_id) REFERENCES workout(id) ON DELETE SET NULL` (00020) |
| `ai_chat_message_status_valid` | `status IN ('streaming', 'completed', 'failed', 'stopped')` (00026) |
| `ai_chat_run_status_valid` | Same four run statuses (00026) |
| `ai_chat_run_generation_status_valid` | `generation_status IN ('queued', 'generating', 'completed', 'failed', 'interrupted', 'stopped')` (00026) |
| `workout_recommendation_context_check` | `CHECK (jsonb_typeof(recommendation_context) = 'array')` (00031) |
| `exercise_user_id_name_key` and its backing index | `UNIQUE NULLS NOT DISTINCT (user_id, name)` (00003). Both columns are already NOT NULL, so this corrects representation without inventing nullable-key behavior. |

The original report's **second missing index and third missing FK were not
reproduced**. In particular, `idx_ai_chat_message_user_conversation`, created in
00016, was deliberately dropped in 00024 when
`ai_chat_message_user_conversation_id_unique` was added. Do not resurrect it.
The training-profile composite FKs and
`exercise_historical_1rm_source_workout_id_fkey` already match. Comparing *all*
storage constraints/indexes avoids guessing which historical claim was meant.

The repaired catalog comparison covers 14 tables, 127 columns, 81 constraints,
38 indexes and eight sequences (268 entries), with no storage differences.

Before repair, the new behavioral SQL failed on the schema-built database when
inserting an assistant message with `status = 'stopped'`, specifically
`ai_chat_message_status_valid`. After repair, the same behavior script passes in
both databases and the storage catalog comparison has no differences. No new
migration is warranted: fresh migrated runtime storage already has the intended
definitions. The existing exercise timestamp default/backfill/NOT NULL fixes in
00034/00035 are preserved. SQLC generation produces no tracked binding changes
because these constraint/index changes do not change query result types.

Focused verification also passed without skips (from `server/`):

```sh
node ../scripts/worktree.ts run test -- go test -p 1 -count=1 -v -run 'TestRepositoryStopRun_|TestRepositoryPrepareMessageStream_AllowsNextPromptWhenStop|TestExerciseListAndCreateReturnPersistedTimestamps|TestExerciseTimestampBackfillPreservesKnownHistory|TestWorkout_Historical1RM_PR_Lifecycle_Integration' ./internal/aichat ./internal/exercise ./internal/workout
node ../scripts/worktree.ts run unit -- go build -p=1 ./...
node ../scripts/worktree.ts run unit -- go vet -p=1 ./...
```

This executed two AI stop/concurrency repository cases, two timestamp cases
(including actual 00035 backfill SQL), and the workout historical-1RM lifecycle.
The two retained restricted-role SQL scripts also passed in the isolated RLS DB.
These complement, rather than replace, the direct status/constraint checks.
Installed tools were Node 24.22.1, SQLC 1.29.0 and Go **1.26.4**; the documented Go
version is 1.26.10. No replacement toolchain was installed. This was a scoped
verification, not the full Go suite or a production/cloud/provider test.
