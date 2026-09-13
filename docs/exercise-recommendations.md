## Review update: sets, reps, and weight

The exercise card shows the suggestion and three numbered faces (1: low energy, 2: okay, 3: feeling good). There is no apply button; the user logs actual sets normally. The previous Details section, set-range editor, and feedback control are removed from this card; existing saved ranges and feedback remain preserved.

The current `exercise-plan-v2` suggestion adds a compact plan to the existing saved recommendation snapshot. Existing `working-sets-v1` snapshots still load unchanged.

- Uses the authenticated owner's goal, experience, avoided exercises, movement limitations, and last two exact-exercise sessions at or before now. History expires after 28 days. Warm-ups are excluded; a latest warm-up-only workout does not fall back to an older load.
- Goal sets a progression target: strength with intermediate/advanced experience uses 5–8 reps; endurance 12–15; other/missing goals or beginner strength use 8–12. These are explicit product defaults, not an individualized training program. No-history starting count is two sets, or three for intermediate/advanced. A saved set range still wins; its minimum is the plan's count.
- Normal repeats the minimum completed reps at the same uniform working weight. Great offers one extra rep below the target ceiling, without simultaneously increasing sets. Sluggish removes up to two reps and one set, flooring both at one.
- Two distinct recent workouts, each with at least two working sets at the same load and all reps at the target ceiling, allow a 2.5 lb increase (only for prior loads 50–2000 lb) and a return to the lower rep target. The planned set count must not exceed the latest count. The user selects the closest available load. Mixed/unknown weights and missing/stale history require the user to choose weight; no load is inferred from body size or experience.
- A listed movement limitation or an exact exercise-name match in avoided exercises suppresses the plan; free-text limitations are not interpreted as medical advice. Duration, equipment, location, and retrospective feedback do not currently change the calculation. This is not a whole-workout volume planner.
- The profile's primary goal is used (first goal if the primary field is absent). The goal and experience used remain part of the stored plan.
- Suggestions are advisory. They never insert or prefill actual sets. Saved snapshots include the complete plan separately from actual work.

Progression design reference: [ACSM progression guidance](https://pubmed.ncbi.nlm.nih.gov/19204579/). The exact thresholds above are conservative product choices and are not a validated coaching algorithm.

---

# Exercise working-set guidance

This first slice separates a durable baseline prescription, today's displayed recommendation, and the sets actually logged. It adds no reps, weights, RIR, muscle classifications, progression, or AI calls. Training Profile remains optional and is not an input to this policy.

## Initial policy: `working-sets-v1`

These are transparent starting heuristics, not proven optimal prescriptions or research-derived volume defaults.

1. Resolve the exact exercise ID within the authenticated user's exercises. A different exercise name or another user's exercise cannot supply history or a prescription.
2. Read the latest session at or before the server's current time, ordered by workout date and then workout ID. Count only `working` sets; warm-ups never contribute. Show its date and working-set count even if it is not eligible for a recommendation.
3. Prefer a user-entered prescription. Supported prescriptions have integer bounds `1 <= minimum <= maximum <= 20`. These are operational input bounds, not a recommended amount of exercise. A prescription remains usable without history or when history is stale.
4. Otherwise, use the latest session's working-set count as both bounds, only when its age is at most 28 × 24 hours (inclusive) and its count is between 1 and 20. Do not skip a more recent unsupported or warm-up-only session to select an older, more convenient count.
5. Normal and Great preserve both baseline bounds. Great never increases sets, reps, or load. Sluggish subtracts one from each bound, with a floor of one. The user remains free to do less, stop, or rest.
6. With no eligible baseline, return a null range and an explanation. Unsupported readiness and invalid prescription inputs do not produce an invented range. Invalid prescription writes receive HTTP 400. Missing/foreign exercise IDs receive 404; missing authentication receives 401.

The 28-day window, 20-set input ceiling, and one-set decrement are deliberately simple product heuristics. They have not been validated as personalized training advice. They do not use profile goals, experience, equipment, or injury information.

## Evidence-derived defaults are deferred

No research-derived default is implemented or claimed. An exact-exercise history and a readiness label cannot establish appropriate weekly per-muscle volume, recovery, or a user's overall program. Per-exercise sets in one session are not interchangeable with per-session totals across exercises or weekly sets per muscle. Before adding a research-derived default, document credible guidelines/systematic reviews, applicable population and training context, units of volume, and limitations. No muscle taxonomy was added to fill that information gap.

A prescription can be entered by the user or transcribed from their coach. The app records it as a user-entered prescription; it does not verify coaching credentials or infer that it is evidence-based. Existing browser-stored exercise goals are unchanged and are never silently migrated into prescriptions.

## API and persistence

- `GET /api/exercises/{id}/recommendation?readiness=normal|great|sluggish` is read-only. It returns the range, baseline snapshot, source, previous session, explanation, and policy version.
- `PUT /api/exercises/{id}/prescription` accepts `{ "baseline": { "min": 3, "max": 4 } }`; explicit `null` clears the prescription. Editing fields alone does not save it.
- `POST /api/workouts` accepts optional `recommendations`, one snapshot per named exercise. Each holds the displayed recommendation and optional `too_little`, `about_right`, or `too_much` feedback.
- `GET /api/workouts/{id}/recommendations` returns saved snapshots without recomputing them.

Migration 00031 adds the owner-protected `exercise_prescription` table and `workout.recommendation_context` JSONB. Workout context is written in the same transaction as actual sets, using the exercise IDs resolved during that save. Validation rejects mismatched exercises and malformed snapshots. Existing workout RLS protects the context; prescription RLS also checks ownership of the target exercise. The migration grants the restricted runtime role access when it exists, and the provisioning allowlist includes the new table.

Readiness clicks make read requests only. They do not create persistent recommendation rows. Guidance is retained in the existing user-scoped local workout draft, then persisted on workout save. Removed exercises are filtered from the submitted context. Repeat-workout behavior retains its existing set-copy semantics and does not copy old recommendation/readiness/feedback snapshots. Editing a saved workout preserves its original guidance record, even if actual sets later change.

Recommendations never populate the actual set list. A user must still log sets through the existing controls. Loading or guidance errors do not prevent ordinary workout entry. Saving a workout while no guidance response has been displayed records no recommendation for that exercise.

## Migration ordering and the earlier local migration 29

Main already contains profile migration 30. New databases and upgrades from main apply `00031_exercise_prescriptions.sql` normally. It has the same schema operations as the earlier, unmerged recommendation migration 29. Never replay those operations on a database that already applied 29.

The source checkout and its database are not modified by this PR. To adopt an existing **local-only** 29 database:

1. Confirm the host is loopback and take a full `pg_dump -Fc` backup. Restore it into a separate disposable database first.
2. In a temporary migration directory, retain the original 29 together with main's migrations through 30. Apply 30 there if it is missing. Do not use the new 31 yet.
3. Compare a `pg_dump --schema-only --no-owner --no-privileges` of that restored database with a fresh database migrated through 31. Ignore only pg_dump's generated `\restrict`/`\unrestrict` keys. Stop on any other difference. This includes constraints and RLS policies, not just table names.
4. After schema equivalence is established, adopt the existing objects by updating only Goose history in a transaction:

```sql
BEGIN;
DO $$
BEGIN
  IF (SELECT count(*) FROM goose_db_version WHERE version_id = 29 AND is_applied) <> 1
     OR (SELECT count(*) FROM goose_db_version WHERE version_id = 30 AND is_applied) <> 1
     OR EXISTS (SELECT 1 FROM goose_db_version WHERE version_id >= 31) THEN
    RAISE EXCEPTION 'Unexpected migration history; stop and inspect';
  END IF;
END $$;
UPDATE goose_db_version SET version_id = 31, id = DEFAULT
WHERE version_id = 29 AND is_applied;
COMMIT;
```

The new history ID matters: Goose obtains the current version from history order. Merely changing `version_id` leaves 30 as the newest history entry. Keep the pre-adoption dump as the audit record of the original 29.

5. Run Goose `version`, `status`, and `up` with the new migration directory. Expect version 31 and no SQL to apply. Verify the existing prescription and workout snapshots remain unchanged, then rerun the restricted-role smoke test. Rehearse rollback only on a disposable copy; down from 31 removes recommendation data.
6. Only after the restored-copy rehearsal succeeds, repeat the backup, schema comparison, and history adoption against the intended local database. Do not use this adoption procedure on production, which never received 29.

## Profile and authentication prerequisites

This branch preserves main's multi-goal profile, owner-scoped queries, settings outlet/index, and profile JSON encoding fix. Profile values are optional and are not recommendation-policy inputs. Local Stack authentication changes from other worktrees are not bundled here.

## Verification

The recommendation integration test uses the same simple-query protocol as the application, a local PostgreSQL database, `SET LOCAL ROLE fittrack_app`, two owners, and rollback-only fixtures. Set `RECOMMENDATION_TEST_DATABASE_URL` to the provisioned local database before running it. CI runs it in the restricted-runtime step, after provisioning, rather than against the ordinary unprovisioned suite database. It verifies baseline/history selection, warm-up exclusion, snapshot persistence, actual-set independence, and account isolation. It must never run against a remote database.

```text
server: go test -short ./...
server: go test ./internal/recommendation -run TestRecommendationPersistenceWithRLS -count=1
server: go vet ./internal/recommendation ./internal/workout ./internal/app
client: bun run test -- src/features/workouts/components/form/__tests__/exercise-recommendation-panel.test.tsx src/features/workouts/hooks/use-new-workout-form-workflow.test.tsx src/lib/local-storage.test.ts
client: bun run tsc
client: bun run lint
client: bun run knip
client: bun run build
```

The September 12 local rehearsal used an isolated PostgreSQL 17 container. Fresh migrations through 31 and the restricted-role SQL smoke test passed. A separate database migrated through the original 29 and profile 30 had an identical schema to fresh 31; its history was adopted without replaying DDL. The original source checkout and its local database were preserved.

Manual walkthrough:

1. Save a synthetic preference in Training Profile and refresh.
2. Start a workout and choose an existing exercise. Inspect the previous date and working-set count; warm-ups do not contribute.
3. Save baseline 3–4. Normal and Great show 3–4; Sluggish shows 2–3.
4. Log actual sets separately and optionally choose feedback. Save the workout.
5. Reopen the saved workout and refresh. Check actual sets and recorded guidance separately.
6. Change the current baseline and reopen the old workout: its snapshot must remain unchanged. Repeat the workout: copied actual sets must not carry old recommendation context.

### Browser persistence evidence (September 12)

Using a synthetic local account and the restricted application database role, Chrome verified profile save/refresh, a previous session with two working sets plus one excluded warm-up, saved baseline 3–4, Normal/Great 3–4, and Sluggish 2–3. One actual 25 lb × 6 working set and About right feedback were saved. Reload displayed one actual set and the original snapshot. After changing the current baseline to 6–8, the saved snapshot stayed 2–3 with baseline 3–4. Repeat then saved one copied actual set and zero snapshots. Anonymous recommendation and saved-context requests returned 401; two-owner isolation was exercised by the database integration and SQL smoke tests.

Desktop screenshots confirmed separate actual-set and saved-guidance sections. At a 390 px viewport, DOM measurements and controls were verified with no horizontal overflow; the browser screenshot API timed out at that override, so no mobile screenshot is claimed.

The PR body must include `<!-- skip-preview -->`: this deliberately skips the preview deployment job and its database migrations while retaining test CI. No production or preview database writes are part of this delivery.

Full backend integration suite (go test -p 1 ./...), backend short suite, targeted vet, frontend typecheck/lint/knip/build, and all 526 frontend tests passed locally. The frontend full-suite pass used two workers after the initial high-parallelism run had three timing failures (523 passed); no product changes were made for that rerun. Local Bun was 1.4.1; CI uses the repository-pinned 1.4.2. Build retained existing chunk-size and Browserslist-age warnings.
