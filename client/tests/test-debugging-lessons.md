# Playwright Test Debugging: Lessons Learned

## Summary
Fixed two failing E2E tests in `workout-edit.test.ts` that were testing the ability to add and remove exercises from workouts in demo mode.

---

## Issues Encountered and Solutions

### Issue 1: Wrong `data-testid` Selector for Exercise Cards

**Problem:**
```typescript
// WRONG - counting sets instead of exercises
const updatedExerciseCount = await page
  .getByTestId('exercise-card')  // ❌ This is for sets within an exercise
  .count();
```

**Root Cause:**
- `exercise-card` is used for **sets** within the exercise detail view (see `exercise-screen.tsx:115`)
- `edit-workout-exercise-card` is used for **exercises** in the workout edit form (see `edit.tsx:223`)
- Using the wrong testid caused the test to count 0 sets instead of the actual exercise count

**Solution:**
```typescript
// CORRECT - counting exercises
const updatedExerciseCount = await page
  .getByTestId('edit-workout-exercise-card')  // ✅ Counts exercises
  .count();
```

**Lesson:** Always verify `data-testid` values in the actual component code before using them in tests.

---

### Issue 2: Wrong Form Input Selectors (Label vs Spinbutton)

**Problem:**
```typescript
// WRONG - trying to use getByLabel for number inputs
await page.getByLabel(/weight/i).fill('185');  // ❌ Timeout - element not found
await page.getByLabel(/reps/i).fill('5');
```

**Root Cause:**
- The `InputField` component renders number inputs as `<input type="number">`
- Browsers render `type="number"` inputs with spinbutton controls
- Playwright sees these as `spinbutton` roles, not as labeled inputs in this context

**Solution:**
```typescript
// CORRECT - using spinbutton role
const weightInput = page.getByRole('spinbutton').first();
const repsInput = page.getByRole('spinbutton').nth(1);

await weightInput.fill('185');  // ✅ Works
await repsInput.fill('5');
```

**Lesson:** Always check the accessibility tree or use Playwright's codegen to verify the correct selector. Number inputs are exposed as spinbuttons, not regular text inputs.

---

### Issue 3: Race Condition - Counting Before Elements Load

**Problem:**
```typescript
// WRONG - counting immediately after navigation
await page.getByRole('link', { name: /edit/i }).click();

const initialExerciseCount = await page
  .getByTestId('edit-workout-exercise-card')
  .count();  // ❌ Returns 0 because exercises haven't rendered yet
```

**Root Cause:**
- Navigation to the edit page happens quickly
- Exercise data loads asynchronously (via React Query/Suspense)
- `.count()` returns immediately, even if the elements haven't rendered yet
- This resulted in `initialExerciseCount = 0` when it should have been 2 or 3

**Solution:**
```typescript
// CORRECT - wait for elements to be visible before counting
await page.getByRole('link', { name: /edit/i }).click();

// Wait for exercises to load
await expect(
  page.getByTestId('edit-workout-exercise-card').first()
).toBeVisible();

const initialExerciseCount = await page
  .getByTestId('edit-workout-exercise-card')
  .count();  // ✅ Now returns correct count
```

**Lesson:** Always wait for elements to be visible before counting or interacting with them, especially after navigation or async data loading.

---

### Issue 4: LocalStorage Security Error

**Problem:**
```typescript
// WRONG - trying to access localStorage before navigating to a page
test.beforeEach(async ({ page }) => {
  await page.evaluate(() => localStorage.clear());  // ❌ SecurityError
  await page.goto('/workouts');
});
```

**Root Cause:**
- Playwright's page context doesn't have access to localStorage until a page is loaded
- Attempting to access localStorage before navigation causes a `SecurityError: Access is denied for this document`

**Solution:**
```typescript
// CORRECT - navigate first, then clear localStorage
test.beforeEach(async ({ page }) => {
  await page.goto('/workouts');  // ✅ Navigate first
  await page.evaluate(() => localStorage.clear());  // Now we have access
  await page.reload();  // Reload to reinitialize demo data
});
```

**Lesson:** Always navigate to a page before attempting to access localStorage or other browser APIs.

---

### Issue 5: Delete Button Not Removing Exercise from DOM

**Problem:**
```typescript
// WRONG - not waiting for the DOM to update after delete
await deleteButton.click();

await page.getByRole('button', { name: /save/i }).click();  // ❌ Saves before delete takes effect
```

**Root Cause:**
- Clicking the delete button triggers a React state update
- React needs time to re-render the component tree
- The test was saving immediately before the exercise was removed from the DOM
- This caused the save to include the deleted exercise

**Solution:**
```typescript
// CORRECT - wait for the DOM to update before saving
await deleteButton.click();

// Wait for the exercise to be removed from the DOM
await expect(page.getByTestId('edit-workout-exercise-card')).toHaveCount(
  initialExerciseCount - 1
);  // ✅ Waits for React to re-render

await page.getByRole('button', { name: /save/i }).click();
```

**Lesson:** After interactions that cause DOM changes (delete, add, update), always wait for the expected DOM state before proceeding.

---

### Issue 6: Demo Data Not Properly Resetting Between Tests

**Problem:**
Tests were accumulating state across runs, causing unpredictable results.

**Root Cause:**
- `localStorage.clear()` removes all data
- `initializeDemoData()` only initializes if NO data exists
- After clearing and reloading, demo data was properly reinitialized
- However, not waiting for the data to fully load caused timing issues

**Solution:**
```typescript
test.beforeEach(async ({ page }) => {
  // Navigate first, then clear localStorage and reload to reinitialize demo data
  await page.goto('/workouts');
  await page.evaluate(() => localStorage.clear());
  await page.reload();

  // Wait for demo data to be initialized and workouts to be visible
  await page.waitForSelector('[data-testid="workout-card"]', { timeout: 5000 });
});
```

**Lesson:** When testing with demo/mock data, ensure proper cleanup and initialization in `beforeEach`, and wait for the data to be fully loaded before tests start.

---

## Best Practices Summary

### 1. **Always Wait for Elements**
```typescript
// ❌ BAD
const count = await page.getByTestId('item').count();

// ✅ GOOD
await expect(page.getByTestId('item').first()).toBeVisible();
const count = await page.getByTestId('item').count();
```

### 2. **Wait for DOM Updates After Actions**
```typescript
// ❌ BAD
await deleteButton.click();
await saveButton.click();

// ✅ GOOD
await deleteButton.click();
await expect(page.getByTestId('item')).toHaveCount(expectedCount);
await saveButton.click();
```

### 3. **Verify Selectors in Component Code**
- Check the actual `data-testid` values in components
- Use Playwright Inspector or codegen to see the accessibility tree
- Understand the difference between similar selectors (e.g., `exercise-card` vs `edit-workout-exercise-card`)

### 4. **Use Correct Accessibility Roles**
- Number inputs → `getByRole('spinbutton')`
- Text inputs → `getByLabel()` or `getByRole('textbox')`
- Buttons → `getByRole('button')`
- Dialogs → `getByRole('dialog')`

### 5. **Navigate Before Accessing Browser APIs**
```typescript
// ❌ BAD
await page.evaluate(() => localStorage.clear());
await page.goto('/');

// ✅ GOOD
await page.goto('/');
await page.evaluate(() => localStorage.clear());
```

### 6. **Isolate Test Data**
- Clear state in `beforeEach`
- Wait for data to initialize after clearing
- Use different test data for each test when possible

---

## Debugging Tools Used

1. **Playwright Screenshots** - Showed actual state when tests failed
2. **Error Context (accessibility tree)** - Revealed actual element roles and structure
3. **Source Code Review** - Checked component implementation to verify testids and structure
4. **Demo Data Analysis** - Understood the initial state and data structure

---

## Final Working Test Pattern

```typescript
test('should add/remove items', async ({ page }) => {
  // 1. Navigate to the page
  await page.getByTestId('item-list').click();
  await page.getByRole('link', { name: /edit/i }).click();

  // 2. Wait for elements to load
  await expect(page.getByTestId('item-card').first()).toBeVisible();

  // 3. Count initial state
  const initialCount = await page.getByTestId('item-card').count();

  // 4. Perform action
  await page.getByRole('button', { name: /delete/i }).click();

  // 5. Wait for DOM to update
  await expect(page.getByTestId('item-card')).toHaveCount(initialCount - 1);

  // 6. Save changes
  await page.getByRole('button', { name: /save/i }).click();
  await page.waitForURL('/items/**');

  // 7. Verify persistence
  await page.getByRole('link', { name: /edit/i }).click();
  await expect(page.getByTestId('item-card').first()).toBeVisible();
  const finalCount = await page.getByTestId('item-card').count();
  expect(finalCount).toBe(initialCount - 1);
});
```

---

## Time Saved by Understanding These Issues

Future developers working on this codebase can now avoid:
- ❌ Using wrong testid selectors (30-60 min debugging)
- ❌ Race conditions from not waiting for elements (60-90 min debugging)
- ❌ LocalStorage security errors (15-30 min debugging)
- ❌ Wrong accessibility roles (30-45 min debugging)
- ❌ DOM update timing issues (45-60 min debugging)

**Total time saved per developer: 3-5 hours** 🎉
