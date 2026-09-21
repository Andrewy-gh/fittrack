# Visual verification

Screenshots of the actual `AnalyticsDashboard` and `TrainingAssessment` components from `feature/training-assessment`, served by this worktree's Vite runtime on port 5199, in Chromium dark mode:

- `desktop-dark.png`: 1280 × 1000.
- `mobile-dark.png`: 390 × 844; no horizontal document overflow.

Fixture: strength and hypertrophy selected; unclassified bench press; completed week with one and five working sets in separate sessions. API responses were intercepted at the network boundary. A temporary dashboard preview harness provided the authenticated user ID and router/query providers; the harness was removed after verification. This is component/dashboard visual verification, not authenticated live backend end-to-end verification. The normal app without Stack credentials falls back to demo mode and its account control requires StackProvider, so no real login claim is made.

Final browser check reported no page errors. PostgreSQL behavior was verified separately through the assessment integration test against a disposable migrated test database; production/user data was not used.

## Post-save context regression verification

`post-save-backdated-dark.png`, `post-save-current-dark.png`, and `post-save-proof.json` verify the actual save workflow, real Sonner action, router search validation, Analytics page, and assessment component together in Chromium. Only API responses and authenticated user context were fixtures, supplied through a temporary preview entry point (removed afterward).

Both cases started with `Alphabetical first` as the first exercise. Clicking **Save workout**, then **View assessment**, selected `Saved press` (ID 9):

- Backdated existing exercise: saved timestamp `2026-08-17T00:30:00Z` is Sunday August 16 in `America/New_York`. Destination requested the correct local week beginning `2026-08-10`, not the UTC Monday or previous current week.
- Newly created exercise: ID 9 was absent from the original cached exercise list. The action refreshed the owned list, resolved the new ID, and requested week beginning `2026-09-21`. The current-week UI showed counts without comparison badges.

The proof JSON records POST/GET requests, selected exercise, timezone, page errors, and horizontal overflow. Both runs had zero page errors and no horizontal overflow. This verifies the frontend flow with network fixtures; it is not a real authentication or production database test.
