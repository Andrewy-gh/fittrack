# Demo Routes Implementation Plan

## Overview
Convert authenticated routes (`/_auth/*`) to demo routes (`/demo/*`) with mock data to allow users to test the app without authentication.

## Implementation Strategy: **Easy + Maintainable**

### 1. Route Structure
- **New Routes**: `/demo/workouts/*`, `/demo/exercises/*`
- **Preserve**: Original `/_auth/*` routes remain unchanged
- **Copy Strategy**: Duplicate route files, change imports to use mock data

### 2. Mock Data Architecture
**Location**: `client/src/lib/demo-data/`
- `exercises.ts` - Mock exercise definitions
- `workouts.ts` - Mock workout history
- `workout-sets.ts` - Mock exercise sets with relationships
- `query-options.ts` - Demo versions of React Query options

**Mock User**: Simple `{ id: 'demo-user', name: 'Demo User' }`

### 3. API Strategy: Mock Query Options
- Create `demoExercisesQueryOptions()` that return static data
- Components stay identical - just import different query functions
- No conditional logic, no API mocking complexity

### 4. Data Persistence
- Reuse existing localStorage with 'demo-user' ID
- Form changes persist during demo session
- Can be cleared/reset easily

## File Changes Required

### New Files (7)
1. `client/docs/demo-implementation-plan.md` - This plan
2. `client/src/lib/demo-data/exercises.ts` - Mock exercise data
3. `client/src/lib/demo-data/workouts.ts` - Mock workout data
4. `client/src/lib/demo-data/query-options.ts` - Demo query functions
5. `client/src/routes/demo/workouts/new.tsx` - Demo workout form
6. `client/src/routes/demo/workouts/index.tsx` - Demo workout list
7. `client/src/routes/demo/exercises/index.tsx` - Demo exercise list

### Modified Files (0)
- No existing files need changes
- Original auth routes unchanged

## Mock Data Schema

### Exercises (5 realistic examples)
```typescript
{ id: 1, name: "Bench Press" }
{ id: 2, name: "Squat" }
{ id: 3, name: "Deadlift" }
{ id: 4, name: "Pull-ups" }
{ id: 5, name: "Overhead Press" }
```

### Workouts (3 recent workouts)
- Push Day (Bench Press, Overhead Press)
- Pull Day (Pull-ups, Deadlifts)
- Leg Day (Squat)

### Sets (Realistic progression data)
- Multiple sets per exercise with varying weights/reps
- Historical data showing progression over time

## Implementation Steps

1. **Create mock data files** with realistic fitness data
2. **Create demo query options** that return mock data
3. **Copy route files** from `/_auth/` to `/demo/`
4. **Update imports** in demo routes to use mock queries
5. **Replace user context** with demo user object
6. **Test demo flows** to ensure full functionality

## Benefits

✅ **Easy Implementation**: ~20% of codebase changes, mostly copying
✅ **Future Maintenance**: Mock data in separate files, easy to update
✅ **No Breaking Changes**: Auth routes untouched
✅ **Modular Design**: Components reused, just different data sources
✅ **Quick Updates**: Change mock data without touching component logic

## Timeline
- **Setup mock data**: 1-2 hours
- **Copy and adapt routes**: 2-3 hours
- **Testing and polish**: 1 hour
- **Total**: ~4-6 hours

This approach prioritizes simplicity and maintainability while providing a fully functional demo experience.