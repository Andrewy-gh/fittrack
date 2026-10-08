import { test, expect } from "@playwright/test";

for (const mode of ["create", "edit"] as const) {
  test(`${mode}: text boundaries preserve drafts and allow Unicode at the server limit`, async ({
    page,
  }) => {
    const valid = "🏋".repeat(256);
    const invalid = valid + "a";
    await page.goto("/workouts");
    await expect(
      page.getByRole("link", { name: /lower body/i }).first(),
    ).toBeVisible();
    if (mode === "create") {
      await page.evaluate(() =>
        localStorage.setItem(
          "workout-entry-form-data",
          JSON.stringify({
            date: new Date().toISOString(),
            notes: "",
            workoutFocus: "",
            exercises: [
              { name: "Bench Press", sets: [{ reps: 10, setType: "working" }] },
            ],
          }),
        ),
      );
    }
    await page.goto(mode === "create" ? "/workouts/new" : "/workouts/1/edit");
    await page.getByRole("button", { name: "Notes", exact: true }).click();
    const notes = page.getByRole("textbox", { name: "Notes", exact: true });
    await notes.fill(invalid);
    await expect(page.getByRole("dialog").getByRole("alert")).toHaveText(
      "Notes must be 256 characters or less.",
    );
    await expect(notes).toHaveValue(invalid);
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close", exact: true })
      .first()
      .click();
    await expect(
      page.getByRole("button", { name: "Save", exact: true }),
    ).toBeDisabled();
    if (mode === "create") {
      await page.reload();
      await expect(
        page.getByRole("button", { name: "Save", exact: true }),
      ).toBeDisabled();
    }
    await page.getByRole("button", { name: "Notes", exact: true }).click();
    await expect(notes).toHaveValue(invalid);
    await notes.fill(valid);
    await expect(
      page.getByText("256/256 characters", { exact: true }),
    ).toBeVisible();
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Close", exact: true })
      .first()
      .click();
    await page.getByLabel("Workout Focus", { exact: true }).click();
    await page.getByRole("combobox", { name: "Workout focus options" }).click();
    const search = page.getByRole("combobox", { name: "Workout focus search" });
    await search.fill(invalid);
    await expect(
      page.getByRole("alert").filter({ hasText: "Workout focus must be 256" }),
    ).toBeVisible();
    await expect(page.getByRole("option", { name: /^Create/ })).toHaveCount(0);
    await expect(search).toHaveValue(invalid);
    await search.fill(valid);
    await page
      .getByRole("option", { name: /^Create/ })
      .first()
      .click();
    await page.getByRole("button", { name: "Add today's focus" }).click();
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await expect
      .poll(async () =>
        page.evaluate((text) => {
          const workouts = JSON.parse(
            localStorage.getItem("fittrack-demo-workouts") ?? "[]",
          );
          return workouts.some(
            (workout: { notes?: string; workout_focus?: string }) =>
              workout.notes === text && workout.workout_focus === text,
          );
        }, valid),
      )
      .toBe(true);
  });

  test(`${mode}: manual exercise name validates before adding without truncation`, async ({
    page,
  }) => {
    await page.goto("/workouts");
    await expect(
      page.getByRole("link", { name: /lower body/i }).first(),
    ).toBeVisible();
    await page.goto(mode === "create" ? "/workouts/new" : "/workouts/1/edit");
    await page.getByRole("link", { name: "Add Exercise", exact: true }).click();
    const search = page.getByRole("textbox", { name: "Search exercises" });
    const valid = "肩".repeat(255) + "🏋";
    await search.fill(valid + "x");
    await expect(page.getByRole("alert")).toHaveText(
      "Exercise name must be 256 characters or less.",
    );
    await expect(
      page.getByRole("button", { name: "Add", exact: true }),
    ).toBeDisabled();
    await expect(search).toHaveValue(valid + "x");
    await search.fill(valid);
    await page.getByRole("button", { name: "Add", exact: true }).click();
    await expect(
      page.getByRole("heading", { name: valid, exact: true }),
    ).toBeVisible();
  });
}

test("notes dialog stays scrollable in a short enlarged-text viewport", async ({
  page,
}) => {
  await page.setViewportSize({ width: 320, height: 360 });
  await page.goto("/workouts/new");
  await page.getByRole("button", { name: "Notes", exact: true }).click();
  await page.evaluate(() => (document.documentElement.style.fontSize = "32px"));
  const dialog = page.getByRole("dialog", { name: "Notes", exact: true });
  await expect
    .poll(() =>
      dialog.evaluate((element) => {
        const box = element.getBoundingClientRect();
        return (
          box.top >= 0 &&
          box.bottom <= innerHeight &&
          element.scrollHeight > element.clientHeight
        );
      }),
    )
    .toBe(true);
  await dialog
    .getByRole("button", { name: "Close", exact: true })
    .first()
    .click();
  await expect(dialog).toHaveCount(0);
});

test("exercise rename uses the same Unicode limit and retains invalid input", async ({
  page,
}) => {
  await page.goto("/exercises/1");
  await page.getByRole("button", { name: "Edit", exact: true }).click();
  const input = page.getByRole("textbox", { name: "Exercise name" });
  const valid = "🏋".repeat(256);
  await input.fill(valid + "x");
  await expect(input).toHaveValue(valid + "x");
  await expect(page.getByRole("alert")).toHaveText(
    "Exercise name must be 256 characters or less.",
  );
  await expect(
    page.getByRole("button", { name: "Save", exact: true }),
  ).toBeDisabled();
  await input.fill(valid);
  await page.getByRole("button", { name: "Save", exact: true }).click();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await expect(
    page.getByRole("heading", { name: valid, exact: true }),
  ).toBeVisible();
});

test("theme metadata follows selection and survives reload", async ({
  page,
}) => {
  await page.goto("/exercises");
  const meta = page.locator('meta[name="theme-color"]');
  await expect(meta).toHaveAttribute("content", "#f9f9f9");
  await page.getByRole("button", { name: "Guest user menu" }).click();
  await page.getByRole("menuitem", { name: "Dark", exact: true }).click();
  await expect(meta).toHaveAttribute("content", "#000000");
  await expect(page.locator("html")).toHaveCSS("color-scheme", "dark");
  await page.reload();
  await expect(meta).toHaveAttribute("content", "#000000");
  await page.getByRole("button", { name: "Guest user menu" }).click();
  await page.getByRole("menuitem", { name: "Light", exact: true }).click();
  await expect(meta).toHaveAttribute("content", "#f9f9f9");
});

test("catalog keeps full names and controls apart at 320px with enlarged text", async ({
  page,
}) => {
  await page.setViewportSize({ width: 320, height: 800 });
  await page.goto("/exercises");
  await expect(
    page.getByRole("heading", { name: "Exercises", exact: true }),
  ).toBeVisible();
  const names = [
    "Single-Arm Half-Kneeling High-to-Low Cable Chop with Contralateral Pause and Extended Range of Motion",
    "X".repeat(256),
    "肩🏋".repeat(100),
  ];
  await page.evaluate((names) => {
    const time = new Date().toISOString();
    localStorage.setItem(
      "fittrack-demo-exercises",
      JSON.stringify(
        names.map((name, index) => ({
          id: index + 1,
          name,
          user_id: "demo-user",
          created_at: time,
          updated_at: time,
        })),
      ),
    );
  }, names);
  await page.reload();
  await expect(
    page.getByRole("link", { name: names[1], exact: true }),
  ).toBeVisible();
  await page.evaluate(() => (document.documentElement.style.fontSize = "32px"));
  const layout = await page
    .getByRole("link", { name: names[1], exact: true })
    .evaluate((link) => {
      const heading = link.querySelector("h3");
      const icon = link.querySelector("svg");
      if (!heading || !icon)
        throw new Error("Missing exercise label or control");
      return {
        textRight: heading.getBoundingClientRect().right,
        iconLeft: icon.getBoundingClientRect().left,
        iconWidth: icon.getBoundingClientRect().width,
        pageWidth: document.documentElement.scrollWidth,
        viewport: innerWidth,
      };
    });
  expect(layout.textRight).toBeLessThanOrEqual(layout.iconLeft);
  expect(layout.iconWidth).toBeGreaterThanOrEqual(40);
  expect(layout.pageWidth).toBe(layout.viewport);
});

test("add-set dialog keeps its footer reachable in a short enlarged-text viewport", async ({
  page,
}) => {
  await page.setViewportSize({ width: 320, height: 360 });
  await page.goto("/workouts/new?addExercise=true");
  await page.getByRole("button", { name: "Bench Press", exact: true }).click();
  await page.getByRole("button", { name: "Add Set", exact: true }).click();
  const dialog = page.getByRole("dialog", { name: "Add Set", exact: true });
  await page.evaluate(() => (document.documentElement.style.fontSize = "32px"));
  await expect
    .poll(() =>
      dialog.evaluate((element) => {
        const box = element.getBoundingClientRect();
        return (
          box.top >= 0 &&
          box.bottom <= innerHeight &&
          element.scrollHeight > element.clientHeight
        );
      }),
    )
    .toBe(true);
  await dialog.getByRole("button", { name: "Cancel", exact: true }).click();
  await expect(dialog).toHaveCount(0);
});
