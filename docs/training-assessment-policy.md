# Strength session reference: first assessment slice

Policy version: `strength-session-reference-v1`.

This implementation proposal follows the user's authorization to begin the first assessment slice. Its precise thresholds are research-informed implementation choices, not a claim that the user individually approved each scientific or product rule. See [research rationale and limitations](training-assessment-research.md).

## Scope

Analytics shows an exercise-level set comparison beside existing Session Metrics when strength is a selected profile goal. It is a general healthy-adult reference, not verified personal suitability, a prescription, a minimum effective dose, or a prediction of progress. There is no combined weekly physiological verdict or cross-goal score.

The reference is 2–3 sets per exercise per session, drawn from ACSM's 2026 guidance, with Krieger's multiple-set research as context. The application compares with its lower edge. Four or more sets are neither excessive nor optimal by this assessment.

## Counting and states

One exact `working` row with positive reps and absent, zero, or finite nonnegative weight contributes one set. Warmups and other types do not qualify. An invalid working row is counted separately and makes that session `cannot_assess`, without silently turning it into a zero. Existing `training-evidence-v1` raw counts remain unchanged.

A session is a distinct saved workout. Distinct local training days are descriptive. No frequency requirement is imposed on every exercise, and no minimum number of sessions is needed to show a numerical comparison. Catalog classification is unnecessary for exercise-level counts; names are not guessed or combined.

- One eligible set: `below_reference`.
- Two or more eligible sets: `reference_met`, explicitly the lower reference in logged training.
- No working sets, invalid working rows, or an unfinished period: `cannot_assess`, with a specific reason.
- No logged exercise sessions: empty evidence, never a below-reference verdict.

A 1-set workout plus a 5-set workout remains one below-reference and one reference-met session. An average of three must not erase the difference. The period summary reports how many sessions reached the lower reference; it does not claim enough weekly training.

## Time

The API accepts a start date and IANA timezone. The window is seven local calendar days with an exclusive end and one captured observation time; future-dated workouts are excluded. Calendar behavior reuses the tested training-evidence period logic, including DST. The UI defaults to the last completed Monday–Sunday week, a product convention rather than a biological boundary. Current-week navigation shows facts only, with no comparison badges. There is no extrapolation or two-week minimum.

## Loading, goals, and limitations

Existing Session Metrics remain informational. Historical 1RM can be manually entered or estimated, its update date is not necessarily a test date, and session fallback is an estimate. No 80% pass/fail assessment is added. Missing weight does not suppress a set-count comparison.

Other selected goals receive concise coverage explanations: hypertrophy attribution/reference compatibility remains unresolved; endurance needs muscular/aerobic distinction and appropriate measurements; general fitness includes activity beyond these logs; weight loss cannot be inferred from workout logs alone; mobility lacks range-of-motion measurements. Absence of strength does not silently substitute a strength assessment. No goals prompts profile setup.

Health, age, effort, real-world log completeness, and deload intent are not inferred. This slice adds no persistent profile or health fields and no AI calls. Assessment responses are computed from existing owned records. Recommendations, personal targets, and new data collection remain deferred.

## Contract and verification

`GET /api/exercises/{id}/assessment?startDate=YYYY-MM-DD&timezone=IANA` requires authentication and exercise ownership. The query includes only owned set/workout/exercise joins, has no row cap, and fails on retrieval errors. Responses expose period/provenance, valid and invalid per-session counts, reason codes, source, applicability, and limitations.

Acceptance cases include 1+5 sessions, a single two-set session, 4+ sets, warmup-only/empty evidence, null/zero weight, invalid legacy data, unclassified exercises, incomplete periods, timezone/DST boundaries, future exclusion, and ownership isolation. Frontend checks cover selected-goal behavior, period navigation, mixed outcomes, current-week suppression, and post-save navigation.

## Validation record (2026-09-21)

- Focused frontend suite: 21 tests passed across assessment, Analytics, and workout-save workflow.
- Frontend TypeScript and lint passed; production build passed (existing chunk-size and stale Browserslist warnings).
- Go workout/app short suites and vet passed. Assessment service/HTTP tests and ownership/window integration passed against a disposable migrated PostgreSQL 17 database.
- Actual dashboard/component dark-mode screenshots at desktop and mobile sizes are in `output/training-assessment/`; no horizontal overflow or final browser page errors. Network fixtures and a temporary preview harness were used; live authenticated frontend-to-backend end-to-end behavior has not been verified.
- No merge, deployment, or publication performed. No new profile fields, migrations, or AI calls.

### Post-save navigation correction

The saved-workout action captures the first submitted exercise name exactly as persisted and the workout's Monday-start local week before clearing the draft. On click, it fetches the current owned exercise list and resolves that exact name, including newly created exercises whose IDs were unavailable before saving. It passes the resolved ID and validated `assessmentWeek` date to Analytics. Failed/missing lookup reports an error without undoing the successful save or opening unrelated evidence. Ordinary Analytics navigation still defaults to the previous completed week; an explicit current-week destination remains counts-only.

Regression verification uses the real form, mutation, query cache, toast, router, and destination UI with the API adapter mocked. Tests cover existing/backdated and new/current exercise destinations plus failed lookup, with separate malformed-date rejection checks. Chromium proof additionally covers a Sunday-local/Monday-UTC boundary in New York; see `output/training-assessment/post-save-proof.json`.

Post-save fix checks: 30 focused frontend tests passed across seven files; the three destination regressions also passed again after test-fixture lint cleanup. TypeScript, lint, and `git diff --check` passed. No backend changes were made for this fix.
