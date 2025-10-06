# Demo Routes Implementation Plan

## Overview
Convert authenticated routes (`/_auth/*`) to demo routes (`/demo/*`) with mock data to allow users to test the app without authentication.

## Implementation Strategy: **Easy + Maintainable**

---

## Phase 1: Data Layer Setup

### 1.1 Type Safety & Data Structures
- [ ] Verify generated types from `@/client` match schema:
  - [ ] `WorkoutWorkoutWithSetsResponse` structure
  - [ ] `ExerciseExerciseResponse` structure
  - [ ] User type: `{ id: number, user_id: string, created_at: string }`
- [ ] Document data relationships:
  - [ ] Workout → Sets (one-to-many via `workout_id`)
  - [ ] Exercise → Sets (one-to-many via `exercise_id`)
  - [ ] Set fields: `exercise_order`, `set_order`, `weight`, `reps`, `set_type`

### 1.2 Mock Data Files (`client/src/lib/demo-data/`)
- [ ] `types.ts` - Import and re-export relevant types from `@/client`
- [ ] `initial-data.ts` - Initial seed data matching real types:
  - [ ] 5 exercises (Bench Press, Squat, Deadlift, Pull-ups, Overhead Press)
  - [ ] 3 workouts (Push Day, Pull Day, Leg Day) with realistic dates
  - [ ] 15-20 sets with proper relationships and ordering
  - [ ] Realistic progression data (weight/reps)
- [ ] `storage.ts` - localStorage utilities:
  - [ ] `DEMO_STORAGE_KEYS` constants
  - [ ] `getDemoWorkouts()` - Load from localStorage or return initial data
  - [ ] `saveDemoWorkouts()` - Persist to localStorage
  - [ ] `getDemoExercises()` - Load from localStorage or return initial data
  - [ ] `saveDemoExercises()` - Persist to localStorage
  - [ ] `getDemoSets()` - Load from localStorage or return initial data
  - [ ] `saveDemoSets()` - Persist to localStorage
  - [ ] `resetDemoData()` - Clear all demo data and restore initial state
  - [ ] `getDemoWorkoutWithSets(id)` - Join workouts + sets + exercises

### 1.3 Demo Query Options (`client/src/lib/demo-data/query-options.ts`)
- [ ] Mirror structure from `@/client/@tanstack/react-query.gen`:
  - [ ] `demoExercisesQueryOptions()` - Return exercises from localStorage
  - [ ] `demoExerciseByIdQueryOptions(id)` - Return single exercise
  - [ ] `demoRecentExerciseSetsQueryOptions(id)` - Return recent sets for exercise
  - [ ] `demoWorkoutsQueryOptions()` - Return workouts sorted by date DESC
  - [ ] `demoWorkoutByIdQueryOptions(id)` - Return workout with joined sets/exercises
  - [ ] `demoWorkoutsFocusValuesQueryOptions()` - Return distinct workout_focus values
- [ ] Demo mutations (update localStorage + invalidate queries):
  - [ ] `useDemoSaveWorkoutMutation()` - Create new workout
  - [ ] `useDemoUpdateWorkoutMutation()` - Update existing workout
  - [ ] `useDemoDeleteWorkoutMutation()` - Delete workout
  - [ ] `useDemoDeleteExerciseMutation()` - Delete exercise

---

## Phase 2: Route Implementation

### 2.1 Identify Auth Routes to Copy
- [ ] List all routes in `client/src/routes/_auth/`:
  - [ ] `workouts/index.tsx` - Workout list
  - [ ] `workouts/new.tsx` - Create workout form
  - [ ] `workouts/$id.tsx` - View/edit workout (if exists)
  - [ ] `exercises/index.tsx` - Exercise list
  - [ ] Other routes?
- [ ] Check for shared components/layouts used by auth routes

### 2.2 Create Demo Routes (`client/src/routes/demo/`)
- [ ] `demo/workouts/index.tsx` - Copy from `_auth/workouts/index.tsx`:
  - [ ] Replace query imports with demo query options
  - [ ] Replace mutation imports with demo mutations
  - [ ] Test: List workouts, delete workout
- [ ] `demo/workouts/new.tsx` - Copy from `_auth/workouts/new.tsx`:
  - [ ] Replace query/mutation imports
  - [ ] Test: Create new workout with exercises/sets
- [ ] `demo/workouts/$id.tsx` - Copy from `_auth/workouts/$id.tsx` (if exists):
  - [ ] Replace query/mutation imports
  - [ ] Test: View and edit workout
- [ ] `demo/exercises/index.tsx` - Copy from `_auth/exercises/index.tsx`:
  - [ ] Replace query/mutation imports
  - [ ] Test: List exercises, delete exercise

### 2.3 Demo Route Layout/Index
- [ ] `demo/index.tsx` - Demo landing/dashboard (if needed)
- [ ] Handle user context (if routes access user data):
  - [ ] Check if `useUser()` hook is used
  - [ ] Create demo user provider or mock context if needed

---

## Phase 3: Integration & Polish

### 3.1 Landing Page Integration
- [ ] Update landing page "Try for free" button:
  - [ ] Link to `/demo/workouts` or `/demo` route
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

### New Files
```
client/src/lib/demo-data/
  ├── types.ts                    # Type imports/exports
  ├── initial-data.ts             # Seed data
  ├── storage.ts                  # localStorage utilities
  └── query-options.ts            # Demo query/mutation options

client/src/routes/demo/
  ├── index.tsx                   # Demo landing (optional)
  ├── workouts/
  │   ├── index.tsx              # Workout list
  │   ├── new.tsx                # Create workout
  │   └── $id.tsx                # Edit workout (if exists)
  └── exercises/
      └── index.tsx              # Exercise list
```

### Modified Files
- [ ] Landing page - Add "Try for free" link to `/demo`

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