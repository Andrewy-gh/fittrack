import { test, expect } from 'vitest';
import { page } from '@vitest/browser/context';

test.describe('Demo Mode - Exercises List', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/exercises');
  });

  test('should load demo exercises', async () => {
    // Verify demo exercises loaded (5 in initial-data.ts)
    const exercises = await page.evaluate(() => {
      const data = localStorage.getItem('demo_exercises');
      return data ? JSON.parse(data) : [];
    });
    expect(exercises).toHaveLength(5);

    // Verify exercise names visible
    await expect.element(page.getByText('Bench Press')).toBeInTheDocument();
    await expect.element(page.getByText('Squat')).toBeInTheDocument();
  });

  test('should navigate to exercise detail', async () => {
    // Click first exercise
    await page.getByText('Bench Press').click();

    // Verify navigation
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/exercises\/\d+/);
    await expect.element(page.getByText('Bench Press')).toBeInTheDocument();
  });
});
