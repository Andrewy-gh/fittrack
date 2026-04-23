import { expect, test } from "@playwright/test";
import {
  authStatePath,
  hasAuthState,
  seedLocalAIChatConversation,
} from "../helpers/local-e2e-auth";

test.describe("Authenticated - Structured workout chat import", () => {
  test.use({ storageState: authStatePath });

  test.beforeEach(() => {
    test.skip(
      !hasAuthState(),
      "Missing auth state. Enable local E2E auth or configure Stack bootstrap credentials for Playwright.",
    );
  });

  test("reopens the latest structured workout draft and imports it into the workout form", async ({
    page,
  }) => {
    const seeded = await seedLocalAIChatConversation({
      title: "Structured pull session",
      user_prompt: "Build me a 45 minute pull workout.",
      assistant_text:
        "I saved a structured pull workout draft so you can tweak it in the form.",
      latest_workout_draft: {
        date: "2026-04-21T12:00:00Z",
        notes: "Keep rest short and focus on controlled reps.",
        workoutFocus: "Structured Pull Day",
        exercises: [
          {
            name: "Barbell Row",
            sets: [
              { reps: 8, setType: "working", weight: 135 },
              { reps: 8, setType: "working", weight: 135 },
            ],
          },
          {
            name: "Lat Pulldown",
            sets: [
              { reps: 12, setType: "working", weight: 90 },
              { reps: 12, setType: "working", weight: 90 },
            ],
          },
        ],
      },
    });

    await page.goto(`/chat?conversationId=${seeded.conversation_id}`);

    await expect(page.getByRole("heading", { name: "AI Chat" })).toBeVisible();
    await expect(
      page.getByText("Latest structured workout draft"),
    ).toBeVisible();

    await page.getByRole("button", { name: "Edit in workout form" }).click();

    await expect(page).toHaveURL(/\/workouts\/new$/);
    await expect(
      page.getByText("Keep rest short and focus on controlled reps."),
    ).toBeVisible();
    await expect(page.getByText("Structured Pull Day")).toBeVisible();
    await expect(
      page.getByRole("button", { name: /edit barbell row/i }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: /edit lat pulldown/i }),
    ).toBeVisible();
  });
});
