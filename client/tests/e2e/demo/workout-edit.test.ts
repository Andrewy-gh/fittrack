import { test, expect } from '@playwright/test';

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

    const updatedExerciseCount = await page
      .getByTestId('edit-workout-exercise-card')
      .count();
    expect(updatedExerciseCount).toBe(initialExerciseCount + 1);
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

    const updatedExerciseCount = await page
      .getByTestId('edit-workout-exercise-card')
      .count();
    expect(updatedExerciseCount).toBe(initialExerciseCount - 1);
  });
});
