# Next Steps for Demo Routes Implementation

**📅 Last Updated**: 2025-10-05 (Session 6)
**👤 Next Agent**: Start here for quick orientation

---

## 🎯 What to Do Next

### Phase 2.3a: View-Only Demo Routes (DO THIS FIRST - Easy!)

Create 4 simple route wrappers in `client/src/routes/demo/`:

**Estimated Time**: 45 minutes total

#### 1. `demo/workouts/index.tsx` (~10 min)
```tsx
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { getDemoWorkoutsQueryOptions } from '@/lib/demo-data/query-options';
import { WorkoutList } from '@/components/workouts/workout-list';

export const Route = createFileRoute('/demo/workouts/')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { data: workouts } = useSuspenseQuery(getDemoWorkoutsQueryOptions());
  return <WorkoutList workouts={workouts} />;
}
```

#### 2. `demo/workouts/$workoutId/index.tsx` (~10 min)
- Copy pattern from `_auth/workouts/$workoutId/index.tsx`
- Use `getDemoWorkoutsByIdQueryOptions(workoutId)` in loader
- Render `<WorkoutDetail>` component

#### 3. `demo/exercises/index.tsx` (~10 min)
- Copy pattern from `_auth/exercises/index.tsx`
- Use `getDemoExercisesQueryOptions()` in loader
- Render `<ExerciseList>` component

#### 4. `demo/exercises/$exerciseId.tsx` (~10 min)
- Copy pattern from `_auth/exercises/$exerciseId.tsx`
- Use `getDemoExercisesByIdQueryOptions(exerciseId)` in loader
- Render `<ExerciseDetail>` component

#### 5. Optional: `demo.tsx` layout (~5 min)
- Create pathless layout route
- Initialize demo data on mount
- Add demo mode indicator/banner

---

## ⚠️ What NOT to Do Yet

### Phase 2.3b: Workout Form Route (DEFERRED - Complex)

**Do NOT implement `demo/workouts/new.tsx` yet!**

**Why?**
- Complex form extraction with user/mutation dependencies
- Requires 1.25 hours of focused work
- Blocks progress on simpler routes

**When ready**:
- Read `client/docs/workout-form-extraction-plan.md` (900 lines)
- Read `client/docs/workout-form-extraction-plan-v2.md` (400 lines)
- Use Hybrid approach (Option C): Extract child components, manual props for parent

---

## 📚 Key Documentation

### Start Here
1. **`client/docs/demo-progress.md`** - Full progress log
   - See Session 6 for workout form analysis
   - See Session 5 for component extraction details

2. **`client/docs/demo-plan.md`** - Implementation checklist
   - Note Phase 2.3a/2.3b split
   - Check off items as you complete them

### Reference Files
- **Shared Components** (already extracted ✅):
  - `client/src/components/workouts/workout-list.tsx`
  - `client/src/components/workouts/workout-detail.tsx`
  - `client/src/components/exercises/exercise-list.tsx`
  - `client/src/components/exercises/exercise-detail.tsx`

- **Demo Data Layer** (already complete ✅):
  - `client/src/lib/demo-data/types.ts`
  - `client/src/lib/demo-data/initial-data.ts`
  - `client/src/lib/demo-data/storage.ts`
  - `client/src/lib/demo-data/query-options.ts`

- **Auth Routes** (reference for demo routes):
  - `client/src/routes/_auth/workouts/index.tsx` (30 lines - thin wrapper)
  - `client/src/routes/_auth/workouts/$workoutId/index.tsx` (34 lines - thin wrapper)
  - `client/src/routes/_auth/exercises/index.tsx` (16 lines - thin wrapper)
  - `client/src/routes/_auth/exercises/$exerciseId.tsx` (31 lines - thin wrapper)

---

## ✅ Success Criteria

After implementing Phase 2.3a, you should be able to:

1. **Navigate to demo routes**:
   - `/demo/workouts` - See list of 3 demo workouts
   - `/demo/workouts/1` - See workout detail with sets/exercises
   - `/demo/exercises` - See list of 5 demo exercises
   - `/demo/exercises/1` - See exercise detail with stats

2. **Verify data loads from localStorage**:
   - Open DevTools → Application → Local Storage
   - See keys: `fittrack-demo-workouts`, `fittrack-demo-exercises`, `fittrack-demo-sets`
   - Data matches types from `@/client/types.gen.ts`

3. **No TypeScript errors**:
   - Run `npm run typecheck` (or equivalent)
   - All routes compile successfully

4. **No console errors**:
   - Open browser console
   - Navigate through demo routes
   - No errors or warnings

---

## 🔄 After Completing Phase 2.3a

Update documentation:
1. Check off Phase 2.3a items in `client/docs/demo-plan.md`
2. Add Session 7 summary to `client/docs/demo-progress.md`
3. Test all 4 demo routes end-to-end
4. Decide: Continue to Phase 2.3b (form) or Phase 3 (integration)?

---

## 💡 Tips

1. **Copy existing auth routes** - They're already thin wrappers (~20-35 lines each)
2. **Only change 3 things per route**:
   - Route path (`/_auth/...` → `/demo/...`)
   - Query options (auth → demo)
   - Remove user dependency
3. **Components are ready** - Just import and render
4. **Demo data auto-initializes** - See `client/src/lib/demo-data/query-options.ts` line 30

---

## 🚨 Common Pitfalls

1. **Don't create new components** - Reuse existing ones from `@/components/`
2. **Don't modify auth routes** - Leave them untouched
3. **Don't start with the form route** - It's complex, do view routes first
4. **Don't forget the loader** - TanStack Router needs `ensureQueryData()` call

---

## 🎉 You Got This!

The hard work is done:
- ✅ Mock data infrastructure built
- ✅ Components extracted and tested
- ✅ Auth routes working as reference

Now it's just wiring up 4 simple routes. Should take ~45 minutes total.

**Questions?** Check the detailed plans or ask the user!

---

**End of Document**
