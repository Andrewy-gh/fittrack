# Weekly muscle contributions

This informative Analytics view is enabled by the hypertrophy profile goal. It covers all owned exercises in the selected local week, independently of the exercise selected for strength or Session Metrics. The existing post-save assessment link supplies the same week to both sections. There are no growth verdicts, fractional totals, targets, or recommendations.

## Reviewed mapping v1

The assessment mapping is separate from catalog primary/secondary metadata. Each row below is a deliberate movement-family match to Pelland et al. Table 1 ([published paper](https://doi.org/10.1007/s40279-025-02344-w)); these matches are product interpretations, not evidence that every variant has identical stimulus. The investigation documents source limitations. Direct means main muscle worked; indirect means assisting muscle in the reviewed role.

| Catalog ID | Reviewed roles | Variant rationale |
| --- | --- | --- |
| Barbell_Bench_Press_-_Medium_Grip | Chest direct, triceps indirect | Medium-grip barbell bench matches the conventional horizontal press family; excludes close-grip and other press variants. |
| Triceps_Pushdown | Triceps direct | Dynamic cable elbow extension matches the triceps-extension family. |
| Barbell_Curl | Elbow flexors direct | Supinated dynamic elbow flexion matches curl family. Group label avoids claiming equal growth of individual flexors. |
| Hammer_Curls | Elbow flexors direct | Neutral-grip dynamic elbow flexion remains a flexor-group task; does not claim equivalent biceps stimulus to supinated curl. |
| Bent_Over_Barbell_Row | Elbow flexors indirect | Elbow flexion assists the reviewed rowing family; back roles remain unreviewed here. |
| Seated_Cable_Rows | Elbow flexors indirect | Same assisting elbow-flexion rationale for cable rowing; no equivalence of back stimulus assumed. |
| Wide-Grip_Lat_Pulldown | Elbow flexors indirect | Dynamic elbow flexion assists vertical pulling; wide grip does not justify direct flexor credit. |
| Barbell_Full_Squat | Quadriceps direct | Loaded knee extension in squat family. Group count does not imply uniform quadriceps-head response or hamstring credit. |
| Leg_Press | Quadriceps direct | Loaded knee extension matches leg-press family; other hip/ankle roles unresolved. |
| Dumbbell_Lunges | Quadriceps direct | Loaded knee extension matches lunge family. Count a logged set once; no assumed per-side doubling. |
| Leg_Extensions | Quadriceps direct | Dynamic isolated knee extension matches extension family. |
| Lying_Leg_Curls | Hamstrings direct | Dynamic knee flexion matches leg-curl family; does not imply all hamstring regions respond equally. |
| Incline_Dumbbell_Press | Triceps indirect | Elbow extension assists incline pressing; incline chest/shoulder roles remain unresolved. |
| Dumbbell_Shoulder_Press | Triceps indirect | Elbow extension assists shoulder pressing; other roles unresolved. |

Every unlisted exercise × muscle pair is unresolved, including other roles of these movements. There are no reviewed exclusions in v1. In particular, squats do not receive automatic hamstring credit, deadlifts do not inherit half-credit from catalog secondary labels, and planks remain unreviewed rather than being equated to dynamic repetition sets. Other muscle groups are outside this mapping. A missing mapping is not zero training.

## Contract and uncertainty

`GET /api/analytics/muscle-contributions?startDate=YYYY-MM-DD&timezone=IANA` is authenticated and owner-scoped. SQL joins enforce matching set, workout and exercise ownership with no result limit. The seven local calendar days use an exclusive end and inclusive sampled observation time, preserving DST and excluding future logs. Future starts and invalid dates/timezones are rejected.

Only working rows with positive reps and absent or finite nonnegative weight count. Invalid rows stay visible separately. Each of five muscle groups exposes direct, indirect, unresolved valid sets and separate invalid mapped/unresolved counts. Unresolved counts are conservative: even an exercise mapped to another muscle remains unreviewed for this muscle. These counts must not be summed into a whole-body score. The exercise list includes contributing and unmapped exercises, preserving custom names.

Known contributions stay visible when mapping is incomplete. No logs produce empty evidence, not a no-training assertion. Current weeks show counts so far; completed weeks also remain counts-only. Catalog version, mapping version and current-link basis are returned: editing a classification recalculates history, not an immutable snapshot. Mutation invalidation covers workout create/update/delete and exercise update/rename/delete.

No new profile data, schema migrations, prescription, AI, inferred effort or range of motion are introduced. Numeric comparison is deferred until counting units and a research reference are compatible.
