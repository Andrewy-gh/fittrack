import { defineConfig } from "@playwright/test";
import base from "./playwright.config";

// Every test under auth/ is mandatory. Keep guest preview coverage separate.
export default defineConfig(base, {
  testDir: "./tests/e2e/auth",
  fullyParallel: false,
  workers: 1,
  forbidOnly: true,
  failOnFlakyTests: true,
  reporter: [
    ["line"],
    ["html", { open: "never", outputFolder: "playwright-auth-report" }],
    ["./scripts/required-auth-reporter.ts"],
  ],
});
