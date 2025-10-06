# Next Steps for Demo Routes Implementation

**📅 Last Updated**: 2025-10-05 (Session 8)
**👤 Next Agent**: Start here for quick orientation

---

## 🚨 Current Status: BLOCKED - Route Redirect Issue

### Problem Discovered (Session 8)

Demo routes are redirecting to login screen when clicking on workout/exercise links.

**Root Cause**:
- Shared components (`WorkoutList`, `ExerciseDetail`, etc.) have hardcoded links to auth routes
- Example: `<Link to="/workouts/$workoutId">` navigates to `/_auth` routes
- `/_auth.tsx:7` has auth check with `or: 'redirect'` → triggers redirect
- TanStack Router requires type-safe hardcoded paths (cannot use dynamic base paths)

**Files Involved**:
- `client/src/components/workouts/workout-list.tsx:90` → hardcoded `/workouts/$workoutId`
- `client/src/components/workouts/workout-detail.tsx:174` → hardcoded `/exercises/$exerciseId`
- `client/src/components/exercises/exercise-list.tsx:57` → hardcoded `/exercises/$exerciseId`
- `client/src/components/demo-header.tsx` → Created to avoid `useUser()` hook redirect

**Attempted Solutions**:
1. ❌ Dynamic base paths → TypeScript errors (TanStack Router requires hardcoded paths)
2. ❌ String template paths → TypeScript errors
3. ✅ Created `DemoHeader` (avoids auth hooks)
4. 🔄 **NEW PROPOSAL**: Merge auth and demo routes (see below)

---

## 📋 Proposed Solution: Route Merge

**See detailed proposal**: `client/docs/route-merge-proposal.md`

### High-Level Plan

Instead of maintaining duplicate routes, **merge `/_auth` and `/demo` routes** into unified root routes:

1. **Move routes** from `/_auth/workouts` → `/workouts`
2. **Conditional data loading** in `beforeLoad`:
   - If user authenticated → use API query options
   - If not authenticated → use demo query options (localStorage)
3. **Single component tree** serves both modes
4. **Hide auth features** (edit/delete) when in demo mode

### Benefits
- ✅ No code duplication
- ✅ Type-safe hardcoded links work correctly
- ✅ Simpler maintenance (one route tree)
- ✅ Seamless UX (demo → authenticated transition)

### Complexity
- **Low**: Query option switching, shared components already support this
- **Medium**: New workout route, URL consistency on login
- **Estimated time**: ~4 hours total

### Key Questions (Pending Review)
1. Route naming: `/workouts` auto-detects mode? Or keep `/demo` explicit?
2. Edit routes: Hide in demo or redirect to sign-in?
3. New workout: Support in demo mode or require auth?
4. Header: Unified with conditional rendering?
5. Data refresh: Auto-refresh on login?

**Next Action**: User to review `route-merge-proposal.md` and approve approach

---

## ✅ What Was Completed (Session 7-8)

### Session 7: View-Only Demo Routes
All view-only demo routes are now live:

1. **`demo.tsx`** - Layout with demo mode banner
2. **`demo/workouts/index.tsx`** - Workout list view
3. **`demo/workouts/$workoutId/index.tsx`** - Workout detail view
4. **`demo/exercises/index.tsx`** - Exercise list view
5. **`demo/exercises/$exerciseId.tsx`** - Exercise detail view

**Files Created**:
- `client/src/routes/demo.tsx` (26 lines)
- `client/src/routes/demo/workouts/index.tsx` (16 lines)
- `client/src/routes/demo/workouts/$workoutId/index.tsx` (31 lines)
- `client/src/routes/demo/exercises/index.tsx` (16 lines)
- `client/src/routes/demo/exercises/$exerciseId.tsx` (32 lines)

**Total**: ~120 lines of route code

### Session 8: Investigation & Proposal
1. **Identified redirect issue** (hardcoded links in shared components)
2. **Created `DemoHeader`** component (avoids `useUser()` hook)
3. **Analyzed route merge approach** (conditional loaders)
4. **Documented comprehensive proposal** (`route-merge-proposal.md`)

---

## 📚 Key Documentation

### Critical Reading (For Route Merge)

1. **`client/docs/route-merge-proposal.md`** ⭐ **START HERE**
   - Comprehensive analysis of merge approach
   - Implementation details with code examples
   - Complexity assessment and blockers
   - Migration steps (4 phases)
   - Questions requiring approval
   - **Action Required**: User review and decision

### Background Context

2. **`client/docs/demo-plan.md`** - Master checklist with all phases
   - Phase 2.3a: ✅ Complete (view-only routes)
   - Phase 2.3b: ⏳ Pending (form route) - **May be superseded by route merge**
   - Phase 3: ⏳ Pending (integration)

3. **`client/docs/demo-progress.md`** - Historical session logs
   - 600+ lines of detailed progress
   - Use as deep reference if needed

### Workout Form Extraction (May Not Be Needed)

4. **`client/docs/workout-form-extraction-plan.md`** - Comprehensive analysis
   - 900 lines covering 3 extraction options
   - **Note**: May be replaced by route merge approach

5. **`client/docs/workout-form-extraction-plan-v2.md`** - Refined approach
   - 400 lines focusing on TanStack Form patterns
   - **Note**: May be replaced by route merge approach

---

## 🗂️ File Structure Summary

### What Exists Now

```
client/src/lib/demo-data/                          ✅ COMPLETE
  ├── types.ts                    # Type imports/exports
  ├── initial-data.ts             # Seed data (5 exercises, 3 workouts, 26 sets)
  ├── storage.ts                  # localStorage utilities
  └── query-options.ts            # Demo query/mutation options

client/src/components/                             ✅ COMPLETE
  ├── demo-header.tsx            # Demo-specific header (no auth)
  ├── workouts/
  │   ├── workout-list.tsx       # Shared workout list component
  │   └── workout-detail.tsx     # Shared workout detail component (supports showEditDelete)
  └── exercises/
      ├── exercise-list.tsx      # Shared exercise list component
      └── exercise-detail.tsx    # Shared exercise detail component

client/src/routes/demo/                            ⚠️ HAS REDIRECT ISSUE
  ├── demo.tsx                    # Layout with banner (uses DemoHeader)
  ├── workouts/
  │   ├── index.tsx              # Workout list ✅
  │   ├── $workoutId/index.tsx   # Workout detail ⚠️ (redirects on link click)
  │   └── new.tsx                # Create workout ⏳ TODO (Phase 2.3b)
  └── exercises/
      ├── index.tsx              # Exercise list ✅
      └── $exerciseId.tsx        # Exercise detail ⚠️ (redirects on link click)
```

### Auth Routes (Reference)

```
client/src/routes/_auth/                           ✅ COMPLETE (Will be moved)
  ├── _auth.tsx                   # Auth layout with redirect
  ├── workouts/
  │   ├── index.tsx              # API-based workout list
  │   ├── $workoutId/index.tsx   # API-based workout detail
  │   ├── $workoutId/edit.tsx    # Edit route (auth required)
  │   └── new.tsx                # Form route (complex, 12KB)
  └── exercises/
      ├── index.tsx              # API-based exercise list
      └── $exerciseId.tsx        # API-based exercise detail
```

---

## 🎯 What to Do Next

### OPTION A: Route Merge Approach (RECOMMENDED)

**Requires**: User approval after reviewing `route-merge-proposal.md`

**Steps after approval**:
1. **Phase 1: Preparation** (30 min)
   - Update `__root.tsx` for nullable user context
   - Create helper for conditional query options
   - Test auth detection

2. **Phase 2: Migrate Exercises** (1 hour)
   - Move `_auth/exercises/index.tsx` → `exercises/index.tsx`
   - Add conditional loaders
   - Test both modes
   - Move `$exerciseId.tsx`

3. **Phase 3: Migrate Workouts** (1.5 hours)
   - Move workout routes with conditional logic
   - Keep edit/new routes auth-protected

4. **Phase 4: Cleanup** (1 hour)
   - Delete `/demo` routes
   - Delete `/_auth.tsx` layout
   - Unified header component
   - Testing

**Estimated Total**: ~4 hours

### OPTION B: Keep Separate Routes (Fallback)

If route merge is rejected:

1. **Create demo-specific components** (duplicate presentation logic)
   - `client/src/components/demo/demo-workout-list.tsx`
   - `client/src/components/demo/demo-exercise-list.tsx`
   - Hardcode `/demo/*` paths in components

2. **Update demo routes** to use demo components

**Downside**: Code duplication, harder maintenance

**Estimated**: ~2 hours

---

## 🔍 Current Test Status

### ✅ Working
- [x] `/demo/workouts` loads workout list
- [x] `/demo/exercises` loads exercise list
- [x] Data loads from localStorage
- [x] Demo banner visible
- [x] TypeScript checks pass (`bun run tsc` after reverting dynamic paths)
- [x] `DemoHeader` renders without auth

### ❌ Broken (Due to Hardcoded Links)
- [ ] Clicking workout from `/demo/workouts` → redirects to login
- [ ] Clicking exercise from `/demo/exercises` → redirects to login
- [ ] Exercise links in workout detail → redirect to login

---

## 💡 Recommendations

### For Next Agent/Session

1. **Review `route-merge-proposal.md`** - This is the critical decision point
2. **Get user approval** on the 5 key questions in the proposal
3. **If approved**: Begin Phase 1 (Preparation) of route merge
4. **If rejected**: Implement Option B (separate components with duplicated code)

### For User

1. **Read**: `client/docs/route-merge-proposal.md`
2. **Decide** on the 5 key questions (route naming, edit behavior, new workout, header, data refresh)
3. **Approve or reject** the route merge approach
4. **If approved**: Agent can proceed with 4-phase migration (~4 hours)

---

## 🚀 Success Metrics

**Current Progress**:
- ✅ Demo data infrastructure (localStorage, query options)
- ✅ Shared components extracted
- ✅ Demo routes created (but have redirect issue)
- ✅ `DemoHeader` component (avoids auth)
- ⏳ Route merge proposal documented

**Remaining for Full Demo Feature**:
- ⏳ Fix redirect issue (route merge or duplicate components)
- ⏳ Landing page integration
- ⏳ End-to-end testing
- ⏳ Documentation cleanup

---

**End of Document**

**NEXT ACTION**: User reviews `client/docs/route-merge-proposal.md` and provides decision on route merge approach
