# Workout List Options Implementation Plan

This plan outlines the steps to add filtering and pagination options to the workouts page.

## Phase 1: Component Installation

- [ ] Install shadcn components:
  - `bunx --bun shadcn@latest add calendar-04`
  - `bunx --bun shadcn@latest add calendar-05`
  - `bunx --bun shadcn@latest add pagination`
  - `bunx --bun shadcn@latest add toggle`
  - `bunx --bun shadcn@latest add accordion`

## Phase 2: Basic UI Integration

- [x] Add state for filters (date range, focus area, sort order, pagination).
- [x] Implement UI for:
  - [x] Date range picker (showing both variants for evaluation).
  - [x] Focus area dropdown.
  - [x] Ascending/Descending toggle/dropdown.
  - [x] Pagination.
  - [x] Items per page dropdown.
- [x] Place these components in `client/src/routes/_layout/workouts/index.tsx`.

## Phase 3: Clustering & Layout Optimization

- [x] Evaluate UI density.
- [x] If clustered, move filters (Date, Focus, Sort) into an `Accordion`.
- [x] Ensure Pagination and Items Per Page remain always visible.

## Phase 4: Evaluation & Cleanup

- [ ] Review UI with the user.
- [ ] Decide on preferred calendar variant.
- [ ] Decide on preferred sort control (toggle vs dropdown).
- [ ] Finalize layout.
