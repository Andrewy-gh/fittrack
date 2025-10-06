# Unified Route Implementation Plan: Demo + Auth Support

**Created**: 2025-10-06
**Updated**: 2025-10-06 (Post-Review)
**Status**: ✅ Approved - Ready for Implementation
**Scope**: `/workouts/new` and `/workouts/$workoutId/edit`

---

## Executive Summary

Implement a unified pattern for both workout creation and editing routes that seamlessly supports:
- **Authenticated users**: API-based persistence with draft localStorage
- **Demo users**: localStorage-based persistence with draft localStorage

Both routes will follow the same architectural pattern already established in view-only routes (`/workouts`, `/exercises`).

---

## Current State

### `/workouts/new` (⚠️ BLOCKED)
- **Status**: TypeScript errors, partially implemented
- **Issues**:
  - Uses `user.id` without optional chaining (4 locations)
  - Uses API-only mutation: `useSaveWorkoutMutation()`
  - Loader only loads API query options
  - Component only uses API queries
- **Draft System**: Uses `@/lib/local-storage` (keyed by `user.id`)
- **Data System**: API-based

### `/workouts/$workoutId/edit` (🔒 AUTH-ONLY)
- **Status**: Working for authenticated users only
- **Issues**:
  - No demo support at all
  - Uses API-only mutation: `useUpdateWorkoutMutation()`
  - Loader only loads API query options
  - Component only uses API queries
- **Draft System**: None (form state only)
- **Data System**: API-based

---

## Design Principles

### 1. **Unified Pattern Across Routes**
Both routes should follow the same conditional logic pattern:
```tsx
const mutation = user ? apiMutation : demoMutation;
const queryOptions = user ? apiQueryOptions() : demoQueryOptions();
```

### 2. **Draft Persistence for Both User Types**
- **Auth users**: Draft keyed by `user.id` → `'workout-entry-form-data-{userId}'`
- **Demo users**: Draft keyed by base key → `'workout-entry-form-data'`
- Existing `@/lib/local-storage` already supports this via `getStorageKey(userId?)`

### 3. **Leverage Existing Infrastructure**
- ✅ Demo mutations already exist in `@/lib/demo-data/query-options.ts`
- ✅ Demo queries already exist
- ✅ localStorage draft system already handles `undefined` userId

---

## Implementation Plan

### Phase 0: Create Factory Functions (NEW - Recommended)

**File**: `client/src/lib/api/unified-query-options.ts` (NEW FILE)

Create the factory functions as shown in "Simpler Implementation Approach" section above. This centralizes all conditional logic in one place.

```tsx
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  exercisesQueryOptions,
  recentExerciseSetsQueryOptions
} from './exercises';
import {
  workoutsQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions
} from './workouts';
import {
  getDemoExercisesQueryOptions,
  getDemoExercisesByIdRecentSetsQueryOptions,
  getDemoWorkoutsQueryOptions,
  getDemoWorkoutsByIdQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from '@/lib/demo-data/query-options';

export function getExercisesQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? exercisesQueryOptions() : getDemoExercisesQueryOptions();
}

export function getRecentSetsQueryOptions(user: CurrentUser | CurrentInternalUser | null, exerciseId: number) {
  return user
    ? recentExerciseSetsQueryOptions(exerciseId)
    : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId);
}

export function getWorkoutsQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions();
}

export function getWorkoutByIdQueryOptions(user: CurrentUser | CurrentInternalUser | null, id: number) {
  return user ? workoutQueryOptions(id) : getDemoWorkoutsByIdQueryOptions(id);
}

export function getWorkoutsFocusQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsFocusValuesQueryOptions() : getDemoWorkoutsFocusValuesQueryOptions();
}
```

---

## Phase 1: Fix `/workouts/new` Route

### File: `client/src/routes/workouts/new.tsx`

#### 1.1 Add Imports
```tsx
// Add at top of file (around line 1-16)
import { useMutation } from '@tanstack/react-query';
import { postDemoWorkoutsMutation } from '@/lib/demo-data/query-options';
import { initializeDemoData } from '@/lib/demo-data/storage';
import {
  getExercisesQueryOptions,
  getWorkoutsFocusQueryOptions
} from '@/lib/api/unified-query-options';
```

#### 1.2 Update `WorkoutTracker` Component

**Lines 45, 49, 64, 96**: Change `user.id` → `user?.id`
```tsx
// Line 45
defaultValues: getInitialValues(user?.id),

// Line 49
saveToLocalStorage(formApi.state.values, user?.id);

// Line 64
clearLocalStorage(user?.id);

// Line 96
clearLocalStorage(user?.id);
```

**Line 42**: Add conditional mutation
```tsx
// Replace:
const saveWorkout = useSaveWorkoutMutation();

// With:
const saveWorkoutApi = useSaveWorkoutMutation();
const saveWorkoutDemo = useMutation(postDemoWorkoutsMutation());
const saveWorkout = user ? saveWorkoutApi : saveWorkoutDemo;
```

#### 1.3 Update Route Loader (Lines 315-319) - SIMPLIFIED

```tsx
// Replace:
loader: async ({ context }): Promise<void> => {
  context.queryClient.ensureQueryData(exercisesQueryOptions());
},

// With (using factory functions):
loader: async ({ context }): Promise<void> => {
  if (!context.user) initializeDemoData();
  await context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
  await context.queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(context.user));
},
```

#### 1.4 Update `RouteComponent` (Lines 321-349) - SIMPLIFIED

```tsx
// Replace:
function RouteComponent() {
  const { user } = Route.useRouteContext();
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({
      queries: [exercisesQueryOptions(), workoutsFocusValuesQueryOptions()],
    });

  // Convert API response to our cleaner DbExercise type
  const exercises: DbExercise[] = exercisesResponse.map((ex) => ({
    id: ex.id,
    name: ex.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <WorkoutTracker
      user={user}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}

// With (using factory functions):
function RouteComponent() {
  const { user } = Route.useRouteContext();

  const { data: exercisesResponse } = useSuspenseQuery(getExercisesQueryOptions(user));
  const { data: workoutsFocusValues } = useSuspenseQuery(getWorkoutsFocusQueryOptions(user));

  // Convert API response to our cleaner DbExercise type
  const exercises: DbExercise[] = exercisesResponse.map((ex) => ({
    id: ex.id,
    name: ex.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <WorkoutTracker
      user={user}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}
```

**Add import at top:**
```tsx
import { useSuspenseQuery } from '@tanstack/react-query';
```

---

### Phase 1.5: Fix RecentSets Component (CRITICAL)

**File**: `client/src/routes/workouts/-components/recent-sets-display.tsx`

#### Update RecentSets Component

```tsx
// Add import at top:
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import { getRecentSetsQueryOptions } from '@/lib/api/unified-query-options';

// Update RecentSetsDisplay interface (line 11):
interface RecentSetsDisplayProps {
  exerciseId: number;
  user: CurrentUser | CurrentInternalUser | null;  // ADD THIS
}

// Update function signature (line 15):
function RecentSetsDisplay({ exerciseId, user }: RecentSetsDisplayProps) {
  // Update query (line 16):
  const { data: recentSets } = useSuspenseQuery(
    getRecentSetsQueryOptions(user, exerciseId)
  );

  // Rest stays the same...
}

// Update RecentSets wrapper (line 99):
export function RecentSets({
  exerciseId,
  user
}: {
  exerciseId: number | null;
  user: CurrentUser | CurrentInternalUser | null;  // ADD THIS
}) {
  if (!exerciseId) {
    return null;
  }

  return (
    <Suspense
      fallback={
        <div className="space-y-4">
          <h2 className="font-semibold text-xl tracking-tight text-foreground mb-4">
            Recent Sets
          </h2>
          <div className="text-center py-8">
            <p className="text-muted-foreground text-sm">
              Loading recent sets...
            </p>
          </div>
        </div>
      }
    >
      <RecentSetsDisplay exerciseId={exerciseId} user={user} />
    </Suspense>
  );
}
```

**File**: `client/src/routes/workouts/new.tsx`

Update RecentSets usage (line 143):
```tsx
// Change:
recentSets={<RecentSets exerciseId={selectedExercise.exerciseId} />}

// To:
recentSets={<RecentSets exerciseId={selectedExercise.exerciseId} user={user} />}
```

---

## Phase 2: Add Demo Support to `/workouts/$workoutId/edit`

### File: `client/src/routes/workouts/$workoutId/edit.tsx`

#### 2.1 Add Imports
```tsx
// Add at top of file
import { useMutation } from '@tanstack/react-query';
import { putDemoWorkoutsByIdMutation } from '@/lib/demo-data/query-options';
import { initializeDemoData } from '@/lib/demo-data/storage';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  getExercisesQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsFocusQueryOptions
} from '@/lib/api/unified-query-options';
```

#### 2.2 Update `EditWorkoutForm` Component Signature

**Line 30-40**: Add `user` parameter
```tsx
// Add to props:
function EditWorkoutForm({
  user, // ADD THIS
  exercises,
  workout,
  workoutId,
  workoutsFocus,
}: {
  user: CurrentUser | CurrentInternalUser | null; // ADD THIS
  exercises: ExerciseExerciseResponse[];
  workout: WorkoutUpdateWorkoutRequest;
  workoutId: number;
  workoutsFocus: WorkoutFocus[];
}) {
```

#### 2.3 Update Mutation (Line 49)

```tsx
// Replace:
const updateWorkoutMutation = useUpdateWorkoutMutation();

// With:
const updateWorkoutApi = useUpdateWorkoutMutation();
const updateWorkoutDemo = useMutation(putDemoWorkoutsByIdMutation());
const updateWorkoutMutation = user ? updateWorkoutApi : updateWorkoutDemo;
```

#### 2.4 Update Route Loader (Lines 316-327) - SIMPLIFIED

```tsx
// Replace:
loader: async ({
  context,
  params,
}): Promise<{
  workoutId: number;
}> => {
  const workoutId = params.workoutId;
  context.queryClient.ensureQueryData(workoutQueryOptions(workoutId));
  context.queryClient.ensureQueryData(exercisesQueryOptions());
  context.queryClient.ensureQueryData(workoutsFocusValuesQueryOptions());
  return { workoutId };
},

// With (using factory functions):
loader: async ({
  context,
  params,
}): Promise<{
  workoutId: number;
}> => {
  const workoutId = params.workoutId;
  if (!context.user) initializeDemoData();

  await context.queryClient.ensureQueryData(getWorkoutByIdQueryOptions(context.user, workoutId));
  await context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
  await context.queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(context.user));

  return { workoutId };
},
```

#### 2.5 Update `RouteComponent` (Lines 331-352) - SIMPLIFIED

```tsx
// Replace:
function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const [{ data: exercises }, { data: workout }, { data: workoutsFocusValues }] = useSuspenseQueries({
    queries: [exercisesQueryOptions(), workoutQueryOptions(workoutId), workoutsFocusValuesQueryOptions()],
  });
  const workoutFormValues: WorkoutUpdateWorkoutRequest = transformToWorkoutFormValues(workout);

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map(
    (wf) => ({
      name: wf,
    })
  );

  return (
    <EditWorkoutForm
      exercises={exercises}
      workout={workoutFormValues}
      workoutId={workoutId}
      workoutsFocus={workoutsFocus}
    />
  );
}

// With (using factory functions):
function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();

  const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions(user));
  const { data: workout } = useSuspenseQuery(getWorkoutByIdQueryOptions(user, workoutId));
  const { data: workoutsFocusValues } = useSuspenseQuery(getWorkoutsFocusQueryOptions(user));

  // Transform workout data for form (works for both API and demo data)
  const workoutFormValues: WorkoutUpdateWorkoutRequest = transformToWorkoutFormValues(workout);

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <EditWorkoutForm
      user={user} // ADD THIS
      exercises={exercises}
      workout={workoutFormValues}
      workoutId={workoutId}
      workoutsFocus={workoutsFocus}
    />
  );
}
```

**Decision**: ✅ Use `useSuspenseQuery` (multiple calls)
**Rationale**:
- Data is tightly coupled (all needed before form renders)
- BUT factory pattern returns conditional query options
- TypeScript inference works better with separate calls
- Proven pattern from existing merged routes
- Consistent with other route implementations

**Note**: While `useSuspenseQueries` might seem better for tightly coupled data, the factory pattern makes multiple `useSuspenseQuery` calls cleaner and more type-safe.

**Add import at top:**
```tsx
import { useSuspenseQuery } from '@tanstack/react-query';
```

#### 2.6 Data Transformation - ✅ VERIFIED

**Status**: No changes needed!

`transformToWorkoutFormValues()` takes `WorkoutWorkoutWithSetsResponse[]` which is **exactly** what both API and demo data return. The function works for both modes without modification.

---

## Phase 3: Type Safety Verification

### 3.1 Run TypeScript Compiler
```bash
bun run tsc
```

**Expected Result**: 0 errors

**If errors occur**, check:
- Query option type compatibility
- Mutation response type compatibility
- User nullable handling

---

## Phase 4: Testing

### 4.1 Demo Mode Testing (`/workouts/new`)

**Setup**: Ensure logged out (demo mode)

- [ ] Navigate to `/workouts/new`
- [ ] Form loads with demo exercises
- [ ] Add an exercise
- [ ] Add sets with reps/weight
- [ ] Fill in date, notes, workout focus
- [ ] **Draft Persistence**: Refresh page → form data persists
- [ ] Save workout → Success message
- [ ] Navigate to `/workouts` → New workout appears
- [ ] **localStorage Check**: Verify demo data in DevTools (`demo_workouts`)

### 4.2 Demo Mode Testing (`/workouts/$workoutId/edit`)

**Setup**: Ensure logged out, have demo workout

- [ ] Navigate to `/workouts`
- [ ] Click on a workout
- [ ] Click "Edit" button
- [ ] Route loads at `/workouts/$workoutId/edit`
- [ ] Form loads with existing workout data
- [ ] Modify sets (add/remove/change)
- [ ] Modify notes/focus
- [ ] Save workout → Success message
- [ ] Navigate back to workout detail → Changes reflected
- [ ] **localStorage Check**: Verify updated data in DevTools

### 4.3 Authenticated Mode Testing (`/workouts/new`)

**Setup**: Login as authenticated user

- [ ] Navigate to `/workouts/new`
- [ ] Form loads with API exercises
- [ ] Add exercise and sets
- [ ] **Draft Persistence**: Refresh page → form data persists
- [ ] Save workout → API call successful
- [ ] Navigate to `/workouts` → New workout from API
- [ ] **localStorage Check**: Draft cleared after save
- [ ] **Demo Data Check**: Demo localStorage cleared on login

### 4.4 Authenticated Mode Testing (`/workouts/$workoutId/edit`)

**Setup**: Login as authenticated user, have API workout

- [ ] Navigate to `/workouts`
- [ ] Click on a workout
- [ ] Click "Edit" button
- [ ] Form loads with API workout data
- [ ] Modify workout
- [ ] Save → API call successful
- [ ] Changes reflected in workout detail

### 4.5 Edge Cases & Critical Scenarios

#### Auth State Transitions
- [ ] **Demo → Login transition**
  - Draft data handling (should clear or preserve?)
  - Query cache invalidation (ensure demo data doesn't leak)
  - Recent sets display switches correctly
- [ ] **Login → Logout → Demo mode**
  - Demo mode still works after logout
  - Demo data reinitializes properly
  - No stale API data in cache

#### RecentSets Component
- [ ] **Demo mode**: Recent sets load and display correctly
- [ ] **Auth mode**: Recent sets load from API
- [ ] **Empty state**: No recent sets shows nothing (not error)
- [ ] **Exercise switch**: Recent sets update when changing exercises

#### Data & Storage
- [ ] **Invalid workout ID** (404 handling in both modes)
- [ ] **Empty demo data** (fresh browser/cleared storage)
- [ ] **localStorage quota exceeded** (graceful degradation)
- [ ] **Multi-tab sync** (changes in one tab reflect in another)

#### Navigation & UX
- [ ] **Edit save navigation**: Correctly navigates to workout detail
- [ ] **New save behavior**: Decide - stay on form or navigate?
- [ ] **Network offline**: Demo mode continues to work

#### Error Scenarios
- [ ] **API error in auth mode**: Displays user-friendly message
- [ ] **Demo mutation error**: Displays consistent error format
- [ ] **Form validation errors**: Work in both modes

---

## Phase 5: Documentation Updates

### 5.1 Update `demo-plan.md`

**Section 2.2 (Lines 54-73)**: Update status
```markdown
#### `/workouts/new` Status: ✅ COMPLETE
- Route supports both auth and demo modes
- Conditional mutations (API vs demo)
- Draft persistence for both user types

#### `/workouts/$workoutId/edit` Status: ✅ COMPLETE (Updated from Auth-Only)
- Route supports both auth and demo modes
- Conditional mutations (API vs demo)
- Demo users can edit workouts
```

### 5.2 Update `NEXT-STEPS.md`

**Section "Priority 1"** (Lines 26-105): Remove blocker, add completion status
```markdown
## ✅ Priority 1: Complete - `/workouts/new` and `/edit` Fixed

Both routes now support demo and authenticated modes using unified pattern.

### Implementation Details
- Conditional mutations (API vs demo)
- Conditional query loaders
- Draft persistence for both user types
- Follows same pattern as view-only routes
```

### 5.3 Update `route-merge-proposal.md` (if exists)

Add section documenting the unified mutation pattern:
```markdown
## Mutation Pattern for Demo Support

All routes requiring mutations follow this pattern:

```tsx
// 1. Call both hooks unconditionally (rules of hooks)
const apiMutation = useApiMutation();
const demoMutation = useMutation(getDemoMutationOptions());

// 2. Use conditionally based on user
const mutation = user ? apiMutation : demoMutation;

// 3. Use mutation as normal
mutation.mutate({ body: data });
```
```

---

## Risk Assessment

### Low Risk ✅
- **Draft localStorage**: Already handles `undefined` userId
- **Demo mutations**: Already exist and tested in other routes
- **Query patterns**: Same as view-only routes (already working)
- **Data transformation**: `transformToWorkoutFormValues()` works with demo data ✅ (verified - takes `WorkoutWorkoutWithSetsResponse[]` which is exactly what demo returns)

### Medium Risk ⚠️
- **Type safety**: Conditional query types
  - **Mitigation**: Use separate `useSuspenseQuery()` calls (proven pattern)

### 🚨 CRITICAL EDGE CASES - MUST ADDRESS

#### 1. **RecentSets Component - BLOCKING ISSUE**
**Location**: `new.tsx:143` and `recent-sets-display.tsx:16-18`
**Problem**: Uses API-only query - will crash in demo mode
```tsx
// Current (broken for demo):
const { data: recentSets } = useSuspenseQuery(
  recentExerciseSetsQueryOptions(exerciseId)  // ❌ API only!
);
```
**Fix Required**: Update `RecentSetsDisplay` component to:
- Accept `user` prop
- Use conditional query: `user ? recentExerciseSetsQueryOptions(id) : getDemoExercisesByIdRecentSetsQueryOptions(id)`
- The demo query function already exists and works

#### 2. **Query Cache Pollution on Auth State Change**
**Problem**: When users login/logout, stale queries from the other mode remain cached
- Demo data showing briefly after login (until refetch)
- API data lingering after logout

**Status**: ⏸️ DEFERRED - Not in Scope
**Reason**:
- Auth handled by Stack Auth (hook-based, no programmatic control)
- Demo data functions use TanStack Query cache keys (e.g., `demo_getExercises`)
- API data uses different cache keys (e.g., `getExercises`)
- Cache keys are already separate, so collision risk is low
- Each route loader fetches fresh data on auth state change
**Future Work**: Monitor for actual cache pollution issues in testing

#### 3. **Navigation Inconsistency**
**Current State**:
- Edit route: Navigates to workout detail after save ✅
- New route: Just resets form, no navigation ❌

**Decision**: ✅ KEEP CURRENT BEHAVIOR
**Rationale**:
- `/workouts/new` should reset form in both auth and demo modes
- Consistent with current auth behavior
- Allows users to quickly log multiple workouts
- Edit navigates because user expects to return to detail view


#### 4. **localStorage Key Handling**
**Problem**: With `user?.id`, demo users use base key but might read wrong key after login
**Current**: `clearLocalStorage(user.id)` → TypeScript error when user is null
**Fix**: Use `user?.id` consistently, but verify key resolution logic

#### 5. **Error Handling Format Differences**
**Problem**: Demo mutations return structured errors, API mutations might return different shapes
**Current**: `alert(error)` approach displays inconsistently

**Status**: ⏸️ DEFERRED - Future Research
**Reason**:
- Need to check `server` folder for API error handling logic
- Current `alert(error)` works for MVP
- Can standardize in follow-up PR after understanding both error shapes
**Action**: Keep current error handling, document inconsistencies if found

#### 6. **localStorage Quota & Multi-tab Issues** (Lower Priority)
- No handling for storage quota exceeded
- No sync between tabs (localStorage events not handled)

**Status**: ⏸️ DEFERRED - Low Priority
**Reason**:
- Edge case - unlikely to hit quota with workout data
- Multi-tab sync not critical for MVP
- Can add localStorage event listeners later if needed
**Action**: Monitor for issues in testing, address if critical

---

## Simpler Implementation Approach

### 🎯 Factory Function Pattern (Recommended)

Instead of inline conditionals in every route, centralize the logic:

**Create**: `client/src/lib/api/unified-query-options.ts`
```tsx
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  exercisesQueryOptions,
  recentExerciseSetsQueryOptions
} from './exercises';
import {
  workoutsQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions
} from './workouts';
import {
  getDemoExercisesQueryOptions,
  getDemoExercisesByIdRecentSetsQueryOptions,
  getDemoWorkoutsQueryOptions,
  getDemoWorkoutsByIdQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from '@/lib/demo-data/query-options';

export function getExercisesQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? exercisesQueryOptions() : getDemoExercisesQueryOptions();
}

export function getRecentSetsQueryOptions(user: CurrentUser | CurrentInternalUser | null, exerciseId: number) {
  return user
    ? recentExerciseSetsQueryOptions(exerciseId)
    : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId);
}

export function getWorkoutsQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions();
}

export function getWorkoutByIdQueryOptions(user: CurrentUser | CurrentInternalUser | null, id: number) {
  return user ? workoutQueryOptions(id) : getDemoWorkoutsByIdQueryOptions(id);
}

export function getWorkoutsFocusQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsFocusValuesQueryOptions() : getDemoWorkoutsFocusValuesQueryOptions();
}
```

**✅ Approved**: Factory functions will be created in `client/src/lib/api/` folder.

### Benefits:
- ✅ **60% less code** in route files
- ✅ Single source of truth for conditional logic
- ✅ Easier to test (test the factory, not every component)
- ✅ Easier to maintain (change once, affects all routes)
- ✅ No duplication of conditional patterns
- ✅ TypeScript inference works perfectly

### Comparison:

**Before (Current Plan - Inline Conditionals)**:
```tsx
// Loader (15 lines):
loader: async ({ context }): Promise<void> => {
  const user = context.user;
  if (user) {
    await context.queryClient.ensureQueryData(exercisesQueryOptions());
    await context.queryClient.ensureQueryData(workoutsFocusValuesQueryOptions());
  } else {
    initializeDemoData();
    await context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
    await context.queryClient.ensureQueryData(getDemoWorkoutsFocusValuesQueryOptions());
  }
}

// Component (8 lines):
const { data: exercises } = user
  ? useSuspenseQuery(exercisesQueryOptions())
  : useSuspenseQuery(getDemoExercisesQueryOptions());

const { data: workoutsFocusValues } = user
  ? useSuspenseQuery(workoutsFocusValuesQueryOptions())
  : useSuspenseQuery(getDemoWorkoutsFocusValuesQueryOptions());
```

**After (Factory Pattern - Clean)**:
```tsx
// Loader (4 lines):
loader: async ({ context }): Promise<void> => {
  if (!context.user) initializeDemoData();
  await context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
  await context.queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(context.user));
}

// Component (2 lines):
const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions(user));
const { data: workoutsFocusValues } = useSuspenseQuery(getWorkoutsFocusQueryOptions(user));
```

---

## Success Criteria

### Functional Requirements ✅
- [ ] Both routes work in demo mode (unauthenticated)
- [ ] Both routes work in auth mode (authenticated)
- [ ] Draft persistence works for both user types
- [ ] Demo data persists across page reloads
- [ ] Auth data goes to API
- [ ] Seamless transition between modes

### Technical Requirements ✅
- [ ] TypeScript compilation: 0 errors
- [ ] No ESLint errors
- [ ] No runtime errors in console
- [ ] Query invalidation works correctly
- [ ] Loading states display properly

### Code Quality Requirements ✅
- [ ] Consistent pattern across both routes
- [ ] No code duplication
- [ ] Clear comments for conditional logic
- [ ] Documentation updated

---

## Rollback Plan

If implementation causes issues:

1. **Revert `/workouts/new`**:
   - Git: `git checkout HEAD -- client/src/routes/workouts/new.tsx`
   - Blockers remain but app is stable

2. **Revert `/workouts/$workoutId/edit`**:
   - Git: `git checkout HEAD -- client/src/routes/workouts/\$workoutId/edit.tsx`
   - Route remains auth-only (current state)

3. **If demo data corrupted**:
   - Clear localStorage: `localStorage.clear()` in DevTools
   - Refresh page → demo data reinitializes

---

## Estimated Time

| Phase | Task | Time | Total |
|-------|------|------|-------|
| 1 | Fix `/workouts/new` | 30 min | 30 min |
| 2 | Add demo to `/edit` | 45 min | 1h 15min |
| 3 | Type safety verification | 15 min | 1h 30min |
| 4 | Testing (both routes, both modes) | 45 min | 2h 15min |
| 5 | Documentation updates | 15 min | 2h 30min |

**Total Estimated Time**: 2.5 hours

---

## Next Steps After Implementation

### Optional Enhancements (Future Work)

1. **Unified Header Component**
   - Show demo banner when not authenticated
   - "You're using demo mode - Sign in to save your data"
   - Link to sign-in page

2. **Landing Page Integration**
   - Update "Try for free" button → `/workouts`
   - Auto-detects demo mode if not authenticated

3. **Form State Refactoring** (Low Priority)
   - Both routes share 90%+ identical UI code
   - Could extract into shared `<WorkoutForm>` component
   - Would reduce maintenance burden
   - Not critical for MVP

4. **Demo Data Export** (Nice to Have)
   - Allow demo users to export their workouts
   - "Download your data before signing up"
   - JSON export feature

---

## Questions for Review

### 🔴 Critical Decisions Needed

1. **Factory Function Pattern Adoption** ⭐ RECOMMENDED
   - **Question**: Should we use the factory function pattern (Phase 0) or stick with inline conditionals?
   - **Recommendation**: Use factory functions - 60% less code, easier maintenance, single source of truth
   - **Trade-off**: One extra file vs. massive code duplication in routes

2. **RecentSets Component Fix** 🚨 BLOCKING
   - **Question**: The `<RecentSets>` component is broken for demo mode. Fix is detailed in Phase 1.5.
   - **Recommendation**: MUST fix - will crash entire `/workouts/new` route in demo mode
   - **Verified**: Demo query already exists, just needs to be wired up

3. **Navigation Behavior After Save**
   - **Current**: Edit route navigates to workout detail, New route just resets form
   - **Question**: Should `/workouts/new` navigate to `/workouts` after successful save?
   - **Recommendation**: Yes, for consistency - or navigate to created workout detail if API returns ID

### 🟡 Important but Not Blocking

4. **Auth State Transition Handling**
   - **Question**: Should we invalidate query cache on login/logout to prevent stale data?
   - **Recommendation**: Yes - add to auth transition logic (not in this PR scope?)
   - **Impact**: Without this, users may see demo data briefly after login

5. **localStorage Key Strategy**
   - **Question**: With `user?.id`, how do we handle draft persistence across auth states?
   - **Current Approach**: Demo uses base key, auth uses user-keyed version
   - **Potential Issue**: Draft from demo mode won't carry over to auth mode
   - **Recommendation**: Document this behavior (or implement migration logic)

**Decision**: ✅ DOCUMENT CURRENT BEHAVIOR
**Current Behavior**:
- Route loaders check auth state and load correct data source
- Demo user draft → Login = Lost (uses different localStorage key)
- This is acceptable for MVP
**Future Enhancement**: Demo-to-auth data migration (out of scope for this PR)
**Action**: Document this behavior for users

6. **Error Handling Standardization**
   - **Question**: Should we standardize error display between API and demo mutations?
   - **Current**: Simple `alert(error)` - works but inconsistent formatting
   - **Recommendation**: Create unified error handler (low priority)

**Decision**: ⏸️ DEFERRED - Research server error handling later
**Action**: Keep current `alert(error)` approach, standardize in future PR

### 🟢 Nice to Have

7. **Testing Scope**
   - **Question**: Full manual testing vs. TypeScript + smoke testing?
   - **Recommendation**: At minimum: TypeScript compilation + test each mode once per route
   - **Full testing**: Use checklist in Phase 4.5

**Decision**: ✅ ADD BASIC TESTING
**Approach**:
- Use `vitest` or `playwright` for demo mode testing
- Focus on demo mode (auth testing complex due to Stack Auth)
- Simple but effective tests for critical paths
**Scope**:
- Demo mode: Create workout, edit workout, draft persistence
- TypeScript compilation (0 errors)
- Manual smoke testing for auth mode
**Future**: Expand auth testing when Stack Auth test patterns established

8. **Draft Persistence for Demo**
   - **Question**: Keep draft persistence for demo users or disable it?
   - **Current**: Enabled with base localStorage key
   - **Recommendation**: Keep it - good UX, already implemented

9. **Implementation Approach**
   - **Question**: Implement both routes in one session or separately?
   - **Recommendation**: Do Phase 0 + Phase 1 + Phase 1.5 first, test thoroughly, then Phase 2
   - **Reason**: RecentSets fix in Phase 1.5 is critical - verify before continuing

### 📊 Data Transformation (RESOLVED ✅)

10. **`transformToWorkoutFormValues()` Compatibility**
    - **Status**: ✅ VERIFIED - Works with both API and demo data
    - **Reason**: Function expects `WorkoutWorkoutWithSetsResponse[]` which both return
    - **Action**: None needed

---

## 📝 Plan Update Summary (2025-10-06)

### What Changed

#### ✅ **Added Critical Edge Cases Section**
Identified 6 critical edge cases that were missing from original plan:
1. 🚨 **RecentSets Component** - Blocking issue, will crash demo mode
2. Query cache pollution on auth state changes
3. Navigation inconsistency between routes
4. localStorage key handling issues
5. Error handling format differences
6. localStorage quota & multi-tab sync

#### ✅ **Added Simpler Implementation (Factory Function Pattern)**
- **Impact**: 60% code reduction in route files
- **New File**: `client/src/lib/api/unified-query-options.ts`
- **Benefits**: Single source of truth, easier testing, better maintainability
- **Comparison**: Before/after code examples showing dramatic simplification

#### ✅ **Updated All Implementation Phases**
- **Phase 0 (NEW)**: Create factory functions first
- **Phase 1**: Simplified using factory functions + added RecentSets fix (Phase 1.5)
- **Phase 2**: Simplified using factory functions
- **Phase 2.6**: Verified data transformation works (removed uncertainty)

#### ✅ **Enhanced Testing Section**
Added comprehensive edge case testing scenarios:
- Auth state transitions
- RecentSets component in both modes
- Data & storage edge cases
- Navigation & UX consistency
- Error scenarios

#### ✅ **Updated Questions for Review**
Reorganized into priority tiers:
- 🔴 Critical (3 questions) - Must decide before starting
- 🟡 Important (3 questions) - Should address
- 🟢 Nice to Have (3 questions) - Optional
- ✅ Resolved (1 question) - Data transformation verified

### Key Recommendations

1. **MUST DO**: Adopt factory function pattern (Phase 0)
2. **MUST FIX**: RecentSets component (Phase 1.5) - blocking issue
3. **SHOULD DO**: Decide on navigation behavior after save
4. **CONSIDER**: Query cache invalidation on auth transitions

### Original Plan Strengths Preserved

- ✅ Unified pattern across routes
- ✅ Draft persistence design
- ✅ Leveraging existing infrastructure
- ✅ Clear success criteria
- ✅ Rollback plan

### Timeline Impact

- Original estimate: 2.5 hours
- **With factory pattern**: Reduces to ~2 hours (saves implementation time)
- **RecentSets fix**: Adds ~15 min (critical, unavoidable)
- **New estimate**: 2-2.5 hours total

---

---

## 📋 Final Approval Summary

### ✅ Decisions Made

| # | Decision | Status | Action |
|---|----------|--------|--------|
| 1 | Factory function pattern | ✅ APPROVED | Create in `client/src/lib/api/` |
| 2 | RecentSets component fix | 🚨 CRITICAL | Must fix in Phase 1.5 |
| 3 | Navigation after save | ✅ APPROVED | Keep current (reset form) |
| 4 | Query cache invalidation | ⏸️ DEFERRED | Monitor, not in scope |
| 5 | Draft persistence across auth | ✅ DOCUMENT | Current behavior acceptable |
| 6 | Error handling standardization | ⏸️ DEFERRED | Research later |
| 7 | localStorage quota/multi-tab | ⏸️ DEFERRED | Low priority |
| 8 | Testing approach | ✅ APPROVED | Basic tests (vitest/playwright) |
| 9 | useSuspenseQuery vs Queries | ✅ APPROVED | Multiple `useSuspenseQuery` |
| 10 | Data transformation | ✅ VERIFIED | Already compatible |