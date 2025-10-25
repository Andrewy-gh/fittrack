import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Workout Create', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate first, then clear localStorage and reload to reinitialize demo data
    await page.goto('/workouts');
    await page.evaluate(() => localStorage.clear());
    await page.reload();

    // Wait for demo data to be initialized and workouts to be visible
    await page.waitForSelector('[data-testid="workout-card"]', {
      timeout: 5000,
    });
  });

  test('should navigate to new workout page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    expect(page.url()).toMatch(/\/workouts\/new/);

    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();
  });

  test('should display empty workout form on new page', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Verify form elements are present
    await expect(
      page.getByRole('link', { name: /add exercise/i })
    ).toBeVisible();
    await expect(page.getByRole('button', { name: /save/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /clear/i })).toBeVisible();

    // Verify no exercises are present initially
    const exerciseCards = page
      .locator('[data-testid*="exercise"]')
      .filter({ hasText: /sets|volume/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should create a workout with single exercise and single set', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add an exercise
    await page.getByRole('link', { name: /add exercise/i }).click();

    // Select an exercise from the list
    await page.getByText('Bench Press', { exact: true }).click();

    // Add a set
    await page.getByRole('button', { name: /add set/i }).click();

    // Wait for the dialog to be visible
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill in set details (using spinbutton role for number inputs)
    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);

    await weightInput.fill('135');
    await repsInput.fill('10');

    await page.getByRole('button', { name: /save set/i }).click();

    // Go back to main form
    await page.getByRole('button', { name: /back/i }).click();

    // Save the workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset (exercises should be cleared)
    await page.waitForTimeout(500);

    // Verify we're still on the new workout page
    expect(page.url()).toContain('/new');

    // Verify the form has been reset (no exercises)
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should persist created workout to localStorage', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add an exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Deadlift', { exact: true }).click();

    // Add a set
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);

    await weightInput.fill('225');
    await repsInput.fill('5');

    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    // Save the workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for save to complete
    await page.waitForTimeout(500);

    // Check localStorage for the new workout (should have 4 workouts now: 3 initial + 1 new)
    const workouts = await page.evaluate(() => {
      const data = localStorage.getItem('fittrack-demo-workouts');
      return data ? JSON.parse(data) : [];
    });
    expect(workouts).toHaveLength(4);
  });

  test('should create workout with multiple exercises', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add first exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    let weightInput = page.getByRole('spinbutton').first();
    let repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    // Add second exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Barbell Squat', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('185');
    await repsInput.fill('8');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    // Verify we have 2 exercises in the form before saving
    let exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(2);

    // Save the workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    // Verify we're still on the new workout page and form is reset
    expect(page.url()).toContain('/new');
    exerciseCards = page.locator('.cursor-pointer').filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should create workout with multiple sets for same exercise', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add an exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Pull-ups', { exact: true }).click();

    // Add first set
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    let weightInput = page.getByRole('spinbutton').first();
    let repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();

    // Add second set
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('8');
    await page.getByRole('button', { name: /save set/i }).click();

    // Add third set
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    weightInput = page.getByRole('spinbutton').first();
    repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('0');
    await repsInput.fill('6');
    await page.getByRole('button', { name: /save set/i }).click();

    // Go back to main form
    await page.getByRole('button', { name: /back/i }).click();

    // Verify exercise shows 3 sets
    const exerciseCard = page.getByTestId('new-workout-exercise-card');
    await expect(exerciseCard.getByText('3')).toBeVisible();
    await expect(exerciseCard.getByText('sets', { exact: true })).toBeVisible();

    // Save the workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    // Verify we're still on the new workout page and form is reset
    expect(page.url()).toContain('/new');
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should clear form data when clear button is clicked', async ({
    page,
  }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add an exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    // Verify exercise is present
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(1);

    // Listen for dialog and accept
    page.on('dialog', (dialog) => dialog.accept());

    // Clear the form
    await page.getByRole('button', { name: /clear/i }).click();

    // Wait a bit for the form to clear
    await page.waitForTimeout(200);

    // Verify exercise is removed
    await expect(exerciseCards).toHaveCount(0);
  });

  test('should add workout with notes and focus', async ({ page }) => {
    await page.getByRole('link', { name: /new workout/i }).click();

    // Add notes - find the notes card and click it
    await page.getByTestId('notes-card').click();

    // Wait for dialog to open
    await expect(page.getByRole('dialog')).toBeVisible();

    // Fill in notes
    const notesTextarea = page.getByTestId('notes-textarea');
    await notesTextarea.fill('Great workout today!');

    // Close notes
    await page.getByTestId('notes-close').click();

    // Add an exercise
    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByText('Bench Press', { exact: true }).click();
    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    const weightInput = page.getByRole('spinbutton').first();
    const repsInput = page.getByRole('spinbutton').nth(1);
    await weightInput.fill('135');
    await repsInput.fill('10');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    // Save the workout
    await page.getByRole('button', { name: /save/i }).click();

    // Wait for form to reset
    await page.waitForTimeout(500);

    // Verify we're still on the new workout page and form is reset
    expect(page.url()).toContain('/new');
    const exerciseCards = page
      .locator('.cursor-pointer')
      .filter({ hasText: /sets/ });
    await expect(exerciseCards).toHaveCount(0);
  });
});
