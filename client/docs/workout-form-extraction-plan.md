# Workout Form Component Extraction Plan

## Document Purpose
This document provides a detailed plan for extracting the `WorkoutTracker` component from `client/src/routes/_auth/workouts/new.tsx` into a shared component at `client/src/components/workouts/workout-form.tsx` so it can be reused by both auth and demo routes.

**Created**: 2025-10-05
**Status**: Planning Phase
**Approach**: Option A - Extract with Parameterization (RECOMMENDED)

---

## Table of Contents
1. [Current State Analysis](#current-state-analysis)
2. [Option A: Extract with Parameterization (RECOMMENDED)](#option-a-extract-with-parameterization-recommended)
3. [Option B: Duplicate Route (FALLBACK)](#option-b-duplicate-route-fallback)
4. [Option C: Hybrid Approach](#option-c-hybrid-approach)
5. [Rollback Strategy](#rollback-strategy)
6. [Success Criteria](#success-criteria)

---

## Current State Analysis

### File Structure
```
client/src/routes/_auth/workouts/new.tsx (354 lines)
├── WorkoutTracker component (lines 25-312)
│   ├── Main form view (workout overview with exercise cards)
│   ├── Add exercise screen (lines 103-119)
│   └── Exercise screen (lines 122-149)
└── RouteComponent (lines 327-353)
    └── Loader (lines 315-325)
```

### Dependencies Breakdown

#### 1. **User Dependency** (4 locations)
- **Line 45**: `getInitialValues(user.id)` - Load saved form from localStorage
- **Line 49**: `saveToLocalStorage(formApi.state.values, user.id)` - Auto-save on change
- **Line 64**: `clearLocalStorage(user.id)` - Clear on successful save
- **Line 96**: `clearLocalStorage(user.id)` - Clear when user clicks "Clear" button

**Analysis**: User is ONLY used for localStorage scoping. All 4 calls use `user.id` string.

#### 2. **Mutation Dependency** (1 location)
- **Line 42**: `useSaveWorkoutMutation()` from `@/lib/api/workouts`
  - Returns: `UseMutationResult<ResponseSuccessResponse, Error, { body: WorkoutCreateWorkoutRequest }>`
  - Called at line 60: `saveWorkout.mutateAsync({ body: trimmedValue })`
  - Success callback (line 63): Clears localStorage and resets form
  - Error callback (line 68): Shows alert

**Analysis**: Mutation is swappable - demo needs `postDemoWorkoutsMutation()` instead.

#### 3. **Query Dependencies** (2 locations)
- **Line 13**: `exercisesQueryOptions()` - Get all exercises for dropdown
- **Line 15**: `workoutsFocusValuesQueryOptions()` - Get workout focus values for combobox
- **Lines 329-332**: Loaded via `useSuspenseQueries` in RouteComponent

**Analysis**: These are passed as props to `WorkoutTracker` (lines 27-28). Already parameterized!

#### 4. **Child Component Dependencies** (3 route-specific components)

##### a. `AddExerciseScreen` (line 112)
**File**: `client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx` (142 lines)

**Props**:
```tsx
{
  form: UseAppForm,           // TanStack Form instance
  exercises: DbExercise[],     // Array of { id, name }
  onAddExercise: (index, exerciseId?) => void,
  onBack: () => void
}
```

**User Dependency**: ❌ NONE - Uses `withForm` HOC with `MOCK_VALUES`, no user prop

**Complexity**: ✅ LOW - Already reusable, just needs to be moved to shared location

**Action Required**: Move to `client/src/components/workouts/add-exercise-screen.tsx`

---

##### b. `ExerciseScreen` + `ExerciseHeader` + `ExerciseSets` (line 135)
**File**: `client/src/routes/_auth/workouts/-components/exercise-screen.tsx` (202 lines)

**Props**:
```tsx
// ExerciseScreen (container)
{
  header: React.ReactNode,
  recentSets?: React.ReactNode,
  sets: React.ReactNode
}

// ExerciseHeader (wrapped with withForm)
{
  form: UseAppForm,
  exerciseIndex: number,
  onBack: () => void
}

// ExerciseSets (wrapped with withForm)
{
  form: UseAppForm,
  exerciseIndex: number
}
```

**User Dependency**: ❌ NONE - Uses `withForm` HOC with `MOCK_VALUES`, no user prop

**Complexity**: ✅ LOW - Already reusable, just needs to be moved

**Action Required**: Move to `client/src/components/workouts/exercise-screen.tsx`

---

##### c. `RecentSets` (line 143)
**File**: `client/src/routes/_auth/workouts/-components/recent-sets-display.tsx` (123 lines)

**Props**:
```tsx
{
  exerciseId: number | null
}
```

**Query Dependency**:
- **Line 16-18**: Uses `recentExerciseSetsQueryOptions(exerciseId)` from `@/lib/api/exercises`
- Returns: `ExerciseRecentSetsResponse[]`

**User Dependency**: ❌ NONE

**Complexity**: ⚠️ MEDIUM - Needs demo version of `recentExerciseSetsQueryOptions`

**Blocker Analysis**:
- Current query: `getExercisesByIdRecentSetsQueryOptions({ path: { id } })` (from API)
- Demo query: `getDemoExercisesByIdRecentSetsQueryOptions(id)` (from demo-data)
- **Already exists!** See `client/src/lib/demo-data/query-options.ts` lines 79-86

**Action Required**:
1. Extract to `client/src/components/workouts/recent-sets-display.tsx`
2. Add `queryOptions` prop to inject either auth or demo query options
3. OR: Create two versions (`RecentSets` for auth, `DemoRecentSets` for demo) if injection is too complex

---

#### 5. **Form System Dependencies** (TanStack Form)

**Hook**: `useAppForm` from `@/hooks/form` (line 43)
- Custom hook created via `createFormHook` from `@tanstack/react-form`
- Provides field components: `DatePicker2`, `NotesTextarea2`, `WorkoutFocusCombobox`, etc.

**HOC**: `withForm` from `@/hooks/form`
- Used by child components: `AddExerciseScreen`, `ExerciseHeader`, `ExerciseSets`, `AddSetDialog`
- Injects form context with `MOCK_VALUES` as `defaultValues`

**MOCK_VALUES**: From `form-options.ts` (line 4-9)
```tsx
export const MOCK_VALUES: WorkoutCreateWorkoutRequest = {
  date: new Date().toISOString(),
  notes: '',
  exercises: [] as Array<WorkoutExerciseInput>,
  workoutFocus: '',
};
```

**Analysis**:
- ✅ `MOCK_VALUES` is already a constant - no user dependency
- ✅ Form system is fully reusable - no changes needed
- ⚠️ Child components use `withForm({ defaultValues: MOCK_VALUES })` - already parameterized correctly

---

#### 6. **LocalStorage Utilities**

**File**: `client/src/lib/local-storage.ts` (56 lines)

**Functions**:
```tsx
saveToLocalStorage(data: FormDataType, userId?: string): void
loadFromLocalStorage(userId?: string): WorkoutCreateWorkoutRequest | null
clearLocalStorage(userId?: string): void
```

**Storage Key Pattern**: `workout-entry-form-data-${userId}` (line 6)

**Analysis**:
- ✅ Already accepts optional `userId` parameter
- ✅ Demo can call with `userId = 'demo-user'`
- ✅ No changes needed to localStorage utilities

**Demo Usage**:
```tsx
// Auth route
saveToLocalStorage(data, user.id);  // "workout-entry-form-data-abc123"

// Demo route
saveToLocalStorage(data, 'demo-user');  // "workout-entry-form-data-demo-user"
```

---

### Key Finding: Child Components are Already User-Agnostic! ✅

**All 3 child components** (`AddExerciseScreen`, `ExerciseScreen`, `RecentSets`) have **ZERO user dependency**:
- They use `withForm` HOC with `MOCK_VALUES` (no user needed)
- They accept form/data via props
- They only handle UI logic, not data fetching (except `RecentSets`)

**This significantly reduces extraction complexity!**

---

## Option A: Extract with Parameterization (RECOMMENDED)

### Strategy
Extract `WorkoutTracker` to `client/src/components/workouts/workout-form.tsx` and inject dependencies via props.

### Implementation Steps

#### Step 1: Move Child Components to Shared Location
**Estimated Time**: 30 minutes
**Risk**: ⚠️ Low - Components already user-agnostic
**Rollback**: Delete new files, revert imports

```bash
# Create new shared component files
client/src/components/workouts/add-exercise-screen.tsx  # Move from routes/-components/
client/src/components/workouts/exercise-screen.tsx      # Move from routes/-components/
client/src/components/workouts/recent-sets-display.tsx  # Move from routes/-components/
client/src/components/workouts/add-set-dialog.tsx       # Move from routes/-components/
```

**Actions**:
1. ✅ Copy `add-exercise-screen.tsx` → `components/workouts/add-exercise-screen.tsx`
2. ✅ Copy `exercise-screen.tsx` → `components/workouts/exercise-screen.tsx`
3. ✅ Copy `add-set-dialog.tsx` → `components/workouts/add-set-dialog.tsx`
4. ⚠️ Handle `RecentSets` (see Step 1.5 below)
5. ✅ Update imports in all files to use `@/components/workouts/` prefix
6. ✅ Keep route-specific copies temporarily for rollback safety

**Verification**:
- Run TypeScript compiler: `npm run typecheck`
- Verify auth route still works: Navigate to `/workouts/new` and test form

---

#### Step 1.5: Handle `RecentSets` Component
**Estimated Time**: 20 minutes
**Risk**: ⚠️ Medium - Needs query parameterization

**Decision Point**: Choose ONE of these approaches:

##### Approach 1A: Inject Query Options (Cleaner, More Complex)
```tsx
// client/src/components/workouts/recent-sets-display.tsx
import { Suspense } from 'react';
import { useSuspenseQuery, type QueryOptions } from '@tanstack/react-query';
import type { ExerciseRecentSetsResponse } from '@/client';
// ... rest of imports

interface RecentSetsProps {
  exerciseId: number | null;
  queryOptions: (id: number) => QueryOptions<ExerciseRecentSetsResponse[]>;
}

function RecentSetsDisplay({ exerciseId, queryOptions }: RecentSetsProps) {
  const { data: recentSets } = useSuspenseQuery(queryOptions(exerciseId!));
  // ... rest of component
}

export function RecentSets({ exerciseId, queryOptions }: RecentSetsProps) {
  if (!exerciseId) return null;
  return (
    <Suspense fallback={<div>Loading recent sets...</div>}>
      <RecentSetsDisplay exerciseId={exerciseId} queryOptions={queryOptions} />
    </Suspense>
  );
}
```

**Usage in routes**:
```tsx
// Auth route
import { recentExerciseSetsQueryOptions } from '@/lib/api/exercises';
<RecentSets exerciseId={id} queryOptions={recentExerciseSetsQueryOptions} />

// Demo route
import { getDemoExercisesByIdRecentSetsQueryOptions } from '@/lib/demo-data/query-options';
<RecentSets exerciseId={id} queryOptions={getDemoExercisesByIdRecentSetsQueryOptions} />
```

**Pros**: Single component, DRY principle
**Cons**: Adds complexity to component API

---

##### Approach 1B: Duplicate Component (Simpler, Faster)
```tsx
// client/src/components/workouts/recent-sets-display.tsx (Auth version)
import { recentExerciseSetsQueryOptions } from '@/lib/api/exercises';
// ... existing implementation

// client/src/components/workouts/demo-recent-sets-display.tsx (Demo version)
import { getDemoExercisesByIdRecentSetsQueryOptions } from '@/lib/demo-data/query-options';
// ... same implementation, different query
```

**Pros**: Simpler, no API changes, faster to implement
**Cons**: ~120 lines duplicated, violates DRY

---

**RECOMMENDATION**: Start with **Approach 1B** (duplicate) for speed. Refactor to 1A later if needed.

**Rollback**: If blocked, keep `RecentSets` in route file and pass as prop to `WorkoutTracker`.

---

#### Step 2: Extract `WorkoutTracker` Component
**Estimated Time**: 45 minutes
**Risk**: ⚠️ Medium - Core form logic
**Rollback**: Revert extraction, restore route file

**New File**: `client/src/components/workouts/workout-form.tsx`

**Component Signature**:
```tsx
import type { UseMutationResult } from '@tanstack/react-query';
import type {
  WorkoutCreateWorkoutRequest,
  ResponseSuccessResponse,
  WorkoutFocus,
} from '@/client/types.gen';
import type { DbExercise } from '@/lib/api/exercises';

export interface WorkoutFormProps {
  // User ID for localStorage scoping
  userId: string;

  // Data passed from route queries
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];

  // Mutation hook (different for auth vs demo)
  saveWorkoutMutation: UseMutationResult<
    ResponseSuccessResponse,
    Error,
    { body: WorkoutCreateWorkoutRequest }
  >;

  // LocalStorage utilities (already parameterized)
  // Option 1: Pass userId and use existing utilities internally
  // Option 2: Inject utilities as functions (more testable)
  // RECOMMENDATION: Option 1 (simpler)
}

export function WorkoutForm({
  userId,
  exercises,
  workoutsFocus,
  saveWorkoutMutation,
}: WorkoutFormProps) {
  // ... existing WorkoutTracker implementation
  // Replace `user.id` with `userId` prop
  // Replace `useSaveWorkoutMutation()` with `saveWorkoutMutation` prop
}
```

**Changes Required**:
1. Replace `user.id` → `userId` (4 locations: lines 45, 49, 64, 96)
2. Replace `const saveWorkout = useSaveWorkoutMutation()` → Use `saveWorkoutMutation` prop
3. Import child components from `@/components/workouts/` instead of `./components/`
4. Export `WorkoutForm` component and `WorkoutFormProps` interface

**Verification**:
- Component compiles without errors
- All child components resolve correctly
- TypeScript types match expected signatures

---

#### Step 3: Update Auth Route (`_auth/workouts/new.tsx`)
**Estimated Time**: 15 minutes
**Risk**: ✅ Low - Simple prop wiring
**Rollback**: Revert to original route file

**New Route Implementation**:
```tsx
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQueries, useMutation } from '@tanstack/react-query';
import { WorkoutForm } from '@/components/workouts/workout-form';
import { exercisesQueryOptions, type DbExercise } from '@/lib/api/exercises';
import {
  workoutsFocusValuesQueryOptions,
  useSaveWorkoutMutation,
  type WorkoutFocus,
} from '@/lib/api/workouts';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';

export const Route = createFileRoute('/_auth/workouts/new')({\n  loader: async ({ context }): Promise<{ user: CurrentUser | CurrentInternalUser }> => {
    const user = context.user;
    context.queryClient.ensureQueryData(exercisesQueryOptions());
    return { user };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();
  const saveWorkoutMutation = useSaveWorkoutMutation();

  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({
      queries: [exercisesQueryOptions(), workoutsFocusValuesQueryOptions()],
    });

  const exercises: DbExercise[] = exercisesResponse.map((ex) => ({
    id: ex.id,
    name: ex.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <WorkoutForm
      userId={user.id}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
      saveWorkoutMutation={saveWorkoutMutation}
    />
  );
}
```

**File Size Reduction**: 354 lines → ~50 lines (85% reduction!)

**Verification**:
- Navigate to `/workouts/new`
- Test all form functionality:
  - Add exercise
  - Add sets
  - Auto-save to localStorage
  - Clear form
  - Submit workout
  - Recent sets display

---

#### Step 4: Create Demo Route (`demo/workouts/new.tsx`)
**Estimated Time**: 15 minutes
**Risk**: ✅ Low - Copy of auth route with swapped dependencies
**Rollback**: Delete demo route file

**New Route Implementation**:
```tsx
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQueries, useMutation } from '@tanstack/react-query';
import { WorkoutForm } from '@/components/workouts/workout-form';
import type { DbExercise } from '@/lib/api/exercises';
import type { WorkoutFocus } from '@/lib/api/workouts';
import {
  getDemoExercisesQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
  postDemoWorkoutsMutation,
} from '@/lib/demo-data/query-options';

export const Route = createFileRoute('/demo/workouts/new')({\n  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const saveWorkoutMutation = useMutation(postDemoWorkoutsMutation());

  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({
      queries: [getDemoExercisesQueryOptions(), getDemoWorkoutsFocusValuesQueryOptions()],
    });

  const exercises: DbExercise[] = exercisesResponse.map((ex) => ({
    id: ex.id,
    name: ex.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <WorkoutForm
      userId="demo-user"
      exercises={exercises}
      workoutsFocus={workoutsFocus}
      saveWorkoutMutation={saveWorkoutMutation}
    />
  );
}
```

**Key Differences from Auth Route**:
1. ✅ Route path: `/demo/workouts/new` instead of `/_auth/workouts/new`
2. ✅ No user from loader (demo doesn't have auth)
3. ✅ Uses `getDemoExercisesQueryOptions()` instead of `exercisesQueryOptions()`
4. ✅ Uses `getDemoWorkoutsFocusValuesQueryOptions()` instead of `workoutsFocusValuesQueryOptions()`
5. ✅ Uses `postDemoWorkoutsMutation()` instead of `useSaveWorkoutMutation()`
6. ✅ `userId="demo-user"` hardcoded instead of `user.id`

**Verification**:
- Navigate to `/demo/workouts/new`
- Test form functionality:
  - Add exercise from demo data
  - Add sets
  - Auto-save to localStorage (key: `workout-entry-form-data-demo-user`)
  - Submit workout
  - Verify workout appears in demo workout list (`/demo/workouts`)

---

#### Step 5: Clean Up Route-Specific Components
**Estimated Time**: 10 minutes
**Risk**: ✅ Low - Safe deletion after verification
**Rollback**: Restore from git

**Actions**:
1. ✅ Delete `client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx`
2. ✅ Delete `client/src/routes/_auth/workouts/-components/exercise-screen.tsx`
3. ✅ Delete `client/src/routes/_auth/workouts/-components/add-set-dialog.tsx`
4. ✅ Delete `client/src/routes/_auth/workouts/-components/recent-sets-display.tsx` (if moved)
5. ✅ Keep `form-options.ts` and `mini-chart.tsx` (still used by other route files)

**Verification**:
- No TypeScript errors
- Run `npm run build` to ensure no import issues
- Both auth and demo routes still work

---

### Total Estimated Time: **2.5 hours**

### Files Created/Modified Summary

**New Files** (6 files):
```
client/src/components/workouts/
  ├── workout-form.tsx              # 320 lines - Extracted WorkoutTracker
  ├── add-exercise-screen.tsx       # 142 lines - Moved from route
  ├── exercise-screen.tsx           # 202 lines - Moved from route
  ├── add-set-dialog.tsx            # 90 lines  - Moved from route
  └── recent-sets-display.tsx       # 123 lines - Moved from route (or duplicated)

client/src/routes/demo/workouts/
  └── new.tsx                       # 50 lines  - New demo route
```

**Modified Files** (1 file):
```
client/src/routes/_auth/workouts/
  └── new.tsx                       # 354 → 50 lines (85% reduction)
```

**Deleted Files** (4 files):
```
client/src/routes/_auth/workouts/-components/
  ├── add-exercise-screen.tsx       # Moved to components/
  ├── exercise-screen.tsx           # Moved to components/
  ├── add-set-dialog.tsx            # Moved to components/
  └── recent-sets-display.tsx       # Moved to components/ (or duplicated)
```

**Total Code Written**: ~930 lines (mostly moved, not new)
**Net New Code**: ~50 lines (demo route wrapper)

---

### Potential Blockers & Solutions

#### Blocker 1: `RecentSets` Query Injection Too Complex
**Symptom**: TypeScript errors when trying to inject query options
**Solution**: Fallback to Approach 1B (duplicate component)
**Time Impact**: +10 minutes
**Rollback**: Use duplicated `DemoRecentSets` component

---

#### Blocker 2: Mutation Type Mismatch
**Symptom**: `saveWorkoutMutation` prop type doesn't match between auth and demo
**Diagnosis**: Check return types of `useSaveWorkoutMutation()` vs `postDemoWorkoutsMutation()`
**Solution**:
- Option 1: Use generic mutation type in `WorkoutFormProps`
- Option 2: Create type alias for mutation result
```tsx
type SaveWorkoutMutation = UseMutationResult<
  ResponseSuccessResponse,
  Error,
  { body: WorkoutCreateWorkoutRequest }
>;
```
**Time Impact**: +15 minutes
**Rollback**: Proceed to Option B (duplicate route)

---

#### Blocker 3: Form Context Issues with `withForm` HOC
**Symptom**: Child components can't access form context
**Diagnosis**: `withForm` creates new form instance instead of using parent form
**Solution**: Child components already use `withForm` correctly - should work as-is
**If Still Blocked**: Pass form instance explicitly to child components
**Time Impact**: +30 minutes
**Rollback**: Proceed to Option B (duplicate route)

---

#### Blocker 4: LocalStorage Key Collision
**Symptom**: Auth and demo forms overwrite each other's localStorage
**Diagnosis**: Should NOT happen - keys are scoped by `userId`
- Auth: `workout-entry-form-data-${user.id}` (e.g., `workout-entry-form-data-abc123`)
- Demo: `workout-entry-form-data-demo-user`
**Solution**: If collision occurs, verify `userId` prop is being passed correctly
**Time Impact**: +10 minutes
**Rollback**: Not needed - this is a bug, not a blocker

---

### Success Criteria

#### Phase 1: Component Extraction ✅
- [ ] All child components moved to `client/src/components/workouts/`
- [ ] TypeScript compiles without errors
- [ ] Auth route (`/_auth/workouts/new`) still works
- [ ] Form auto-save persists across page reloads
- [ ] Workout submission succeeds and redirects
- [ ] Recent sets display for existing exercises

#### Phase 2: Demo Route Creation ✅
- [ ] Demo route (`/demo/workouts/new`) renders without errors
- [ ] Demo exercises appear in "Add Exercise" screen
- [ ] Demo workout focus values appear in combobox
- [ ] Form auto-saves to `workout-entry-form-data-demo-user` key
- [ ] Workout submission creates new workout in demo localStorage
- [ ] New workout appears in `/demo/workouts` list
- [ ] Recent sets display for demo exercises (if implemented)

#### Phase 3: Code Quality ✅
- [ ] No code duplication (except `RecentSets` if using Approach 1B)
- [ ] TypeScript types are accurate and strict
- [ ] No `any` types introduced
- [ ] Comments explain key decisions (especially prop injection)
- [ ] File structure is logical and follows existing patterns

#### Phase 4: Documentation ✅
- [ ] Update `demo-plan.md` to check off Phase 2.3 tasks
- [ ] Update `demo-progress.md` with extraction details
- [ ] Document decision to extract vs duplicate `RecentSets`
- [ ] Note any deviations from original plan

---

## Option B: Duplicate Route (FALLBACK)

### When to Use This Option
- Option A encounters multiple blockers (2+ blockers)
- Time constraints require faster delivery
- Risk tolerance is low (cannot risk breaking auth flow)
- You want to ship demo routes ASAP and refactor later

### Strategy
Copy `_auth/workouts/new.tsx` → `demo/workouts/new.tsx` and swap dependencies inline.

### Implementation Steps

#### Step 1: Copy Route File
**Estimated Time**: 5 minutes
**Risk**: ✅ Very Low

```bash
cp client/src/routes/_auth/workouts/new.tsx client/src/routes/demo/workouts/new.tsx
```

---

#### Step 2: Update Demo Route Dependencies
**Estimated Time**: 20 minutes
**Risk**: ✅ Low

**Changes Required**:
1. Update route path (line 314):
```tsx
// Before
export const Route = createFileRoute('/_auth/workouts/new')({

// After
export const Route = createFileRoute('/demo/workouts/new')({
```

2. Remove user from loader (lines 315-323):
```tsx
// Before
loader: async ({ context }): Promise<{ user: CurrentUser | CurrentInternalUser }> => {
  const user = context.user;
  context.queryClient.ensureQueryData(exercisesQueryOptions());
  return { user };
},

// After
loader: async ({ context }) => {
  context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
},
```

3. Update imports (lines 4, 13, 15):
```tsx
// Before
import { useSaveWorkoutMutation, type WorkoutFocus } from '@/lib/api/workouts';
import { exercisesQueryOptions, type DbExercise } from '@/lib/api/exercises';
import { workoutsFocusValuesQueryOptions } from '@/lib/api/workouts';

// After
import { type WorkoutFocus } from '@/lib/api/workouts';
import { type DbExercise } from '@/lib/api/exercises';
import {
  getDemoExercisesQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
  postDemoWorkoutsMutation,
} from '@/lib/demo-data/query-options';
```

4. Update RouteComponent to use demo queries (lines 327-353):
```tsx
// Before
function RouteComponent() {
  const { user } = Route.useLoaderData();
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({
      queries: [exercisesQueryOptions(), workoutsFocusValuesQueryOptions()],
    });
  // ...
  return (
    <WorkoutTracker
      user={user}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}

// After
function RouteComponent() {
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({
      queries: [getDemoExercisesQueryOptions(), getDemoWorkoutsFocusValuesQueryOptions()],
    });
  // ...
  return (
    <WorkoutTracker
      userId="demo-user"  // <-- Changed from user prop
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}
```

5. Update `WorkoutTracker` to accept `userId` string instead of `user` object:
```tsx
// Before (line 25-32)
function WorkoutTracker({
  user,
  exercises,
  workoutsFocus,
}: {
  user: CurrentUser | CurrentInternalUser;
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];
}) {

// After
function WorkoutTracker({
  userId,
  exercises,
  workoutsFocus,
}: {
  userId: string;
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];
}) {
```

6. Replace all `user.id` with `userId` (lines 45, 49, 64, 96):
```tsx
// Before
defaultValues: getInitialValues(user.id),
saveToLocalStorage(formApi.state.values, user.id);
clearLocalStorage(user.id);

// After
defaultValues: getInitialValues(userId),
saveToLocalStorage(formApi.state.values, userId);
clearLocalStorage(userId);
```

7. Replace mutation hook (line 42):
```tsx
// Before
const saveWorkout = useSaveWorkoutMutation();

// After
const saveWorkout = useMutation(postDemoWorkoutsMutation());
```

8. Update `RecentSets` import and usage (line 143):
```tsx
// Before
import { RecentSets } from './-components/recent-sets-display';
// ... later
<RecentSets exerciseId={selectedExercise.exerciseId} />

// After - Option 1: Create demo version
import { DemoRecentSets } from './-components/demo-recent-sets-display';
<DemoRecentSets exerciseId={selectedExercise.exerciseId} />

// After - Option 2: Disable for now
{/* <RecentSets exerciseId={selectedExercise.exerciseId} /> */}
```

---

#### Step 3: Handle Route-Specific Components
**Estimated Time**: 15 minutes
**Risk**: ✅ Low

**Option 3A**: Keep using auth route components
- Import from `@/routes/_auth/workouts/-components/`
- Works as-is, no changes needed
- Creates coupling between auth and demo routes

**Option 3B**: Copy components to demo route folder
```bash
mkdir -p client/src/routes/demo/workouts/-components
cp -r client/src/routes/_auth/workouts/-components/* client/src/routes/demo/workouts/-components/
```
- Update imports in `new.tsx` to use `./-components/`
- Fully independent from auth route
- More duplication (~600 lines)

**RECOMMENDATION**: Option 3A (reuse auth components) for speed. Components are already user-agnostic.

---

#### Step 4: Create Demo `RecentSets` Component
**Estimated Time**: 10 minutes
**Risk**: ✅ Low

**File**: `client/src/routes/demo/workouts/-components/demo-recent-sets-display.tsx`

```tsx
import { Suspense } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChevronRight } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import type { ExerciseRecentSetsResponse } from '@/client';
import { formatDate } from '@/lib/utils';
import { getDemoExercisesByIdRecentSetsQueryOptions } from '@/lib/demo-data/query-options';
import { sortByExerciseAndSetOrder } from '@/lib/utils';

interface DemoRecentSetsDisplayProps {
  exerciseId: number;
}

function DemoRecentSetsDisplay({ exerciseId }: DemoRecentSetsDisplayProps) {
  const { data: recentSets } = useSuspenseQuery(
    getDemoExercisesByIdRecentSetsQueryOptions(exerciseId)
  );

  // ... rest of implementation is IDENTICAL to auth version
  // Just copy from client/src/routes/_auth/workouts/-components/recent-sets-display.tsx
}

export function DemoRecentSets({ exerciseId }: { exerciseId: number | null }) {
  if (!exerciseId) return null;
  return (
    <Suspense fallback={<div>Loading recent sets...</div>}>
      <DemoRecentSetsDisplay exerciseId={exerciseId} />
    </Suspense>
  );
}
```

**Shortcut**: If you want to skip this for MVP, just comment out `<RecentSets>` in demo route.

---

### Total Estimated Time: **50 minutes**

### Files Created Summary

**New Files** (2 files):
```
client/src/routes/demo/workouts/
  ├── new.tsx                                    # 354 lines - Copy of auth route with demo deps
  └── -components/demo-recent-sets-display.tsx  # 123 lines - Copy of auth RecentSets
```

**Total Code Duplication**: ~477 lines

---

### Pros & Cons

**✅ Pros**:
- Fast to implement (< 1 hour)
- Zero risk to auth flow (no changes to existing files)
- Easy to understand (no abstraction complexity)
- Can be refactored to Option A later

**❌ Cons**:
- Code duplication (~477 lines)
- Future UI changes require updates in 2 places
- Violates DRY principle
- Not as maintainable long-term

---

### When to Use Option B

Use this option if:
1. **Timeline is critical** - Need demo routes working today
2. **Risk-averse environment** - Cannot afford to break auth flow
3. **Option A encounters blockers** - Multiple issues during extraction
4. **Short-term solution acceptable** - Plan to refactor later (Phase 3+)

---

## Option C: Hybrid Approach

### Strategy
Extract only the child components to shared location, but keep `WorkoutTracker` duplicated in routes.

**Rationale**: Child components are already user-agnostic and easy to extract. The main `WorkoutTracker` component has complex user/mutation dependencies - keep it duplicated for now.

### Implementation Steps

1. ✅ Extract child components (Step 1 from Option A) - **30 minutes**
2. ✅ Update auth route to import from `@/components/workouts/` - **10 minutes**
3. ✅ Copy auth route to demo route (Step 1 from Option B) - **5 minutes**
4. ✅ Update demo route dependencies (Step 2 from Option B) - **20 minutes**
5. ✅ Create demo `RecentSets` (Step 4 from Option B) - **10 minutes**

### Total Estimated Time: **1.25 hours**

### Pros & Cons

**✅ Pros**:
- Balances DRY and pragmatism
- Shared child components (~600 lines reused)
- Lower risk than full extraction
- Faster than Option A, better than Option B

**❌ Cons**:
- Still duplicates `WorkoutTracker` main logic (~150 lines)
- Future changes to form flow need 2 updates
- Not as clean as Option A

---

### When to Use Option C

Use this option if:
1. **Moderate time constraints** - Need demo working in 1-2 hours
2. **Want to minimize duplication** - But not at the cost of complexity
3. **Child components are reusable elsewhere** - E.g., if you plan to use `ExerciseScreen` in other contexts
4. **Option A feels too risky** - But Option B feels too lazy

**VERDICT**: Option C is a **solid middle ground** if you're unsure about Option A but want better than Option B.

---

## Rollback Strategy

### If Blocked During Option A

#### Rollback Point 1: After Step 1 (Child Components Moved)
**Trigger**: TypeScript errors, child components don't work
**Action**:
1. Delete new files in `client/src/components/workouts/`
2. Revert `client/src/routes/_auth/workouts/new.tsx` to original imports
3. `git checkout client/src/routes/_auth/workouts/new.tsx`
**Next Step**: Proceed to Option B or Option C

---

#### Rollback Point 2: After Step 2 (WorkoutTracker Extracted)
**Trigger**: Form doesn't work, mutation errors, localStorage issues
**Action**:
1. Delete `client/src/components/workouts/workout-form.tsx`
2. Restore original `client/src/routes/_auth/workouts/new.tsx` from git
3. `git checkout client/src/routes/_auth/workouts/new.tsx`
4. Verify auth route works again
**Next Step**: Proceed to Option B

---

#### Rollback Point 3: After Step 3 (Auth Route Updated)
**Trigger**: Auth route broken, users can't create workouts
**Action**:
1. `git checkout client/src/routes/_auth/workouts/new.tsx`
2. Restart dev server
3. Test auth route works
4. Keep extracted components in `client/src/components/workouts/` for future use
**Next Step**: Proceed to Option B for demo route only

---

### If Blocked During Option B

#### Rollback Point: Demo Route Broken
**Trigger**: Demo route errors, form doesn't submit
**Action**:
1. Delete `client/src/routes/demo/workouts/new.tsx`
2. Delete `client/src/routes/demo/workouts/-components/` (if created)
3. Auth route unaffected - no rollback needed
**Next Step**: Debug specific error or postpone demo workout creation

---

### Emergency Rollback (Nuclear Option)

If everything is broken:
```bash
# Discard all changes
git reset --hard HEAD

# Or rollback to specific commit
git log --oneline
git reset --hard <commit-hash>
```

**Prevention**: Commit after each successful step!

```bash
# After Step 1
git add .
git commit -m "Extract child components to shared location"

# After Step 2
git add .
git commit -m "Extract WorkoutTracker to WorkoutForm component"

# After Step 3
git add .
git commit -m "Update auth route to use WorkoutForm component"

# After Step 4
git add .
git commit -m "Create demo workout form route"
```

---

## Success Criteria

### Functional Requirements ✅
- [ ] Auth users can create workouts at `/_auth/workouts/new`
- [ ] Demo users can create workouts at `/demo/workouts/new`
- [ ] Form auto-saves to localStorage (different keys for auth vs demo)
- [ ] Exercise search works (shows demo exercises for demo route)
- [ ] Set entry dialog works (weight, reps, set type)
- [ ] Recent sets display (optional - can be skipped for MVP)
- [ ] Workout submission succeeds and redirects to workout list
- [ ] Created workout appears in respective list (auth or demo)

### Technical Requirements ✅
- [ ] TypeScript compiles with zero errors
- [ ] No `any` types introduced
- [ ] No console errors in browser
- [ ] No breaking changes to existing auth routes
- [ ] Code follows existing patterns and conventions

### Quality Requirements ✅
- [ ] Code is readable and well-commented
- [ ] Component props are clearly documented
- [ ] Commit history is clean with meaningful messages
- [ ] Documentation updated (demo-plan.md, demo-progress.md)

### Performance Requirements ✅
- [ ] Page load time < 2 seconds
- [ ] Form renders without jank
- [ ] Auto-save debounce works (500ms delay)
- [ ] No unnecessary re-renders

---

## Decision Matrix

| Criteria | Option A | Option B | Option C |
|----------|----------|----------|----------|
| **Time to Implement** | 2.5 hours | 50 min | 1.25 hours |
| **Code Duplication** | ~120 lines | ~477 lines | ~270 lines |
| **Risk to Auth Flow** | Medium | Low | Low |
| **Long-term Maintainability** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| **Complexity** | High | Low | Medium |
| **Rollback Difficulty** | Medium | Easy | Easy |
| **DRY Principle** | ✅ Best | ❌ Worst | ⚠️ Okay |
| **Testability** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |

### Recommendation by Scenario

| Scenario | Recommended Option | Reasoning |
|----------|-------------------|-----------|
| **"I want the best long-term solution"** | **Option A** | Clean architecture, no duplication |
| **"I need demo working in 1 hour"** | **Option B** | Fast, low risk |
| **"I want balance of speed and quality"** | **Option C** | Good compromise |
| **"I'm not confident in my TypeScript skills"** | **Option B** | Simpler, less abstraction |
| **"I expect to add more demo features later"** | **Option A** | Pays off over time |
| **"This is a prototype/MVP"** | **Option B** | Ship fast, refactor later |

---

## Next Steps After Implementation

### Phase 2.3 Continuation
After workout form extraction, continue with:
1. ✅ Create demo layout route (`demo.tsx`)
2. ✅ Create demo workout list route (`demo/workouts/index.tsx`)
3. ✅ Create demo workout detail route (`demo/workouts/$workoutId/index.tsx`)
4. ✅ Create demo exercise routes (`demo/exercises/*`)
5. ✅ Wire up navigation between demo routes

### Phase 3: Integration & Polish
1. Add "Try Demo" button on landing page → `/demo/workouts`
2. Add "Reset Demo Data" button in demo routes
3. Add demo mode indicator/banner
4. Add link from demo to sign up for real account

---

## Appendix: File Reference

### Key Files to Review

**Form System**:
- `client/src/hooks/form.ts` - Custom form hook setup
- `client/src/routes/_auth/workouts/-components/form-options.ts` - MOCK_VALUES

**LocalStorage**:
- `client/src/lib/local-storage.ts` - Form state persistence

**Auth Route**:
- `client/src/routes/_auth/workouts/new.tsx` - Current implementation

**API Layer**:
- `client/src/lib/api/workouts.ts` - Workout queries/mutations
- `client/src/lib/api/exercises.ts` - Exercise queries

**Demo Data Layer**:
- `client/src/lib/demo-data/query-options.ts` - Demo query/mutation options
- `client/src/lib/demo-data/storage.ts` - localStorage CRUD operations

**Child Components**:
- `client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx`
- `client/src/routes/_auth/workouts/-components/exercise-screen.tsx`
- `client/src/routes/_auth/workouts/-components/recent-sets-display.tsx`
- `client/src/routes/_auth/workouts/-components/add-set-dialog.tsx`

---

## Questions & Answers

### Q: Can we parameterize the localStorage utilities instead of passing userId?
**A**: The utilities already accept optional `userId` parameter. Passing `userId` prop is simpler than injecting function refs. ✅ Recommended approach.

---

### Q: Should we use dependency injection for the mutation?
**A**: Yes! Passing `saveWorkoutMutation` as a prop is clean dependency injection. Makes testing easier too. ✅

---

### Q: Do we need to extract `MOCK_VALUES`?
**A**: No. `MOCK_VALUES` is a constant with no user dependency. Can stay in `form-options.ts` and be imported by both routes. ✅

---

### Q: What if demo and auth form behavior diverges in the future?
**A**: If significant differences emerge, you can:
1. Add optional props to `WorkoutForm` for conditional behavior
2. Split into `AuthWorkoutForm` and `DemoWorkoutForm` components
3. Keep shared UI components, duplicate business logic

For now, assume they'll stay similar. Don't over-engineer for hypothetical futures. ✅

---

### Q: Should we move `form-options.ts` and `mini-chart.tsx` to shared location?
**A**:
- `form-options.ts`: Keep in route folder - it's specific to workout form
- `mini-chart.tsx`: Move to `@/components/ui/` if used elsewhere, otherwise keep in route

For this extraction, **no need to move these files**. ✅

---

### Q: How do we test that localStorage keys don't collide?
**A**:
```tsx
// Test in browser console
// 1. Go to /_auth/workouts/new (logged in as user abc123)
localStorage.getItem('workout-entry-form-data-abc123'); // Should have data

// 2. Go to /demo/workouts/new
localStorage.getItem('workout-entry-form-data-demo-user'); // Should have different data

// 3. Verify they're isolated
localStorage.getItem('workout-entry-form-data-abc123'); // Auth data unchanged
localStorage.getItem('workout-entry-form-data-demo-user'); // Demo data unchanged
```

---

## Final Recommendation

**Start with Option A** (Extract with Parameterization).

**Why?**
1. You're a developer who values good architecture ✅
2. You're willing to do the hard work ✅
3. Child components are already user-agnostic (low risk) ✅
4. Demo data infrastructure is solid (Phase 1 complete) ✅
5. You have this detailed plan to guide you ✅

**If blocked:**
1. First blocker → Continue debugging (use plan's "Blockers & Solutions")
2. Second blocker → Fall back to Option C (hybrid)
3. Third blocker → Fall back to Option B (duplicate)

**Estimated timeline with Option A**: 2.5 hours of focused work.

You've got this! 🚀

---

**End of Document**
