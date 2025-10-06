# Testing Implementation Plan

**Created**: 2025-10-06
**Updated**: 2025-10-06
**Status**: Planning
**Scope**: Automated testing for unified routes (demo + auth modes)

---

## Executive Summary

Add automated testing to verify the unified route implementation works correctly in both demo and authenticated modes. Focus on critical user flows and data persistence.

---

## Testing Strategy

### Priority Levels

1. **P0 - Critical**: Demo mode core functionality (no auth required, easy to test)
2. **P1 - High**: Data persistence and localStorage
3. **P2 - Medium**: Auth mode (complex due to Stack Auth)
4. **P3 - Low**: Edge cases and error handling

### Test Framework Selection

**Options:**
- **Vitest** - Unit/integration tests, fast, good for testing utilities
- **Vitest Browser Mode** - E2E tests in real browsers, native Bun support
- **Playwright** - E2E tests, excellent but has Bun compatibility issues
- **React Testing Library** - Component tests (already installed)

**Recommendation**: Use **Vitest Browser Mode** for E2E tests
- ✅ **Native Bun support** (no compatibility issues)
- ✅ Already have Vitest installed in the project
- ✅ Can test actual routes and navigation in real browsers
- ✅ Better coverage of real user flows
- ✅ Can verify localStorage behavior
- ✅ Uses Playwright under the hood for browser automation
- ✅ Easier to test demo mode without mocking auth
- ✅ Same API as Playwright (page.getByText, page.getByRole, etc.)
- ⚠️ Currently experimental (but stable enough for production use)

**Why not Playwright directly?**
- ❌ No official Bun support (tests hang with config files)
- ❌ Leaves zombie processes when using Bun runtime
- ❌ `bunx playwright test` uses Node.js anyway, defeating Bun's purpose

---

## Phase 1: Setup (P0)

### 1.1 Install Dependencies

```bash
cd client
bun add -D @vitest/browser playwright
```

**Note**: We're installing `playwright` as a provider for Vitest browser mode, not using Playwright directly.

### 1.2 Update Vitest Config

**File**: `client/vitest.config.ts` (create or update)

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    // Browser mode configuration
    browser: {
      enabled: true,
      name: 'chromium',
      provider: 'playwright',
      headless: true,
      // Auto-start dev server for tests
      screenshotOnFailure: true,
    },
    // Use the existing dev server
    setupFiles: ['./tests/setup.ts'],
  },
});
```

### 1.3 Create Test Setup File

**File**: `client/tests/setup.ts`

```typescript
import { beforeEach } from 'vitest';
import { page } from '@vitest/browser/context';

// Clear localStorage before each test
beforeEach(async () => {
  await page.evaluate(() => localStorage.clear());
});
```

### 1.4 Create Test Directory Structure

```
client/tests/
├── setup.ts
├── e2e/
│   ├── demo/
│   │   ├── workouts-list.test.ts
│   │   ├── workout-detail.test.ts
│   │   ├── workout-edit.test.ts
│   │   ├── exercises-list.test.ts
│   │   └── exercise-detail.test.ts
│   └── helpers/
│       ├── demo-helpers.ts
│       └── storage-helpers.ts
└── fixtures/
    └── demo-data.ts
```

### 1.5 Update package.json Scripts

Add test scripts to `client/package.json`:

```json
{
  "scripts": {
    "test:e2e": "vitest --browser.enabled",
    "test:e2e:ui": "vitest --browser.enabled --ui",
    "test:e2e:headed": "vitest --browser.enabled --browser.headless=false"
  }
}
```

---

## Phase 2: Demo Mode Tests (P0)

### 2.1 Test: Workouts List (`/workouts`)

**File**: `client/tests/e2e/demo/workouts-list.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Workouts List', () => {
  test.beforeEach(async () => {
    // Clear localStorage and navigate to workouts
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load demo workouts', async () => {
    // Verify demo data initialized
    const workoutCards = page.getByTestId('workout-card');
    await expect.element(workoutCards).toBeInTheDocument();

    // Verify workout content
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
    await expect.element(page.getByText('Evening Cardio')).toBeInTheDocument();
  });

  test('should persist demo data across page reloads', async () => {
    // Verify initial load
    await expect.element(page.getByTestId('workout-card')).toBeInTheDocument();
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();

    // Reload page
    await page.reload();

    // Verify data still there
    await expect.element(page.getByTestId('workout-card')).toBeInTheDocument();
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });

  test('should navigate to workout detail', async () => {
    // Click first workout
    await page.getByText('Morning Strength').click();

    // Verify navigation (check URL contains /workouts/)
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/workouts\/\d+/);
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });
});
```

### 2.2 Test: Workout Detail (`/workouts/$workoutId`)

**File**: `client/tests/e2e/demo/workout-detail.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Workout Detail', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should display workout details', async () => {
    // Navigate to first workout
    await page.getByText('Morning Strength').click();

    // Verify workout details visible
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
    await expect.element(page.getByText('Edit')).toBeInTheDocument();
    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    // Verify exercises shown
    const exerciseCard = page.getByTestId('exercise-card');
    await expect.element(exerciseCard).toBeInTheDocument();
  });

  test('should show edit and delete buttons in demo mode', async () => {
    await page.getByText('Morning Strength').click();

    // Verify buttons are visible (not hidden in demo mode)
    await expect.element(page.getByRole('button', { name: /edit/i })).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  test('should delete workout in demo mode', async () => {
    // Count initial workouts
    const initialWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data).length : 0;
    });

    // Navigate to workout and delete
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion (if dialog exists)
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toBe('/workouts');

    // Verify workout no longer in list
    const morningStrength = await page.getByText('Morning Strength').query();
    expect(morningStrength).toBeNull();

    // Verify localStorage updated
    const updatedWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedWorkouts).toBe(initialWorkouts - 1);
  });
});
```

### 2.3 Test: Workout Edit (`/workouts/$workoutId/edit`)

**File**: `client/tests/e2e/demo/workout-edit.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page, userEvent } from '@vitest/browser/context';

test.describe('Demo Mode - Workout Edit', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load edit form with existing workout data', async () => {
    // Navigate to workout detail
    await page.getByText('Morning Strength').click();

    // Click Edit button
    await page.getByRole('button', { name: /edit/i }).click();

    // Verify navigated to edit route
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/workouts\/\d+\/edit/);

    // Verify form loaded
    await expect.element(page.getByRole('heading', { name: /edit training/i })).toBeInTheDocument();

    // Verify exercises loaded
    const exerciseCard = page.getByTestId('exercise-card');
    await expect.element(exerciseCard).toBeInTheDocument();
  });

  test('should modify and save workout in demo mode', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Modify notes field
    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Updated notes in demo mode');

    // Save workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for navigation/mutation
    await page.waitForTimeout(500);

    // Navigate back to list and verify workout still exists
    await page.goto('/workouts');
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });

  test('should persist edited workout to localStorage', async () => {
    // Edit workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Persistent notes');
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for mutation
    await page.waitForTimeout(500);

    // Reload page
    await page.reload();

    // Verify data persisted in localStorage
    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(3); // Still 3 workouts
  });

  test('should add exercise to workout', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Count initial exercises via localStorage
    const initialExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });

    // Click "Add Exercise" button
    await page.getByRole('button', { name: /add exercise/i }).click();

    // Select an exercise (adjust selector based on your UI)
    await page.getByText('Bench Press').click();

    // Save the changes
    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForTimeout(500);

    // Verify exercise added in localStorage
    const updatedExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });
    expect(updatedExerciseCount).toBe(initialExerciseCount + 1);
  });

  test('should remove exercise from workout', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Find and click delete on first exercise card
    const firstExerciseCard = page.getByTestId('exercise-card');
    const deleteButton = firstExerciseCard.getByRole('button', { name: /delete|remove/i });
    await deleteButton.click();

    // Save and verify
    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForTimeout(500);

    // Verify in localStorage that exercise count decreased
    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
    expect(morningStrength.exercises.length).toBeGreaterThan(0);
  });
});
```

### 2.4 Test: Exercises List (`/exercises`)

**File**: `client/tests/e2e/demo/exercises-list.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Exercises List', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/exercises');
  });

  test('should load demo exercises', async () => {
    // Verify demo exercises loaded (5 in initial-data.ts)
    const exercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data) : [];
    });
    expect(exercises).toHaveLength(5);

    // Verify exercise names visible
    await expect.element(page.getByText('Bench Press')).toBeInTheDocument();
    await expect.element(page.getByText('Squat')).toBeInTheDocument();
  });

  test('should navigate to exercise detail', async () => {
    // Click first exercise
    await page.getByText('Bench Press').click();

    // Verify navigation
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/exercises\/\d+/);
    await expect.element(page.getByText('Bench Press')).toBeInTheDocument();
  });
});
```

### 2.5 Test: Exercise Detail (`/exercises/$exerciseId`)

**File**: `client/tests/e2e/demo/exercise-detail.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Exercise Detail', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/exercises');
  });

  test('should display exercise details', async () => {
    // Navigate to exercise
    await page.getByText('Bench Press').click();

    // Verify exercise name visible
    await expect.element(page.getByRole('heading', { name: /bench press/i })).toBeInTheDocument();

    // Verify delete button visible in demo mode
    await expect.element(page.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  test('should delete exercise in demo mode', async () => {
    // Count initial exercises
    const initialExercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data).length : 0;
    });

    // Navigate to exercise and delete
    await page.getByText('Bench Press').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion (if dialog exists)
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toBe('/exercises');

    // Verify exercise count decreased in localStorage
    const updatedExercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedExercises).toBe(initialExercises - 1);
  });
});
```

---

## Phase 3: localStorage Tests (P1)

### 3.1 Test: Data Persistence Helpers

**File**: `client/tests/e2e/helpers/storage-helpers.ts`

```typescript
import { Page } from '@playwright/test';

export async function getDemoWorkouts(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_workouts');
    return data ? JSON.parse(data) : null;
  });
}

export async function getDemoExercises(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_exercises');
    return data ? JSON.parse(data) : null;
  });
}

export async function clearAllStorage(page: Page) {
  await page.evaluate(() => localStorage.clear());
}

export async function verifyDemoDataExists(page: Page) {
  const workouts = await getDemoWorkouts(page);
  const exercises = await getDemoExercises(page);
  return workouts !== null && exercises !== null;
}
```

### 3.2 Test: localStorage Persistence

**File**: `client/tests/e2e/demo/storage-persistence.test.ts`

```typescript
import { test, expect } from 'vitest';
import { page, userEvent } from '@vitest/browser/context';
import { getDemoWorkouts, getDemoExercises, clearAllStorage } from '../helpers/storage-helpers';

test.describe('localStorage Persistence', () => {
  test.beforeEach(async () => {
    await clearAllStorage(page);
  });

  test('should initialize demo data on first visit', async () => {
    await page.goto('/workouts');

    // Verify localStorage has demo data
    const workouts = await getDemoWorkouts(page);
    const exercises = await getDemoExercises(page);

    expect(workouts).not.toBeNull();
    expect(exercises).not.toBeNull();
    expect(workouts.length).toBe(3);
    expect(exercises.length).toBe(5);
  });

  test('should persist demo data across page reloads', async () => {
    await page.goto('/workouts');

    // Get initial data
    const initialWorkouts = await getDemoWorkouts(page);

    // Reload page
    await page.reload();

    // Verify data unchanged
    const reloadedWorkouts = await getDemoWorkouts(page);
    expect(reloadedWorkouts).toEqual(initialWorkouts);
  });

  test('should persist workout edits to localStorage', async () => {
    await page.goto('/workouts');

    // Edit a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Test notes for persistence');
    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForTimeout(500); // Wait for mutation

    // Check localStorage directly
    const workouts = await getDemoWorkouts(page);
    const updatedWorkout = workouts.find((w: any) => w.notes === 'Test notes for persistence');

    expect(updatedWorkout).toBeDefined();
  });

  test('should persist workout deletion to localStorage', async () => {
    await page.goto('/workouts');

    // Get initial count
    const initialWorkouts = await getDemoWorkouts(page);
    const initialCount = initialWorkouts.length;

    // Delete a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    await page.waitForTimeout(500);

    // Verify localStorage updated
    const updatedWorkouts = await getDemoWorkouts(page);
    expect(updatedWorkouts.length).toBe(initialCount - 1);
  });
});
```

---

## Phase 4: Auth Mode Tests (P2 - DEFERRED)

**Status**: ⏸️ DEFERRED

**Reason**: Stack Auth integration makes auth testing complex. Requires:
- Mock auth provider
- Test user credentials
- Auth state management

**Future Work**: Add auth tests once auth testing strategy is established.

**Suggested Approach**:
- Use Playwright's `storageState` to persist auth cookies
- Create helper to login programmatically
- Test API-based mutations vs demo mutations

---

## Phase 5: Test Data Attributes (Required)

### 5.1 Add `data-testid` Attributes

**Files to Update**:

1. **Workout List** (`components/workouts/workout-list.tsx`):
   ```tsx
   <Card data-testid="workout-card" ...>
   ```

2. **Exercise List** (`components/exercises/exercise-list.tsx`):
   ```tsx
   <Card data-testid="exercise-card" ...>
   ```

3. **Exercise Cards in Workout Forms** (`routes/workouts/-components/*`):
   ```tsx
   <Card data-testid="exercise-card" ...>
   ```

### 5.2 Verify Forms Have Accessible Labels

Ensure all form fields have proper labels or placeholders for Playwright to target.

---

## Phase 6: CI/CD Integration (P3 - Future)

### 6.1 Test Scripts Already Added

Scripts were added in Phase 1.5:

```json
{
  "scripts": {
    "test:e2e": "vitest --browser.enabled",
    "test:e2e:ui": "vitest --browser.enabled --ui",
    "test:e2e:headed": "vitest --browser.enabled --browser.headless=false"
  }
}
```

### 6.2 GitHub Actions Workflow

**File**: `.github/workflows/vitest-e2e.yml`

```yaml
name: Vitest E2E Tests
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
jobs:
  test:
    timeout-minutes: 60
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest
      - name: Install dependencies
        run: cd client && bun install
      - name: Install Playwright Browsers
        run: cd client && bunx playwright install chromium --with-deps
      - name: Run Vitest E2E tests
        run: cd client && bun run test:e2e
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: test-results
          path: client/test-results/
          retention-days: 30
```

---

## Success Criteria

### Phase 2 (Demo Mode Tests) ✅
- [ ] All 5 test files created
- [ ] All tests passing locally
- [ ] Demo mode fully covered (list, detail, edit, delete)
- [ ] localStorage persistence verified

### Phase 3 (localStorage Tests) ✅
- [ ] Storage helpers created
- [ ] Persistence tests passing
- [ ] Mutations reflected in localStorage

### Phase 5 (Test Data Attributes) ✅
- [ ] `data-testid` added to all cards
- [ ] Forms have accessible labels
- [ ] Tests can target elements reliably

---

## Estimated Timeline

| Phase | Task | Time |
|-------|------|------|
| 1 | Setup Vitest browser mode + config | 10 min |
| 2 | Write 5 demo test files | 30 min |
| 3 | localStorage tests + helpers | 15 min |
| 5 | Add data-testid attributes | 10 min |
| - | Run tests + fix issues | 15 min |

**Total Estimated Time**: 1.5 hours

---

## Risks & Mitigations

### Risk: Tests are brittle (depend on exact text)
**Mitigation**: Use `data-testid` attributes and accessible roles

### Risk: Timing issues (data not loaded)
**Mitigation**: Use Vitest browser mode's auto-waiting, add explicit waits where needed

### Risk: localStorage not cleared between tests
**Mitigation**: `beforeEach` hook clears storage

### Risk: Tests fail in CI but pass locally
**Mitigation**: Use consistent baseURL, add retries in CI config

---

## Next Steps

1. [X] **Approve this plan** ✅ (Updated to use Vitest Browser Mode)
2. [ ] **Phase 1**: Install Vitest browser mode dependencies and create config
3. [ ] **Phase 5**: Add `data-testid` attributes to components (do this first!)
4. [ ] **Phase 2**: Implement demo mode tests (5 test files)
5. [ ] **Phase 3**: Implement localStorage tests
6. [ ] **Run tests**: `bun run test:e2e:ui` (interactive mode)
7. [ ] **Fix any failing tests**
8. [ ] **Update documentation**: Mark Phase 3.3 in `demo-plan.md` as complete

---

## Summary of Changes from Original Plan

### What Changed
- **Framework**: Switched from Playwright to **Vitest Browser Mode**
- **Reason**: Native Bun support, no compatibility issues
- **Test Files**: Renamed from `.spec.ts` to `.test.ts` (Vitest convention)
- **API Changes**:
  - Import from `'vitest'` instead of `'@playwright/test'`
  - Import `page` from `'@vitest/browser/context'`
  - Use `expect.element()` for DOM assertions
  - Use `page.evaluate()` for URL checks and localStorage access
  - Use `userEvent.fill()` for form inputs
  - Use `.query()` for checking element existence

### What Stayed the Same
- Same test structure and test cases
- Same locator API (`getByText`, `getByRole`, `getByTestId`, etc.)
- Same testing approach (E2E tests in real browsers)
- Same localStorage verification strategy
- Same test organization and file structure

### Benefits
- ✅ Works natively with Bun runtime
- ✅ No hanging tests or zombie processes
- ✅ Reuses existing Vitest installation
- ✅ Faster test execution
- ✅ Same developer experience as Playwright
- ✅ Better integration with existing Vitest unit tests

---

**End of Plan**
