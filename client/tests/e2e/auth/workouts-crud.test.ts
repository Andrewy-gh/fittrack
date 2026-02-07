import { test, expect } from '@playwright/test';
import { existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const authStatePath = path.join(__dirname, '..', '.auth', 'stack.json');

test.describe('Authenticated - Workouts CRUD', () => {
  test.use({ storageState: authStatePath });
  test.describe.configure({ mode: 'serial' });

  test.beforeEach(() => {
    test.skip(
      !existsSync(authStatePath),
      'Missing auth state. Ensure server/.env has PROJECT_ID + SECRET_SERVER_KEY or set E2E_STACK_EMAIL/E2E_STACK_PASSWORD.'
    );
  });

  test('supports create, read, update, delete', async ({ page }, testInfo) => {
    const suffix = `${Date.now()}-${testInfo.workerIndex}`;
    const focus = `E2E Focus ${suffix}`;
    const updatedFocus = `E2E Focus Updated ${suffix}`;
    const exerciseName = `E2E Exercise ${suffix}`;

    await page.goto('/workouts');
    await expect(page.getByRole('heading', { name: /workouts/i })).toBeVisible();

    await page.getByRole('link', { name: /new workout/i }).click();
    await expect(
      page.getByRole('heading', { name: /today's training/i })
    ).toBeVisible();

    await page.getByLabel('Workout Focus').click();
    await page.getByRole('combobox', { name: /workout focus options/i }).click();
    await page
      .getByRole('combobox', { name: /workout focus search/i })
      .fill(focus);
    await page
      .getByRole('option', { name: `Create "${focus}"` })
      .click();
    await page.getByRole('button', { name: /add today's focus/i }).click();

    await page.getByRole('link', { name: /add exercise/i }).click();
    await page.getByLabel(/search exercises/i).fill(exerciseName);
    await page.getByRole('button', { name: /^add$/i }).click();

    await page.getByRole('button', { name: /add set/i }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    await page.getByLabel(/^weight$/i).fill('135');
    await page.getByLabel(/^reps$/i).fill('8');
    await page.getByRole('button', { name: /save set/i }).click();
    await page.getByRole('button', { name: /back/i }).click();

    const createResponsePromise = page.waitForResponse((response) =>
      response.url().includes('/api/workouts') &&
      response.request().method() === 'POST'
    );
    await page.getByRole('button', { name: /^save$/i }).click();
    const createResponse = await createResponsePromise;
    expect(createResponse.ok()).toBeTruthy();

    await page.goto('/workouts');

    const workoutCard = page
      .getByTestId('workout-card')
      .filter({ hasText: focus.toUpperCase() })
      .first();
    await expect(workoutCard).toBeVisible();
    await workoutCard.click();

    await expect(page).toHaveURL(/\/workouts\/\d+\/?$/);
    await expect(
      page
        .getByRole('main')
        .locator('[data-slot="badge"]', { hasText: focus.toUpperCase() })
    ).toBeVisible();

    await page.getByRole('link', { name: /edit/i }).click();
    await expect(
      page.getByRole('heading', { name: /edit training/i })
    ).toBeVisible();

    await page.getByLabel('Workout Focus').click();
    await page.getByRole('combobox', { name: /workout focus options/i }).click();
    await page
      .getByRole('combobox', { name: /workout focus search/i })
      .fill(updatedFocus);
    await page
      .getByRole('option', { name: `Create "${updatedFocus}"` })
      .click();
    await page.getByRole('button', { name: /add today's focus/i }).click();

    const updateResponsePromise = page.waitForResponse((response) =>
      response.url().includes('/api/workouts/') &&
      response.request().method() === 'PUT'
    );
    await page.getByRole('button', { name: /^save$/i }).click();
    const updateResponse = await updateResponsePromise;
    expect(updateResponse.ok()).toBeTruthy();

    await expect(page).toHaveURL(/\/workouts\/\d+\/?$/);
    await expect(
      page
        .getByRole('main')
        .locator('[data-slot="badge"]', { hasText: updatedFocus.toUpperCase() })
    ).toBeVisible();

    const deleteResponsePromise = page.waitForResponse((response) =>
      response.url().includes('/api/workouts/') &&
      response.request().method() === 'DELETE'
    );
    await page.getByRole('button', { name: /^delete$/i }).click();
    await page.getByRole('button', { name: /delete workout/i }).click();
    const deleteResponse = await deleteResponsePromise;
    expect(deleteResponse.ok()).toBeTruthy();

    await expect(page).toHaveURL(/\/workouts\/?$/);
    await expect(
      page.getByTestId('workout-card').filter({
        hasText: updatedFocus.toUpperCase(),
      })
    ).toHaveCount(0);
  });
});
