import { test, expect } from '@playwright/test';
import { getDemoWorkouts, getDemoExercises, clearAllStorage } from '../helpers/storage-helpers';

test.describe('localStorage Persistence', () => {
  const lowerBodyWorkout = (page: import('@playwright/test').Page) =>
    page.getByRole('link', { name: /lower body/i }).first();

  test.beforeEach(async ({ page }) => {
    await clearAllStorage(page);
  });

  test('should initialize demo data on first visit', async ({ page }) => {
    await page.goto('/workouts');

    const workouts = await getDemoWorkouts(page);
    const exercises = await getDemoExercises(page);

    expect(workouts).not.toBeNull();
    expect(exercises).not.toBeNull();
    expect(workouts.length).toBe(3);
    expect(exercises.length).toBe(5);
  });

  test('should persist demo data across page reloads', async ({ page }) => {
    await page.goto('/workouts');

    const initialWorkouts = await getDemoWorkouts(page);

    await page.reload();

    const reloadedWorkouts = await getDemoWorkouts(page);
    expect(reloadedWorkouts).toEqual(initialWorkouts);
  });

  test('should persist workout edits to localStorage', async ({ page }) => {
    await page.goto('/workouts');

    await lowerBodyWorkout(page).click();
    await page.getByRole('link', { name: /edit/i }).click();

    await page.getByRole('button', { name: /^notes$/i }).click();

    const notesDialog = page.getByRole('dialog', { name: /^notes$/i });
    const notesField = notesDialog.getByRole('textbox', { name: /^notes$/i });
    await notesField.fill('Test notes for persistence');

    await notesDialog.getByRole('button', { name: /^close$/i }).first().click();

    await page.getByRole('button', { name: /^save$/i }).click();

    await page.waitForURL('/workouts/**', { timeout: 5000 });

    // Wait for the save to complete and data to be persisted
    await page.waitForTimeout(500);

    const workouts = await getDemoWorkouts(page);
    const updatedWorkout = workouts.find((w: any) => w.notes === 'Test notes for persistence');

    expect(updatedWorkout).toBeDefined();
  });

  test('should persist workout deletion to localStorage', async ({ page }) => {
    await page.goto('/workouts');

    const initialWorkouts = await getDemoWorkouts(page);
    const initialCount = initialWorkouts.length;

    await lowerBodyWorkout(page).click();
    await page.getByRole('button', { name: /delete/i }).click();

    const confirmButton = page.getByRole('button', { name: /confirm|yes|delete/i });
    const isVisible = await confirmButton.isVisible().catch(() => false);
    if (isVisible) {
      await confirmButton.click();
    }

    await page.waitForURL('/workouts', { timeout: 5000 });

    const updatedWorkouts = await getDemoWorkouts(page);
    expect(updatedWorkouts.length).toBe(initialCount - 1);
  });
});
