# Database Migrations

This directory holds the PostgreSQL schema migrations for the server. Migrations are managed with [Goose](https://github.com/pressly/goose) and are applied in numeric order.

## File Conventions

- Keep filenames zero-padded and sequential, for example `00015_add_example_column.sql`.
- Write both `-- +goose Up` and `-- +goose Down` sections so rollbacks stay possible during development.
- Prefer small, focused migrations. If a change needs data backfill or follow-up constraints, split that into a separate migration.

## Common Commands

Run these from [`server/Makefile`](/C:/E/projects/fittrack/server/Makefile):

```bash
make migrate-create NAME=add_example_column
make migrate-up
make migrate-down
```

`make migrate-up` and `make migrate-down` load `DATABASE_URL` from `server/setenv.sh`.

## Current Migration Set

As of now, this directory contains migrations `00001` through `00014`, covering:

- initial user, workout, exercise, and set tables
- user ownership and row-level security
- set and exercise ordering columns
- workout focus indexing
- cascade delete cleanup
- numeric weight storage
- historical 1RM tracking

For the authoritative history, read the migration files themselves rather than maintaining a per-file summary here.
