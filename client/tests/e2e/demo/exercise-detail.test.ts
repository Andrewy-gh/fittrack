import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Exercise Detail', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/exercises');
  });

  test('should display exercise details', async () => {
    // Navigate to exercise
    await page.getByText('Bench Press').click();

    // Verify exercise name visible
    await expect.element(page.getByRole('heading', { name: /bench press/i })).toBeInTheDocument();

    // Verify delete button visible in demo mode
    await expect.element(page.getByRole('button', { name: /delete/i })).toBeInTheDocument();
  });

  test('should delete exercise in demo mode', async () => {
    // Count initial exercises
    const initialExercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data).length : 0;
    });

    // Navigate to exercise and delete
    await page.getByText('Bench Press').click();
    await page.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion (if dialog exists)
    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    // Verify navigated back to list
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toBe('/exercises');

    // Verify exercise count decreased in localStorage
    const updatedExercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data).length : 0;
    });
    expect(updatedExercises).toBe(initialExercises - 1);
  });
});
