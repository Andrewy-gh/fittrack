import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Exercise Detail', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/exercises');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should display exercise details', async ({ page }) => {
    await page.getByText('Bench Press').click();

    await expect(page.getByRole('heading', { name: /bench press/i })).toBeVisible();

    await expect(page.getByRole('button', { name: /delete/i })).toBeVisible();
  });

  test('should delete exercise in demo mode', async ({ page }) => {
    const initialExercises = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-exercises');
      return data ? JSON.parse(data).length : 0;
    });

    await page.getByText('Bench Press').click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.isVisible().catch(() => false);
    if (isVisible) {
      await confirmButton.click();
    }

    await expect(page).toHaveURL('/exercises');

    const updatedExercises = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-exercises');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedExercises).toBe(initialExercises - 1);
  });

  test('should render zero-set exercise detail', async ({ page }) => {
    const exerciseId = await page.evaluate(() => {
      const now = new Date().toISOString();
      const exercisesRaw = localStorage.getItem('fittrack-demo-exercises');
      const exercises = exercisesRaw ? JSON.parse(exercisesRaw) as Array<{
        id: number;
        name: string;
        user_id: string;
        created_at: string;
        updated_at: string;
      }> : [];
      const nextId = exercises.length > 0
        ? Math.max(...exercises.map((exercise) => exercise.id)) + 1
        : 1;

      exercises.push({
        id: nextId,
        name: 'No Sets Yet',
        user_id: 'demo-user',
        created_at: now,
        updated_at: now,
      });

      localStorage.setItem('fittrack-demo-exercises', JSON.stringify(exercises));
      return nextId;
    });

    await page.goto(`/exercises/${exerciseId}`);

    await expect(page.getByRole('heading', { name: /no sets yet/i })).toBeVisible();
    await expect(page.getByText('No workouts logged for this exercise yet.')).toBeVisible();
  });
});
