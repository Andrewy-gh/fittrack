# Weekly Workout Streak Implementation Plan

**Project:** FitTrack - Weekly Workout Streak Feature
**Created:** 2026-01-08
**Status:** 🔨 Planning - Ready for Implementation

---

## Overview

This plan addresses the bug in the "This Week" metric calculation that shows 8 workouts instead of a maximum of 7, and implements a more engaging "Weekly Workout Streak" feature based on fitness app engagement research.

**Key Goals:**

1. Fix the off-by-one error in "This Week" calculation (8-day window instead of 7)
2. Replace "This Week" metric with "Weekly Workout Streak" for better user engagement
3. Implement weekly streak calculation logic that counts consecutive weeks meeting a workout threshold
4. Use research-backed approach: weekly goals show 20% better 90-day retention vs daily streaks
5. Reduce user pressure while maintaining motivation (weekly > daily for long-term retention)

**Current Bug:**
- **File:** `client/src/components/workouts/workout-summary-cards.tsx:11-16`
- **Issue:** Creates 8-day window instead of 7 (includes both 7-days-ago AND today)
- **Impact:** Shows "This Week: 8" which is mathematically impossible

**Proposed Solution:**
- Replace "This Week" card with "Weekly Workout Streak"
- Shows consecutive weeks where user hit workout threshold (e.g., 3+ workouts per week)
- Can grow indefinitely (more motivating than max of 7)
- Aligns with research: weekly goals reduce pressure and improve retention

---

## Implementation Summary

**Status:** 0/6 tasks completed

- [ ] Task 1: Fix "This Week" Bug (Quick Fix Option)
- [ ] Task 2: Create Weekly Streak Calculation Utility
- [ ] Task 3: Update WorkoutSummaryCards Component
- [ ] Task 4: Add User Configuration for Streak Threshold (Optional)
- [ ] Task 5: Update Tests
- [ ] Task 6: Verify Streak Calculation Accuracy

---

## Task 1: Fix "This Week" Bug (Quick Fix Option)

**Priority:** High
**Status:** Pending
**Files Affected:**
- `client/src/components/workouts/workout-summary-cards.tsx`

**Description:** Quick fix for the 8-day bug. This provides a fallback option if weekly streaks take longer than expected.

### Subtasks

#### 1.1 Fix the Off-By-One Error

- [ ] Read `client/src/components/workouts/workout-summary-cards.tsx`
- [ ] Locate lines 11-16 with buggy calculation
- [ ] Replace with correct 7-day window calculation:
  ```typescript
  const sixDaysAgo = new Date();
  sixDaysAgo.setDate(sixDaysAgo.getDate() - 6);
  sixDaysAgo.setHours(0, 0, 0, 0);
  return workoutDate >= sixDaysAgo;
  ```
- [ ] Verify calculation includes today + last 6 days = 7 total days

#### 1.2 Test Fix Manually

- [ ] Test with workouts spanning 8 days
- [ ] Verify count shows maximum of 7
- [ ] Test edge cases (workouts at midnight, timezone boundaries)

---

## Task 2: Create Weekly Streak Calculation Utility

**Priority:** High
**Status:** Pending
**Files Affected:**
- `client/src/lib/workout-utils.ts` (new file)
- `client/src/lib/workout-utils.test.ts` (new file)

**Description:** Implement the core logic for calculating weekly workout streaks.

### Subtasks

#### 2.1 Design Streak Calculation Algorithm

- [ ] Define week boundaries (Sunday-Saturday or Monday-Sunday)
- [ ] Define workout threshold per week (default: 3 workouts)
- [ ] Handle incomplete current week (exclude from streak count)
- [ ] Sort workouts by date (oldest to newest)
- [ ] Group workouts by week
- [ ] Count consecutive weeks meeting threshold

**Algorithm Pseudocode:**
```
1. Sort workouts by date (ascending)
2. Group by week number
3. For each week (excluding current incomplete week):
   - Count workouts in that week
   - If count >= threshold: increment streak
   - If count < threshold: reset streak to 0
4. Return current streak length
```

#### 2.2 Create workout-utils.ts File

- [ ] Create `client/src/lib/workout-utils.ts`
- [ ] Add TypeScript type definitions:
  ```typescript
  export interface WorkoutStreakOptions {
    weekStartsOn?: 0 | 1; // 0 = Sunday, 1 = Monday
    weeklyThreshold?: number;
    includeCurrentWeek?: boolean;
  }
  ```

#### 2.3 Implement calculateWeeklyStreak Function

- [ ] Implement main function signature:
  ```typescript
  export function calculateWeeklyStreak(
    workouts: Array<{ date: string }>,
    options: WorkoutStreakOptions = {}
  ): number
  ```
- [ ] Parse workout dates from ISO 8601 strings
- [ ] Handle timezone conversions properly
- [ ] Group workouts by week using date-fns or native Date
- [ ] Calculate consecutive weeks meeting threshold
- [ ] Return streak count as number

#### 2.4 Implement Helper Functions

- [ ] `getWeekKey(date: Date, weekStartsOn: 0 | 1): string` - returns "2026-W02" format
- [ ] `getWeekStart(date: Date, weekStartsOn: 0 | 1): Date` - returns start of week
- [ ] `isCurrentWeek(date: Date, weekStartsOn: 0 | 1): boolean` - checks if date is in current week
- [ ] `groupWorkoutsByWeek(workouts, weekStartsOn): Map<string, number>` - groups and counts

#### 2.5 Handle Edge Cases

- [ ] Empty workouts array returns 0
- [ ] Single workout returns 0 or 1 (depending on threshold)
- [ ] Workouts in current incomplete week handled correctly
- [ ] Handle workouts with invalid dates gracefully
- [ ] Handle gaps in workout history (reset streak)
- [ ] Handle consecutive weeks correctly (no gaps = increment)

---

## Task 3: Update WorkoutSummaryCards Component

**Priority:** High
**Status:** Pending
**Files Affected:**
- `client/src/components/workouts/workout-summary-cards.tsx`

**Description:** Replace "This Week" card with "Weekly Workout Streak" card.

### Subtasks

#### 3.1 Read Current Implementation

- [ ] Read `client/src/components/workouts/workout-summary-cards.tsx`
- [ ] Understand current card structure and props
- [ ] Identify "This Week" card rendering logic

#### 3.2 Import New Utility

- [ ] Import `calculateWeeklyStreak` from `@/lib/workout-utils`
- [ ] Import or define `WorkoutStreakOptions` type

#### 3.3 Replace "This Week" Logic

- [ ] Remove lines 11-16 (buggy thisWeekWorkouts calculation)
- [ ] Add new calculation:
  ```typescript
  const weeklyStreak = calculateWeeklyStreak(workouts, {
    weekStartsOn: 0, // Sunday (or make configurable)
    weeklyThreshold: 3, // 3 workouts per week (or make configurable)
    includeCurrentWeek: false
  });
  ```

#### 3.4 Update Card UI

- [ ] Change card title from "This Week" to "Weekly Workout Streak"
- [ ] Change icon from calendar to fire/streak icon (if available)
- [ ] Update displayed value from `thisWeekWorkouts` to `weeklyStreak`
- [ ] Add unit label: "{weeklyStreak} weeks" or "{weeklyStreak} week streak"
- [ ] Consider adding tooltip explaining streak rules

#### 3.5 Handle Streak Display

- [ ] Show "0 weeks" when streak is 0
- [ ] Show "1 week" (singular) when streak is 1
- [ ] Show "N weeks" (plural) when streak > 1
- [ ] Consider adding congratulatory message for long streaks (optional)

---

## Task 4: Add User Configuration for Streak Threshold (Optional)

**Priority:** Low
**Status:** Pending
**Files Affected:**
- `client/src/components/workouts/workout-summary-cards.tsx`
- User preferences store/context (if exists)
- Backend user preferences endpoint (if implementing server-side storage)

**Description:** Allow users to configure their weekly workout goal threshold.

### Subtasks

#### 4.1 Decide Configuration Approach

- [ ] **Option A:** Hardcode threshold to 3 workouts/week (simplest)
- [ ] **Option B:** Store in localStorage (client-side only)
- [ ] **Option C:** Add to user profile in database (requires backend changes)
- [ ] Document decision in plan

#### 4.2 If Option A (Hardcoded) - RECOMMENDED for MVP

- [ ] Use fixed threshold of 3 workouts per week
- [ ] Document in code comment
- [ ] Skip remaining subtasks

#### 4.3 If Option B (localStorage)

- [ ] Create settings UI component for threshold selection
- [ ] Store preference in localStorage
- [ ] Read preference in WorkoutSummaryCards
- [ ] Provide sensible default (3 workouts)
- [ ] Validate range (2-7 workouts per week)

#### 4.4 If Option C (Database)

- [ ] Add `weekly_goal` field to user profile schema
- [ ] Create migration to add column
- [ ] Update user model and API types
- [ ] Create settings API endpoint (GET/PUT)
- [ ] Update frontend to fetch and update preference
- [ ] Handle users without configured preference (default to 3)

---

## Task 5: Update Tests

**Priority:** High
**Status:** Pending
**Files Affected:**
- `client/src/lib/workout-utils.test.ts` (new file)
- `client/src/components/workouts/workout-summary-cards.test.tsx` (if exists)

**Description:** Comprehensive test coverage for streak calculation logic.

### Subtasks

#### 5.1 Create workout-utils.test.ts

- [ ] Create test file `client/src/lib/workout-utils.test.ts`
- [ ] Set up test framework (Vitest/Jest)
- [ ] Import functions to test

#### 5.2 Test calculateWeeklyStreak - Basic Cases

- [ ] Test empty array returns 0
- [ ] Test single workout returns 0 (below threshold)
- [ ] Test 3 workouts in one week returns 0 (current week excluded)
- [ ] Test 3 workouts in previous week returns 1
- [ ] Test two consecutive weeks with 3+ workouts returns 2

#### 5.3 Test calculateWeeklyStreak - Edge Cases

- [ ] Test gap in weeks resets streak to 0
- [ ] Test current incomplete week is excluded from streak
- [ ] Test workouts at week boundaries (Sunday/Monday edge)
- [ ] Test different weekStartsOn values (0 vs 1)
- [ ] Test different threshold values (2, 3, 4, 5)
- [ ] Test invalid dates are handled gracefully
- [ ] Test workouts spanning multiple months/years

#### 5.4 Test calculateWeeklyStreak - Real World Scenarios

- [ ] Test workout pattern from screenshot: 110 total, 8 this week
- [ ] Test user with consistent 3x/week schedule (4+ week streak)
- [ ] Test user with inconsistent schedule (0 streak despite many workouts)
- [ ] Test new user with only current week workouts (0 streak)

#### 5.5 Test Helper Functions

- [ ] Test `getWeekKey` returns correct format
- [ ] Test `getWeekStart` returns correct Sunday/Monday
- [ ] Test `isCurrentWeek` correctly identifies current week
- [ ] Test `groupWorkoutsByWeek` groups correctly

#### 5.6 Update Component Tests

- [ ] Update WorkoutSummaryCards tests if they exist
- [ ] Mock calculateWeeklyStreak function
- [ ] Test streak value displays correctly
- [ ] Test pluralization (1 week vs N weeks)
- [ ] Test with zero streak
- [ ] Test with large streak (52+ weeks)

#### 5.7 Test Coverage Goals

- [ ] Aim for >90% coverage on workout-utils.ts
- [ ] Test all code paths
- [ ] Test all error conditions
- [ ] Run coverage report: `npm run test:coverage`

---

## Task 6: Verify Streak Calculation Accuracy

**Priority:** Medium
**Status:** Pending
**Files Affected:**
- Manual testing checklist

**Description:** Manually verify streak calculations with real data and edge cases.

### Subtasks

#### 6.1 Test with Production-Like Data

- [ ] Create test account with 110+ workouts (like screenshot)
- [ ] Manually calculate expected streak
- [ ] Verify UI shows correct streak
- [ ] Test with current week having 8 workouts (like bug report)

#### 6.2 Test Edge Cases in Browser

- [ ] Test at week boundary (Saturday 11:59 PM → Sunday 12:01 AM)
- [ ] Test with timezone changes (if applicable)
- [ ] Test with exactly threshold workouts (e.g., exactly 3)
- [ ] Test with threshold - 1 workouts (e.g., 2 when threshold is 3)
- [ ] Test with no workouts (empty state)

#### 6.3 Compare with Activity Graph

- [ ] Verify streak aligns with activity graph heatmap
- [ ] Check that weeks with bright squares count toward streak
- [ ] Check that weeks with dim/no squares reset streak

#### 6.4 User Acceptance Testing

- [ ] Show to 2-3 users for feedback
- [ ] Verify streak calculation makes intuitive sense
- [ ] Gather feedback on threshold (is 3 workouts/week reasonable?)
- [ ] Adjust if needed based on feedback

---

## Implementation Order

### Phase 1: Core Implementation (Do First)

1. **Task 2**: Create Weekly Streak Calculation Utility
   - Foundation for everything else
   - Can be developed and tested independently
   - ~1-2 hours

2. **Task 5.1-5.5**: Write Tests for Utility Functions
   - TDD approach: write tests as you implement
   - Ensures correctness before UI integration
   - ~30-60 minutes

### Phase 2: UI Integration

3. **Task 3**: Update WorkoutSummaryCards Component
   - Integrate utility into UI
   - Replace buggy "This Week" card
   - ~30 minutes

4. **Task 5.6**: Update Component Tests
   - Ensure component integration works
   - ~15-30 minutes

### Phase 3: Validation

5. **Task 6**: Verify Streak Calculation Accuracy
   - Manual testing with real data
   - Edge case validation
   - ~30-45 minutes

### Phase 4: Optional Enhancement

6. **Task 4**: Add User Configuration (Optional)
   - Only if time permits and user demand exists
   - Start with Option A (hardcoded) for MVP
   - ~0-2 hours depending on approach

### Fallback: Quick Fix Only

If timeline is constrained:
- **Task 1**: Fix "This Week" Bug Only
  - 5-minute fix
  - Doesn't add new features but fixes the bug
  - Can implement streaks later as enhancement

---

## Testing Strategy

### Unit Tests (Vitest/Jest)

- [ ] Test all pure functions in workout-utils.ts
- [ ] Mock workout data with various patterns
- [ ] Test boundary conditions (week edges, timezone)
- [ ] Test error handling (invalid dates, empty arrays)
- [ ] Aim for >90% code coverage

### Integration Tests

- [ ] Test WorkoutSummaryCards with mocked workout data
- [ ] Test calculation with real workout response format
- [ ] Verify UI updates when workouts change

### Manual Testing

- [ ] Test with real user account
- [ ] Test at week boundaries
- [ ] Test with different workout patterns
- [ ] Verify against activity graph visualization
- [ ] Cross-browser testing (Chrome, Firefox, Safari)

### Performance Testing

- [ ] Test with 1000+ workouts (heavy user)
- [ ] Verify calculation is fast (<50ms)
- [ ] Check for memory leaks in streak calculation
- [ ] Ensure no unnecessary re-renders

---

## Success Criteria

- [ ] "This Week" bug fixed (no longer shows 8)
- [ ] Weekly streak calculation returns correct values for all test cases
- [ ] UI displays streak with correct pluralization
- [ ] Tests achieve >90% coverage on new code
- [ ] Manual testing confirms accuracy with real data
- [ ] Streak aligns with activity graph visualization
- [ ] Code review approved
- [ ] No performance regressions
- [ ] Documentation updated with streak calculation rules

---

## Rollback Plan

If issues are discovered:

1. **Immediate**: Revert to showing "Total Workouts" only (remove second card)
2. **Short-term**: Revert to Task 1 quick fix (corrected "This Week" display)
3. **Investigation**: Debug streak calculation in dev environment
4. **Fix Forward**: Patch calculation bug and redeploy

---

## Research & Decisions

### Why Weekly Streaks Over Daily Streaks?

**Research Findings:**
1. **Down Dog app**: 20% better 90-day retention with weekly goals vs daily streaks
2. **Psychology**: Daily streaks create pressure and discourage necessary rest days
3. **Loss Aversion**: 58% of users continue after broken daily streak vs 66% with intact
4. **Flexibility**: Weekly goals allow life to happen without punishment

**Decision:** Implement weekly streak for better long-term user retention

### Week Start Day: Sunday vs Monday?

**Options:**
- **Sunday (0)**: Standard in US, aligns with calendar weeks
- **Monday (1)**: ISO 8601 standard, common in fitness tracking

**Decision:** Use **Sunday (0)** for MVP to align with US conventions
- Can make configurable later if international users request it
- Matches most US fitness apps

### Weekly Workout Threshold?

**Options:**
- 2 workouts/week: Too easy, less motivating
- 3 workouts/week: Evidence-based (Ladder app uses this)
- 4 workouts/week: Moderate challenge
- 5+ workouts/week: Too strict, discourages beginners

**Decision:** Use **3 workouts/week** as default
- Supported by research (Ladder app success)
- Reasonable for beginners and advanced users
- Can be made configurable in future

### Include Current Incomplete Week?

**Options:**
- **Include**: Shows progress toward this week's goal
- **Exclude**: Only count complete weeks

**Decision:** **Exclude current week** from streak count
- More accurate (week isn't over yet)
- Prevents confusion (streak jumping mid-week)
- Can show "X/3 this week" separately if desired

---

## User-Facing Changes

### Before (Buggy)
```
┌─────────────────┐  ┌─────────────────┐
│ Total Workouts  │  │   This Week     │
│                 │  │                 │
│      110        │  │       8         │ ← BUG!
└─────────────────┘  └─────────────────┘
```

### After (Fixed)
```
┌─────────────────┐  ┌──────────────────────┐
│ Total Workouts  │  │ Weekly Workout Streak│
│                 │  │                      │
│      110        │  │      4 weeks         │ ← New!
└─────────────────┘  └──────────────────────┘
```

### Streak Calculation Example

**User Workout History:**
```
Week of Dec 15-21: 4 workouts ✓ (meets threshold)
Week of Dec 22-28: 3 workouts ✓ (meets threshold)
Week of Dec 29-Jan 4: 5 workouts ✓ (meets threshold)
Week of Jan 5-11: 2 workouts ✗ (below threshold)
Week of Jan 12-18: 4 workouts ✓ (meets threshold)
Week of Jan 19-25: 8 workouts ← Current week (excluded)
```

**Streak = 1 week** (only week of Jan 12-18, since Jan 5-11 broke the streak)

---

## Open Questions (To Be Resolved)

### Design Questions

1. **What icon should we use for weekly streak?**
   - 🔥 Fire emoji (classic streak icon)
   - 📊 Chart/progress icon
   - 🏆 Trophy icon
   - 📅 Calendar check icon
   - **Recommend:** Fire icon (🔥) - universally recognized for streaks

2. **Should we show "X/3 this week" alongside streak?**
   - **Option A:** Only show streak (simpler UI)
   - **Option B:** Show both streak and current week progress
   - **Recommend:** Option A for MVP (keep UI clean)

3. **Should we show streak even if it's 0?**
   - **Yes:** Shows the feature exists, motivates user to start
   - **No:** Only show when streak > 0
   - **Recommend:** Yes, always show (even "0 weeks")

### Technical Questions

4. **Should we use a date library (date-fns) or native Date?**
   - **date-fns:** More reliable, handles edge cases better
   - **Native Date:** No dependencies, smaller bundle
   - **Recommend:** Check if date-fns is already in project. If yes, use it. If no, use native Date to avoid new dependency.

5. **Where should week start configuration live?**
   - Hardcoded constant
   - Environment variable
   - User preference (database)
   - **Recommend:** Hardcoded for MVP (Sunday = 0)

6. **Should backend calculate streak or keep client-side?**
   - **Client-side:** Simpler, no backend changes needed
   - **Server-side:** More consistent, can cache results
   - **Recommend:** Client-side for MVP (all data already loaded)

### User Experience Questions

7. **How do we explain streak rules to users?**
   - Tooltip on hover
   - Help icon with modal
   - Settings page documentation
   - **Recommend:** Tooltip on hover with brief explanation

8. **What happens if user changes threshold later?**
   - Recalculate streak with new threshold
   - Keep historical streak, apply new threshold going forward
   - **Recommend:** Recalculate (if we add configuration)

---

## Timeline Estimate

**Total Estimated Time:** 3-5 hours

### Breakdown:
- Task 2 (Streak Utility): 1-2 hours
- Task 5 (Tests): 1-1.5 hours
- Task 3 (UI Update): 30 minutes
- Task 6 (Manual Verification): 30-45 minutes
- Buffer for debugging: 30-45 minutes

**Fastest Path (Bug Fix Only):**
- Task 1: 5-10 minutes (just fix the 8-day bug)

---

## Post-Implementation

After completion:

- [ ] Monitor user feedback on streak feature
- [ ] Track engagement metrics (do streaks improve retention?)
- [ ] Consider adding streak "freeze" feature (like Duolingo)
- [ ] Consider adding weekly goal configuration
- [ ] A/B test: streak vs "This Week" to validate decision
- [ ] Update user documentation/help center
- [ ] Share implementation learnings with team

---

## Related Files Reference

### Frontend
- `client/src/components/workouts/workout-summary-cards.tsx` - Main component with bug
- `client/src/routes/_layout/workouts/index.tsx` - Workouts page
- `client/src/lib/api/workouts.ts` - Workout API hooks
- `client/src/client/types.gen.ts` - TypeScript types

### Backend (No changes needed)
- `server/internal/database/models.go` - Workout model
- `server/internal/workout/` - Workout service/handlers

### New Files to Create
- `client/src/lib/workout-utils.ts` - Streak calculation logic
- `client/src/lib/workout-utils.test.ts` - Unit tests

---

## References

### Research Sources
- [Down Dog's 20% retention increase](https://www.lucid.now/blog/retention-metrics-for-fitness-apps-industry-insights/)
- [Strategies to Increase Fitness App Engagement](https://orangesoft.co/blog/strategies-to-increase-fitness-app-engagement-and-retention)
- [Psychology of Gamification Streaks](https://www.plotline.so/blog/streaks-for-gamification-in-mobile-apps)
- [How Broken Streaks Sap Motivation](https://www.psychologytoday.com/ca/blog/ulterior-motives/202306/how-broken-streaks-sap-motivation)
- [Designing Streaks for Long-term User Growth](https://www.mindtheproduct.com/designing-streaks-for-long-term-user-growth/)

### Code Examples
- Ladder app: Weekly workout count streaks (3 workouts/week)
- Gentler Streak: Flexible daily goals including rest
- Duolingo: Daily streaks with freeze/emergency reserves

---

**End of Implementation Plan**
