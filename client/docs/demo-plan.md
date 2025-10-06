# Demo Routes Implementation Plan

## Overview
Convert authenticated routes (`/_auth/*`) to demo routes (`/demo/*`) with mock data to allow users to test the app without authentication.

## Implementation Strategy: **Easy + Maintainable**

---

## Phase 1: Data Layer Setup

### 1.1 Type Safety & Data Structures ✅ COMPLETE
- [x] Verify generated types from `@/client` match schema:
  - [x] `WorkoutWorkoutWithSetsResponse` structure - Verified (flattened/joined structure)
  - [x] `ExerciseExerciseResponse` structure - Verified (simple entity with metadata)
  - [x] User type analysis - No explicit User type; entities include `user_id: string`
- [x] Document data relationships:
  - [x] Workout → Sets (one-to-many via `workout_id`)
  - [x] Exercise → Sets (one-to-many via `exercise_id`)
  - [x] Set fields: `exercise_order`, `set_order`, `weight`, `reps`, `set_type`
- [x] **Created**: `client/docs/phase-1-1-type-verification.md` (230 lines - full verification report)

### 1.2 Mock Data Files (`client/src/lib/demo-data/`) ✅ COMPLETE
- [x] `types.ts` - Import and re-export relevant types from `@/client`
- [x] `initial-data.ts` - Initial seed data matching real types:
  - [x] 5 exercises (Barbell Squat, Bench Press, Deadlift, Overhead Press, Pull-ups)
  - [x] 3 workouts with realistic dates (7 days ago, 5 days ago, 2 days ago)
  - [x] 26 sets with proper relationships and ordering
  - [x] Realistic progression data (weight/reps showing progression)
- [x] `storage.ts` - localStorage utilities:
  - [x] `STORAGE_KEYS` constants
  - [x] `getAllWorkouts()` / `getWorkoutById()` - Load from localStorage
  - [x] `createWorkout()` / `updateWorkout()` / `deleteWorkout()` - CRUD operations
  - [x] `getAllExercises()` / `getExerciseById()` - Load from localStorage
  - [x] `createExercise()` / `deleteExercise()` - CRUD operations
  - [x] `getExerciseWithSets()` / `getExerciseRecentSets()` - Join operations
  - [x] `resetDemoData()` - Clear all demo data and restore initial state
  - [x] `initializeDemoData()` - Initialize on first load
  - [x] `getWorkoutFocusValues()` - Get distinct focus values

### 1.3 Demo Query Options (`client/src/lib/demo-data/query-options.ts`) ✅ COMPLETE
- [x] Mirror structure from `@/client/@tanstack/react-query.gen`:
  - [x] `getDemoExercisesQueryOptions()` - Return exercises from localStorage
  - [x] `getDemoExercisesByIdQueryOptions(id)` - Return exercise with sets
  - [x] `getDemoExercisesByIdRecentSetsQueryOptions(id)` - Return recent sets for exercise
  - [x] `getDemoWorkoutsQueryOptions()` - Return workouts sorted by date DESC
  - [x] `getDemoWorkoutsByIdQueryOptions(id)` - Return workout with joined sets/exercises
  - [x] `getDemoWorkoutsFocusValuesQueryOptions()` - Return distinct workout_focus values
- [x] Demo mutations (update localStorage + invalidate queries):
  - [x] `postDemoWorkoutsMutation()` - Create new workout
  - [x] `putDemoWorkoutsByIdMutation()` - Update existing workout
  - [x] `deleteDemoWorkoutsByIdMutation()` - Delete workout
  - [x] `deleteDemoExercisesByIdMutation()` - Delete exercise
  - [x] `postDemoExercisesMutation()` - Create exercise

---

## Phase 2: Route Implementation

### 2.0 Component Reuse Strategy Analysis ✅ COMPLETE

**Key Finding**: User dependency is minimal - only used for localStorage scoping in 3 route files, zero usage in child components.

**Decision**: Extract shared display components instead of duplicating routes.

**Rationale**:
- All UI components (`WorkoutsDisplay`, `ExerciseDisplay`, etc.) are already user-agnostic
- User only used for: `user.id` in localStorage keys and auth validation in loaders
- Extracting components maintains DRY principle without conditional complexity
- Future UI updates only require one change instead of two

**Strategy**: Component Extraction Pattern (Option A)
1. Extract display components to `client/src/components/`
2. Both `/_auth/*` and `/demo/*` routes import same components
3. Routes handle data fetching (different query options), components handle display

### 2.1 Extract Shared Components (`client/src/components/`) ✅ COMPLETE
- [x] Create `client/src/components/workouts/` directory
- [x] Extract `WorkoutsDisplay` → `components/workouts/workout-list.tsx`
  - [x] Move from `_auth/workouts/index.tsx` (lines 13-142)
  - [x] Export as standalone component
  - [x] Remove user prop (not needed by component)
  - [x] Added optional `hasWorkoutInProgress` and `newWorkoutLink` props for flexibility
- [x] Extract `IndividualWorkoutPage` → `components/workouts/workout-detail.tsx`
  - [x] Move from `_auth/workouts/$workoutId/index.tsx` (lines 14-224)
  - [x] Export as standalone component
- [ ] Extract `WorkoutTracker` → `components/workouts/workout-form.tsx` (if needed for new workout flow)
  - [ ] Analyze `_auth/workouts/new-2.tsx` dependencies
  - [ ] May need to keep in route file due to complex state/localStorage integration
- [x] Create `client/src/components/exercises/` directory
- [x] Extract `ExercisesDisplay` → `components/exercises/exercise-list.tsx`
  - [x] Move from `_auth/exercises/index.tsx` (lines 11-80)
  - [x] Export as standalone component
- [x] Extract `ExerciseDisplay` → `components/exercises/exercise-detail.tsx`
  - [x] Move from `_auth/exercises/$exerciseId.tsx` (lines 25-254)
  - [x] Export as standalone component

### 2.2 Update Auth Routes to Use Extracted Components ✅ COMPLETE
- [x] Update `_auth/workouts/index.tsx`:
  - [x] Import `WorkoutList` from `@/components/workouts/workout-list`
  - [x] Update `RouteComponent` to render imported component
  - [x] Keep loader and query logic unchanged
  - [x] User prop handled in route file for localStorage check
- [x] Update `_auth/workouts/$workoutId/index.tsx`:
  - [x] Import `WorkoutDetail` from `@/components/workouts/workout-detail`
  - [x] Update `RouteComponent` to render imported component
  - [x] Keep loader and query logic unchanged
- [x] Update `_auth/exercises/index.tsx`:
  - [x] Import `ExerciseList` from `@/components/exercises/exercise-list`
  - [x] Update `RouteComponent` to render imported component
- [x] Update `_auth/exercises/$exerciseId.tsx`:
  - [x] Import `ExerciseDetail` from `@/components/exercises/exercise-detail`
  - [x] Update `RouteComponent` to render imported component

### 2.3 Create Demo Routes (`client/src/routes/demo/`)
- [ ] `demo.tsx` - Demo layout route (pathless layout with Header + initialize demo data)
- [ ] `demo/workouts/index.tsx`:
  - [ ] Import `WorkoutList` from `@/components/workouts/workout-list`
  - [ ] Use `getDemoWorkoutsQueryOptions()` in loader
  - [ ] Render same component as auth route
- [ ] `demo/workouts/$workoutId/index.tsx`:
  - [ ] Import `WorkoutDetail` from `@/components/workouts/workout-detail`
  - [ ] Use `getDemoWorkoutsByIdQueryOptions(id)` in loader
  - [ ] Render same component as auth route
- [ ] `demo/workouts/new-2.tsx` (or similar):
  - [ ] Analyze if `WorkoutTracker` can be extracted or needs route-specific version
  - [ ] Use demo query/mutation options
  - [ ] Use `'demo-user'` for localStorage operations
- [ ] `demo/exercises/index.tsx`:
  - [ ] Import `ExerciseList` from `@/components/exercises/exercise-list`
  - [ ] Use `getDemoExercisesQueryOptions()` in loader
  - [ ] Render same component as auth route
- [ ] `demo/exercises/$exerciseId.tsx`:
  - [ ] Import `ExerciseDetail` from `@/components/exercises/exercise-detail`
  - [ ] Use `getDemoExercisesByIdQueryOptions(id)` in loader
  - [ ] Render same component as auth route

---

## Phase 3: Integration & Polish

### 3.1 Landing Page Integration
- [ ] Update landing page "Try for free" button:
  - [ ] Link to `/demo/workouts/new`
  - [ ] Ensure navigation works correctly

### 3.2 Demo UX Enhancements
- [ ] Add "Reset Demo" button/option:
  - [ ] Placement: Demo settings, or persistent banner
  - [ ] Calls `resetDemoData()` and invalidates all queries
- [ ] Add demo indicator:
  - [ ] Banner/badge showing "Demo Mode" (optional)
  - [ ] Link to sign up for real account (optional)

### 3.3 Testing Checklist
- [ ] **Workouts**:
  - [ ] Create new workout with multiple exercises
  - [ ] View workout list (sorted by date)
  - [ ] Edit existing workout
  - [ ] Delete workout
  - [ ] Verify localStorage persistence across page reloads
- [ ] **Exercises**:
  - [ ] View exercise list
  - [ ] Create exercise (via workout form)
  - [ ] Delete exercise
  - [ ] Verify exercise relationships with sets
- [ ] **Data Integrity**:
  - [ ] Sets properly linked to workouts and exercises
  - [ ] Exercise/set ordering preserved
  - [ ] Date sorting works correctly
- [ ] **Reset Flow**:
  - [ ] Reset demo data
  - [ ] Verify initial data restored
  - [ ] Confirm localStorage cleared

### 3.4 Type Safety Verification
- [ ] Run TypeScript compiler: `npm run typecheck` (or equivalent)
- [ ] Ensure no type errors in demo files
- [ ] Verify mock data matches generated types from `@/client`

---

## File Structure Summary

### New Files Created (Phase 1 & 2.1/2.2)
```
client/src/lib/demo-data/                          ✅ COMPLETE
  ├── types.ts                    # Type imports/exports
  ├── initial-data.ts             # Seed data (5 exercises, 3 workouts, 26 sets)
  ├── storage.ts                  # localStorage utilities (390 lines)
  └── query-options.ts            # Demo query/mutation options (237 lines)

client/src/components/                             ✅ COMPLETE (Phase 2.1)
  ├── workouts/
  │   ├── workout-list.tsx       # Extracted from _auth/workouts/index.tsx
  │   └── workout-detail.tsx     # Extracted from _auth/workouts/$workoutId/index.tsx
  └── exercises/
      ├── exercise-list.tsx      # Extracted from _auth/exercises/index.tsx
      └── exercise-detail.tsx    # Extracted from _auth/exercises/$exerciseId.tsx

client/src/routes/demo/                            ⏳ TODO (Phase 2.3)
  ├── index.tsx                   # Demo landing (optional)
  ├── workouts/
  │   ├── index.tsx              # Workout list
  │   ├── new.tsx                # Create workout
  │   └── $id.tsx                # Edit workout (if exists)
  └── exercises/
      └── index.tsx              # Exercise list
```

### Modified Files
- [x] `_auth/workouts/index.tsx` - Now uses `<WorkoutList>` component (Phase 2.2)
- [x] `_auth/workouts/$workoutId/index.tsx` - Now uses `<WorkoutDetail>` component (Phase 2.2)
- [x] `_auth/exercises/index.tsx` - Now uses `<ExerciseList>` component (Phase 2.2)
- [x] `_auth/exercises/$exerciseId.tsx` - Now uses `<ExerciseDetail>` component (Phase 2.2)
- [ ] Landing page - Add "Try for free" link to `/demo` (Phase 3.1)

---

## Data Schema Reference

### Database Structure (from schema.sql)
```sql
users (id, user_id, created_at)
workout (id, date, notes, workout_focus, user_id)
exercise (id, name, user_id)
set (id, exercise_id, workout_id, weight, reps, set_type,
     exercise_order, set_order, user_id)
```

### localStorage Keys
```typescript
'fittrack-demo-workouts'  // Workout[]
'fittrack-demo-exercises' // Exercise[]
'fittrack-demo-sets'      // Set[]
```

### Data Relationships
- Workout (1) → Sets (N) via `set.workout_id`
- Exercise (1) → Sets (N) via `set.exercise_id`
- Sets link workouts and exercises (junction table pattern)

---

## Benefits

✅ **Type Safe**: Uses generated types, survives OpenAPI regeneration
✅ **Persistent**: localStorage maintains demo state across sessions
✅ **Resettable**: Users can restore initial demo data anytime
✅ **Maintainable**: Mock data separated from component logic
✅ **No Breaking Changes**: Auth routes completely untouched
✅ **Realistic**: Full CRUD operations with proper data relationships

---

## Timeline Estimate
- **Phase 1 (Data Layer)**: 2-3 hours
- **Phase 2 (Routes)**: 2-3 hours
- **Phase 3 (Integration & Testing)**: 2-3 hours
- **Total**: 6-9 hours

This approach ensures careful implementation with comprehensive type safety and testing.