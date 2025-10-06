# Demo Routes Implementation Plan - REVISED

## Overview
~~Convert authenticated routes (`/_auth/*`) to demo routes (`/demo/*`) with mock data~~

**ACTUAL IMPLEMENTATION**: Merged auth and demo routes into unified root routes (`/workouts`, `/exercises`) with conditional data loading based on authentication state.

## Implementation Strategy: **Route Merge (Unified Routes)**

See `route-merge-proposal.md` for detailed rationale and architecture.

---

## Phase 1: Data Layer Setup ✅ COMPLETE

### 1.1 Type Safety & Data Structures ✅ COMPLETE
- [x] Verify generated types from `@/client` match schema
- [x] Document data relationships
- [x] **Created**: `client/docs/phase-1-1-type-verification.md`

### 1.2 Mock Data Files (`client/src/lib/demo-data/`) ✅ COMPLETE
- [x] `types.ts` - Import and re-export relevant types from `@/client`
- [x] `initial-data.ts` - Seed data (5 exercises, 3 workouts, 26 sets)
- [x] `storage.ts` - localStorage utilities (390 lines)
  - [x] Full CRUD operations for workouts and exercises
  - [x] `clearDemoData()` - NEW: Clear demo data when user authenticates
- [x] `query-options.ts` - Demo query/mutation options (237 lines)

### 1.3 Demo Query Options ✅ COMPLETE
- [x] All query options mirroring API structure
- [x] All mutations (create, update, delete) for both workouts and exercises

---

## Phase 2: Route Merge Implementation ✅ COMPLETE

### 2.1 Extract Shared Components ✅ COMPLETE
- [x] `components/workouts/workout-list.tsx`
- [x] `components/workouts/workout-detail.tsx`
- [x] `components/exercises/exercise-list.tsx`
- [x] `components/exercises/exercise-detail.tsx`

### 2.2 Create Unified Routes ✅ COMPLETE

**Strategy**: Single route tree serves both authenticated and demo users via conditional loaders.

#### Completed Routes:
- [x] `/exercises/` - Conditional: API queries vs demo queries
- [x] `/exercises/$exerciseId` - Conditional: API vs demo data
- [x] `/workouts/` - Conditional: API queries vs demo queries
- [x] `/workouts/$workoutId/` - Conditional: API vs demo data
  - Edit/delete buttons **shown in both modes** (demo has mutations)
- [x] `/workouts/$workoutId/edit` - Auth-only (complex, uses API-specific hooks)
- [x] `/workouts/new` - **PARTIALLY COMPLETE** (see below)

#### `/workouts/new` Status: ⚠️ BLOCKED

**Current State:**
- Route created at `/workouts/new` (not auth-only)
- User type changed to nullable: `CurrentUser | CurrentInternalUser | null`
- Uses API-only mutation: `useSaveWorkoutMutation()`
- Uses `user.id` for localStorage (breaks when user is null)

**Remaining Work:**
1. Add conditional mutation (API vs demo)
2. Handle `user?.id` for localStorage operations (use `undefined` for demo users)
3. Update `WorkoutTracker` to use conditional save mutation
4. Update all `user.id` references to `user?.id`

**Files Involved:**
- `routes/workouts/new.tsx` (lines 42, 45, 49, 64, 96)
- Need to conditionally use `useSaveWorkoutMutation()` vs `useMutation(postDemoWorkoutsMutation())`

### 2.3 Cleanup ✅ COMPLETE
- [x] Deleted `/demo/*` routes (never fully implemented)
- [x] Deleted `/_auth/*` routes (migrated to root)
- [x] Moved `-components` folders to new route locations
- [x] Updated import paths in shared components

---

## Phase 3: Type Safety & Integration ⏳ IN PROGRESS

### 3.1 TypeScript Compilation ⏳ PARTIALLY COMPLETE
- [x] Fixed query option type conflicts (used conditional `useSuspenseQuery` calls)
- [x] Fixed main.tsx context (user set to `null` instead of `undefined!`)
- [x] TanStack Router routes regenerated
- [ ] **REMAINING**: Fix `/workouts/new` TypeScript errors (nullable user issues)

### 3.2 Data Cleanup on Login ✅ IMPLEMENTED
- [x] `clearDemoData()` function created in `storage.ts`
- [x] Called in route loaders when `user` exists (clears demo localStorage)
- [ ] Test: Verify demo data cleared when user logs in

### 3.3 Testing Checklist ⏳ TODO
- [ ] **Demo Mode (Unauthenticated)**:
  - [ ] View workouts list with demo data
  - [ ] View workout detail
  - [ ] Delete workout (demo mutation)
  - [ ] View exercises list
  - [ ] View exercise detail
  - [ ] Verify edit/delete buttons visible
  - [ ] ~~Create new workout~~ (blocked until `/workouts/new` completed)
- [ ] **Authenticated Mode**:
  - [ ] View workouts list with API data
  - [ ] Create new workout
  - [ ] Edit existing workout
  - [ ] Delete workout
  - [ ] Verify demo data cleared on login
- [ ] **Data Persistence**:
  - [ ] Demo data persists across page reloads
  - [ ] Demo data cleared when user authenticates

---

## Current File Structure

### Routes (Unified - No Separation)
```
client/src/routes/
├── __root.tsx                       # Updated: user is nullable
├── exercises/
│   ├── index.tsx                   # ✅ Conditional: API or demo queries
│   ├── $exerciseId.tsx             # ✅ Conditional: API or demo data
│   └── -components/
│       └── exercise-delete-dialog.tsx
└── workouts/
    ├── index.tsx                   # ✅ Conditional: API or demo queries
    ├── $workoutId/
    │   ├── index.tsx              # ✅ Conditional: API or demo data
    │   └── edit.tsx               # ✅ Auth-only (complex)
    ├── new.tsx                    # ⚠️ BLOCKED: Needs conditional mutations
    └── -components/               # Shared components for workout forms
        ├── add-exercise-screen.tsx
        ├── exercise-screen.tsx
        ├── form-options.ts
        ├── mini-chart.tsx
        └── recent-sets-display.tsx
```

### Shared Components
```
client/src/components/
├── workouts/
│   ├── workout-list.tsx           # ✅ Used by both modes
│   └── workout-detail.tsx         # ✅ Shows edit/delete in both modes
└── exercises/
    ├── exercise-list.tsx          # ✅ Used by both modes
    └── exercise-detail.tsx        # ✅ Used by both modes
```

### Demo Data Infrastructure
```
client/src/lib/demo-data/
├── types.ts                       # ✅ Type exports
├── initial-data.ts                # ✅ Seed data
├── storage.ts                     # ✅ localStorage CRUD + clearDemoData()
└── query-options.ts               # ✅ Query/mutation options
```

---

## Benefits of Route Merge Approach

✅ **No Code Duplication**: Single route tree, shared components
✅ **Type-Safe Links**: Hardcoded TanStack Router paths work correctly
✅ **Simpler Maintenance**: One set of routes to update
✅ **Better UX**: Seamless transition from demo to authenticated
✅ **Full CRUD in Demo**: Users can create/edit/delete in demo mode

---

## Blockers

### 🔴 Critical: `/workouts/new` Route
**Issue**: Route uses API-only mutation and assumes user exists.

**Solution Required**:
1. Conditional mutation based on `user` existence
2. Handle `user?.id` for localStorage operations
3. Pass correct mutation to `WorkoutTracker` component

**Estimated Time**: 1-2 hours

---

## Next Steps for Next Agent

1. **Fix `/workouts/new` route** (see BLOCKED section above)
2. **Run `bun run tsc`** to verify all TypeScript errors resolved
3. **Manual testing** of both demo and auth modes
4. **Update landing page** to link to `/workouts` (auto-detects mode)
5. **Optional**: Create unified header with demo banner

See `NEXT-STEPS.md` for detailed handoff.
