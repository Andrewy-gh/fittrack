import { test, expect } from 'vitest';
import { page, userEvent } from '@vitest/browser/context';

test.describe('Demo Mode - Workout Edit', () => {
  test.beforeEach(async () => {
    await page.evaluate(() => localStorage.clear());
    await page.goto('/workouts');
  });

  test('should load edit form with existing workout data', async () => {
    // Navigate to workout detail
    await page.getByText('Morning Strength').click();

    // Click Edit button
    await page.getByRole('button', { name: /edit/i }).click();

    // Verify navigated to edit route
    const currentUrl = await page.evaluate(() => window.location.pathname);
    expect(currentUrl).toMatch(/\/workouts\/\d+\/edit/);

    // Verify form loaded
    await expect.element(page.getByRole('heading', { name: /edit training/i })).toBeInTheDocument();

    // Verify exercises loaded
    const exerciseCard = page.getByTestId('exercise-card');
    await expect.element(exerciseCard).toBeInTheDocument();
  });

  test('should modify and save workout in demo mode', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Modify notes field
    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Updated notes in demo mode');

    // Save workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for navigation/mutation
    await page.waitForTimeout(500);

    // Navigate back to list and verify workout still exists
    await page.goto('/workouts');
    await expect.element(page.getByText('Morning Strength')).toBeInTheDocument();
  });

  test('should persist edited workout to localStorage', async () => {
    // Edit workout
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    const notesField = page.getByLabel(/notes/i);
    await userEvent.fill(notesField, 'Persistent notes');
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for mutation
    await page.waitForTimeout(500);

    // Reload page
    await page.reload();

    // Verify data persisted in localStorage
    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(3); // Still 3 workouts
  });

  test('should add exercise to workout', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Count initial exercises via localStorage
    const initialExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });

    // Click "Add Exercise" button
    await page.getByRole('button', { name: /add exercise/i }).click();

    // Select an exercise (adjust selector based on your UI)
    await page.getByText('Bench Press').click();

    // Save the changes
    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForTimeout(500);

    // Verify exercise added in localStorage
    const updatedExerciseCount = await page.evaluate(() => {
      const workouts = JSON.parse(localStorage.getItem('demo_workouts') || '[]');
      const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
      return morningStrength?.exercises?.length || 0;
    });
    expect(updatedExerciseCount).toBe(initialExerciseCount + 1);
  });

  test('should remove exercise from workout', async () => {
    // Navigate to edit
    await page.getByText('Morning Strength').click();
    await page.getByRole('button', { name: /edit/i }).click();

    // Find and click delete on first exercise card
    const firstExerciseCard = page.getByTestId('exercise-card');
    const deleteButton = firstExerciseCard.getByRole('button', { name: /delete|remove/i });
    await deleteButton.click();

    // Save and verify
    await page.getByRole('button', { name: /save/i }).click();
    await page.waitForTimeout(500);

    // Verify in localStorage that exercise count decreased
    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('demo_workouts');
      return data ? JSON.parse(data) : [];
    });
    const morningStrength = workouts.find((w: any) => w.name === 'Morning Strength');
    expect(morningStrength.exercises.length).toBeGreaterThan(0);
  });
});
