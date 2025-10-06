import { beforeEach } from 'vitest';
import { page } from '@vitest/browser/context';

// Clear localStorage before each test
beforeEach(async () => {
  await page.evaluate(() => localStorage.clear());
});
