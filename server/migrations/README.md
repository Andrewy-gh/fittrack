# Database Migrations

This directory holds the PostgreSQL schema migrations for the server. Migrations are managed with [Goose](https://github.com/pressly/goose) and are applied in numeric order.

## File Conventions

- Keep filenames zero-padded and sequential, for example `00015_add_example_column.sql`.
- Write both `-- +goose Up` and `-- +goose Down` sections so rollbacks stay possible during development.
- Prefer small, focused migrations. If a change needs data backfill or follow-up constraints, split that into a separate migration.

## Common Commands

Run these from the `server` directory (see [Makefile](../Makefile)):

```bash
make migrate-create NAME=add_example_column
make migrate-up
make migrate-down
```

`make migrate-up` and `make migrate-down` load `DATABASE_URL` from `server/setenv.sh`.

## Runtime Privileges

Migrations run as the owning/admin role; production queries run as restricted `fittrack_app`. See [production triage](../../docs/production-triage.md) for credential locations and rollout guidance. Do not use the runtime credential to run Goose.

[Runtime provisioning](../scripts/provision-runtime-role.sql) revokes broad grants and sets deny-by-default privileges for objects created by the migration owner. Creating a table, sequence, or function does **not** automatically make it usable by the app.

For each new runtime object:

1. Enable RLS and add the appropriate user policies for user-owned tables.
2. Grant only required operations to `fittrack_app` in the migration, including sequence `USAGE` and function `EXECUTE` where needed. Guard grants with a role-existence check so fresh databases can migrate before provisioning creates the role.
3. Update the provisioning script's explicit allowlist too; rerunning it resets existing grants. Do not grant broad access to `PUBLIC` or use blanket default grants.
4. Keep `SECURITY DEFINER` functions narrowly scoped, pin their search path, and revoke default `PUBLIC` execution. Webhook idempotency tables remain inaccessible directly.
5. Extend the restricted-role smoke test for the new behavior. Verify both a fresh migration/provisioning path and an upgrade with the role already provisioned; the latter catches missing migration grants that fresh provisioning can hide.

CI uses a separate disposable `fittrack_rls_test` database, applies migrations and provisioning, then connects as `fittrack_app` to run [role verification](../scripts/verify-runtime-role.sql), [two-user and webhook smoke tests](../scripts/test-runtime-role.sql), and the app's connection-context integration test. The SQL smoke test rolls back its fixtures but can advance sequences; never run it against production.

For the authoritative schema history, read the migration files themselves rather than maintaining a per-file summary here.
