import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workouts List', () => {
  test.beforeEach(async ({ page }) => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load demo workouts', async ({ page }) => {
    const workoutCards = page.getByTestId('workout-card');
    await expect(workoutCards.first()).toBeVisible();

    await expect(page.getByText('Morning Strength')).toBeVisible();
    await expect(page.getByText('Evening Cardio')).toBeVisible();
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await expect(page.getByTestId('workout-card').first()).toBeVisible();
    await expect(page.getByText('Morning Strength')).toBeVisible();

    await page.reload();

    await expect(page.getByTestId('workout-card').first()).toBeVisible();
    await expect(page.getByText('Morning Strength')).toBeVisible();
  });

  test('should navigate to workout detail', async ({ page }) => {
    await page.getByText('Morning Strength').click();

    expect(page.url()).toMatch(/\/workouts\/\d+/);
    await expect(page.getByText('Morning Strength')).toBeVisible();
  });
});
