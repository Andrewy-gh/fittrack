import { test, expect, type Locator, type Page } from '@playwright/test';

async function dragHandleToCardTop(page: Page, handle: Locator, targetCard: Locator) {
  const handleBox = await handle.boundingBox();
  const targetBox = await targetCard.boundingBox();

  if (!handleBox || !targetBox) {
    throw new Error('Missing drag handle or target card box');
  }

  await page.mouse.move(
    handleBox.x + handleBox.width / 2,
    handleBox.y + handleBox.height / 2
  );
  await page.mouse.down();
  await page.mouse.move(
    handleBox.x + handleBox.width / 2,
    handleBox.y + handleBox.height / 2 - 12,
    { steps: 4 }
  );
  await page.mouse.move(
    targetBox.x + targetBox.width / 2,
    targetBox.y + targetBox.height / 3,
    { steps: 16 }
  );
  await page.mouse.up();
}

test.describe('Demo Mode - Workout Edit', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate first, then clear localStorage and reload to reinitialize demo data
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();

    // Wait for demo data to be initialized and workouts to be visible
    await page.waitForSelector('[data-testid="workout-card"]', { timeout: 5000 });
  });

  test('should load edit form with existing workout data', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();

    await page.getByRole('link', { name: /edit/i }).click();

    expect(page.url()).toMatch(/\/workouts\/\d+\/edit/);

    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();

    const exerciseCard = page.getByTestId('edit-workout-exercise-card');
    await expect(exerciseCard.first()).toBeVisible();
  });

  test('should modify and save workout in demo mode', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByTestId('notes-card').click();

    const notesField = page.getByTestId('notes-textarea');
    await notesField.fill('Updated notes in demo mode');

    await page.getByTestId('notes-close').click();

    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.goto('/workouts');

    // Wait for the save to complete and data to be persisted
    await page.waitForTimeout(500);

    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
  });

  test('should persist edited workout to localStorage', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByTestId('notes-card').click();
    
    const notesField = page.getByTestId('notes-textarea');
    await notesField.fill('Test notes for persistence');

    await page.getByTestId('notes-close').click();

    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.reload();

    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(3);
  });

  test('should add exercise to workout', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    // Wait for exercises to load
    await expect(
      page.getByTestId('edit-workout-exercise-card').first()
    ).toBeVisible();

    const initialExerciseCount = await page
      .getByTestId('edit-workout-exercise-card')
      .count();

    await page.getByRole('button', { name: /add exercise/i }).click();

    // Add an exercise that doesn't exist in workout 3 (which has Barbell Squat, Deadlift, Pull-ups)
    await page.getByText('Bench Press').click();

    await page.getByRole('button', { name: /add set/i }).click();

    // Wait for the dialog to be visible
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);

    await weightInput.fill('185');
    await repsInput.fill('5');

    await page.getByRole('button', { name: /save set/i }).click();

    await page.getByRole('button', { name: /back/i }).click();

    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.getByRole('link', { name: /edit/i }).click();

    // Wait for exercises to load after navigation
    await expect(
      page.getByTestId('edit-workout-exercise-card').first()
    ).toBeVisible();

    await expect(page.getByTestId('edit-workout-exercise-card')).toHaveCount(
      initialExerciseCount + 1
    );
  });

  test('should remove exercise from workout', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    // Wait for exercises to load
    await expect(
      page.getByTestId('edit-workout-exercise-card').first()
    ).toBeVisible();

    const initialExerciseCount = await page
      .getByTestId('edit-workout-exercise-card')
      .count();

    const firstExerciseCard = page
      .getByTestId('edit-workout-exercise-card')
      .first();
    const deleteButton = firstExerciseCard.getByRole('button', {
      name: /delete|remove/i,
    });
    await deleteButton.click();

    // Wait for the exercise to be removed from the DOM
    await expect(page.getByTestId('edit-workout-exercise-card')).toHaveCount(
      initialExerciseCount - 1
    );

    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.getByRole('link', { name: /edit/i }).click();

    // Wait for exercises to load after navigation
    await expect(
      page.getByTestId('edit-workout-exercise-card').first()
    ).toBeVisible();

    await expect(page.getByTestId('edit-workout-exercise-card')).toHaveCount(
      initialExerciseCount - 1
    );
  });

  test('should reorder exercises and persist the saved order', async ({
    page,
  }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    const exerciseCards = page.getByTestId('edit-workout-exercise-card');
    await expect(exerciseCards.first()).toBeVisible();
    const initialExerciseCount = await exerciseCards.count();
    expect(initialExerciseCount).toBeGreaterThan(1);

    const firstExerciseName = await exerciseCards
      .nth(0)
      .locator('span.text-primary')
      .first()
      .textContent();
    const secondExerciseName = await exerciseCards
      .nth(1)
      .locator('span.text-primary')
      .first()
      .textContent();

    expect(firstExerciseName).toBeTruthy();
    expect(secondExerciseName).toBeTruthy();

    await page.getByTestId('edit-exercise-order').click();

    const dragHandles = page.getByTestId('edit-workout-exercise-card-drag-handle');
    await expect(dragHandles).toHaveCount(initialExerciseCount);
    await expect(exerciseCards.first()).toHaveClass(/workout-card-wiggle/);

    await page.emulateMedia({ reducedMotion: 'reduce' });
    await dragHandleToCardTop(page, dragHandles.nth(1), exerciseCards.first());
    await expect(page.getByTestId('save-exercise-order')).toBeEnabled();
    await page.getByTestId('save-exercise-order').click();

    await expect(exerciseCards.nth(0)).toContainText(secondExerciseName!);
    await expect(exerciseCards.nth(1)).toContainText(firstExerciseName!);

    await page.getByRole('button', { name: /^save$/i }).click();
    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.getByRole('link', { name: /edit/i }).click();

    await expect(exerciseCards.first()).toBeVisible();
    await expect(exerciseCards).toHaveCount(initialExerciseCount);
    await expect(exerciseCards.nth(0)).toContainText(secondExerciseName!);
    await expect(exerciseCards.nth(1)).toContainText(firstExerciseName!);
  });

  // URL Navigation Tests
  test('should update URL when navigating to add exercise', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    expect(page.url()).toContain('addExercise=true');
  });

  test('should support browser back button from add exercise', async ({
    page,
  }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.goBack();

    expect(page.url()).not.toContain('addExercise');
    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();
  });

  test('should update URL when navigating to exercise detail', async ({
    page,
  }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByTestId('edit-workout-exercise-card').first().click();
    expect(page.url()).toContain('exerciseIndex=0');
  });

  test('should support browser back button from exercise detail', async ({
    page,
  }) => {
    await page.getByTestId('workout-card').first().click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByTestId('edit-workout-exercise-card').first().click();
    await page.goBack();

    expect(page.url()).not.toContain('exerciseIndex');
    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();
  });

  test('should handle invalid exercise index gracefully', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    const editUrl = page.url().replace('/workouts/', '/workouts/') + '/edit';

    await page.goto(editUrl + '?exerciseIndex=999');

    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();

    expect(page.url()).not.toContain('exerciseIndex');
  });

  test('should handle negative exercise index gracefully', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();
    const editUrl = page.url().replace('/workouts/', '/workouts/') + '/edit';

    await page.goto(editUrl + '?exerciseIndex=-1');

    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();

    expect(page.url()).not.toContain('exerciseIndex');
  });
});
