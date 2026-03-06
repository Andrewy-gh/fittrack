# Dependency Upgrade – Next Batch

This tracks the larger-but-safe-later frontend upgrades intentionally deferred from the March safe refresh.

## Scope

- `@tanstack/react-router` / `@tanstack/react-router-devtools` / `@tanstack/router-plugin`
- `@tanstack/react-form`
- `recharts`

## Why deferred

These are higher-surface-area upgrades likely to impact routing behavior, form ergonomics, and chart rendering.

## Plan (safe)

1. Create branch: `chore/deps-next-batch`
2. Upgrade one family at a time:
   - A) TanStack Router stack
   - B) TanStack Form
   - C) Recharts
3. After each family:
   - `cd client && bun run tsc`
   - `cd client && bun run test`
   - smoke-check key pages:
     - `/analytics`
     - `/workouts`
     - auth/route transitions
4. Commit per family (small rollback surface).
5. Merge only when all pass.

## Acceptance criteria

- Typecheck passes
- Unit tests pass
- No regressions in analytics chart interaction
- No route navigation regressions
- No form submission/validation regressions

## Suggested commit split

- `chore(deps): upgrade tanstack router stack`
- `chore(deps): upgrade tanstack react-form`
- `chore(deps): upgrade recharts`
