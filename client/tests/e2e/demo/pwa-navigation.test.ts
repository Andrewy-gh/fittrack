import { test, expect } from "@playwright/test";

test.use({ viewport: { width: 393, height: 852 }, colorScheme: "dark" });

test("PWA navigation shrinks while browsing and expands at the top", async ({
  page,
}) => {
  await page.addInitScript(() => {
    Object.defineProperty(navigator, "standalone", { value: true });
    localStorage.setItem("vite-ui-theme", "dark");
  });
  await page.goto("/workouts");
  await expect(
    page.getByRole("link", { name: /lower body/i }).first(),
  ).toBeVisible();
  const nav = page.getByRole("navigation", { name: "PWA navigation" });
  const tab = nav.getByRole("link", { name: "Workouts" });
  const expanded = (await tab.boundingBox())!;
  await page.evaluate(() => window.scrollTo(0, 300));
  await expect.poll(async () => (await tab.boundingBox())!.height).toBe(44);
  const compact = (await tab.boundingBox())!;
  expect(compact.width).toBeLessThan(expanded.width);
  expect(compact.width).toBeGreaterThanOrEqual(44);
  await page.evaluate(() => window.scrollTo(0, 500));
  await expect
    .poll(async () => (await tab.boundingBox())!.width)
    .toBe(compact.width);
  await page.evaluate(() => window.scrollTo(0, 100));
  await expect.poll(async () => (await tab.boundingBox())!.height).toBe(44);
  await page.evaluate(() => window.scrollTo(0, 0));
  await expect
    .poll(async () => (await tab.boundingBox())!.height)
    .toBe(expanded.height);
  await page.emulateMedia({ reducedMotion: "reduce" });
  await page.evaluate(() => window.scrollTo(0, 300));
  await expect.poll(async () => (await tab.boundingBox())!.height).toBe(44);
  expect(
    await tab.evaluate(
      (element) => getComputedStyle(element).transitionProperty,
    ),
  ).toBe("none");
  await nav.getByRole("link", { name: "Exercises" }).click();
  await expect(page).toHaveURL(/\/exercises/);
  await expect(nav.getByRole("link", { name: "Exercises" })).toHaveAttribute(
    "aria-current",
    "page",
  );
});
