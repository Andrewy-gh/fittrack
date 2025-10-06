# Next Steps for Demo Routes Implementation

**📅 Last Updated**: 2025-10-06 (Session 9 - Route Merge Complete)
**👤 Next Agent**: Start here for quick orientation

---

## 🎯 Current Status: 90% Complete - One Route Blocked

### ✅ What's Working

The **route merge approach** is implemented and mostly functional:

1. **Unified Routes** (`/workouts`, `/exercises`) serve both authenticated and demo users
2. **Demo data infrastructure** complete (localStorage CRUD + query options)
3. **Conditional loaders** implemented in all view routes
4. **Shared components** extracted and working
5. **Type safety** mostly resolved (query option conflicts fixed)

### 🔴 What's Broken

**One critical blocker**: `/workouts/new` route has TypeScript errors due to nullable `user`.

---

## 🚨 Priority 1: Fix `/workouts/new` Route

### Problem

The route was updated to support demo mode (user can be `null`), but the `WorkoutTracker` component still assumes user exists:

**TypeScript Errors (from `bun run tsc`):**
```
Line 45: user.id - user is possibly null
Line 49: user.id - user is possibly null
Line 64: user.id - user is possibly null
Line 96: user.id - user is possibly null
Line 42: useSaveWorkoutMutation() - API-only, needs conditional mutation
```

### Solution

Update `routes/workouts/new.tsx` to handle nullable user:

1. **Change `user.id` to `user?.id`** in 4 places:
   - Line 45: `getInitialValues(user?.id)`
   - Line 49: `saveToLocalStorage(formApi.state.values, user?.id)`
   - Line 64: `clearLocalStorage(user?.id)`
   - Line 96: `clearLocalStorage(user?.id)`

2. **Add conditional mutation** (line 42):
   ```tsx
   // Current (API-only):
   const saveWorkout = useSaveWorkoutMutation();

   // Should be:
   const saveWorkout = user
     ? useSaveWorkoutMutation()
     : useMutation(postDemoWorkoutsMutation());
   ```

3. **Add imports** at top of file:
   ```tsx
   import { useMutation } from '@tanstack/react-query';
   import { postDemoWorkoutsMutation } from '@/lib/demo-data/query-options';
   ```

4. **Update loader** to load demo data when not authenticated:
   ```tsx
   loader: async ({ context }) => {
     const user = context.user;
     if (user) {
       // Authenticated
       await context.queryClient.ensureQueryData(exercisesQueryOptions());
       await context.queryClient.ensureQueryData(workoutsFocusValuesQueryOptions());
     } else {
       // Demo mode
       initializeDemoData();
       await context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
       await context.queryClient.ensureQueryData(getDemoWorkoutsFocusValuesQueryOptions());
     }
   },
   ```

5. **Update RouteComponent** to use conditional queries:
   ```tsx
   function RouteComponent() {
     const { user } = Route.useRouteContext();

     const { data: exercisesResponse } = user
       ? useSuspenseQuery(exercisesQueryOptions())
       : useSuspenseQuery(getDemoExercisesQueryOptions());

     const { data: workoutsFocusValues } = user
       ? useSuspenseQuery(workoutsFocusValuesQueryOptions())
       : useSuspenseQuery(getDemoWorkoutsFocusValuesQueryOptions());

     // ... rest of component
   }
   ```

6. **Verify with**: `bun run tsc` (should have 0 errors)

**Estimated Time**: 30-45 minutes

---

## 📋 Completed Work (Session 9)

### Route Merge Implementation

**What We Did**: Merged `/_auth` and `/demo` routes into unified root routes.

**Why**:
- Avoids code duplication
- Type-safe hardcoded links work correctly
- Simpler maintenance (one route tree)
- Seamless UX transition from demo → authenticated

**Routes Created:**
- ✅ `/exercises/` - List exercises (conditional: API or demo)
- ✅ `/exercises/$exerciseId` - Exercise detail (conditional)
- ✅ `/workouts/` - List workouts (conditional)
- ✅ `/workouts/$workoutId/` - Workout detail (conditional, shows edit/delete in both modes)
- ✅ `/workouts/$workoutId/edit` - Edit workout (auth-only, complex)
- ⚠️ `/workouts/new` - Create workout (BLOCKED - see above)

**Deleted:**
- ❌ `/demo/*` routes (never fully implemented)
- ❌ `/_auth/*` routes (migrated to root)
- ❌ `/_auth.tsx` layout (no longer needed)
- ❌ `demo.tsx` layout (no longer needed)

### Type Safety Fixes

1. **Query option type conflicts**: Fixed by using conditional `useSuspenseQuery` calls instead of conditional query options
   ```tsx
   // ❌ Before (type error):
   const queryOptions = user ? apiOptions() : demoOptions();
   const { data } = useSuspenseQuery(queryOptions);

   // ✅ After (works):
   const { data } = user
     ? useSuspenseQuery(apiOptions())
     : useSuspenseQuery(demoOptions());
   ```

2. **Root context**: Updated `__root.tsx` to make user nullable: `user: CurrentUser | CurrentInternalUser | null`

3. **main.tsx**: Changed initial context from `user: undefined!` to `user: null`

4. **Component imports**: Fixed paths after moving `-components` folders

---

## 🗂️ File Structure (Post-Merge)

### Current Routes
```
client/src/routes/
├── __root.tsx                       # ✅ User nullable
├── exercises/
│   ├── index.tsx                   # ✅ Conditional queries
│   ├── $exerciseId.tsx             # ✅ Conditional queries
│   └── -components/
│       └── exercise-delete-dialog.tsx
└── workouts/
    ├── index.tsx                   # ✅ Conditional queries
    ├── $workoutId/
    │   ├── index.tsx              # ✅ Conditional queries
    │   └── edit.tsx               # ✅ Auth-only
    ├── new.tsx                    # ⚠️ BLOCKED
    └── -components/
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
│   ├── workout-list.tsx           # ✅ Accepts optional props
│   └── workout-detail.tsx         # ✅ showEditDelete prop
└── exercises/
    ├── exercise-list.tsx
    └── exercise-detail.tsx
```

### Demo Data (Complete)
```
client/src/lib/demo-data/
├── types.ts                       # ✅ Type exports
├── initial-data.ts                # ✅ 5 exercises, 3 workouts, 26 sets
├── storage.ts                     # ✅ CRUD + clearDemoData()
└── query-options.ts               # ✅ All queries + mutations
```

---

## 📚 Key Documentation

### Primary Docs (Read First)
1. **`route-merge-proposal.md`** - Architecture decisions and rationale
2. **`demo-plan.md`** - Revised implementation plan (matches what was built)
3. **THIS FILE** - Immediate next steps

### Background Context
- `demo-progress.md` - Session logs (600+ lines, historical reference)
- `workout-form-extraction-plan.md` - Not needed (didn't extract form)
- `phase-1-1-type-verification.md` - Type analysis (reference)

---

## ✅ Testing Checklist (After Fix)

### Demo Mode (Unauthenticated)
- [ ] Navigate to `/workouts` → See demo data
- [ ] Click a workout → See workout detail
- [ ] Click "Edit" or "Delete" → Mutations work
- [ ] Navigate to `/exercises` → See demo data
- [ ] Click an exercise → See exercise detail
- [ ] Navigate to `/workouts/new` → Can create workout
- [ ] Reload page → Demo data persists (localStorage)

### Authenticated Mode
- [ ] Login → Navigate to `/workouts` → See API data
- [ ] Demo localStorage cleared (verify in DevTools)
- [ ] Create new workout → Saves to API
- [ ] Edit workout → Updates API
- [ ] Delete workout → Deletes from API

### Edge Cases
- [ ] TypeScript compilation: `bun run tsc` (0 errors)
- [ ] Demo → Login transition (data refresh)
- [ ] Login → Logout → Demo data still available

---

## 🔍 Common Issues & Solutions

### TypeScript Errors

**Issue**: `Property 'id' does not exist on type ... | null`
**Fix**: Use optional chaining: `user?.id` instead of `user.id`

**Issue**: Query option type incompatibility
**Fix**: Use conditional `useSuspenseQuery` calls (see Type Safety Fixes section)

### Route Not Found

**Issue**: TanStack Router doesn't recognize new routes
**Fix**: Run `bun run generate-routes` (you likely already did this)

### Demo Data Not Loading

**Issue**: `initializeDemoData()` not called
**Fix**: Add to route loader when `!user`

---

## 🎯 After `/workouts/new` is Fixed

### Optional Enhancements

1. **Unified Header Component**
   - Show demo banner when not authenticated
   - Link to sign-in from banner
   - File: `components/header.tsx`

2. **Landing Page Integration**
   - Update "Try for free" button → `/workouts`
   - Auto-detects demo mode if not authenticated

3. **Edit Route Demo Support**
   - Currently auth-only (`/workouts/$workoutId/edit`)
   - Could add demo support similar to `/workouts/new`
   - Lower priority (users can delete/recreate in demo)

---

## 📝 Summary for Next Agent

**What's Done:**
- ✅ Demo data infrastructure (localStorage + queries/mutations)
- ✅ Shared components extracted
- ✅ 5 out of 6 routes fully working with conditional loading
- ✅ Type safety mostly resolved

**What's Blocked:**
- 🔴 `/workouts/new` - TypeScript errors (nullable user)

**Your Mission:**
1. Fix `/workouts/new` (see Priority 1 section)
2. Run `bun run tsc` → 0 errors
3. Manual test both modes
4. Update this file when complete

**Estimated Time**: 1-2 hours total

---

**End of Handoff**

Good luck! The hard part (architecture decision + route merge) is done. This is just cleanup. 🚀
