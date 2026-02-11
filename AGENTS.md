# FitTrack Agent Notes

## Snapshot (2026-02-11)

- Exercise metrics history is backend-powered: `GET /api/exercises/{id}/metrics-history?range=W|M|6M|Y`
- Bucketing semantics:
  - `W/M`: per-workout points
  - `6M`: weekly buckets
  - `Y`: monthly buckets
  - Intensity allows `>100`
- Historical 1RM lifecycle is live:
  - Endpoint: `PATCH /api/exercises/{id}/historical-1rm` (`manual|recompute`)
  - Auto PR detection on workout create; recompute on workout update/delete when PR was sourced from that workout
  - UI: historical 1RM tile + dialog on exercise detail (demo supports manual override + recompute via localStorage)
  - Integration coverage: `server/internal/workout/historical_1rm_pr_integration_test.go`
- Exercise detail API simplified:
  - `GET /api/exercises/{id}` now returns `{ exercise, sets }`
  - `exercise` includes `historical_1rm`, `historical_1rm_updated_at`, `historical_1rm_source_workout_id`
  - Works for zero-set exercises (metadata still present)
- Zero-set exercise detail regressions covered:
  - API shape regression: `server/internal/exercise/handler_zero_sets_regression_test.go`
  - UI render-path regression: `client/tests/e2e/demo/exercise-detail.test.ts`
- Exercise detail now includes computed best e1RM:
  - `GET /api/exercises/{id}` now includes `exercise.best_e1rm`
  - Historical 1RM card consumes API `best_e1rm` in authed mode (set-derived fallback only when field is missing)
- Historical 1RM read endpoint removed:
  - Removed `GET /api/exercises/{id}/historical-1rm`
  - Regenerated swagger + TS client (`server/docs/*`, `client/src/client/*`)
- Charting:
  - Legacy client-side `ChartBarVol` removed
  - Volume now comes from backend metrics-history as `Working-Set Volume` (`total_volume_working`) under Session Metrics
- Demo-mode parity:
  - localStorage key: `fittrack-demo-historical-1rm`
  - bootstrapped from working sets; lifecycle hooks on workout create/update/delete + exercise delete
  - Test: `client/tests/e2e/demo/exercise-historical-1rm.test.ts`
- Dev ergonomics:
  - Stack Auth optional for demo-mode (`VITE_PROJECT_ID` / `VITE_PUBLISHABLE_CLIENT_KEY` can be absent)

## Infra / Gotchas

- `server/docker-compose.yaml` volume path changed to `server/_db-data` to avoid Go tooling permission issues on `server/db-data`
- Local `.env` files created for tests/dev; gitignored; do not commit

## Next Chunk

- Optional: include `best_e1rm_source_workout_id` in `GET /api/exercises/{id}` to avoid set-derived fallback for workout link
- Add authenticated UI regression for `best_e1rm` render path when Stack Auth e2e env is configured
