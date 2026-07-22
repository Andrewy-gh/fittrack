import { expect, test } from "@playwright/test";
import {
  authStatePath,
  hasAuthState,
  restoreLocalAIChatAccess,
  seedLocalAIChatConversation,
} from "../helpers/local-e2e-auth";

test.describe("Authenticated - former subscriber AI chat privacy", () => {
  test.use({ storageState: authStatePath });
  let didExpireAIChatAccess = false;

  test.beforeEach(() => {
    test.skip(
      !hasAuthState(),
      "Missing auth state. Enable local E2E auth for the former-subscriber fixture.",
    );
    test.skip(
      !process.env.CI && process.env.E2E_EXCLUSIVE_MUTATION !== "true",
      "This destructive shared-user fixture requires E2E_EXCLUSIVE_MUTATION=true so Playwright uses one worker.",
    );
  });

  test.afterEach(async () => {
    if (didExpireAIChatAccess) {
      await restoreLocalAIChatAccess();
      didExpireAIChatAccess = false;
    }
  });

  test("deletes all chat history without active AI chat access and stays signed in", async ({
    page,
  }) => {
    await seedLocalAIChatConversation({
      title: "Former subscriber private history",
      user_prompt: "Private former subscriber prompt",
      assistant_text: "Private former subscriber response",
      latest_workout_draft: {
        date: "2026-07-22",
        exercises: [
          {
            name: "Goblet Squat",
            sets: [{ reps: 8, setType: "working", weight: 50 }],
          },
        ],
      },
      expire_ai_chat_access_after_seed: true,
    });
    didExpireAIChatAccess = true;

    await page.goto("/");
    const formerSubscriberAccess = await page.evaluate(async () => {
      const session = JSON.parse(
        window.localStorage.getItem("fittrack-local-e2e-auth") ?? "null",
      ) as { userId?: string } | null;
      if (!session?.userId) {
        throw new Error("Local E2E auth session is missing");
      }
      const headers: Record<string, string> = {
        "x-fittrack-dev-e2e-user": session.userId,
      };
      const [features, conversations] = await Promise.all([
        fetch("/api/features/access", { headers }),
        fetch("/api/ai/conversations", { headers }),
      ]);
      return {
        featuresStatus: features.status,
        featureKeys: (await features.json()) as Array<{
          feature_key?: string;
        }>,
        conversationsStatus: conversations.status,
      };
    });
    expect(formerSubscriberAccess.featuresStatus).toBe(200);
    expect(formerSubscriberAccess.featureKeys).not.toContainEqual(
      expect.objectContaining({ feature_key: "ai_chatbot" }),
    );
    expect(formerSubscriberAccess.conversationsStatus).toBe(403);

    await page.goto("/settings");
    await expect(
      page.getByRole("heading", { name: "Account settings" }),
    ).toBeVisible();
    await expect(
      page.getByText(/without an active AI chat subscription/i),
    ).toBeVisible();

    await page
      .getByRole("checkbox", {
        name: /I understand this permanently deletes all of my AI chat history/i,
      })
      .check();
    await page
      .getByRole("button", { name: "Delete all AI chat history" })
      .click();

    await expect(page.getByText("All AI chat history deleted.")).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "Account settings" }),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "AI chat history deleted" }),
    ).toBeDisabled();
  });
});
