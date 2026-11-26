import { expect, test } from "@playwright/test";
import type {
  AichatSaveLatestWorkoutDraftResponse,
  WorkoutCreateWorkoutRequest,
  WorkoutWorkoutDetailResponse,
  WorkoutWorkoutResponse,
} from "../../../src/client/types.gen";
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

  test("preserves stored chat text and recommendations through import, edit, reload and undo", async ({
    page,
    baseURL,
  }, testInfo) => {
    test.skip(
      process.env.E2E_LOCAL_AUTH_ENABLED !== "true",
      "Uses synthetic chat and exercise fixtures in the isolated local E2E database",
    );
    for (const endpoint of [baseURL, process.env.E2E_LOCAL_AUTH_API_BASE_URL]) {
      expect(["127.0.0.1", "localhost", "[::1]"]).toContain(
        new URL(endpoint ?? "").hostname,
      );
    }
    const headers = {
      "x-fittrack-dev-e2e-user":
        process.env.E2E_LOCAL_AUTH_USER_ID ?? "local-e2e-user",
    };
    // The server validates that a snapshot belongs to the exercise being saved.
    const suffix = `${Date.now()}-${testInfo.workerIndex}`;
    const name = `Chat row ${suffix}`;
    const exercise = await page.request.post("/api/exercises", {
      headers,
      data: { name },
    });
    expect(exercise.ok()).toBeTruthy();
    const { id: exerciseId }: { id: number } = await exercise.json();
    expect(exerciseId).toBeGreaterThan(0);
    const notes = "\uFEFFKeep rest short\uFEFF";
    const focus = `\uFEFFStructured Pull Day ${suffix}\uFEFF`;
    const draft = {
      date: "2026-04-21T12:00:00Z",
      notes,
      workoutFocus: focus,
      exercises: [
        {
          name,
          sets: [
            { reps: 5, setType: "warmup" },
            { reps: 8, setType: "working", weight: 0 },
          ],
        },
      ],
      recommendations: [
        {
          exerciseName: name,
          recommendation: {
            exerciseId,
            readiness: "normal",
            explanation: "Synthetic historical context",
            policyVersion: "exercise-plan-v2",
            plan: { sets: 3, reps: 8, weight: null },
          },
        },
      ],
    } satisfies WorkoutCreateWorkoutRequest;
    const seeded = await seedLocalAIChatConversation({
      title: "Structured pull session",
      user_prompt: "Build me a 45 minute pull workout.",
      assistant_text: "I saved a structured pull workout draft.",
      latest_workout_draft: draft,
    });

    await page.goto(`/chat?conversationId=${seeded.conversation_id}`);

    await expect(
      page.getByText("Build me a 45 minute pull workout.", { exact: true }),
    ).toBeVisible();
    await expect(
      page.getByText("Latest structured workout draft"),
    ).toBeVisible();

    const directSave = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname ===
          `/api/ai/conversations/${seeded.conversation_id}/latest-workout-draft/save` &&
        response.request().method() === "POST",
    );
    await page.getByRole("button", { name: "Save now", exact: true }).click();
    const directResponse = await directSave;
    expect(directResponse.ok()).toBeTruthy();
    const { workout_id: directId }: AichatSaveLatestWorkoutDraftResponse =
      await directResponse.json();
    expect(directId).toBeGreaterThan(0);

    await page.getByRole("button", { name: "Edit in workout form" }).click();
    await expect(page).toHaveURL(/\/workouts\/new$/);
    const eligible = page.getByText(
      /original recommendation context will be saved/i,
    );
    const edited = page.getByText(
      /original recommendation context will not be saved/i,
    );
    const notesCard = page.getByRole("button", { name: "Notes", exact: true });
    const dialog = page.getByRole("dialog", { name: "Notes" });
    const notesInput = dialog.getByRole("textbox", { name: "Notes" });
    await expect(eligible).toBeVisible();
    await notesCard.click();
    // Playwright text matchers normalize whitespace; read the actual value.
    expect(await notesInput.inputValue()).toBe(notes);
    await notesInput.fill("Keep rest short");
    await notesInput.press("Escape");
    await expect(dialog).toBeHidden();
    await expect(edited).toBeVisible();

    await page.reload();
    await expect(edited).toBeVisible();
    await notesCard.click();
    expect(await notesInput.inputValue()).toBe("Keep rest short");
    await notesInput.fill(notes);
    await notesInput.press("Escape");
    await expect(dialog).toBeHidden();
    await expect(eligible).toBeVisible();

    const importSave = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === "/api/workouts" &&
        response.request().method() === "POST",
    );
    await page.getByRole("button", { name: "Save", exact: true }).click();
    expect((await importSave).ok()).toBeTruthy();

    const list = await page.request.get("/api/workouts", { headers });
    expect(list.ok()).toBeTruthy();
    const rows: WorkoutWorkoutResponse[] = await list.json();
    const imported = rows.filter(
      (row) => row.workout_focus === focus && row.id !== directId,
    );
    expect(imported).toHaveLength(1);

    async function readWorkout(id: number | undefined) {
      const response = await page.request.get(`/api/workouts/${id}`, {
        headers,
      });
      expect(response.ok()).toBeTruthy();
      const { recommendations, sets }: WorkoutWorkoutDetailResponse =
        await response.json();
      return {
        recommendations,
        sets: sets.map(
          ({
            exercise_id,
            workout_date,
            workout_notes,
            workout_focus,
            set_type,
            reps,
            weight,
          }) => ({
            exercise_id,
            workout_date: Date.parse(workout_date),
            workout_notes,
            workout_focus,
            set_type,
            reps,
            weight,
          }),
        ),
      };
    }
    const direct = await readWorkout(directId);
    expect(direct).toEqual({
      recommendations: draft.recommendations,
      sets: [
        {
          exercise_id: exerciseId,
          workout_date: Date.parse(draft.date),
          workout_notes: notes,
          workout_focus: focus,
          set_type: "warmup",
          reps: 5,
          weight: undefined,
        },
        {
          exercise_id: exerciseId,
          workout_date: Date.parse(draft.date),
          workout_notes: notes,
          workout_focus: focus,
          set_type: "working",
          reps: 8,
          weight: 0,
        },
      ],
    });
    expect(await readWorkout(imported[0]?.id)).toEqual(direct);
  });
});
