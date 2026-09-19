# Exercise catalog: first metadata slice

## Provenance and review

The 25 checked-in records in [`catalog.json`](../server/internal/exercisecatalog/catalog.json) are a field-preserving subset of [yuhonas/free-exercise-db](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f), pinned to commit `a859101d633a01c4a1a920d6a8ce41dabba0705f`. Source: [`dist/exercises.json`](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/dist/exercises.json). Reviewed on 2026-09-19 for identity, equipment, primary/secondary roles, and suitability for this initial common-exercise subset.

The pinned [`LICENSE.md`](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/LICENSE.md) is the **Unlicense**, dedicating the software to the public domain and permitting copying, modification, and distribution. Its text is retained unchanged in [`server/internal/exercisecatalog/LICENSE.md`](../server/internal/exercisecatalog/LICENSE.md). The upstream [README](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/README.md) credits the original dataset to [Ollie Jennings / exercises.json](https://github.com/wrkout/exercises.json). This import includes only identifiers, names, equipment and muscle labels; it does not import images, instructions, site code or branding.

The local catalog SHA-256 is `e8eb652d64598edcaaed6c0fd8bbf48af2a9bf3d9ffe576542b1e85ea501a36f`.

## Mapping choices

- Keep upstream IDs as stable catalog IDs. A future display-name change must not rename an ID or repoint an existing ID to a different movement.
- Copy `name`, `equipment`, `primaryMuscles`, and `secondaryMuscles` exactly. Empty secondary lists mean “none listed in the source.” They do not assert that no other muscles participate.
- Retain source vocabulary such as `body only`, `middle back`, and `lower back`. Equipment is one source label, not a complete equipment inventory (a dumbbell press can also require a bench).
- Retain source role assignments, including **Barbell Deadlift → lower back as primary** and **Hammer Curls → biceps with no listed secondary muscles**. These are catalog descriptions, not independently validated physiological rankings. No weights, set credits, weekly calculations, targets, or training recommendations are derived from them.
- Preserve specificity in names (medium-grip bench press, full squat, wide-grip pulldown). Do not infer these variants from names such as “Bench,” “Squat,” or “Pulldown.” No mandatory variation fields.
- Keep user-owned exercise IDs and names independent of catalog identity. Linking, changing or clearing classification modifies only `catalog_id` and the exercise's `updated_at`. It does not replace exercises or alter sets, workout IDs, historical 1RM values/sources, or history.

## Reviewed selection

Each row is copied unchanged from the pinned record with the corresponding catalog ID.

| Catalog ID / source | Name | Equipment | Primary | Secondary |
| --- | --- | --- | --- | --- |
| [Barbell_Bench_Press_-_Medium_Grip](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Barbell_Bench_Press_-_Medium_Grip.json) | Barbell Bench Press - Medium Grip | barbell | chest | shoulders, triceps |
| [Barbell_Curl](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Barbell_Curl.json) | Barbell Curl | barbell | biceps | forearms |
| [Barbell_Deadlift](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Barbell_Deadlift.json) | Barbell Deadlift | barbell | lower back | calves, forearms, glutes, hamstrings, lats, middle back, quadriceps, traps |
| [Barbell_Full_Squat](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Barbell_Full_Squat.json) | Barbell Full Squat | barbell | quadriceps | calves, glutes, hamstrings, lower back |
| [Bent_Over_Barbell_Row](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Bent_Over_Barbell_Row.json) | Bent Over Barbell Row | barbell | middle back | biceps, lats, shoulders |
| [Crunches](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Crunches.json) | Crunches | body only | abdominals | None listed |
| [Dumbbell_Bench_Press](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Dumbbell_Bench_Press.json) | Dumbbell Bench Press | dumbbell | chest | shoulders, triceps |
| [Dumbbell_Lunges](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Dumbbell_Lunges.json) | Dumbbell Lunges | dumbbell | quadriceps | calves, glutes, hamstrings |
| [Dumbbell_Shoulder_Press](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Dumbbell_Shoulder_Press.json) | Dumbbell Shoulder Press | dumbbell | shoulders | triceps |
| [Hammer_Curls](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Hammer_Curls.json) | Hammer Curls | dumbbell | biceps | None listed |
| [Incline_Dumbbell_Press](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Incline_Dumbbell_Press.json) | Incline Dumbbell Press | dumbbell | chest | shoulders, triceps |
| [Leg_Extensions](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Leg_Extensions.json) | Leg Extensions | machine | quadriceps | None listed |
| [Leg_Press](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Leg_Press.json) | Leg Press | machine | quadriceps | calves, glutes, hamstrings |
| [Lying_Leg_Curls](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Lying_Leg_Curls.json) | Lying Leg Curls | machine | hamstrings | None listed |
| [One-Arm_Dumbbell_Row](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/One-Arm_Dumbbell_Row.json) | One-Arm Dumbbell Row | dumbbell | middle back | biceps, lats, shoulders |
| [Plank](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Plank.json) | Plank | body only | abdominals | None listed |
| [Pullups](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Pullups.json) | Pullups | body only | lats | biceps, middle back |
| [Pushups](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Pushups.json) | Pushups | body only | chest | shoulders, triceps |
| [Romanian_Deadlift](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Romanian_Deadlift.json) | Romanian Deadlift | barbell | hamstrings | calves, glutes, lower back |
| [Seated_Cable_Rows](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Seated_Cable_Rows.json) | Seated Cable Rows | cable | middle back | biceps, lats, shoulders |
| [Seated_Calf_Raise](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Seated_Calf_Raise.json) | Seated Calf Raise | machine | calves | None listed |
| [Side_Lateral_Raise](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Side_Lateral_Raise.json) | Side Lateral Raise | dumbbell | shoulders | None listed |
| [Standing_Calf_Raises](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Standing_Calf_Raises.json) | Standing Calf Raises | machine | calves | None listed |
| [Triceps_Pushdown](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Triceps_Pushdown.json) | Triceps Pushdown | cable | triceps | None listed |
| [Wide-Grip_Lat_Pulldown](https://github.com/yuhonas/free-exercise-db/blob/a859101d633a01c4a1a920d6a8ce41dabba0705f/exercises/Wide-Grip_Lat_Pulldown.json) | Wide-Grip Lat Pulldown | cable | lats | biceps, middle back, shoulders |

## Persistence and API

Migration `00033_exercise_catalog.sql` adds a nullable `exercise.catalog_id` with a check constraint containing the reviewed IDs. Existing rows remain NULL; there is no matching migration or bulk backfill. The schema snapshot matches the migration. Catalog JSON is embedded into the API binary; runtime operation requires no upstream requests. The catalog integrity test checks the selected IDs against the migration. Future catalog additions require an explicit reviewed catalog change and a migration extending the constraint.

- `GET /api/exercise-catalog`: authenticated, returns the 25 entries.
- `GET /api/exercises/{id}`: includes optional `exercise.catalog`, resolved from the saved link. An unclassified exercise omits it.
- `PUT /api/exercises/{id}/catalog`: body `{"catalog_id":"Pushups"}` links or changes; `{"catalog_id":""}` clears. Missing/null/unknown IDs are rejected. A foreign or missing exercise returns 404. The UPDATE itself includes the current user ID.
- Workout create/update inputs accept optional `exercises[].catalog_id`. New exercises are linked in the workout transaction. Reusing an existing `(user_id, name)` preserves its ID and classification, even if a different catalog ID is submitted. Reclassification requires the explicit exercise endpoint. Repeated names with conflicting selections are rejected before writes.
- Free-text exercise creation, imported workouts and existing clients continue to work with no catalog ID. A name identical to a catalog entry is still unclassified unless the user explicitly selects it.

## Frontend behavior

Signed-in workout entry offers **Browse catalog**, with search by name, equipment or muscle and metadata shown before choosing. Selection fills the draft name and catalog ID; server-owned metadata resolves from that ID. No exercise is persisted until the workout is saved. Free-text entry remains available. If the name is already in the user's exercise list, the existing exercise is reused and retains its classification.

The exercise detail page offers **Classify exercise**, **Change classification**, and **Clear**, with equipment and both muscle roles displayed automatically. The custom exercise name stays unchanged. Loading, empty search, failed catalog load/retry, pending save and failed save states are handled. Catalog controls are omitted in demo mode, which has no user-owned backend exercises.

## Validation

- Catalog integrity and pinned-source equality, including the retained license.
- Real Postgres handler tests: unauthenticated access, foreign ownership, missing exercise, malformed/unknown ID, link/change/clear, unchanged custom name/ID/sets/history/1RM and database constraint enforcement.
- Real Postgres workout service tests: explicit catalog selection through create and edit, unclassified free-text names, existing-name reuse and rejection of invalid/conflicting IDs before writes.
- Frontend behavior tests: search, metadata, classify/change/clear, failed saves, failed catalog fetch/retry, no-match behavior, explicit catalog draft selection and typed-name non-classification.

Validation run on 2026-09-19: the full Go suite passed with a disposable Postgres database (`go test -p 1 ./...`); all 532 frontend tests passed. Client lint, knip, TypeScript and production build passed. The affected Go packages and seven focused frontend tests passed again after final cleanup. Two live Chromium tests passed at 1280px and 390px, including saved catalog selection, change/clear/reclassify, custom-name/history preservation after reload, and dark-mode screenshots. Browser coverage used local test authentication, not an external identity-provider login.

Reproduce browser coverage with the local-auth setup from [development.md](development.md), then run `E2E_LOCAL_AUTH_ENABLED=true bun run test:e2e -- tests/e2e/auth/exercise-catalog.test.ts`. Use an isolated migrated database: the test creates and cleans up its own exercises and workouts. Both Vite Stack configuration values must be nonempty for the existing client local-auth path to mount.
