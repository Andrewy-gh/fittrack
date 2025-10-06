# Workout Form Component Extraction Plan v2 (SIMPLIFIED)

## Document Purpose
After reviewing TanStack Form documentation, we discovered that the extraction is **significantly simpler** than initially estimated. The child components are already using `withForm` HOC correctly, which means they're fully reusable.

**Created**: 2025-10-05 (Updated after TanStack Form docs review)
**Status**: Planning Phase - SIMPLIFIED APPROACH
**Approach**: Use `withForm` HOC (TanStack Form's built-in composition pattern)

---

## 🎯 Key Insights from TanStack Form Docs

### 1. `withForm` HOC Pattern (Form Composition Guide)

TanStack Form provides `withForm` specifically for breaking large forms into smaller, reusable pieces:

```tsx
const ChildForm = withForm({
  // These values are only used for type-checking, NOT at runtime
  defaultValues: { firstName: 'John', lastName: 'Doe' },

  // Optional: adds props to the render function
  props: { title: 'Child Form' },

  // Render function receives both form instance and custom props
  render: function Render({ form, title }) {
    return (
      <form.AppField name="firstName" />
    );
  },
});

// Usage: Pass form instance from parent
function Parent() {
  const form = useAppForm({ defaultValues: { ... } });
  return <ChildForm form={form} title="Test" />;
}
```

**Critical Quote from Docs (line 484-485)**:
> "These values are only used for type-checking, and are not used at runtime"

**This means:**
- ✅ `defaultValues` in `withForm` is just for TypeScript
- ✅ The actual form instance comes from the `form` prop
- ✅ Multiple components can share the same form instance
- ✅ **Your child components are already using this pattern correctly!**

### 2. Current Implementation Analysis

Looking at your existing child components:

```tsx
// client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx
export const AddExerciseScreen = withForm({
  defaultValues: MOCK_VALUES,  // ✅ Type-checking only!
  props: {} as AddExerciseScreenProps,
  render: function Render({ form, exercises, onAddExercise, onBack }) {
    // Uses form.AppField, form fields from PARENT form instance
  },
});
```

**Analysis**:
- ✅ Already uses `withForm` HOC
- ✅ `MOCK_VALUES` is just for type inference (not creating a new form)
- ✅ Receives `form` prop from parent
- ✅ **NO REFACTORING NEEDED for child components!**

### 3. What This Means for Extraction

**Original Estimate**: Extract with complex prop injection (2.5 hours)

**New Estimate**: Use `withForm` for `WorkoutTracker` (1 hour)

**Why the reduction?**
1. No need to extract child components - already reusable ✅
2. No need to manually inject form context - `withForm` handles it ✅
3. No need to worry about form instance creation - TanStack Form handles it ✅
4. Just wrap `WorkoutTracker` with same pattern as child components ✅

---

## Simplified Option A: Use `withForm` HOC

### Strategy
Wrap `WorkoutTracker` component with `withForm` HOC, following the exact same pattern as existing child components.

### Implementation Steps

#### Step 1: Extract `WorkoutTracker` Using `withForm`
**Estimated Time**: 30 minutes
**Risk**: ✅ Low - Using TanStack Form's built-in pattern
**Rollback**: Revert file, use Option B

**New File**: `client/src/components/workouts/workout-form.tsx`

```tsx
import { withForm } from '@/hooks/form';
import { useMutation } from '@tanstack/react-query';
import type { DbExercise } from '@/lib/api/exercises';
import type { WorkoutFocus } from '@/lib/api/workouts';
import { clearLocalStorage, saveToLocalStorage } from '@/lib/local-storage';
import { MOCK_VALUES } from '@/routes/_auth/workouts/-components/form-options';
import { AddExerciseScreen } from './add-exercise-screen';
import { ExerciseScreen, ExerciseHeader, ExerciseSets } from './exercise-screen';
import { RecentSets } from './recent-sets-display';
// ... other imports

interface WorkoutFormProps {
  userId: string;
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];
  saveMutation: () => UseMutationResult<ResponseSuccessResponse, Error, { body: WorkoutCreateWorkoutRequest }>;
}

export const WorkoutForm = withForm({
  // Type-checking only - NOT used at runtime
  defaultValues: MOCK_VALUES,

  // Custom props injected into render function
  props: {} as WorkoutFormProps,

  // Render function receives form instance + custom props
  render: function Render({ form, userId, exercises, workoutsFocus, saveMutation }) {
    const [currentView, setCurrentView] = useState<'main' | 'exercise' | 'add-exercise'>('main');
    const [selectedExercise, setSelectedExercise] = useState<{ index: number; exerciseId: number | null } | null>(null);

    const saveWorkout = saveMutation();  // Call the mutation factory

    // ... rest of WorkoutTracker implementation
    // Replace all instances of:
    //   - getInitialValues(user.id) → getInitialValues(userId)
    //   - saveToLocalStorage(..., user.id) → saveToLocalStorage(..., userId)
    //   - clearLocalStorage(user.id) → clearLocalStorage(userId)
    //   - useSaveWorkoutMutation() → saveMutation()

    return (
      <Suspense fallback={<Spinner />}>
        <div className="max-w-md mx-auto space-y-6 px-4 pb-8">
          {/* ... exact same JSX as current WorkoutTracker ... */}
        </div>
      </Suspense>
    );
  },
});
```

**Changes Required**:
1. ✅ Copy `WorkoutTracker` component body into `render` function
2. ✅ Add props interface with `userId`, `exercises`, `workoutsFocus`, `saveMutation`
3. ✅ Replace `user.id` → `userId` (4 locations)
4. ✅ Replace `useSaveWorkoutMutation()` → `saveMutation()` (1 location)
5. ✅ Import child components from `@/components/workouts/` (after moving them)

**Verification**:
- TypeScript compiles without errors
- No ESLint warnings about hooks in render (use `function Render` syntax)

---

#### Step 2: Move Child Components to Shared Location
**Estimated Time**: 15 minutes
**Risk**: ✅ Very Low - Just file moves
**Rollback**: Delete new files

**Actions**:
```bash
# Move components
mv client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx \
   client/src/components/workouts/add-exercise-screen.tsx

mv client/src/routes/_auth/workouts/-components/exercise-screen.tsx \
   client/src/components/workouts/exercise-screen.tsx

mv client/src/routes/_auth/workouts/-components/add-set-dialog.tsx \
   client/src/components/workouts/add-set-dialog.tsx

mv client/src/routes/_auth/workouts/-components/recent-sets-display.tsx \
   client/src/components/workouts/recent-sets-display.tsx
```

**Update imports**:
- In moved files: Update relative imports to use `@/` prefix
- In `workout-form.tsx`: Import from `@/components/workouts/`

**Keep in route folder**:
- `form-options.ts` (used by route loaders)
- `mini-chart.tsx` (might be route-specific)

---

#### Step 3: Update Auth Route to Use `WorkoutForm`
**Estimated Time**: 10 minutes
**Risk**: ✅ Low - Simple prop wiring
**Rollback**: Revert file

**File**: `client/src/routes/_auth/workouts/new.tsx`

```tsx
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQueries } from '@tanstack/react-query';
import { WorkoutForm } from '@/components/workouts/workout-form';
import { exercisesQueryOptions, type DbExercise } from '@/lib/api/exercises';
import {
  workoutsFocusValuesQueryOptions,
  useSaveWorkoutMutation,
  type WorkoutFocus,
} from '@/lib/api/workouts';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';

export const Route = createFileRoute('/_auth/workouts/new')({
  loader: async ({ context }): Promise<{ user: CurrentUser | CurrentInternalUser }> => {
    const user = context.user;
    context.queryClient.ensureQueryData(exercisesQueryOptions());
    return { user };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] = useSuspenseQueries({
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
      saveMutation={useSaveWorkoutMutation}  // Pass mutation factory function
    />
  );
}
```

**Key Points**:
- Pass `userId` instead of full `user` object
- Pass `useSaveWorkoutMutation` function (not the result of calling it)
- `WorkoutForm` will call the mutation factory internally

**File Size**: 354 lines → ~50 lines (85% reduction!)

---

#### Step 4: Create Demo Route
**Estimated Time**: 10 minutes
**Risk**: ✅ Very Low - Copy of auth route
**Rollback**: Delete file

**File**: `client/src/routes/demo/workouts/new.tsx`

```tsx
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQueries } from '@tanstack/react-query';
import { WorkoutForm } from '@/components/workouts/workout-form';
import type { DbExercise } from '@/lib/api/exercises';
import type { WorkoutFocus } from '@/lib/api/workouts';
import {
  getDemoExercisesQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
  postDemoWorkoutsMutation,
} from '@/lib/demo-data/query-options';

export const Route = createFileRoute('/demo/workouts/new')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] = useSuspenseQueries({
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
      saveMutation={() => postDemoWorkoutsMutation()}  // Wrap in arrow function
    />
  );
}
```

**Key Difference from Auth Route**:
- `userId="demo-user"` (hardcoded)
- Uses demo query options
- Uses demo mutation wrapped in arrow function

---

#### Step 5: Handle `RecentSets` Component
**Estimated Time**: 15 minutes
**Risk**: ⚠️ Medium - Query parameterization needed
**Rollback**: Duplicate component (Approach 1B from v1 plan)

**Decision**: Duplicate for MVP, refactor later

**Action**:
1. Copy `recent-sets-display.tsx` → `demo-recent-sets-display.tsx`
2. Update import: `getDemoExercisesByIdRecentSetsQueryOptions`
3. Update component name: `DemoRecentSets`
4. Use in `WorkoutForm` based on prop (add `isDemo` boolean prop)

**OR** (cleaner but more complex):
Add `recentSetsQueryOptions` prop to `WorkoutFormProps` and inject it.

---

### Total Estimated Time: **1 hour** (down from 2.5 hours!)

---

## Comparison with Original Plan

| Aspect | Original Plan (v1) | Simplified Plan (v2) |
|--------|-------------------|---------------------|
| **Approach** | Extract with manual prop injection | Use `withForm` HOC |
| **Child Components** | Extract to shared location | **Already reusable!** Just move files |
| **Form Context** | Manually inject via props | **Built-in to `withForm`** |
| **Complexity** | High (custom abstraction) | Low (TanStack pattern) |
| **Time Estimate** | 2.5 hours | 1 hour |
| **Risk** | Medium (custom code) | Low (framework pattern) |
| **Maintainability** | Good | **Excellent** (idiomatic) |

---

## Key Advantages of v2 Plan

1. ✅ **Idiomatic TanStack Form** - Uses official composition pattern
2. ✅ **Less custom code** - Leverages framework features
3. ✅ **Child components already done** - No refactoring needed
4. ✅ **Better documentation** - Follows TanStack Form guides
5. ✅ **Easier onboarding** - Future devs will recognize the pattern
6. ✅ **Type-safe** - TanStack Form handles type inference
7. ✅ **Performance** - Context uses TanStack Store (signals, not reactive values)

---

## Updated Success Criteria

### Phase 1: Extract `WorkoutTracker` with `withForm` ✅
- [ ] `WorkoutForm` component created using `withForm` HOC
- [ ] TypeScript compiles without errors
- [ ] No ESLint warnings (use `function Render` syntax)
- [ ] Props interface includes `userId`, `exercises`, `workoutsFocus`, `saveMutation`

### Phase 2: Move Child Components ✅
- [ ] All 4 child components moved to `client/src/components/workouts/`
- [ ] Imports updated in all files
- [ ] Auth route still works

### Phase 3: Update Routes ✅
- [ ] Auth route uses `WorkoutForm` component (50 lines)
- [ ] Demo route created (50 lines)
- [ ] Both routes pass correct props
- [ ] Both routes work end-to-end

### Phase 4: Verification ✅
- [ ] Auth workout creation works
- [ ] Demo workout creation works
- [ ] localStorage scoped correctly (different keys)
- [ ] No code duplication (except `RecentSets` if duplicated)

---

## Potential Blockers & Solutions

### Blocker 1: Mutation Factory Type Mismatch
**Symptom**: TypeScript error when passing `useSaveWorkoutMutation` as prop
**Diagnosis**: Mutation factory vs mutation result
**Solution**:
```tsx
// Option 1: Pass factory function
saveMutation={() => useSaveWorkoutMutation()}

// Option 2: Change prop type to accept either
saveMutation: () => UseMutationResult<...> | UseMutationOptions<...>
```
**Time Impact**: +10 minutes

---

### Blocker 2: `getInitialValues` Requires `userId`
**Symptom**: Form doesn't load saved data from localStorage
**Diagnosis**: `getInitialValues(userId)` is called in route, not in component
**Current Code (new.tsx line 45)**:
```tsx
const form = useAppForm({
  defaultValues: getInitialValues(user.id),  // ❌ Called in route
  // ...
});
```

**Solution**: Move `getInitialValues` call to route loader
```tsx
// Auth route
function RouteComponent() {
  const { user } = Route.useLoaderData();
  const initialValues = getInitialValues(user.id);

  return (
    <WorkoutForm
      userId={user.id}
      initialValues={initialValues}  // ✅ Pass as prop
      // ...
    />
  );
}

// WorkoutForm
render: function Render({ form, userId, initialValues, ... }) {
  // form is already created by parent with correct defaultValues
  // No need to call getInitialValues here
}
```

**Wait, there's an issue here!** Looking closer at the code:

```tsx
// new.tsx lines 43-51
const form = useAppForm({
  defaultValues: getInitialValues(user.id),
  listeners: {
    onChange: ({ formApi }) => {
      saveToLocalStorage(formApi.state.values, user.id);
    },
  },
  // ...
});
```

**The form is created in the ROUTE, not in `WorkoutTracker`!**

Let me re-read the current implementation...

**AH! I see now. Looking at lines 327-353:**

```tsx
function RouteComponent() {
  const { user } = Route.useLoaderData();
  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({ ... });

  // ... data mapping ...

  return (
    <WorkoutTracker
      user={user}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}
```

**And `WorkoutTracker` creates the form internally (line 43):**
```tsx
const form = useAppForm({
  defaultValues: getInitialValues(user.id),
  // ...
});
```

**So the form IS created inside `WorkoutTracker`!** This is perfect for `withForm` approach.

**CORRECTION**: No blocker here. The `withForm` render function can call hooks (including `useAppForm`), BUT we don't need to because **`withForm` receives the form instance from parent**.

**Actually, let me check the TanStack docs again...**

Looking at lines 1252-1279 of the Form Composition docs:

```tsx
const ChildForm = withForm({
  ...formOpts,
  props: { title: 'Child Form' },
  render: ({ form, title }) => {  // ✅ Receives form from parent
    return (
      <form.AppField name="firstName" />
    );
  },
});

const Parent = () => {
  const form = useAppForm({ ...formOpts });  // ✅ Parent creates form
  return <ChildForm form={form} title={'Testing'} />;
};
```

**SO**: The parent creates the form, and `withForm` component receives it.

**This means we need to MOVE the form creation from `WorkoutTracker` to the route!**

**Updated Solution**:

```tsx
// Auth route
function RouteComponent() {
  const { user } = Route.useLoaderData();

  // ✅ Create form in route
  const form = useAppForm({
    defaultValues: getInitialValues(user.id),
    listeners: {
      onChange: ({ formApi }) => {
        saveToLocalStorage(formApi.state.values, user.id);
      },
      onChangeDebounceMs: 500,
    },
    onSubmit: async ({ value }) => {
      const trimmedValue = {
        ...value,
        notes: value.notes?.trim() || undefined,
        workoutFocus: value.workoutFocus?.trim() || undefined,
      };
      await saveWorkout.mutateAsync({ body: trimmedValue }, {
        onSuccess: () => {
          clearLocalStorage(user.id);
          form.reset();
        },
        onError: (error) => alert(error),
      });
    },
  });

  const saveWorkout = useSaveWorkoutMutation();

  const [{ data: exercisesResponse }, { data: workoutsFocusValues }] =
    useSuspenseQueries({ ... });

  return (
    <WorkoutForm
      form={form}  // ✅ Pass form instance
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}
```

```tsx
// WorkoutForm
export const WorkoutForm = withForm({
  defaultValues: MOCK_VALUES,  // Type-checking only
  props: {} as WorkoutFormProps,
  render: function Render({ form, exercises, workoutsFocus }) {
    // ✅ Use form from parent
    // No useAppForm call here

    return (
      <form.AppField name="..." />
    );
  },
});
```

**Time Impact**: +15 minutes (need to move form creation logic to route)

---

### Blocker 3: Form Lifecycle Hooks in Route
**Symptom**: `onSubmit`, `listeners` need access to mutation
**Diagnosis**: Mutation is component-level, but form creation is route-level now
**Solution**: Move mutation creation to route level too

```tsx
function RouteComponent() {
  const { user } = Route.useLoaderData();
  const saveWorkout = useSaveWorkoutMutation();  // ✅ Create here

  const form = useAppForm({
    defaultValues: getInitialValues(user.id),
    listeners: {
      onChange: ({ formApi }) => {
        saveToLocalStorage(formApi.state.values, user.id);
      },
    },
    onSubmit: async ({ value }) => {
      // ✅ Can access saveWorkout here
      await saveWorkout.mutateAsync({ body: value });
    },
  });

  return <WorkoutForm form={form} ... />;
}
```

**Time Impact**: Already included in Blocker 2 solution

---

## Updated Implementation Steps (Revised)

### Step 1: Move Child Components
Same as before - just file moves.

### Step 2: Extract `WorkoutTracker` UI to `WorkoutForm`
Extract ONLY the render logic (JSX), not the form creation.

```tsx
export const WorkoutForm = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as { exercises: DbExercise[]; workoutsFocus: WorkoutFocus[] },
  render: function Render({ form, exercises, workoutsFocus }) {
    const [currentView, setCurrentView] = useState<'main' | 'exercise' | 'add-exercise'>('main');
    const [selectedExercise, setSelectedExercise] = useState<...>(null);

    // ✅ Helper functions (no form creation)
    const getExerciseId = (exerciseName: string): number | null => { ... };
    const handleAddExercise = (index: number, exerciseId?: number) => { ... };
    const handleExerciseClick = (index: number) => { ... };
    const handleClearForm = () => {
      if (confirm('Are you sure?')) {
        // ❌ Can't call clearLocalStorage here - no userId!
        // ✅ Need to pass as callback prop
      }
    };

    // ✅ All the JSX from WorkoutTracker
    return <div>...</div>;
  },
});
```

**WAIT - Another issue!** `handleClearForm` needs `userId` and `form.reset()`.

**Solution**: Pass callbacks as props:

```tsx
props: {} as {
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];
  onClearForm: () => void;  // ✅ Callback from parent
},

render: function Render({ form, exercises, workoutsFocus, onClearForm }) {
  const handleClearForm = () => {
    if (confirm('Are you sure?')) {
      onClearForm();  // ✅ Call parent callback
    }
  };
  // ...
}
```

**This is getting complicated...**

---

## 🤔 Decision Point: Is `withForm` the Right Approach?

After deeper analysis, I see that `withForm` works great for:
- ✅ Pure presentational components (like child components)
- ✅ Components that RECEIVE form instance from parent
- ✅ Components without complex state management

But `WorkoutTracker` has:
- ❌ Form creation logic with lifecycle hooks
- ❌ LocalStorage integration tied to userId
- ❌ Mutation handling with complex callbacks
- ❌ Multi-screen navigation state

**Conclusion**: `withForm` might not be the best fit for `WorkoutTracker` itself, even though it's perfect for the child components.

---

## Revised Recommendation: Hybrid Approach (v2b)

After analyzing the TanStack Form docs and current code:

1. ✅ **Child components** - Already using `withForm` perfectly, just move files
2. ❌ **WorkoutTracker** - Too complex for `withForm`, use manual extraction (Option A from v1)

**Best Strategy**:
- Use `withForm` pattern for child components (already done ✅)
- Use manual props injection for `WorkoutTracker` (Option A from v1 plan)

**This gives us**:
- ✅ Idiomatic TanStack Form for child components
- ✅ Flexibility for complex parent component
- ✅ Best of both worlds

**Time Estimate**: 1.5 hours (less than v1's 2.5 hours because child components are already done)

---

## Final Recommendation

**Use the v1 plan (Option A) but with this key insight**:
- ✅ Child components are already reusable (using `withForm`) - **just move files**
- ✅ `WorkoutForm` parent needs manual extraction with props (userId, mutation, etc.)

This is actually **Option C (Hybrid)** from the v1 plan!

**See v1 plan for detailed steps, with this modification**:
- Skip "refactoring child components" - they're already perfect
- Focus extraction effort on `WorkoutTracker` only

---

## Next Steps

1. ✅ Review both plans (v1 and v2)
2. ✅ Decide: Use Hybrid approach (Option C from v1, informed by v2 insights)
3. ✅ Move child components to `@/components/workouts/` (15 min)
4. ✅ Extract `WorkoutForm` with manual props injection (45 min)
5. ✅ Update auth and demo routes (25 min)

**Total: 1.25 hours** (v1 Option C estimate, now validated by v2 analysis)

---

**End of Document**
