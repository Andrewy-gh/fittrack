# FitTrack Agent Notes

## Progress (2026-02-10)

- Implemented backend-powered exercise metrics history: `GET /api/exercises/{id}/metrics-history?range=W|M|6M|Y`
- Per-workout points for `W/M`; weekly buckets for `6M`; monthly buckets for `Y`; intensity allowed `>100`
- SQL added to `server/query.sql`; regenerated sqlc (`server/internal/database/query.sql.go`)
- Added migration for `exercise.historical_1rm*`: `server/migrations/00014_add_exercise_historical_1rm.sql`; updated `server/schema.sql`
- Swagger regenerated (`server/docs/swagger.json|yaml`); TS client regenerated (`client/src/client/*`)
- Frontend: exercise detail renders session-metrics charts under "Session Metrics" with same range selector; bar-click navigates to workout when `workout_id` present
- Demo-mode supported: metrics-history computed client-side from `exerciseSets`
- Decision documented: no SUM across workouts for bucket rollups (see `docs/new-metrics/metrics-history-bucketing.md`)
- Added regression test for new handler: `server/internal/exercise/metrics_history_handler_test.go`
- Implemented historical 1RM lifecycle:
  - Endpoints: `GET /api/exercises/{id}/historical-1rm`, `PATCH /api/exercises/{id}/historical-1rm` (`manual|recompute`)
  - Auto PR detection on workout create; recompute on workout update/delete when PR was sourced from that workout
  - UI: "Historical 1RM" tile + dialog on exercise detail (demo supports manual override + recompute via localStorage)
  - Tests: handler tests + integration PR lifecycle (`server/internal/workout/historical_1rm_pr_integration_test.go`)
- Simplified exercise detail fetching:
  - `GET /api/exercises/{id}` (sets payload) now includes stored historical 1RM fields per row: `historical_1rm`, `historical_1rm_updated_at`, `historical_1rm_source_workout_id`
  - UI reads stored values from `exerciseSets[0]` and avoids `GET /historical-1rm` when sets exist
- Volume chart migrated to backend `metrics-history` points:
  - Removed legacy client-side "Daily Volume" (`ChartBarVol`)
  - Added `Working-Set Volume` series under "Session Metrics" (uses `total_volume_working`)
- Demo-mode parity for historical 1RM:
  - localStorage key: `fittrack-demo-historical-1rm`; bootstrapped from working sets; lifecycle hooks on workout create/update/delete + exercise delete
  - Test: `client/tests/e2e/demo/exercise-historical-1rm.test.ts`
- Stack Auth optional (dev/test ergonomics): app runs in demo-mode even if `VITE_PROJECT_ID` / `VITE_PUBLISHABLE_CLIENT_KEY` missing

## Infra / Gotchas

- `server/docker-compose.yaml` volume path changed to `server/_db-data` to avoid Go tooling permission issues on `server/db-data`
- Local `.env` files created for tests/dev; gitignored; do not commit

## Next Chunk

- Optional: finish API cleanup: return exercise meta alongside sets for `GET /api/exercises/{id}` so historical 1RM is available even when exercise has zero sets; then delete `GET /api/exercises/{id}/historical-1rm`
