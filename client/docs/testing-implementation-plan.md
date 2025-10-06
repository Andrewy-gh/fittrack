# Testing Implementation Plan

**Created**: 2025-10-06
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
- **Playwright** - E2E tests, can test full user flows, better for route testing
- **React Testing Library** - Component tests

**Recommendation**: Start with **Playwright** for E2E tests
- Can test actual routes and navigation
- Better coverage of real user flows
- Can verify localStorage behavior
- Easier to test demo mode without mocking auth

---

## Phase 1: Setup (P0)

### 1.1 Install Dependencies

```bash
cd client
bun add -D @playwright/test
bunx playwright install
```

### 1.2 Create Playwright Config

**File**: `client/playwright.config.ts`

```typescript
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
```

### 1.3 Create Test Directory Structure

```
client/tests/
├── e2e/
│   ├── demo/
│   │   ├── workouts-list.spec.ts
│   │   ├── workout-detail.spec.ts
│   │   ├── workout-edit.spec.ts
│   │   ├── exercises-list.spec.ts
│   │   └── exercise-detail.spec.ts
│   └── helpers/
│       ├── demo-helpers.ts
│       └── storage-helpers.ts
└── fixtures/
    └── demo-data.ts
```

---

## Phase 2: Demo Mode Tests (P0)

### 2.1 Test: Workouts List (`/workouts`)

**File**: `client/tests/e2e/demo/workouts-list.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workouts List', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage and start fresh
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should load demo workouts', async ({ page }) => {
    await page.goto('/workouts');

    // Verify demo data initialized
    const workoutCards = page.locator('[data-testid="workout-card"]');
    await expect(workoutCards).toHaveCount(3); // 3 demo workouts

    // Verify workout content
    await expect(page.getByText('Morning Strength')).toBeVisible();
    await expect(page.getByText('Evening Cardio')).toBeVisible();
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await page.goto('/workouts');

    // Verify initial load
    const workoutCards = page.locator('[data-testid="workout-card"]');
    await expect(workoutCards).toHaveCount(3);

    // Reload page
    await page.reload();

    // Verify data still there
    await expect(workoutCards).toHaveCount(3);
    await expect(page.getByText('Morning Strength')).toBeVisible();
  });

  test('should navigate to workout detail', async ({ page }) => {
    await page.goto('/workouts');

    // Click first workout
    await page.getByText('Morning Strength').click();

    // Verify navigation
    await expect(page).toHaveURL(/\/workouts\/\d+/);
    await expect(page.getByText('Morning Strength')).toBeVisible();
  });
});
```

### 2.2 Test: Workout Detail (`/workouts/$workoutId`)

**File**: `client/tests/e2e/demo/workout-detail.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Detail', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should display workout details', async ({ page }) => {
    await page.goto('/workouts');

    // Navigate to first workout
    await page.getByText('Morning Strength').click();

    // Verify workout details visible
    await expect(page.getByText('Morning Strength')).toBeVisible();
    await expect(page.getByText('Edit')).toBeVisible();
    await expect(page.getByText('Delete')).toBeVisible();

    // Verify exercises shown
    const exerciseCards = page.locator('[data-testid="exercise-card"]');
    await expect(exerciseCards.first()).toBeVisible();
  });

  test('should show edit and delete buttons in demo mode', async ({ page }) => {
    await page.goto('/workouts');
    await page.getByText('Morning Strength').click();

    // Verify buttons are visible (not hidden in demo mode)
    await expect(page.getByRole('button', { name: /edit/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /delete/i })).toBeVisible();
  });

  test('should delete workout in demo mode', async ({ page }) => {
    await page.goto('/workouts');

    // Count initial workouts
    let workoutCards = page.locator('[data-testid="workout-card"]');
    const initialCount = await workoutCards.count();

    // Navigate to workout and delete
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion (if dialog exists)
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    if (await confirmButton.isVisible()) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    await expect(page).toHaveURL('/workouts');

    // Verify workout count decreased
    workoutCards = page.locator('[data-testid="workout-card"]');
    await expect(workoutCards).toHaveCount(initialCount - 1);

    // Verify workout no longer in list
    await expect(page.getByText('Morning Strength')).not.toBeVisible();
  });
});
```

### 2.3 Test: Workout Edit (`/workouts/$workoutId/edit`)

**File**: `client/tests/e2e/demo/workout-edit.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Edit', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should load edit form with existing workout data', async ({ page }) => {
    await page.goto('/workouts');

    // Navigate to workout detail
    await page.getByText('Morning Strength').click();

    // Click Edit button
    await page.getByRole('button', { name: /edit/i }).click();

    // Verify navigated to edit route
    await expect(page).toHaveURL(/\/workouts\/\d+\/edit/);

    // Verify form loaded
    await expect(page.getByRole('heading', { name: /edit training/i })).toBeVisible();

    // Verify exercises loaded
    const exerciseCards = page.locator('[data-testid="exercise-card"]');
    await expect(exerciseCards.first()).toBeVisible();
  });

  test('should modify and save workout in demo mode', async ({ page }) => {
    await page.goto('/workouts');

    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Modify notes field
    const notesField = page.getByLabel(/notes/i).or(page.getByPlaceholder(/notes/i));
    await notesField.fill('Updated notes in demo mode');

    // Save workout
    await page.getByRole('button', { name: /save/i }).click();

    // Verify saved (should navigate to detail or show success)
    await page.waitForTimeout(500); // Wait for mutation

    // Navigate back to detail
    await page.goto('/workouts');
    await page.getByText('Morning Strength').click();

    // Verify notes updated (if visible in detail view)
    // This depends on your UI - adjust accordingly
  });

  test('should persist edited workout to localStorage', async ({ page }) => {
    await page.goto('/workouts');

    // Edit workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i).or(page.getByPlaceholder(/notes/i));
    await notesField.fill('Persistent notes');
    await page.getByRole('button', { name: /save/i }).click();

    // Reload page
    await page.reload();

    // Verify data persisted
    await page.goto('/workouts');
    const workoutCards = page.locator('[data-testid="workout-card"]');
    await expect(workoutCards).toHaveCount(3); // Still 3 workouts
  });

  test('should add exercise to workout', async ({ page }) => {
    await page.goto('/workouts');

    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Count initial exercises
    const exerciseCards = page.locator('[data-testid="exercise-card"]');
    const initialCount = await exerciseCards.count();

    // Click "Add Exercise" button
    await page.getByRole('button', { name: /add exercise/i }).click();

    // Select an exercise (adjust selector based on your UI)
    await page.getByText('Bench Press').click();

    // Verify exercise added
    await expect(exerciseCards).toHaveCount(initialCount + 1);
  });

  test('should remove exercise from workout', async ({ page }) => {
    await page.goto('/workouts');

    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Count initial exercises
    const exerciseCards = page.locator('[data-testid="exercise-card"]');
    const initialCount = await exerciseCards.count();

    // Click delete on first exercise
    const deleteButton = exerciseCards.first().getByRole('button', { name: /delete|remove/i });
    await deleteButton.click();

    // Verify exercise removed
    await expect(exerciseCards).toHaveCount(initialCount - 1);
  });
});
```

### 2.4 Test: Exercises List (`/exercises`)

**File**: `client/tests/e2e/demo/exercises-list.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Exercises List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/exercises');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should load demo exercises', async ({ page }) => {
    await page.goto('/exercises');

    // Verify demo exercises loaded (5 in initial-data.ts)
    const exerciseCards = page.locator('[data-testid="exercise-card"]');
    await expect(exerciseCards).toHaveCount(5);

    // Verify exercise names
    await expect(page.getByText('Bench Press')).toBeVisible();
    await expect(page.getByText('Squat')).toBeVisible();
  });

  test('should navigate to exercise detail', async ({ page }) => {
    await page.goto('/exercises');

    // Click first exercise
    await page.getByText('Bench Press').click();

    // Verify navigation
    await expect(page).toHaveURL(/\/exercises\/\d+/);
    await expect(page.getByText('Bench Press')).toBeVisible();
  });
});
```

### 2.5 Test: Exercise Detail (`/exercises/$exerciseId`)

**File**: `client/tests/e2e/demo/exercise-detail.spec.ts`

```typescript
import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Exercise Detail', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/exercises');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should display exercise details', async ({ page }) => {
    await page.goto('/exercises');

    // Navigate to exercise
    await page.getByText('Bench Press').click();

    // Verify exercise name visible
    await expect(page.getByRole('heading', { name: /bench press/i })).toBeVisible();

    // Verify delete button visible in demo mode
    await expect(page.getByRole('button', { name: /delete/i })).toBeVisible();
  });

  test('should delete exercise in demo mode', async ({ page }) => {
    await page.goto('/exercises');

    // Count initial exercises
    let exerciseCards = page.locator('[data-testid="exercise-card"]');
    const initialCount = await exerciseCards.count();

    // Navigate to exercise and delete
    await page.getByText('Bench Press').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    if (await confirmButton.isVisible()) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    await expect(page).toHaveURL('/exercises');

    // Verify exercise count decreased
    exerciseCards = page.locator('[data-testid="exercise-card"]');
    await expect(exerciseCards).toHaveCount(initialCount - 1);
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

**File**: `client/tests/e2e/demo/storage-persistence.spec.ts`

```typescript
import { test, expect } from '@playwright/test';
import { getDemoWorkouts, getDemoExercises, clearAllStorage } from '../helpers/storage-helpers';

test.describe('localStorage Persistence', () => {
  test.beforeEach(async ({ page }) => {
    await clearAllStorage(page);
  });

  test('should initialize demo data on first visit', async ({ page }) => {
    await page.goto('/workouts');

    // Verify localStorage has demo data
    const workouts = await getDemoWorkouts(page);
    const exercises = await getDemoExercises(page);

    expect(workouts).not.toBeNull();
    expect(exercises).not.toBeNull();
    expect(workouts.length).toBe(3);
    expect(exercises.length).toBe(5);
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await page.goto('/workouts');

    // Get initial data
    const initialWorkouts = await getDemoWorkouts(page);

    // Reload page
    await page.reload();

    // Verify data unchanged
    const reloadedWorkouts = await getDemoWorkouts(page);
    expect(reloadedWorkouts).toEqual(initialWorkouts);
  });

  test('should persist workout edits to localStorage', async ({ page }) => {
    await page.goto('/workouts');

    // Edit a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i).or(page.getByPlaceholder(/notes/i));
    await notesField.fill('Test notes for persistence');
    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForTimeout(500); // Wait for mutation

    // Check localStorage directly
    const workouts = await getDemoWorkouts(page);
    const updatedWorkout = workouts.find((w: any) => w.notes === 'Test notes for persistence');

    expect(updatedWorkout).toBeDefined();
  });

  test('should persist workout deletion to localStorage', async ({ page }) => {
    await page.goto('/workouts');

    // Get initial count
    const initialWorkouts = await getDemoWorkouts(page);
    const initialCount = initialWorkouts.length;

    // Delete a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    if (await confirmButton.isVisible()) {
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

### 6.1 Add Test Script to `package.json`

```json
{
  "scripts": {
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:e2e:debug": "playwright test --debug"
  }
}
```

### 6.2 GitHub Actions Workflow

**File**: `.github/workflows/playwright.yml`

```yaml
name: Playwright Tests
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
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - name: Install dependencies
        run: cd client && bun install
      - name: Install Playwright Browsers
        run: cd client && bunx playwright install --with-deps
      - name: Run Playwright tests
        run: cd client && bun run test:e2e
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: client/playwright-report/
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
| 1 | Setup Playwright + config | 10 min |
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
**Mitigation**: Use Playwright's auto-waiting, add explicit waits where needed

### Risk: localStorage not cleared between tests
**Mitigation**: `beforeEach` hook clears storage

### Risk: Tests fail in CI but pass locally
**Mitigation**: Use consistent baseURL, add retries in CI config

---

## Next Steps

1. **Approve this plan**
2. **Phase 1**: Install Playwright and create config
3. **Phase 5**: Add `data-testid` attributes to components (do this first!)
4. **Phase 2**: Implement demo mode tests
5. **Phase 3**: Implement localStorage tests
6. **Run tests**: `bun run test:e2e:ui` (interactive mode)
7. **Fix any failing tests**
8. **Update documentation**: Mark Phase 3.3 in `demo-plan.md` as complete

---

**End of Plan**
