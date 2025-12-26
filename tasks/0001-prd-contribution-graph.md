# PRD: Workout Contribution Graph

## 1. Introduction/Overview

The Workout Contribution Graph is a GitHub-style activity visualization that displays a user's workout consistency over time. Each square represents a day, with color intensity indicating workout activity level based on the number of working sets performed.

This feature helps users visualize their training consistency at a glance, motivating them to maintain workout streaks and identify patterns in their training habits.

## 2. Goals

1. Provide users with a visual representation of their workout consistency over the past 52 weeks
2. Motivate users to maintain regular training habits through visual feedback
3. Help users identify gaps or patterns in their training schedule
4. Integrate seamlessly with the existing app design using the primary (orange) color scheme

## 3. User Stories

1. **As a user**, I want to see my workout activity over the past year so that I can understand my training consistency.

2. **As a user**, I want the activity levels to reflect my personal workout habits so that the visualization is meaningful to my training style (dynamic thresholds).

3. **As a user**, I want to hover over a day to see details so that I can quickly check what I did on a specific date.

4. **As a user**, I want to click on a day to navigate to that workout so that I can review or edit the workout details.

5. **As a user**, I want the graph to work on mobile devices so that I can check my progress on any device.

6. **As a new user with no workouts**, I want to see an encouraging message so that I'm motivated to log my first workout.

## 4. Functional Requirements

### 4.1 Data & Backend

1. The system must create a new API endpoint `GET /api/workouts/contribution-data` that returns daily workout aggregations for the past 52 weeks.

2. The endpoint must return for each day with activity:
   - `date` (ISO date string, e.g., "2025-01-15")
   - `count` (number of working sets performed that day)
   - `level` (0-4, calculated using dynamic thresholds)
   - `workout_ids` (array of workout IDs for that day, for navigation)

3. The system must count only sets with `set_type = 'working'` (exclude warmup sets).

4. The system must calculate dynamic level thresholds based on the user's personal data:
   - Level 0: No workout
   - Level 1: Below 25th percentile of user's daily set counts
   - Level 2: 25th-50th percentile
   - Level 3: 50th-75th percentile
   - Level 4: Above 75th percentile

5. The system must fall back to static thresholds when user has fewer than 10 workout days:
   - Level 0: 0 sets
   - Level 1: 1-5 sets
   - Level 2: 6-10 sets
   - Level 3: 11-15 sets
   - Level 4: 16+ sets

6. The endpoint must be protected by authentication and return only the authenticated user's data.

### 4.2 Frontend Display

7. The contribution graph must be displayed above the workout list on the `/workouts` page.

8. The graph must be wrapped in a collapsible accordion component.

9. The graph must NOT render its DOM content when the accordion is collapsed (lazy rendering for performance).

10. The accordion must be expanded by default on desktop and collapsed by default on mobile (to prioritize workout list visibility on small screens).

11. The graph must display the last 52 weeks of data, with the most recent week on the right.

12. The graph must be responsive and horizontally scrollable on mobile devices.

13. The graph must use the app's primary color (orange) with opacity levels for activity intensity:
    - Level 0: `muted` (background color)
    - Level 1: `primary/20` (20% opacity)
    - Level 2: `primary/40` (40% opacity)
    - Level 3: `primary/60` (60% opacity)
    - Level 4: `primary/80` (80% opacity)

14. The graph must display month labels above the calendar grid.

15. The graph must include a Less/More legend footer to help users understand the visualization.

### 4.3 Interactivity

16. The system must display a tooltip on hover showing:
    - The date (formatted, e.g., "Monday, Jan 15, 2025")
    - Number of working sets (e.g., "12 working sets")

17. The system must allow clicking on a day with workout(s) to navigate to that workout.

18. When a day has a single workout, clicking must navigate directly to `/workouts/{id}`.

19. When a day has multiple workouts, clicking must display a popover with a list of workouts showing:
    - Time of workout (if available, otherwise order)
    - Workout focus (if set)
    - User clicks an option to navigate to that specific workout

### 4.4 Empty State

20. When the user has no workouts in the 52-week period, the system must display an encouraging message instead of an empty graph (e.g., "Start your fitness journey! Log your first workout to see your progress here.").

### 4.5 Caching

21. Contribution data must be cached for the session duration.

22. The cache must be invalidated when a workout is created, updated, or deleted.

### 4.6 Demo Mode (Optional - Lower Priority)

23. The contribution graph should work in demo mode using localStorage demo data if implementation complexity is low. Otherwise, this may be deferred to a future PR.

## 5. Non-Goals (Out of Scope)

1. **Volume-based metrics**: The graph will not use total volume (weight × reps) as the metric due to high variance between workout types.

2. **Exercise-specific graphs**: No breakdown by exercise or muscle group.

3. **Historical data beyond 52 weeks**: The graph will only show the past year.

4. **Customizable date ranges**: Users cannot change the time period displayed.

5. **Export functionality**: No ability to export or share the graph.

6. **Streak notifications**: No push notifications or badges for streaks.

7. **Comparison features**: No comparison with other users or previous periods.

## 6. Design Considerations

### 6.1 Component Structure

The existing `ContributionGraph` component from Kibo UI (`client/src/components/kibo-ui/contribution-graph/index.tsx`) should be used with customizations:

- `ContributionGraph` - Root provider component
- `ContributionGraphCalendar` - The main grid with month labels
- `ContributionGraphBlock` - Individual day squares (customize colors)
- `ContributionGraphFooter` - Optional footer with legend
- `ContributionGraphLegend` - Less/More indicator

### 6.2 Color Customization

Override the default gray colors with primary color using Tailwind classes:

```tsx
// Replace muted-foreground with primary color
'data-[level="0"]:fill-muted'
'data-[level="1"]:fill-primary/20'
'data-[level="2"]:fill-primary/40'
'data-[level="3"]:fill-primary/60'
'data-[level="4"]:fill-primary/80'
```

### 6.3 Accordion Implementation

Use shadcn/ui `Accordion` or `Collapsible` component with conditional rendering:

```tsx
{isOpen && <ContributionGraph ... />}
```

### 6.4 Mobile Responsiveness

The `ContributionGraphCalendar` component already supports horizontal scrolling via `overflow-x-auto`. Ensure the parent container allows this.

## 7. Technical Considerations

### 7.1 Backend (Go/sqlc)

1. **New SQL Query** in `server/query.sql`:
   ```sql
   -- name: GetContributionData :many
   SELECT
     DATE(w.date) as date,
     COUNT(s.id) as set_count,
     ARRAY_AGG(DISTINCT w.id) as workout_ids
   FROM workout w
   LEFT JOIN "set" s ON s.workout_id = w.id AND s.set_type = 'working'
   WHERE w.user_id = $1
     AND w.date >= NOW() - INTERVAL '52 weeks'
   GROUP BY DATE(w.date)
   ORDER BY date;
   ```

2. **New Handler** in `server/internal/workout/handler.go`:
   - Calculate percentiles for dynamic thresholds
   - Map set counts to levels
   - Return structured response

3. **New Route** in `server/cmd/api/routes.go`:
   - `GET /api/workouts/contribution-data`

### 7.2 Frontend (React/TanStack)

1. **New Query Options** in `client/src/lib/api/workouts.ts`:
   - `contributionDataQueryOptions()`

2. **New Component** in `client/src/components/workouts/`:
   - `workout-contribution-graph.tsx` - Wrapper with accordion, data fetching, click handling

3. **Integration** in `client/src/routes/_layout/workouts/index.tsx`:
   - Add component above `WorkoutList`
   - Ensure data is prefetched in loader

### 7.3 Dependencies

- Existing: `date-fns` (already used by contribution-graph component)
- Existing: shadcn/ui `Popover`, `Accordion`/`Collapsible` components
- Existing: TanStack Query for data fetching

### 7.4 Performance Considerations

- Lazy render accordion content to avoid unnecessary DOM nodes
- Use `useSuspenseQuery` with prefetching in route loader
- Backend query uses existing index on `(user_id, date)`

## 8. Success Metrics

1. **Adoption**: >50% of active users view the contribution graph within the first week of release.

2. **Engagement**: Users who interact with the graph (hover/click) show higher workout logging frequency.

3. **Performance**: Graph loads within 200ms on desktop, 500ms on mobile (excluding network latency).

4. **Accessibility**: Component passes basic accessibility checks (keyboard navigation, screen reader support).

## 9. Resolved Decisions

The following questions were resolved during PRD review:

1. **Accordion default state**: Expanded on desktop, collapsed on mobile (to prioritize workout list visibility on small screens).

2. **Legend inclusion**: Include the Less/More legend footer, as it helps new users understand the visualization.

3. **Caching strategy**: Invalidate on workout create/update/delete, otherwise cache for session duration.

4. **Demo mode priority**: Include if not complex, otherwise defer to a future PR.
