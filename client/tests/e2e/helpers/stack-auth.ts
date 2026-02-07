import type { Page } from '@playwright/test';

export async function signInWithStack(page: Page, email: string, password: string) {
  await page.goto('/handler/sign-in');

  const passwordTab = page.getByRole('tab', { name: /email & password/i });
  if (await passwordTab.count()) {
    await passwordTab.click();
  }

  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/password/i).fill(password);

  await Promise.all([
    page.waitForURL((url) => !url.pathname.startsWith('/handler'), {
      timeout: 15000,
    }),
    page.getByRole('button', { name: /sign in/i }).click(),
  ]);
}
