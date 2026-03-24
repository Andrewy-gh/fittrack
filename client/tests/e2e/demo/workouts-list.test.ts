import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workouts List', () => {
  const lowerBodyWorkout = (page: import('@playwright/test').Page) =>
    page.getByRole('link', { name: /lower body/i }).first();

  test.beforeEach(async ({ page }) => {
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should load demo workouts', async ({ page }) => {
    await expect(lowerBodyWorkout(page)).toBeVisible();

    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
    await expect(page.getByText('UPPER BODY PUSH').first()).toBeVisible();
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await expect(lowerBodyWorkout(page)).toBeVisible();
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();

    await page.reload();

    await expect(lowerBodyWorkout(page)).toBeVisible();
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
  });

  test('should navigate to workout detail', async ({ page }) => {
    await lowerBodyWorkout(page).click();

    expect(page.url()).toMatch(/\/workouts\/\d+/);
    await expect(page.getByText('LOWER BODY').first()).toBeVisible();
  });
});
