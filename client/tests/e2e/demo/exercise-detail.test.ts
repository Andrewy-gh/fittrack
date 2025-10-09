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
});
