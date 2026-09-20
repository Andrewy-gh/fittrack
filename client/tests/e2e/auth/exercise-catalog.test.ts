import { expect, test } from "@playwright/test";
import { authStatePath } from "../helpers/local-e2e-auth";

// Run with local E2E authentication against an isolated migrated database.
test.use({ storageState: authStatePath });
test.describe.configure({ mode: "serial" });

test.beforeEach(() => {
  test.skip(
    process.env.E2E_LOCAL_AUTH_ENABLED !== "true",
    "Requires isolated local E2E authentication",
  );
});

for (const viewport of [
  { width: 1280, height: 1000 },
  { width: 390, height: 844 },
]) {
  test(`catalog selection and classification preserve history at ${viewport.width}px`, async ({
    page,
  }, testInfo) => {
    await page.setViewportSize(viewport);
    await page.addInitScript(() =>
      localStorage.setItem("vite-ui-theme", "dark"),
    );
    const headers = {
      "x-fittrack-dev-e2e-user":
        process.env.E2E_LOCAL_AUTH_USER_ID ?? "local-e2e-user",
    };
    const catalogName = viewport.width === 390 ? "Pushups" : "Plank";
    const list = await page.request.get("/api/exercises", { headers });
    expect(list.ok()).toBeTruthy();
    const existing = (await list.json()) as Array<{ id: number; name: string }>;
    expect(
      existing.find((entry) => entry.name === catalogName),
      "Use an isolated database without this fixture exercise",
    ).toBeUndefined();
    let exerciseId: number | undefined;
    let workoutId: number | undefined;
    try {
      await page.goto("/workouts/new?addExercise=true");
      await page.getByRole("button", { name: "Browse catalog" }).click();
      await page.getByLabel("Search catalog").fill(catalogName);
      await expect(
        page.getByRole("button", { name: catalogName, exact: true }),
      ).toBeVisible();
      await page.screenshot({
        path: testInfo.outputPath("catalog-picker.png"),
        fullPage: true,
      });
      await page
        .getByRole("button", { name: catalogName, exact: true })
        .click();
      await page.getByRole("button", { name: "Add Set", exact: true }).click();
      await page.getByLabel("Reps", { exact: true }).fill("10");
      await page.getByRole("button", { name: "Save Set", exact: true }).click();
      await page.getByRole("button", { name: "Back", exact: true }).click();
      const saved = page.waitForResponse(
        (response) =>
          response.url().endsWith("/api/workouts") &&
          response.request().method() === "POST",
      );
      await page.getByRole("button", { name: "Save", exact: true }).click();
      expect((await saved).ok()).toBeTruthy();
      const rows = (await (
        await page.request.get("/api/exercises", { headers })
      ).json()) as Array<{ id: number; name: string }>;
      exerciseId = rows.find((entry) => entry.name === catalogName)?.id;
      expect(exerciseId).toBeDefined();
      const original = await (
        await page.request.get(`/api/exercises/${exerciseId}`, { headers })
      ).json();
      expect(original.exercise.catalog.id).toBe(catalogName);
      expect(original.sets).toHaveLength(1);
      workoutId = original.sets[0].workout_id;
      const customName = `My ${catalogName.toLowerCase()} ${viewport.width}`;
      expect(
        (
          await page.request.patch(`/api/exercises/${exerciseId}`, {
            headers,
            data: { name: customName },
          })
        ).ok(),
      ).toBeTruthy();
      await page.goto(`/exercises/${exerciseId}`);
      await expect(
        page.getByRole("heading", { name: customName, exact: true }),
      ).toBeVisible();
      await page.getByRole("button", { name: "Change classification" }).click();
      await page.getByLabel("Search catalog").fill("dumbbell bench");
      await page
        .getByRole("button", { name: "Dumbbell Bench Press", exact: true })
        .click();
      await expect(
        page.getByText("Equipment: dumbbell", { exact: true }),
      ).toBeVisible();
      await expect(page.getByRole("dialog")).not.toBeVisible();
      await page.screenshot({
        path: testInfo.outputPath("classification.png"),
        fullPage: true,
      });
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= window.innerWidth,
        ),
      ).toBeTruthy();
      await page.getByRole("button", { name: "Clear", exact: true }).click();
      await expect(
        page.getByText("Unclassified", { exact: true }),
      ).toBeVisible();
      await page.getByRole("button", { name: "Classify exercise" }).click();
      await page.getByLabel("Search catalog").fill("no such exercise");
      await expect(page.getByText(/No catalog matches/)).toBeVisible();
      await page.getByLabel("Search catalog").fill(catalogName);
      await page
        .getByRole("button", { name: catalogName, exact: true })
        .click();
      await expect(
        page.getByRole("button", { name: "Change classification" }),
      ).toBeVisible();
      await page.reload();
      await expect(
        page.getByRole("heading", { name: customName, exact: true }),
      ).toBeVisible();
      const after = await (
        await page.request.get(`/api/exercises/${exerciseId}`, { headers })
      ).json();
      expect(after.exercise.id).toBe(original.exercise.id);
      expect(after.exercise.name).toBe(customName);
      expect(after.exercise.catalog.id).toBe(catalogName);
      expect(
        after.sets.map((set: Record<string, unknown>) => ({
          ...set,
          exercise_name: catalogName,
        })),
      ).toEqual(original.sets);
    } finally {
      if (workoutId)
        await page.request.delete(`/api/workouts/${workoutId}`, { headers });
      if (exerciseId)
        await page.request.delete(`/api/exercises/${exerciseId}`, { headers });
    }
  });
}
