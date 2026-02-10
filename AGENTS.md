# FitTrack Agent Notes

## Progress (2026-02-10)

- Implemented backend-powered exercise metrics history: `GET /api/exercises/{id}/metrics-history?range=W|M|6M|Y`
- Per-workout points for `W/M`; weekly buckets for `6M`; monthly buckets for `Y`; intensity allowed `>100`
- SQL added to `server/query.sql`; regenerated sqlc (`server/internal/database/query.sql.go`)
- Added migration for `exercise.historical_1rm*`: `server/migrations/00014_add_exercise_historical_1rm.sql`; updated `server/schema.sql`
- Swagger regenerated (`server/docs/swagger.json|yaml`); TS client regenerated (`client/src/client/*`)
- Frontend: exercise detail now renders 4 new charts under "Session Metrics" with same range selector; bar-click navigates to workout when `workout_id` present
- Demo-mode supported: metrics-history computed client-side from `exerciseSets`
- Decision documented: no SUM across workouts for bucket rollups (see `docs/new-metrics/metrics-history-bucketing.md`)
- Added regression test for new handler: `server/internal/exercise/metrics_history_handler_test.go`
- Implemented historical 1RM lifecycle:
  - Endpoints: `GET /api/exercises/{id}/historical-1rm`, `PATCH /api/exercises/{id}/historical-1rm` (`manual|recompute`)
  - Auto PR detection on workout create; recompute on workout update/delete when PR was sourced from that workout
  - UI: "Historical 1RM" tile + dialog on exercise detail (demo shows computed only)
  - Tests: handler tests + integration PR lifecycle (`server/internal/workout/historical_1rm_pr_integration_test.go`)

## Infra / Gotchas

- `server/docker-compose.yaml` volume path changed to `server/_db-data` to avoid Go tooling permission issues on `server/db-data`
- Local `.env` files created for tests/dev; gitignored; do not commit

## Next Chunk

- Optional: demo-mode parity for historical 1RM (localStorage)
- Optional: simplify API: fold historical 1RM payload into `GET /api/exercises/{id}` to avoid extra query
- Optional: migrate existing volume chart to consume backend `metrics-history` range semantics (avoid client-side slice-by-points divergence)
