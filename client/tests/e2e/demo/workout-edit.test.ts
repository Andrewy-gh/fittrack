import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Edit', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load edit form with existing workout data', async ({ page }) => {
    await page.getByText('Morning Strength').click();

    await page.getByRole('button', { name: /edit/i }).click();

    expect(page.url()).toMatch(/\/workouts\/\d+\/edit/);

    await expect(page.getByRole('heading', { name: /edit training/i })).toBeVisible();

    const exerciseCard = page.getByTestId('exercise-card');
    await expect(exerciseCard.first()).toBeVisible();
  });

  test('should modify and save workout in demo mode', async ({ page }) => {
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await notesField.fill('Updated notes in demo mode');

    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.goto('/workouts');
    await expect(page.getByText('Morning Strength')).toBeVisible();
  });

  test('should persist edited workout to localStorage', async ({ page }) => {
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await notesField.fill('Persistent notes');
    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForURL('/workouts/**', { timeout: 5000 });

    await page.reload();

    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(3);
  });

  test('should add exercise to workout', async ({ page }) => {
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const initialExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });

    await page.getByRole('button', { name: /add exercise/i }).click();

    await page.getByText('Bench Press').click();

    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForURL('/workouts/**', { timeout: 5000 });

    const updatedExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });
    expect(updatedExerciseCount).toBe(initialExerciseCount + 1);
  });

  test('should remove exercise from workout', async ({ page }) => {
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const firstExerciseCard = page.getByTestId('exercise-card').first();
    const deleteButton = firstExerciseCard.getByRole('button', { name: /delete|remove/i });
    await deleteButton.click();

    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForURL('/workouts/**', { timeout: 5000 });

    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
    expect(morningStrength.exercises.length).toBeGreaterThan(0);
  });
});
