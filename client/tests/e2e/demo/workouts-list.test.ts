import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workouts List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should load demo workouts', async ({ page }) => {
    const workoutCards = page.getByTestId('workout-card');
    await expect(workoutCards.first()).toBeVisible();

    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
    await expect(page.getByText('UPPER BODY PUSH').first()).toBeVisible();
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await expect(page.getByTestId('workout-card').first()).toBeVisible();
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();

    await page.reload();

    await expect(page.getByTestId('workout-card').first()).toBeVisible();
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
  });

  test('should navigate to workout detail', async ({ page }) => {
    await page.getByTestId('workout-card').first().click();

    expect(page.url()).toMatch(/\/workouts\/\d+/);
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
  });
});
