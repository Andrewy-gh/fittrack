# Training evidence

`workout.EvidenceService` is a backend-only, callable summary of logged training evidence. It has no HTTP route, UI, chat-tool integration, target verdict, or recommendation behavior.

```go
evidenceService := workout.NewEvidenceService(queries, time.Now)
evidence, err := evidenceService.Summarize(ctx, workout.TrainingEvidenceRequest{
    StartDate: "2026-09-20",
    Timezone:  "America/New_York",
})
```

`ctx` must contain the authenticated user through the existing `user.Current` context pattern. The service returns an unauthorized error without querying when that context is absent.

## Period and observation semantics

The request requires both:

- `StartDate`: exactly `YYYY-MM-DD`.
- `Timezone`: an explicit IANA timezone. `UTC` is valid; empty, `Local`, malformed, and non-representable local-midnight requests are rejected.

The period is exactly seven **local calendar days**, `[start local midnight, start + 7 local calendar days midnight)`. It is not a 168-hour interval, so its UTC duration changes across daylight-saving transitions. The response records the requested `StartDate` and derived exclusive `EndDate`, timezone, `StartAt`, `EndAt`, one captured `ObservedAt`, and `Partial`.

The injected server clock is captured exactly once per call. A period whose local start instant is after `ObservedAt` is rejected. When `EndAt > ObservedAt`, `Partial` is true and SQL includes only workouts at or before `ObservedAt`. The SQL interval is otherwise inclusive at `StartAt` and exclusive at `EndAt`.

## Counts and retrieval

A single uncapped generated SQL query reads every owned workout in the period. It retains workouts without sets through outer joins, filters workouts, sets, and exercises by the authenticated user, and excludes corrupt cross-user set/exercise joins instead of exposing or counting them. Query failures return an error, never an empty or partial success.

`Activity` reports separately:

- `LoggedWorkoutCount`: all logged workouts in the bounded, observed period, including empty and warmup-only workouts.
- `WorkingSetSessionCount`: distinct workouts that contain at least one exact `set_type = 'working'` row.
- `WorkingSetLocalDayCount`: distinct requested-timezone local dates with at least one working set.

Every eligible stored row with exact `set_type = 'working'` counts once. Weight (including null or zero) and reps do not alter the count. Warmups and every other set type do not become working sets.

## Classification and muscles

`Classification` contains total, classified, and unclassified working-set counts plus one of these statuses:

- `not_applicable` when there are zero working sets; it is not `100%` coverage.
- `unclassified` when working sets exist but none have a current catalog link.
- `partial` when both classified and unclassified working sets exist.
- `complete` when every working set has a current embedded-catalog link.

Classification uses only the current `exercise.catalog_id`, resolved with `exercisecatalog.Find`; names are never guessed. A non-null catalog ID missing from the embedded catalog fails explicitly as inconsistent data.

Each `MuscleEvidence` item is one catalog muscle label in either `primary` or `secondary` role. It reports working-set count, distinct working-set sessions, and distinct local working-set days. A qualifying set is counted once for each label listed in that role. Session and day membership are deduplicated across exercises. Items are deterministic: primary roles first and alphabetized by muscle, then secondary roles alphabetized by muscle.

Historical periods intentionally recalculate against current exercise links and the current embedded catalog. `Basis.ClassificationBasis` is `current_exercise_catalog_link`; this service makes no immutable historical-assessment promise. `Basis.CalculationPolicyVersion` is `training-evidence-v1`.

`Basis.CatalogVersion` comes from `exercisecatalog.Version()`. It is SHA-256 over canonical catalog content: Go's `json.Marshal` of the decoded, ordered `[]exercisecatalog.Entry`. This makes the version stable across CRLF/LF and formatting changes while changing it when entry content or list order changes.

## Limitations

- Catalog muscle roles are descriptive upstream metadata, not validated stimulus.
- Unclassified working sets are unknown muscle training, not zero training for every muscle.
- Classified coverage describes current links in logged data, not complete real-world training logs.
- Empty secondary lists mean no secondary muscles are listed in the catalog, not that no secondary participation occurred.
- Counts across muscle labels and roles are not additive: one set can appear under multiple labels.
- The summary makes no load-volume, effort, unilateral-doubling, fractional-credit, progress, plateau, target-attainment, or hypertrophy claim.
- Zero logged workouts means no workouts were logged in this period; it does not prove no training occurred.
