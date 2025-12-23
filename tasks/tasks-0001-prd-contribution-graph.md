# Tasks: Workout Contribution Graph

## Relevant Files

### Backend
- `server/query.sql` - Add new `GetContributionData` query for fetching 52-week workout aggregations
- `server/internal/workout/models.go` - Add response types for contribution data
- `server/internal/workout/repository.go` - Add repository method for contribution data with RLS handling
- `server/internal/workout/service.go` - Add service method with dynamic threshold/percentile calculation
- `server/internal/workout/handler.go` - Add HTTP handler with Swagger documentation
- `server/cmd/api/routes.go` - Register the new route

### Frontend - Data Layer
- `client/src/lib/api/workouts.ts` - Add query options and update mutations for cache invalidation

### Frontend - Components
- `client/src/components/ui/collapsible.tsx` - New shadcn/ui Collapsible component
- `client/src/components/workouts/workout-contribution-graph.tsx` - Main wrapper component with accordion, interactivity, and styling

### Frontend - Integration
- `client/src/routes/_layout/workouts/index.tsx` - Integrate component and add to route loader

### Notes

- Backend follows Handler -> Service -> Repository pattern with sqlc-generated code
- Frontend uses TanStack Query with `useSuspenseQuery` for data fetching
- Existing `ContributionGraph` component from Kibo UI (`client/src/components/kibo-ui/contribution-graph/`) will be customized
- Use `go test ./...` for backend tests
- The contribution graph uses `data-[level="X"]` attributes for color styling via Tailwind

---

## Tasks

- [x] 1.0 Create Backend API Endpoint for Contribution Data
  - [x] 1.1 Add `GetContributionData` SQL query to `server/query.sql` that aggregates daily working set counts, groups by date, and returns workout IDs array for the past 52 weeks
  - [x] 1.2 Run `sqlc generate` to generate the Go code from the new query
  - [x] 1.3 Add `ContributionDataResponse` and `ContributionDay` types to `server/internal/workout/models.go` with fields: date, count, level (0-4), and workout_ids
  - [x] 1.4 Add `GetContributionData` method to the repository interface and implement it in `repository.go` with proper context timeout and RLS error handling
  - [x] 1.5 Add `GetContributionData` method to the service layer that calculates dynamic level thresholds using percentiles (25th, 50th, 75th) when user has 10+ workout days, otherwise use static thresholds (0, 1-5, 6-10, 11-15, 16+)
  - [x] 1.6 Add `GetContributionData` handler in `handler.go` with Swagger documentation comments
  - [x] 1.7 Register the route `GET /api/workouts/contribution-data` in `server/cmd/api/routes.go`
  - [x] 1.8 Test the endpoint manually or with unit tests to verify correct data aggregation and level calculation

- [x] 2.0 Add Frontend Query Options and Cache Management
  - [x] 2.1 Add `contributionDataQueryOptions()` function to `client/src/lib/api/workouts.ts` wrapping the generated TanStack Query options
  - [x] 2.2 Update `useSaveWorkoutMutation` to invalidate contribution data query key on success
  - [x] 2.3 Update `useDeleteWorkoutMutation` to invalidate contribution data query key on success
  - [x] 2.4 Verify TypeScript types are correctly generated for the new endpoint response

- [x] 3.0 Add Collapsible UI Component
  - [x] 3.1 Add the shadcn/ui Collapsible component to `client/src/components/ui/collapsible.tsx` (use `npx shadcn@latest add collapsible` or manually create based on Radix UI)
  - [x] 3.2 Verify the component exports `Collapsible`, `CollapsibleTrigger`, and `CollapsibleContent`

- [ ] 4.0 Create WorkoutContributionGraph Component
  - [x] 4.1 Create `client/src/components/workouts/workout-contribution-graph.tsx` with the basic component structure and data fetching using `useSuspenseQuery`
  - [ ] 4.2 Implement the Collapsible wrapper with a header/trigger showing "Activity" or similar label with expand/collapse icon
  - [ ] 4.3 Implement lazy rendering so the ContributionGraph DOM is only rendered when the collapsible is open (`{isOpen && <ContributionGraph ... />}`)
  - [ ] 4.4 Customize ContributionGraphBlock colors to use the primary (orange) color scheme: level 0 = muted, levels 1-4 = primary with 20%/40%/60%/80% opacity
  - [ ] 4.5 Add ContributionGraphFooter with ContributionGraphLegend showing Less/More indicator
  - [ ] 4.6 Implement tooltip on hover displaying formatted date (e.g., "Monday, Jan 15, 2025") and working set count (e.g., "12 working sets")
  - [ ] 4.7 Implement click handler for days with a single workout to navigate directly to `/workouts/{id}`
  - [ ] 4.8 Implement click handler for days with multiple workouts to show a Popover listing each workout with time/focus, allowing user to select which workout to navigate to
  - [ ] 4.9 Implement empty state: when user has no workouts in 52-week period, display an encouraging message (e.g., "Start your fitness journey! Log your first workout to see your progress here.")
  - [ ] 4.10 Ensure the component is responsive with horizontal scroll on mobile (leverage existing `overflow-x-auto` in ContributionGraphCalendar)

- [ ] 5.0 Integrate Contribution Graph into Workouts Page
  - [ ] 5.1 Import `WorkoutContributionGraph` in `client/src/routes/_layout/workouts/index.tsx`
  - [ ] 5.2 Add the component above the `WorkoutList` component in the page layout
  - [ ] 5.3 Add `contributionDataQueryOptions()` to the route loader to prefetch contribution data
  - [ ] 5.4 Implement responsive default state: collapsible expanded by default on desktop, collapsed by default on mobile (use media query or viewport width check)
  - [ ] 5.5 Test the full integration: verify data loads, accordion works, tooltips display, navigation works for single and multiple workout days, and empty state displays correctly for new users
