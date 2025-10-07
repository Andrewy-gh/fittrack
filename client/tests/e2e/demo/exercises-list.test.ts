import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Exercises List', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/exercises');
  });

  test('should load demo exercises', async ({ page }) => {
    const exercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data) : [];
    });
    expect(exercises).toHaveLength(5);

    await expect(page.getByText('Bench Press')).toBeVisible();
    await expect(page.getByText('Squat')).toBeVisible();
  });

  test('should navigate to exercise detail', async ({ page }) => {
    await page.getByText('Bench Press').click();

    expect(page.url()).toMatch(/\/exercises\/\d+/);
    await expect(page.getByText('Bench Press')).toBeVisible();
  });
});
