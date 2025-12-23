# Tasks: Workout Contribution Graph

## Relevant Files

*To be completed after sub-task generation*

### Notes

- Unit tests should typically be placed alongside the code files they are testing
- Backend follows Handler -> Service -> Repository pattern with sqlc-generated code
- Frontend uses TanStack Query with `useSuspenseQuery` for data fetching
- Existing `ContributionGraph` component from Kibo UI will be customized
- Use `npx jest [optional/path/to/test/file]` for frontend tests
- Use `go test ./...` for backend tests

## Tasks

- [ ] 1.0 Create Backend API Endpoint for Contribution Data
  - Implement the `GET /api/workouts/contribution-data` endpoint that returns daily workout aggregations for the past 52 weeks, including set counts, levels (0-4), and workout IDs for navigation.

- [ ] 2.0 Add Frontend Query Options and Cache Management
  - Create TanStack Query options for fetching contribution data and set up cache invalidation when workouts are created, updated, or deleted.

- [ ] 3.0 Add Accordion/Collapsible UI Component
  - Add the shadcn/ui Accordion or Collapsible component to the project if not already present, as it's required for the collapsible contribution graph wrapper.

- [ ] 4.0 Create WorkoutContributionGraph Component
  - Build the main wrapper component that displays the contribution graph in a collapsible accordion, handles tooltips, click navigation (including popover for multiple workouts), and shows an encouraging empty state for new users.

- [ ] 5.0 Integrate Contribution Graph into Workouts Page
  - Add the WorkoutContributionGraph component above the workout list on the `/workouts` page, configure default accordion state (expanded on desktop, collapsed on mobile), and ensure data is prefetched in the route loader.
