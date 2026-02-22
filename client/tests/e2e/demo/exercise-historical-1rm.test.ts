import { test, expect } from '@playwright/test';

test.describe('Demo Mode - Historical 1RM', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/exercises');
    await page.evaluate(() => localStorage.clear());
    await page.reload();
  });

  test('should allow setting and persisting manual historical 1RM', async ({ page }) => {
    await page.getByText('Bench Press').click();

    await page.getByRole('button', { name: 'Set' }).click();

    const input = page.locator('#historical-1rm-input');
    await expect(input).toBeVisible();
    await input.fill('300');

    await page.getByRole('button', { name: 'Save' }).click();

    await expect(page.getByText('300.0 lb', { exact: true })).toBeVisible();
    await expect(page.getByText(/Updated .*Manual/i)).toBeVisible();

    await page.reload();

    await expect(page.getByText('300.0 lb', { exact: true })).toBeVisible();

    const stored = await page.evaluate(() => localStorage.getItem('fittrack-demo-historical-1rm'));
    expect(stored).not.toBeNull();
  });
});
