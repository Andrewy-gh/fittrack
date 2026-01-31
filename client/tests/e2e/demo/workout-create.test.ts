import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Create', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate first, then clear localStorage and reload to reinitialize demo data
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();

    // Wait for demo data to be initialized and workouts to be visible
    await page.waitForSelector('[data-testid="workout-card"]', {
      timeout: 5000,
    });
  });

  test('should navigate to new workout page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    expect(page.url()).toMatch(/\/workouts\/new/);

    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();
  });

  test('should display empty workout form on new page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await expect(
      page.getByRole('link', { name: /add exercise/i })
    ).toBeVisible();
    await expect(page.getByRole('button', { name: /save/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /clear/i })).toBeVisible();

    const exerciseCards = page
      .locator('[data-testid*="exercise"]')
      .filter({ hasText: /sets|volume/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should create a workout with single exercise and single set', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();

    await page.getByText('Bench Press', { exact: true }).click();

    await page.getByRole('button', { name: /add set/i }).click();

    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill in set details (weight and reps use spinbutton role for number inputs)
    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);

    await weightInput.fill('135');
    await repsInput.fill('10');

    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /back/i }).click();

    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');

    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should not add empty sets when dialog is dismissed', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();

    await page.getByRole('button', { name: /close/i }).click();
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(page.locator('[data-testid="exercise-card"]')).toHaveCount(0);

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();

    await page
      .locator('[data-slot="dialog-overlay"]')
      .click({ position: { x: 10, y: 10 } });
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(page.locator('[data-testid="exercise-card"]')).toHaveCount(0);
  });

  test('should keep set when only set type changes', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.locator('[data-slot="select-trigger"]').click();
    await page
      .locator('[data-slot="select-item"]', { hasText: /warmup/i })
      .click();
    await expect(
      page.getByRole('button', { name: /remove set/i })
    ).toBeVisible();

    await page.getByRole('button', { name: /close/i }).click();
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(page.locator('[data-testid="exercise-card"]')).toHaveCount(1);
  });

  test('should persist created workout to localStorage', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Deadlift', { exact: true }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);

    await weightInput.fill('225');
    await repsInput.fill('5');

    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    await page.getByRole('button', { name: /save/i }).click();

    // Wait for save to complete
    await page.waitForTimeout(500);

    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(4);
  });

  test('should create workout with multiple exercises', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    let weightInput = page.getByRole('spinbutton').first();
    let repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Barbell Squat', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('185');
    await repsInput.fill('8');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    let exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(2);

    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    exerciseCards = page.locator('.cursor-pointer').filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should create workout with multiple sets for same exercise', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Pull-ups', { exact: true }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    let weightInput = page.getByRole('spinbutton').first();
    let repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('8');
    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('6');
    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /back/i }).click();

    const exerciseCard = page.getByTestId('new-workout-exercise-card');
    await expect(exerciseCard.getByText('3')).toBeVisible();
    await expect(exerciseCard.getByText('sets', { exact: true })).toBeVisible();
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should clear form data when clear button is clicked', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(1);

    page.on('dialog', (dialog) => dialog.accept());

    await page.getByRole('button', { name: /clear/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(200);

    await expect(exerciseCards).toHaveCount(0);
  });

  test('should add workout with notes and focus', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByTestId('notes-card').click();

    await expect(page.getByRole('dialog')).toBeVisible();

    const notesTextarea = page.getByTestId('notes-textarea');
    await notesTextarea.fill('Great workout today!');

    await page.getByTestId('notes-close').click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  // URL Navigation Tests
  test('should update URL when navigating to add exercise', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();
    await page.getByRole('link', { name: /add exercise/i }).click();
    expect(page.url()).toContain('addExercise=true');
  });

  test('should support browser back button from add exercise', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.goBack();
    expect(page.url()).not.toContain('addExercise');
    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();
  });

  test('should support deep linking to add exercise screen', async ({
    page,
  }) => {
    await page.goto('/workouts/new?addExercise=true');
    await expect(
      page.getByRole('heading', { name: /choose exercise/i })
    ).toBeVisible();
  });

  test('should update URL when navigating to exercise detail', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    expect(page.url()).toContain('exerciseIndex=0');
  });

  test('should support browser back button from exercise detail', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    await page.goBack();

    expect(page.url()).toContain('addExercise=true');
    expect(page.url()).not.toContain('exerciseIndex');
    await expect(
      page.getByRole('heading', { name: /choose exercise/i })
    ).toBeVisible();
  });

  test('should support browser forward button', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    expect(page.url()).toContain('addExercise=true');

    await page.goBack();
    expect(page.url()).not.toContain('addExercise');

    await page.goForward();
    expect(page.url()).toContain('addExercise=true');
    await expect(
      page.getByRole('heading', { name: /choose exercise/i })
    ).toBeVisible();
  });

  test('should support deep linking to exercise detail with localStorage', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /back/i }).click();

    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await exerciseCards.first().click();
    const detailUrl = page.url();

    await page.goto(detailUrl);

    await expect(page.getByText('Bench Press')).toBeVisible();
  });

  test('should handle invalid exercise index gracefully', async ({ page }) => {
    await page.goto('/workouts/new?exerciseIndex=999');

    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();

    expect(page.url()).not.toContain('exerciseIndex');
  });

  test('should handle negative exercise index gracefully', async ({ page }) => {
    await page.goto('/workouts/new?exerciseIndex=-1');

    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();

    expect(page.url()).not.toContain('exerciseIndex');
  });

  test('should preserve URL state on page refresh - add exercise screen', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();
    await page.getByRole('link', { name: /add exercise/i }).click();

    await page.reload();

    expect(page.url()).toContain('addExercise=true');
    await expect(
      page.getByRole('heading', { name: /choose exercise/i })
    ).toBeVisible();
  });

  test('should preserve URL state on page refresh - exercise detail', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    // Wait for localStorage to be updated (debounced at 500ms)
    await page.waitForTimeout(600);

    await page.reload();

    expect(page.url()).toContain('exerciseIndex=0');
    await expect(page.getByText('Bench Press')).toBeVisible();
  });
});
