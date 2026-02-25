import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
const isCI = Boolean(process.env.CI);
const e2ePort = Number(process.env.E2E_PORT ?? '5173');
const baseURL = process.env.E2E_BASE_URL ?? `http://127.0.0.1:${e2ePort}`;

export default defineConfig({
  testDir: './tests/e2e',

  globalSetup: './tests/e2e/global-setup.ts',

  /* Avoid jobs running forever in CI when setup gets stuck. */
  globalTimeout: isCI ? 20 * 60 * 1000 : undefined,

  /* Run tests in files in parallel */
  fullyParallel: true,

  /* Increase timeout for CI */
  timeout: isCI ? 60000 : 30000,

  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,

  /* Retry on CI only */
  retries: isCI ? 2 : 0,

  /* Prefer stability over throughput in shared CI runners. */
  workers: isCI ? 1 : undefined,

  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: isCI ? [['line'], ['html', { open: 'never' }]] : 'html',

  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL,

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',

    /* Screenshot on failure */
    screenshot: 'only-on-failure',

  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: {
    command: isCI ? 'bun run serve:test' : 'bun run dev',
    url: baseURL,
    reuseExistingServer: !isCI,
    timeout: 120000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
