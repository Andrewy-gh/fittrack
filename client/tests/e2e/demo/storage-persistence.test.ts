import { test, expect } from 'vitest';
import { page, userEvent } from '@vitest/browser/context';
import { getDemoWorkouts, getDemoExercises, clearAllStorage } from '../helpers/storage-helpers';

test.describe('localStorage Persistence', () => {
  test.beforeEach(async () => {
    await clearAllStorage(page);
  });

  test('should initialize demo data on first visit', async () => {
    await page.goto('/workouts');

    // Verify localStorage has demo data
    const workouts = await getDemoWorkouts(page);
    const exercises = await getDemoExercises(page);

    expect(workouts).not.toBeNull();
    expect(exercises).not.toBeNull();
    expect(workouts.length).toBe(3);
    expect(exercises.length).toBe(5);
  });

  test('should persist demo data across page reloads', async () => {
    await page.goto('/workouts');

    // Get initial data
    const initialWorkouts = await getDemoWorkouts(page);

    // Reload page
    await page.reload();

    // Verify data unchanged
    const reloadedWorkouts = await getDemoWorkouts(page);
    expect(reloadedWorkouts).toEqual(initialWorkouts);
  });

  test('should persist workout edits to localStorage', async () => {
    await page.goto('/workouts');

    // Edit a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Test notes for persistence');
    await page.getByRole('button', { name: /save/i }).click();

    await page.waitForTimeout(500); // Wait for mutation

    // Check localStorage directly
    const workouts = await getDemoWorkouts(page);
    const updatedWorkout = workouts.find((w: any) => w.notes === 'Test notes for persistence');

    expect(updatedWorkout).toBeDefined();
  });

  test('should persist workout deletion to localStorage', async () => {
    await page.goto('/workouts');

    // Get initial count
    const initialWorkouts = await getDemoWorkouts(page);
    const initialCount = initialWorkouts.length;

    // Delete a workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.query() !== null;
    if (isVisible) {
      await confirmButton.click();
    }

    await page.waitForTimeout(500);

    // Verify localStorage updated
    const updatedWorkouts = await getDemoWorkouts(page);
    expect(updatedWorkouts.length).toBe(initialCount - 1);
  });
});
