# Route Merge Proposal: Unified Auth/Demo Routes

**Date**: 2025-10-05
**Status**: Pending Review
**Context**: Currently maintaining duplicate route structures (`/_auth/workouts` vs `/demo/workouts`) causing redirect issues and code duplication

---

## Problem Statement

Demo routes redirect to login because:
1. Shared components have hardcoded links to `/_auth` routes (`/workouts/$workoutId`, `/exercises/$exerciseId`)
2. TanStack Router requires type-safe hardcoded paths (cannot use dynamic base paths)
3. Two separate route trees create maintenance burden and duplication

**Current Redirect Issue**: When clicking a workout in `/demo/workouts`, it navigates to `/workouts/$workoutId` which is under `/_auth`, triggering authentication redirect.

---

## Proposed Solution

**Merge auth and demo routes into unified root routes** with conditional data loading based on authentication state.

### High-Level Approach

1. **Move routes** from `/_auth/workouts` and `/_auth/exercises` to `/workouts` and `/exercises`
2. **Conditional loaders**: Check user authentication in `beforeLoad`
   - **Authenticated** → Load data from API (`workoutsQueryOptions()`)
   - **Unauthenticated** → Load data from localStorage (`getDemoWorkoutsQueryOptions()`)
3. **Single component tree**: Shared components work for both auth and demo modes
4. **Hide auth-only features** in demo mode (edit/delete buttons, user-specific features)

---

## Architecture Changes

### Current Structure
```
routes/
├── _auth.tsx (layout with auth check → redirect if no user)
├── _auth/
│   ├── workouts/
│   │   ├── index.tsx (uses API query options)
│   │   ├── $workoutId/index.tsx
│   │   ├── $workoutId/edit.tsx
│   │   └── new.tsx
│   └── exercises/
│       ├── index.tsx
│       └── $exerciseId.tsx
├── demo.tsx (layout with demo banner)
└── demo/
    ├── workouts/
    │   ├── index.tsx (uses demo query options)
    │   └── $workoutId/index.tsx
    └── exercises/
        ├── index.tsx
        └── $exerciseId.tsx
```

### Proposed Structure
```
routes/
├── workouts/
│   ├── index.tsx (conditional: API or demo query options)
│   ├── $workoutId/index.tsx (conditional: API or demo data)
│   ├── $workoutId/edit.tsx (requires auth → redirect to sign-in)
│   └── new.tsx (conditional: API mutations or demo mutations)
└── exercises/
    ├── index.tsx (conditional: API or demo query options)
    └── $exerciseId.tsx (conditional: API or demo data)
```

---

## Implementation Details

### 1. Route Loader Pattern

**Before** (`_auth/workouts/index.tsx`):
```tsx
export const Route = createFileRoute('/_auth/workouts/')({
  loader: async ({ context }) => {
    const user = context.user; // Always exists due to _auth layout
    context.queryClient.ensureQueryData(workoutsQueryOptions());
    return { user };
  },
  component: RouteComponent,
});
```

**After** (`workouts/index.tsx`):
```tsx
export const Route = createFileRoute('/workouts/')({
  beforeLoad: async ({ context }) => {
    const user = await stackClientApp.getUser(); // No redirect

    if (user) {
      // Authenticated: use API
      context.queryClient.ensureQueryData(workoutsQueryOptions());
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions());
    }

    return { user }; // user is nullable
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();

  // Use conditional query options
  const queryOptions = user
    ? workoutsQueryOptions()
    : getDemoWorkoutsQueryOptions();

  const { data: workouts } = useSuspenseQuery(queryOptions);

  const hasWorkoutInProgress = user
    ? loadFromLocalStorage(user.id) !== null
    : false; // Demo mode: no workout in progress

  return (
    <WorkoutList
      workouts={workouts}
      hasWorkoutInProgress={hasWorkoutInProgress}
      newWorkoutLink="/workouts/new"
    />
  );
}
```

**Feedback**
- I would want the ensureQueryData to be called in not in the `beforeLoad` function as according to tanstack router doc's the `beforeLoad` functions happens in serial. We can move the ensureQueryData to the `loader` function to not block ui as it happens in parallel. 
- It seems like we are storing demo updates into storage as check `lib/demo-data/storage.ts` so the above example `hasWorkoutInProgress` always returns false even though we a demo user might have a workout in progress. You'll have to keep an eye on this as we progress through the PR.
- I made your change in `app.tsx` where user can be used through context now `const user = await stackClientApp.getUser();` so is not needed 

### 2. Header Component

**Unified header** that adapts to auth state:

```tsx
export function Header() {
  const user = useUser({ or: 'return-null' }); // No redirect

  return (
    <header className="flex justify-between p-2 border-b">
      <nav>
        <Link to="/">Home</Link>
        <Link to="/workouts">Workouts</Link>
        <Link to="/exercises">Exercises</Link>
      </nav>

      {/* Show demo banner if not authenticated */}
      {!user && (
        <div className="bg-yellow-100 px-4 py-2">
          Demo Mode: <Link to="/handler/sign-in">Sign In</Link>
        </div>
      )}

      {/* Show user button if authenticated */}
      {user && <CustomUserButton />}
    </header>
  );
}
```

### 3. Edit/Delete Routes

**Auth-required routes** stay protected:

```tsx
// workouts/$workoutId/edit.tsx
export const Route = createFileRoute('/workouts/$workoutId/edit')({
  beforeLoad: async () => {
    const user = await stackClientApp.getUser({ or: 'redirect' }); // Redirects if no user
    return { user };
  },
  // ... rest of route
});
```

**Shared components** hide edit/delete in demo mode:

```tsx
// Already implemented via showEditDelete prop
<WorkoutDetail
  workout={workout}
  showEditDelete={!!user} // Hide in demo mode
/>
```

---

## Complexity & Blockers

### ✅ Low Complexity

1. **Query Options Switching**
   - Auth routes already use `workoutsQueryOptions()` vs `getDemoWorkoutsQueryOptions()`
   - Just need conditional logic in loaders

2. **User Context**
   - Change `context.user` from required to nullable
   - Only 2 routes currently use `context.user`:
     - `workouts/index.tsx` → `loadFromLocalStorage(user.id)`
     - `workouts/new.tsx` → passes user to form

3. **Shared Components**
   - Already support both modes via props (`showEditDelete`, `newWorkoutLink`)
   - No changes needed to components

### ⚠️ Medium Complexity

4. **New Workout Route** (`/workouts/new`)
   - **Current**: 12KB file with complex TanStack Form, requires user for mutations
   - **Options**:
     - A) Support demo mode (use demo mutations to localStorage)
     - B) Require auth (redirect to sign-in if no user)
   - **Recommendation**: Option B (simpler, demo is view-only)

5. **URL Consistency**
   - User logs in while viewing `/workouts` (demo data)
   - Need to invalidate/refresh queries to show real data
   - **Solution**: Listen to auth state changes, invalidate queries on login

### 🔴 Potential Issues

6. **TypeScript Context Types**
   - Root route context needs nullable user: `user?: CurrentUser | CurrentInternalUser | null`
   - Auth-required routes override to require user: `user: CurrentUser | CurrentInternalUser`
   - May need custom route contexts per route

7. **Demo Data Persistence**
   - Currently demo uses `initializeDemoData()` in layout
   - After merge, need to call on every route or in root layout
   - **Solution**: Call in beforeLoad for all routes when no user

---

## Migration Steps

### Phase 1: Preparation
1. Update `__root.tsx` to support nullable user in context
2. Create helper function for conditional query options
3. Test auth state detection (`stackClientApp.getUser()` behavior)

### Phase 2: Move Routes (One at a time)
1. **Start with exercises** (simpler, no edit route)
   - Move `_auth/exercises/index.tsx` → `exercises/index.tsx`
   - Add conditional query options
   - Test both auth and demo modes
   - Move `$exerciseId.tsx`

2. **Then workouts**
   - Move `workouts/index.tsx` with conditional logic
   - Move `$workoutId/index.tsx`
   - Keep `$workoutId/edit.tsx` auth-protected
   - Keep `new.tsx` auth-protected (or implement demo support)

### Phase 3: Cleanup
1. Delete `/demo` routes
2. Delete `/_auth.tsx` layout (auth checks now in individual routes)
3. Delete `demo.tsx` layout (banner now in unified header)
4. Update `DemoHeader` → `Header` with conditional rendering

### Phase 4: Testing
1. Verify unauthenticated users see demo data
2. Verify authenticated users see their data
3. Verify edit/delete hidden in demo mode
4. Verify protected routes redirect to sign-in
5. Verify login → data refresh

---

## Questions for Review

### 1. Route Naming
- **Option A**: `/workouts` (auto-detects auth vs demo) ✅ **Recommended**
- **Option B**: Keep `/demo/workouts` for explicit demo, `/workouts` requires auth

**Feedback**
Use options A

### 2. Edit/Delete Routes
- **Option A**: Hide completely in demo mode ✅ **Recommended**
- **Option B**: Show but redirect to sign-in when clicked
- **Option C**: Keep under `/_auth` requiring authentication

**Feedback**
- Believe that we might have created a lib file in `lib/demo-data/query-options.ts` that handles editing and deleting in demo mode. If so we should not hide the buttons. Do more research.

### 3. New Workout Route
- **Option A**: Support demo mode (allow creating to localStorage)
- **Option B**: Require auth (redirect to sign-in) ✅ **Recommended**
- **Option C**: Separate demo creation route with limited features

**Feedback**
- I believe that we might have created a lib file in `lib/demo-data/query-options.ts` that handles creating in demo mode. If so we should not hide the buttons. Do more research.

### 4. Header Component
- **Option A**: Unified header with conditional user button/demo banner ✅ **Recommended**
- **Option B**: Keep separate headers, conditionally render based on auth

**Feedback**
- I like option A

### 5. Data Refresh on Login
- **Option A**: Auto-refresh queries when user logs in ✅ **Recommended**
- **Option B**: Redirect user to home, require manual navigation
- **Option C**: Show prompt to refresh

**Feedback**
- This is a scenario that I would like to work through. 
- The demo mode contains a placeholder data, so when a user logs while on `/workouts` `exercises` and `exercises/exercisesId` any placeholder data should be refreshed. However what if the user has created something in the demo mode and logs in, should we keep that data? Let's say they were in the process of creating a new workout or exercise and they decide to sign up, I don't think it would be a good user experience to have the data they created in the demo mode to be lost. However, if to enable the newly created data to persist into that user's account, we would need to implement a way to transfer the data from localStorage to the database which is a bit of a challenge. Let me know your thoughts on this.

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Type errors from nullable user | High | Create route-specific contexts, test thoroughly |
| Demo data not initializing | Medium | Call `initializeDemoData()` in beforeLoad for all routes |
| Auth state change not detected | Medium | Add auth state listener, invalidate queries on login |
| Protected routes bypass | High | Test redirect behavior, ensure `or: 'redirect'` works |
| Breaking existing auth routes | High | **Migrate incrementally** (exercises first, then workouts) |

---

## Benefits

1. **No Code Duplication**: Single route tree serves both modes
2. **Type-Safe Links**: Hardcoded TanStack Router paths work correctly
3. **Simpler Maintenance**: One set of routes to maintain
4. **Better UX**: Seamless transition from demo to authenticated
5. **Cleaner Architecture**: Conditional data loading is explicit

---

## Rollback Plan

If migration fails:
1. Keep `/_auth` routes as-is
2. Revert moved routes back to `/_auth`
3. Keep separate `/demo` routes
4. Alternative: Create demo-specific components (more duplication but safer)

---

## Timeline Estimate

- **Preparation**: 30 minutes
- **Exercise routes migration**: 1 hour
- **Workout routes migration**: 1.5 hours
- **Cleanup & testing**: 1 hour
- **Total**: ~4 hours

---

## Recommendation

**Proceed with the merge** using:
- Unified `/workouts` and `/exercises` routes (Option A)
- Hide edit/delete in demo mode (Option A)
- Require auth for new workout route (Option B)
- Unified header with conditional rendering (Option A)
- Auto-refresh on login (Option A)

This provides the cleanest architecture and best UX while maintaining reasonable complexity.

---

**Next Steps**: Review questions, get approval, begin Phase 1 (Preparation)
