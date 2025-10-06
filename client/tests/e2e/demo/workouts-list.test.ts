import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Workouts List', () => {
  test.beforeEach(async () => {
    // Clear localStorage and navigate to workouts
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load demo workouts', async () => {
    // Verify demo data initialized
    const workoutCards = page.getByTestId('workout-card');
    await expect.element(workoutCards).toBeInTheDocument();

    // Verify workout content
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
    await expect.element(page.getByText('Evening Cardio')).toBeInTheDocument();
  });

  test('should persist demo data across page reloads', async () => {
    // Verify initial load
    await expect.element(page.getByTestId('workout-card')).toBeInTheDocument();
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();

    // Reload page
    await page.reload();

    // Verify data still there
    await expect.element(page.getByTestId('workout-card')).toBeInTheDocument();
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });

  test('should navigate to workout detail', async () => {
    // Click first workout
    await page.getByText('Morning Strength').click();

    // Verify navigation (check URL contains /workouts/)
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/workouts\/\d+/);
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });
});
