# PR #71 - Contribution Graph Implementation Tracker

**Date Started:** 2025-12-23
**Status:** Not Started
**Related Files:**
- Handoff Document: `PR-71-HANDOFF.md`
- Implementation Plan: `~/.claude/plans/deep-marinating-treehouse.md`

---

## Progress Overview

- [x] Task 1: Fix Frontend Performance - API Changes (6/6 subtasks)
- [x] Task 2: Verify RLS and Fix Data Leakage (3/3 subtasks)
- [x] Task 3: Fix Test Cleanup Pattern (2/2 subtasks)
- [x] Task 4: Change Threshold 10 → 30 Days (2/2 subtasks)
- [x] Task 5: Add Error Handling (3/3 subtasks)
- [ ] Task 6: Add Edge Case Tests (0/3 subtasks)

**Total Progress:** 16/19 subtasks completed

---

## Task 1: Fix Frontend Performance - Include Workout Metadata in API

**Priority:** MUST-FIX #2 (Highest Impact)
**Status:** Completed
**Files:** `server/query.sql`, `server/internal/workout/models.go`, `server/internal/workout/service.go`, `server/internal/workout/handler_test.go`, `client/src/client/types.gen.ts`, `client/src/components/workouts/workout-contribution-graph.tsx`

### Subtasks

- [x] **1.1: Modify SQL query to include workout metadata**
  - File: `server/query.sql` (lines 166-176)
  - Change: Replace `ARRAY_AGG(DISTINCT w.id)` with `JSON_AGG(DISTINCT JSONB_BUILD_OBJECT(...))`
  - Include: workout id, time, and focus in JSON object
  - Status: Completed

- [x] **1.2: Update backend models**
  - File: `server/internal/workout/models.go` (lines 221-231)
  - Add: `WorkoutSummary` struct with ID, Time, Focus fields
  - Update: `ContributionDay` to use `Workouts []WorkoutSummary` instead of `WorkoutIDs []int32`
  - Status: Completed

- [x] **1.3: Update service conversion logic**
  - File: `server/internal/workout/service.go` (lines 213-239)
  - Replace: `extractWorkoutIDs` function with new `parseWorkouts` function
  - Parse: JSON bytes into `[]WorkoutSummary` with proper error handling
  - Handle: Null focus values correctly
  - Status: Completed

- [x] **1.4: Update handler tests**
  - File: `server/internal/workout/handler_test.go` (lines 662-761)
  - Update: Mock responses to use `Workouts` instead of `WorkoutIDs`
  - Add: Test cases for workouts with/without focus
  - Verify: Time format and focus field in assertions
  - Status: Completed

- [x] **1.5: Run sqlc generate and verify**
  - Command: `cd server && sqlc generate`
  - Verify: Generated types in `server/internal/database/query.sql.go`
  - Check: `GetContributionDataRow` has `Workouts []byte` field
  - Status: Completed

- [x] **1.6: Update frontend component**
  - File: `client/src/components/workouts/workout-contribution-graph.tsx`
  - Remove: Line 46 - `useSuspenseQuery(workoutsQueryOptions())`
  - Update: `workoutDetailsById` map to extract from `data.days.workouts`
  - Update: `workoutIdsByDate` map to extract IDs from workouts array
  - Remove: `workoutsQueryOptions` import if unused
  - Status: Completed

---

## Task 2: Verify RLS Context and Fix Data Leakage Risk

**Priority:** MUST-FIX #1 (Security)
**Status:** Completed
**Files:** `server/query.sql`, `server/internal/workout/repository.go`, `server/internal/workout/handler_test.go`

### Subtasks

- [x] **2.1: Review RLS implementation pattern**
  - Files: `server/query.sql` (line 169), `server/internal/workout/repository.go` (lines 555-583)
  - Verify: Pattern is consistent with other endpoints
  - Check: RLS policies exist in migrations
  - Confirm: userID parameter always from authenticated context
  - Status: Completed

- [x] **2.2: Add security comments in SQL**
  - File: `server/query.sql`
  - Add: Comment explaining RLS security model
  - Document: Why application-level filtering is sufficient
  - Verify: GROUP BY + WHERE prevents cross-user data leakage
  - Status: Completed

- [x] **2.3: Add RLS integration test**
  - File: `server/internal/workout/handler_test.go`
  - Create: `TestContributionData_Integration_RLS` function
  - Test: Two users with workouts on same day don't mix
  - Verify: User A only sees their own workout metadata
  - Use: `setupTestUserContext` from testutils
  - Status: Completed

---

## Task 3: Fix Test Cleanup Pattern (Remove SQL Injection Risk)

**Priority:** MUST-FIX #3 (Code Quality)
**Status:** Completed
**Files:** `server/internal/workout/handler_test.go`

### Subtasks

- [x] **3.1: Refactor cleanupTestData function**
  - File: `server/internal/workout/handler_test.go` (lines 1444-1493)
  - Replace: LIKE pattern `test-user-%` with exact user ID list
  - Use: Iteration over known test user IDs
  - Add: Logging for cleanup failures (don't fail tests)
  - Pattern: Follow `cleanupSpecificTestUsers` from `delete_integration_test.go`
  - Status: Completed

- [x] **3.2: (Optional) Update test setup to track users**
  - File: `server/internal/workout/handler_test.go` (lines 32-62 registry, 1404-1424 setup, 1478-1531 cleanup)
  - Created: Test user registry with thread-safe functions
  - Updated: `setupTestUsers` to register created users (including RLS test users)
  - Updated: `cleanupTestData` to use registry instead of hardcoded list
  - Status: Completed

---

## Task 4: Change Dynamic Threshold from 10 to 30 Days

**Priority:** SHOULD-FIX #4 (UX Improvement)
**Status:** Completed
**Files:** `server/internal/workout/service.go`, `server/internal/workout/handler_test.go`

### Subtasks

- [x] **4.1: Update threshold check**
  - File: `server/internal/workout/service.go` (line 254)
  - Change: `if len(nonZeroCounts) < 10` to `if len(nonZeroCounts) < 30`
  - Update: Code comment to reflect new threshold
  - Status: Completed

- [x] **4.2: Update threshold tests**
  - File: `server/internal/workout/handler_test.go` (lines 822-867)
  - Update: Test names and descriptions for 30-day threshold
  - Add: Test case for exactly 29 days (should use static)
  - Add: Test case for exactly 30 days (should use dynamic)
  - Add: Test case for 10 workout days (verifies old threshold now uses static)
  - Verify: All tests passing with new 30-day threshold
  - Status: Completed

---

## Task 5: Add Error Handling to Contribution Graph

**Priority:** SHOULD-FIX #5 (UX/Resilience)
**Status:** Completed
**Files:** `client/src/routes/_layout/workouts/index.tsx`, `client/src/components/workouts/contribution-graph-error.tsx`

**Implementation Approach:** Query-level error handling instead of ErrorBoundary (better fit with TanStack Router patterns)

### Subtasks

- [x] **5.1: Implement query-level error handling**
  - File: `client/src/routes/_layout/workouts/index.tsx`
  - Changed: `useSuspenseQuery` → `useQuery` for contribution data
  - Added: Explicit `isError` and `isSuccess` state handling
  - Render: `<ContributionGraphError />` when query fails
  - Status: Completed
  - Note: Initial ErrorBoundary approach was committed but replaced with this approach

- [x] **5.2: Create error fallback component**
  - File: `client/src/components/workouts/contribution-graph-error.tsx` (new file)
  - Created: Error component matching design system
  - Includes: Card with AlertCircle icon and user-friendly message
  - Status: Completed

- [x] **5.3: Handle loading and error states**
  - File: `client/src/routes/_layout/workouts/index.tsx`
  - Configured: Query enabled only for authenticated users
  - Handled: Loading (no display), Error (ContributionGraphError), Success (graph)
  - Status: Completed

---

## Task 6: Add Edge Case Tests for Level Calculation

**Priority:** SHOULD-FIX #6 (Test Coverage)
**Status:** Not Started
**Files:** `server/internal/workout/handler_test.go`

### Subtasks

- [ ] **6.1: Add boundary condition tests**
  - File: `server/internal/workout/handler_test.go`
  - Add: Test for exactly 29 days (should use static thresholds)
  - Add: Test for exactly 30 days (should use dynamic thresholds)
  - Add: Test for 31 days (confirm dynamic continues)
  - Location: Add to `TestWorkoutService_CalculateLevelThresholds`
  - Status: Not Started

- [ ] **6.2: Add identical counts edge case tests**
  - File: `server/internal/workout/handler_test.go`
  - Add: Test all identical counts (e.g., fifteen 5's)
  - Add: Test mostly identical with one outlier
  - Verify: Thresholds remain strictly increasing
  - Verify: No overflow or invalid thresholds
  - Location: Add to `TestWorkoutService_CalculateLevelThresholds`
  - Status: Not Started

- [ ] **6.3: Add integration test for GetContributionData**
  - File: `server/internal/workout/handler_test.go`
  - Create: `TestContributionData_Integration` function
  - Test: Actual SQL query execution with real database
  - Verify: 52-week date range filtering
  - Verify: Multiple workouts per day aggregation
  - Verify: Level calculation with real data
  - Verify: Workout metadata correctly aggregated (after Task 1)
  - Status: Not Started

---

## Testing Checklist

After all tasks complete:

### Backend Tests
- [ ] Run `go test ./server/internal/workout/...`
- [ ] All existing tests pass
- [ ] New RLS integration test passes
- [ ] New edge case tests pass
- [ ] New contribution data integration test passes

### Frontend Tests
- [ ] Frontend types regenerated correctly
- [ ] No TypeScript compilation errors
- [ ] Component imports resolve correctly

### Manual Testing
- [ ] Contribution graph displays without fetching all workouts
- [ ] Network tab shows only contribution data API call, not workouts list
- [ ] Popover shows workout time in localized format
- [ ] Popover shows workout focus when present
- [ ] Multiple workouts per day display correctly in popover
- [ ] Days with no workouts don't crash
- [ ] Error boundary catches failures gracefully
- [ ] Loading states work correctly

### Integration Testing
- [ ] RLS prevents cross-user data leakage
- [ ] 52-week date range is correct
- [ ] Edge case: 29 days uses static thresholds
- [ ] Edge case: 30+ days uses dynamic thresholds
- [ ] Edge case: Identical counts don't break thresholds

---

## Notes & Decisions

### Implementation Order
1. Task 1 (backend changes first, then frontend)
2. Task 2 (RLS verification - can run in parallel with Task 1)
3. Task 3 (test cleanup - independent)
4. Task 4 (threshold change - after Task 1 to avoid conflicts)
5. Task 5 (error boundaries - after Task 1.6)
6. Task 6 (edge case tests - after Task 4 for new threshold)

### Key Dependencies
- Task 1.5 must complete before Task 1.6 (backend types → frontend types)
- Task 5 should wait for Task 1.6 (avoid conflicts in same component)
- Task 6 should wait for Task 4 (test new 30-day threshold)

### Rollback Points
If issues arise:
- After Task 1.5: Backend changes complete, can regenerate frontend types
- After Task 1.6: Full API contract change complete
- After Task 2.3: RLS verification complete
- After all tasks: Full implementation complete

---

## Handoff Information

**For Next Agent:**
- Read `PR-71-HANDOFF.md` for full context
- Check this file for current progress
- Uncompleted subtasks marked with `[ ]`
- Follow implementation order above
- Verify all tests pass before marking task complete
- Update this file as you complete subtasks

**Questions/Blockers:**
- None at this time

**Last Updated:** 2025-12-23 (Initial creation)
