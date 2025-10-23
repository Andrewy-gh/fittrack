# Workout New Page - Query Params Refactor Plan

## Goal
Replace React state with TanStack Router search params for proper browser navigation (back/forward buttons work, deep linking supported).

## Search Schema
```typescript
z.object({
  addExercise: z.boolean().default(false),
  exerciseIndex: z.coerce.number().int().optional(),
})
```

## Key Changes

### 1. Route Configuration
- Add `validateSearch` with Zod schema to route definition

### 2. Replace State with Search Params
**Remove:**
- `useState` for `currentView`
- `useState` for `selectedExercise`

**Add:**
- `const { addExercise, exerciseIndex } = Route.useSearch()`
- `const navigate = useNavigate({ from: Route.fullPath })`

### 3. Exercise Index Validation
```typescript
if (exerciseIndex !== undefined) {
  const exercises = form.state.values.exercises;
  if (exerciseIndex < 0 || exerciseIndex >= exercises.length) {
    // Silently redirect to main
    navigate({ search: {} });
    return null;
  }
}
```

### 4. Update Navigation

**Before:** `setCurrentView('add-exercise')`
**After:** `navigate({ search: { addExercise: true } })`

**Before:** `setCurrentView('exercise')` + `setSelectedExercise(...)`
**After:** `navigate({ search: { exerciseIndex: idx } })`

**Before:** `setCurrentView('main')`
**After:** `navigate({ search: {} })`

### 5. Update Components

#### Exercise Cards (new.tsx)
- Convert to `<Link>` components with `search={{ exerciseIndex }}`
- Keep delete button click handler with `e.stopPropagation()`

#### Add Exercise Button (new.tsx)
- Use `<Link to="." search={{ addExercise: true }}>`

#### AddExerciseScreen
- `onAddExercise` should call `navigate({ search: { exerciseIndex } })` internally
- Remove `onAddExercise` from parent, handle navigation inside component

#### ExerciseHeader & ExerciseScreen
- `onBack` callbacks stay but call `navigate({ search: {} })`

### 6. Conditional Rendering
```typescript
if (addExercise) return <AddExerciseScreen ... />;

if (exerciseIndex !== undefined) {
  // Validation + render ExerciseScreen
}

// Default main view
return <main>...</main>;
```

## Edge Cases Handled

1. **Invalid Index**: Silently redirect to main
2. **Delete Exercise**: Stays in main view (delete only available there)
3. **Clear Form**: Add `navigate({ search: {} })` to handler
4. **Race Conditions**: Validation catches stale indices

## Benefits
✅ Browser back/forward buttons work
✅ URL reflects current state
✅ Deep linking supported
✅ Refresh preserves view (if form data exists)
