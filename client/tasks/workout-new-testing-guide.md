# Workout New Page - Testing Guide

## Overview
This guide covers testing for the query params refactor that replaced React state with TanStack Router search params.

## Quick Start

### 1. Start Development Server
```bash
cd client
npm run dev
```

### 2. Run Existing E2E Tests
```bash
# Run all tests
npm run test:e2e

# Or run with UI for interactive debugging
npm run test:e2e:ui
```

---

## Automated Testing

### E2E Tests to Run

The existing Playwright tests should mostly pass. Run these specific test files:

```bash
# Workout creation tests (primary affected area)
npx playwright test tests/e2e/demo/workout-create.test.ts

# All demo tests
npx playwright test tests/e2e/demo/
```

### Expected Test Results

- ✅ Most tests should pass without changes
- ⚠️ Some navigation assertions might need minor updates if they're checking specific button clicks vs Link clicks
- ✅ Form functionality (adding exercises, sets, saving) should work identically

---

## Manual Testing Checklist

### 🔗 Browser Navigation (Most Important)

**Test Case 1: Back Button from Add Exercise Screen**
- [x] Navigate to `/workouts/new`
- [x] Click "Add Exercise" button
- [x] URL should show `?addExercise=true`
- [x] Click browser **back button**
- [x] ✅ Should return to main workout form
- [ ] URL should be `/workouts/new` (no params) **Not expected Url: http://localhost:5173/workouts/new?addExercise=false**

**Test Case 2: Back Button from Exercise Detail**
- [x] Navigate to `/workouts/new`
- [x] Click "Add Exercise"
- [x] Select any exercise (e.g., "Bench Press")
- [ ] URL should show `?exerciseIndex=0` **Not expected Url: http://localhost:5173/workouts/new?exerciseIndex=0&addExercise=false**
- [x] Click browser **back button**
- [x] ✅ Should return to main workout form
- [ ] URL should be `/workouts/new` (no params) **Not expected Url: http://localhost:5173/workouts/new?addExercise=false**

**Test Case 3: Forward Button**
- [x] Follow Test Case 1 or 2 above
- [x] After clicking back, click browser **forward button**
- [x] ✅ Should navigate forward to the previous screen
- [ ] URL params should be restored **Not expected Url: http://localhost:5173/workouts/new?exerciseIndex=1&addExercise=false**

**Test Case 4: Multiple Navigation Steps**
- [ ] Add 2-3 exercises to the workout
- [ ] Click on first exercise card → Check URL has `?exerciseIndex=0`
- [ ] Click back → Should return to main view
- [ ] Click on second exercise card → Check URL has `?exerciseIndex=1`
- [ ] Click back → Should return to main view
- [ ] Use browser forward/back buttons multiple times
- [ ] ✅ Navigation should work smoothly in both directions

---

### 🔗 URL Deep Linking

**Test Case 5: Deep Link to Add Exercise Screen**
- [x] Navigate to `/workouts/new?addExercise=true` directly in browser
- [x] ✅ Should show "Choose Exercise" screen
- [x] Or copy URL when on Add Exercise screen and open in new tab


**Test Case 6: Deep Link to Exercise Detail**
- [x] Add an exercise to the workout (e.g., "Bench Press")
- [x] Click on the exercise card to view details
- [x] Copy the URL (should be `/workouts/new?exerciseIndex=0`)
- [x] Open that URL in a **new tab** or **incognito window**
- [x] ✅ Should show the exercise detail screen (if localStorage has the exercise)

**Test Case 7: Invalid Exercise Index**
- [x] Navigate to `/workouts/new?exerciseIndex=999` directly
- [x] ✅ Should silently redirect to main view (`/workouts/new`)
- [x] No errors should appear in console

**Test Case 8: Negative Exercise Index**
- [x] Navigate to `/workouts/new?exerciseIndex=-1` directly
- [x] ✅ Should silently redirect to main view
- [x] No errors in console

---

### 🔄 Page Refresh

**Test Case 9: Refresh on Add Exercise Screen**
- [x] Click "Add Exercise" button
- [x] Refresh the page (F5 or Ctrl+R)
- [x] ✅ Should remain on "Choose Exercise" screen
- [x] URL should still show `?addExercise=true`

**Test Case 10: Refresh on Exercise Detail**
- [x] Add an exercise and click on it to view details
- [x] Refresh the page
- [x] ✅ Should remain on exercise detail screen
- [x] Exercise data should persist (from localStorage)

**Test Case 11: Refresh on Main View**
- [x] Be on main workout form view
- [x] Refresh the page
- [x] ✅ Should remain on main view
- [x] Any in-progress workout data should persist (from localStorage)

---

### ✂️ Form Actions

**Test Case 12: Clear Form Action**
- [ ] Add an exercise to the workout
- [ ] Click on exercise to view details (URL: `?exerciseIndex=0`)
- [ ] Go back to main view
- [ ] Click "Clear" button and confirm
- [ ] ✅ URL should reset to `/workouts/new` (no params)
- [ ] ✅ Form should be completely empty
- [ ] ✅ localStorage should be cleared

**Test Case 13: Save Workout Action**
- [ ] Add an exercise with at least one set
- [ ] Navigate to exercise detail (URL has `?exerciseIndex=0`)
- [ ] Go back and click "Save"
- [ ] ✅ URL should reset to `/workouts/new` (no params)
- [ ] ✅ Form should be reset (no exercises)
- [ ] ✅ Should remain on `/workouts/new` page

**Test Case 14: Delete Exercise**
- [ ] Add 2 exercises to the workout
- [ ] On main view, click the trash icon on first exercise
- [ ] ✅ Exercise should be deleted immediately
- [ ] ✅ Should remain on main view (no URL navigation)
- [ ] ✅ Remaining exercise indices should update correctly

---

### 🎯 Exercise Card Interactions

**Test Case 15: Click Exercise Card**
- [x] Add an exercise to the workout
- [x] Click anywhere on the exercise card (not the delete button)
- [x] ✅ Should navigate to exercise detail screen
- [x] ✅ URL should show `?exerciseIndex=0`
- [x] ✅ Correct exercise details should appear

**Test Case 16: Delete Button on Exercise Card**
- [x] Add an exercise to the workout
- [x] Click the **trash/delete icon** on the exercise card
- [x] ✅ Exercise should be deleted
- [x] ✅ Should NOT navigate away from main view
- [x] ✅ URL should remain `/workouts/new` (no params)

**Test Case 17: Add Exercise Button**
- [x] Click "Add Exercise" button on main form
- [x] ✅ Should navigate to "Choose Exercise" screen
- [x] ✅ URL should show `?addExercise=true`

---

### 📝 Add Exercise Flow

**Test Case 18: Select Exercise from List**
- [x] Click "Add Exercise"
- [x] Click on any exercise from the list (e.g., "Deadlift")
- [x] ✅ Should navigate to exercise detail screen
- [x] ✅ URL should show `?exerciseIndex=N` where N is the new exercise index
- [x] ✅ Exercise name should appear in header
- [x] ✅ "Add Set" button should be visible

**Test Case 19: Create New Exercise via Search**
- [x] Click "Add Exercise"
- [x] Type a new exercise name in search (e.g., "My Custom Exercise")
- [x] Click the "Add" button that appears
- [x] ✅ Should navigate to exercise detail screen for the new exercise
- [x] ✅ URL should show `?exerciseIndex=N`
- [x] ✅ New exercise name should appear in header

**Test Case 20: Back from Add Exercise Screen**
- [x] Click "Add Exercise"
- [x] Click the back arrow (ChevronLeft) in header
- [x] ✅ Should return to main workout form
- [x] ✅ URL should be `/workouts/new` (no params)
- [x] ✅ No exercise should be added

---

### 🎨 Edge Cases

**Test Case 21: Race Condition - Delete Then Navigate**
- [x] Add 3 exercises
- [x] Open exercise at index 2 in detail view
- [x] Use back button to return to main view
- [x] Delete the first two exercises
- [x] Click browser forward button
- [x] ✅ Should handle gracefully (either show valid exercise or redirect to main)

**Test Case 22: Empty Workout Navigation**
- [x] Start with empty workout form
- [x] Try navigating to `?exerciseIndex=0` manually
- [x] ✅ Should redirect to main view
- [x] ✅ No errors in console

**Test Case 23: Multiple Quick Navigation Clicks**
- [ ] Add 3 exercises
- [ ] Rapidly click on different exercise cards
- [ ] ✅ Navigation should be smooth
- [ ] ✅ URL should update correctly
- [ ] ✅ Correct exercise should be displayed

---

## Expected Behaviors

### ✅ What Should Work

1. **Browser Back/Forward Buttons**: Should navigate between views correctly
2. **URL Deep Links**: Copying and pasting URLs should work
3. **Page Refresh**: Should preserve the current view (if data exists in localStorage)
4. **URL Updates**: URL should always reflect the current view state
5. **Form Persistence**: In-progress workout data persists via localStorage
6. **Validation**: Invalid exercise indices should redirect to main view silently

### ❌ Known Limitations

1. **Deep Links Without Data**: Opening `?exerciseIndex=0` without localStorage data will redirect to main
2. **URL Pollution**: URLs now contain search params (this is intentional and beneficial)

---

## Troubleshooting

### Issue: Tests Failing on Navigation
**Solution**: Update test assertions to use `getByRole('link')` instead of `getByRole('button')` for navigation elements

### Issue: URL Not Updating
**Solution**: Check browser console for errors. Ensure TanStack Router is configured correctly.

### Issue: Back Button Not Working
**Solution**: Verify that all navigation uses either `<Link>` components or `navigate({ search: {...} })`

### Issue: Exercise Index Out of Bounds
**Solution**: Validation should catch this and redirect. Check lines 125-129 in `new.tsx`

---

## Next Steps (Optional)

### Future E2E Test Cases to Add

If you want to add automated tests for the new URL features:

```typescript
// Example test to add to workout-create.test.ts

test('should update URL when navigating to add exercise', async ({ page }) => {
  await page.goto('/workouts/new');
  await page.getByRole('link', { name: /add exercise/i }).click();
  expect(page.url()).toContain('addExercise=true');
});

test('should support browser back button from add exercise', async ({ page }) => {
  await page.goto('/workouts/new');
  await page.getByRole('link', { name: /add exercise/i }).click();
  await page.goBack();
  expect(page.url()).not.toContain('addExercise');
  await expect(page.getByRole('heading', { name: /today's training/i })).toBeVisible();
});

test('should support deep linking to add exercise screen', async ({ page }) => {
  await page.goto('/workouts/new?addExercise=true');
  await expect(page.getByRole('heading', { name: /choose exercise/i })).toBeVisible();
});
```

---

## Summary

**Priority Testing Order:**

1. ⭐ **Manual Testing** - Browser back/forward buttons (most important new feature)
2. ⭐ **Manual Testing** - URL deep linking
3. ⭐ **E2E Tests** - Run existing tests to catch regressions
4. **Optional** - Add new E2E tests for URL features

**Time Estimate:**
- Manual testing: ~15-20 minutes
- E2E test run: ~2-5 minutes
- Total: ~20-25 minutes for comprehensive testing
