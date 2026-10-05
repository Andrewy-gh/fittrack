import { defineConfig, devices } from "@playwright/test";
import { resolveE2EServerConfig } from "./scripts/local-server-config";

/**
 * See https://playwright.dev/docs/test-configuration.
 */
const isCI = Boolean(process.env.CI);
const runsExclusiveMutationTests =
  process.env.E2E_EXCLUSIVE_MUTATION === "true";
const { baseURL, webServer } = resolveE2EServerConfig();

export default defineConfig({
  testDir: "./tests/e2e",

  globalSetup: "./tests/e2e/global-setup.ts",

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
  workers: isCI || runsExclusiveMutationTests ? 1 : undefined,

  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: isCI ? [["line"], ["html", { open: "never" }]] : "html",

  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL,

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: "on-first-retry",

    /* Screenshot on failure */
    screenshot: "only-on-failure",
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer,
});
