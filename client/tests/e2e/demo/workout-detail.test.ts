import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Detail', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should display workout details', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();

    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
    await expect(page.getByRole('link', { name: /edit/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /delete/i })).toBeVisible();

    const exerciseCard = page.getByTestId('workout-detail-exercise-card');
    await expect(exerciseCard.first()).toBeVisible();
  });

  test('should show edit and delete buttons in demo mode', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();

    await expect(page.getByRole('link', { name: /edit/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /delete/i })).toBeVisible();
  });

  test('should delete workout in demo mode', async ({ page }) => {
    const initialWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-workouts');
      return data ? JSON.parse(data).length : 0;
    });

    await page.getByTestId('workout-card').first().click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.isVisible().catch(() => false);
    if (isVisible) {
      await confirmButton.click();
    }

    await expect(page).toHaveURL('/workouts');

    const updatedWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-workouts');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedWorkouts).toBe(initialWorkouts - 1);
  });
});
