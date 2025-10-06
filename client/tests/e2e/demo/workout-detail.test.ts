import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Workout Detail', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should display workout details', async () => {
    // Navigate to first workout
    await page.getByText('Morning Strength').click();

    // Verify workout details visible
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
    await expect.element(page.getByText('Edit')).toBeInTheDocument();
    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    // Verify exercises shown
    const exerciseCard = page.getByTestId('exercise-card');
    await expect.element(exerciseCard).toBeInTheDocument();
  });

  test('should show edit and delete buttons in demo mode', async () => {
    await page.getByText('Morning Strength').click();

    // Verify buttons are visible (not hidden in demo mode)
    await expect.element(page.getByRole('button', { name: /edit/i })).toBeInTheDocument();
    await expect.element(page.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  test('should delete workout in demo mode', async () => {
    // Count initial workouts
    const initialWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data).length : 0;
    });

    // Navigate to workout and delete
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion (if dialog exists)
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toBe('/workouts');

    // Verify workout no longer in list
    const morningStrength = await page.getByText('Morning Strength').query();
    expect(morningStrength).toBeNull();

    // Verify localStorage updated
    const updatedWorkouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedWorkouts).toBe(initialWorkouts - 1);
  });
});
