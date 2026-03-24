import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Create', () => {
  const lowerBodyWorkout = (page: import('@playwright/test').Page) =>
    page.getByRole('link', { name: /lower body/i }).first();
  const exerciseCards = (page: import('@playwright/test').Page) =>
    page.getByTestId('new-workout-exercise-card');
  const addExercise = (page: import('@playwright/test').Page) =>
    page.getByRole('link', { name: /add exercise/i });
  const saveWorkout = (page: import('@playwright/test').Page) =>
    page.getByRole('button', { name: /^save$/i });

  test.beforeEach(async ({ page }) => {
    // Navigate first, then clear localStorage and reload to reinitialize demo data
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();

    // Wait for demo data to be initialized and workouts to be visible
    await expect(lowerBodyWorkout(page)).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to new workout page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await expect(page).toHaveURL(/\/workouts\/new/);

    await expect(page.getByRole('button', { name: /^save$/i })).toBeVisible();
  });

  test('should display empty workout form on new page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await expect(saveWorkout(page)).toBeVisible();
    await expect(addExercise(page)).toBeVisible();
    await expect(page.getByRole('button', { name: /clear/i })).toBeVisible();

    await expect(exerciseCards(page)).toHaveCount(0);
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

    await saveWorkout(page).click();

    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');

    await expect(exerciseCards(page)).toHaveCount(0);
  });

  test('should discard a new exercise when its empty set dialog is closed', async ({
    page,
  }) => {
    const openNewExercise = async () => {
      await page.getByRole('link', { name: /add exercise/i }).click();
      await expect(
        page.getByRole('heading', { name: /choose exercise/i })
      ).toBeVisible();

      await page.getByRole('button', { name: /bench press/i }).first().click();

      await expect(page).toHaveURL(/exerciseIndex=\d+/);
      await expect(
        page.getByRole('heading', { name: 'Bench Press', exact: true })
      ).toBeVisible();
      await expect(page.getByRole('button', { name: /add set/i })).toBeVisible();
    };

    await page.getByRole('link', { name: /new workout/i }).click();

    await openNewExercise();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();

    await page.getByRole('button', { name: /close/i }).click();
    await expect(page.getByRole('dialog')).toHaveCount(0);
    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();
    await expect(page).toHaveURL(/\/workouts\/new$/);
    await expect(exerciseCards(page)).toHaveCount(0);
  });

  test.fixme('overlay dismissal should discard an empty new-exercise set', async () => {
    // TODO(issue-81): overlay dismissal currently preserves the zeroed set.
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
    await expect(page.getByTestId('exercise-card')).toHaveCount(1);
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

    await saveWorkout(page).click();

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

    await expect(exerciseCards(page)).toHaveCount(2);

    await saveWorkout(page).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    await expect(exerciseCards(page)).toHaveCount(0);
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

    await expect(exerciseCards(page).filter({ hasText: /pull-ups/i })).toBeVisible();
    await saveWorkout(page).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    await expect(exerciseCards(page)).toHaveCount(0);
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

    await expect(exerciseCards(page)).toHaveCount(1);

    await page.getByRole('button', { name: /clear/i }).click();
    await page.getByRole('button', { name: /clear draft/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(200);

    await expect(exerciseCards(page)).toHaveCount(0);
  });

  test('should add workout with notes and focus', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    await page.getByRole('button', { name: /^notes$/i }).click();

    await expect(page.getByRole('dialog')).toBeVisible();

    const notesDialog = page.getByRole('dialog', { name: /^notes$/i });
    const notesTextarea = notesDialog.getByRole('textbox', {
      name: /^notes$/i,
    });
    await notesTextarea.fill('Great workout today!');

    await notesDialog.getByRole('button', { name: /^close$/i }).first().click();

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

    await saveWorkout(page).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    expect(page.url()).toContain('/new');
    await expect(exerciseCards(page)).toHaveCount(0);
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

    await exerciseCards(page).first().click();
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

    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();

    // Wait for localStorage to be updated (debounced at 500ms)
    await page.waitForTimeout(600);

    await page.reload();

    expect(page.url()).toContain('exerciseIndex=0');
    await expect(page.getByText('Bench Press')).toBeVisible();
  });
});
