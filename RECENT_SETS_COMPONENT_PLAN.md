# Recent Sets Component Implementation Plan

## Overview
This pull request aims to implement a component that displays recent workout sets for exercises in the new workout creation flow (`client/src/routes/_auth/workouts/new-2.tsx`). This component will be similar to the workout history cards shown in `client/src/routes/_auth/exercises/$exerciseId.tsx` (lines 166-218).

## Current State Analysis

### Existing Exercise Detail View (Reference Implementation)
**File**: `client/src/routes/_auth/exercises/$exerciseId.tsx` (lines 166-218)
- Shows workout history cards with date, time, notes, reps, volume
- Displays individual sets with weight, reps, and volume
- Uses `useSuspenseQuery` with `exerciseQueryOptions`
- Has access to exercise ID from route params

### Current Workout Creation Flow
**File**: `client/src/routes/_auth/workouts/new-2.tsx`
- Uses form with exercise array: `exercises[exerciseIndex]`
- Only stores exercise name, no exercise ID
- Exercise selection from dropdown adds exercise by name only
- No recent sets display currently implemented

### Data Structures
- `workout_ExerciseInput`: Only has `name` and `sets` fields (no ID)
- `exercise_RecentSetsResponse`: Contains `set_id`, `weight`, `reps`, `workout_date`, `created_at`
- Available API endpoint: `getExercisesRecentSets(id: number)`

## Implementation Strategy

### State Management Approach
**Decision**: Consolidate into a single state object instead of separate `exerciseId` and `exerciseIndex` states.

```typescript
// Current state
const [selectedExerciseIndex, setSelectedExerciseIndex] = useState<number | null>(null);

// Enhanced state
const [selectedExercise, setSelectedExercise] = useState<{
  index: number;
  exerciseId: number | null;
} | null>(null);
```

**Rationale**: 
- Keeps related data together
- Makes it clear when we have both index and ID
- Easier to pass both pieces of data to components
- Prevents state synchronization issues

## Tasks Breakdown

### 1. API Integration
**File**: `client/src/lib/api/exercises.ts`

Add query options for recent sets:

```typescript
import type { exercise_RecentSetsResponse } from '@/generated';

export function recentExerciseSetsQueryOptions(exerciseId: number, user: User) {
  const validatedUser = ensureUser(user);
  return queryOptions<exercise_RecentSetsResponse[], Error>({
    queryKey: ['exercises', 'recent-sets', exerciseId],
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercisesRecentSets(exerciseId);
    },
  });
}
```

### 2. Update Workout Form Data Structure

**Challenge**: The `workout_ExerciseInput` type doesn't include exercise ID, but we need it to fetch recent sets.

**Solution**: Extend the form state to track exercise IDs alongside the workout data:

```typescript
// Enhanced state in WorkoutTracker component
const [selectedExercise, setSelectedExercise] = useState<{
  index: number;
  exerciseId: number | null;
} | null>(null);

// Helper to get exercise ID from exercises list
const getExerciseId = (exerciseName: string): number | null => {
  const exercise = exercises.find(ex => ex.name === exerciseName);
  return exercise?.id || null;
};
```

### 3. Update Exercise Selection Flow

**File**: `client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx`

Modify the exercise addition to capture and pass exercise ID:

```typescript
onClick={() => {
  field.pushValue({
    name: exercise.name,
    sets: [],
  });
  const exerciseIndex = field.state.value.length - 1;
  onAddExercise(exerciseIndex, exercise.id); // Pass exercise ID too
}}
```

**File**: `client/src/routes/_auth/workouts/new-2.tsx`

Update handler signatures:

```typescript
const handleAddExercise = (index: number, exerciseId: number) => {
  setSelectedExercise({ index, exerciseId });
  setCurrentView('exercise');
};

const handleExerciseClick = (index: number) => {
  const exerciseName = form.state.values.exercises[index]?.name;
  const exerciseId = getExerciseId(exerciseName);
  setSelectedExercise({ index, exerciseId });
  setCurrentView('exercise');
};
```

### 4. Create Recent Sets Component

**File**: `client/src/routes/_auth/workouts/-components/recent-sets-display.tsx`

```typescript
import { useSuspenseQuery } from '@tanstack/react-query';
import { recentExerciseSetsQueryOptions } from '@/lib/api/exercises';
import type { User } from '@/lib/api/auth';

interface RecentSetsDisplayProps {
  exerciseId: number;
  user: User;
}

export function RecentSetsDisplay({ exerciseId, user }: RecentSetsDisplayProps) {
  const { data: recentSets } = useSuspenseQuery(
    recentExerciseSetsQueryOptions(exerciseId, user)
  );

  if (recentSets.length === 0) {
    return (
      <div className="text-center py-4">
        <p className="text-muted-foreground text-sm">No sets yet</p>
      </div>
    );
  }

  // Group sets by workout_date (similar to existing implementation)
  const groupedSets = /* grouping logic */;

  return (
    <div className="space-y-4">
      <h3 className="text-lg font-semibold">Recent Sets</h3>
      {/* Render workout cards similar to lines 166-218 from exerciseId.tsx */}
      {/* Exclude lines 188-193 (reps and volume summary) as specified */}
    </div>
  );
}
```

### 5. Update Exercise Screen Component

**File**: `client/src/routes/_auth/workouts/-components/exercise-screen.tsx`

Add recent sets section:

```typescript
// Add props
type ExerciseScreenProps = {
  exerciseIndex: number;
  exerciseId: number | null;
  user: User;
  onBack: () => void;
};

// In the render function, add Recent Sets section
{exerciseId && (
  <Suspense fallback={<div>Loading recent sets...</div>}>
    <RecentSetsDisplay exerciseId={exerciseId} user={user} />
  </Suspense>
)}
```

### 6. Update Main Workout Component

**File**: `client/src/routes/_auth/workouts/new-2.tsx`

Pass the exercise ID and user to ExerciseScreen2:

```typescript
if (
  currentView === 'exercise' &&
  selectedExercise !== null &&
  form.state.values.exercises.length > 0
) {
  return (
    <Suspense fallback={/* loading */}>
      <ExerciseScreen2
        form={form}
        exerciseIndex={selectedExercise.index}
        exerciseId={selectedExercise.exerciseId}
        user={user}
        onBack={() => setCurrentView('main')}
      />
    </Suspense>
  );
}
```

## Edge Cases & Considerations

### 1. No Exercise ID Available
**Scenario**: User types custom exercise name not in the database
**Solution**: Show "No sets yet" placeholder when `exerciseId` is null

### 2. Network Errors
**Scenario**: API call fails
**Solution**: Use React Query's error boundary or error state handling

### 3. Empty Recent Sets
**Scenario**: Exercise exists but has no previous sets
**Solution**: Show "No sets yet" message as specified

### 4. Loading States
**Scenario**: Fetching recent sets data
**Solution**: Use Suspense boundaries with appropriate loading indicators

## File Changes Summary

```
client/src/lib/api/exercises.ts                           [MODIFIED]
client/src/routes/_auth/workouts/new-2.tsx               [MODIFIED]  
client/src/routes/_auth/workouts/-components/exercise-screen.tsx [MODIFIED]
client/src/routes/_auth/workouts/-components/add-exercise-screen.tsx [MODIFIED]
client/src/routes/_auth/workouts/-components/recent-sets-display.tsx [NEW]
```

## Testing Strategy

### Unit Tests
- Test `recentExerciseSetsQueryOptions` with mocked API calls
- Test `RecentSetsDisplay` component with various data states
- Test state transitions in the main workout component

### Integration Tests
- Test complete flow from exercise selection to recent sets display
- Test error handling when exercise ID is not available
- Test loading states and Suspense boundaries

### Manual Testing
- Verify recent sets appear when selecting existing exercises
- Verify "No sets yet" appears for new/custom exercises
- Verify navigation flow works correctly
- Test with and without network connectivity

## Migration Considerations

### Backward Compatibility
- Existing workout creation flow should work unchanged
- Form data structure changes are additive (no breaking changes)
- Recent sets feature is purely additive

### Performance
- React Query caching will prevent redundant API calls
- Suspense ensures UI remains responsive
- Component lazy loading where appropriate

## Future Enhancements

### Possible Extensions
- Cache recent sets data across navigation
- Add "use these weights" quick-fill feature
- Show recent sets trend/progression
- Add recent sets filtering/sorting options

---

## Implementation Checklist

- [ ] Add `recentExerciseSetsQueryOptions` to exercises API
- [ ] Update state management in main workout component
- [ ] Modify exercise selection to capture exercise IDs  
- [ ] Create `RecentSetsDisplay` component
- [ ] Update `ExerciseScreen2` to show recent sets
- [ ] Handle edge cases (no ID, no data, errors)
- [ ] Add loading states and error boundaries
- [ ] Test complete user flow
- [ ] Update TypeScript types as needed
- [ ] Document component interfaces
