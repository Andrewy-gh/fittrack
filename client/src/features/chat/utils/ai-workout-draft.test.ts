import { describe, expect, it } from "vitest";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import { createWorkoutDraftStorage } from "@/lib/local-storage";
import { saveAIWorkoutDraftToWorkoutForm } from "@/features/chat/utils/ai-workout-draft";
import { getInitialValues } from "@/features/workouts/components/form/form-options";

function createMemoryStorage() {
  const values = new Map<string, string>();

  return {
    getItem(key: string) {
      return values.get(key) ?? null;
    },
    setItem(key: string, value: string) {
      values.set(key, value);
    },
    removeItem(key: string) {
      values.delete(key);
    },
  };
}

describe("ai workout draft import", () => {
  it("does not replace an existing form draft when imported context is unsupported", () => {
    const storage = createWorkoutDraftStorage(createMemoryStorage());
    const original = { date: "2026-04-20T12:00:00Z", exercises: [] };
    storage.save(original, "user-123");
    expect(
      saveAIWorkoutDraftToWorkoutForm(
        {
          date: "2026-04-21T12:00:00Z",
          exercises: [
            { name: "Pushups", sets: [{ reps: 8, setType: "working" }] },
          ],
          unsupportedRecommendations: true,
        },
        "user-123",
        storage,
      ),
    ).toBeNull();
    expect(storage.load("user-123")).toEqual(original);
  });
  it("overwrites the saved workout-form draft with the imported ai draft", () => {
    const storage = createMemoryStorage();
    const draftStorage = createWorkoutDraftStorage(storage);
    const initialDraft: AIWorkoutDraft = {
      date: "2026-04-20T12:00:00Z",
      notes: "First draft",
      workoutFocus: "push",
      exercises: [
        {
          name: "Bench Press",
          sets: [{ reps: 8, setType: "working", weight: 185 }],
        },
      ],
    };
    const regeneratedDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    saveAIWorkoutDraftToWorkoutForm(initialDraft, "user-123", draftStorage);
    const savedDraft = saveAIWorkoutDraftToWorkoutForm(
      regeneratedDraft,
      "user-123",
      draftStorage,
    );

    expect(savedDraft).toMatchObject({
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    });
    expect(draftStorage.load("user-123")).toMatchObject({
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    });
  });

  it("preserves catalog, absent weight and zero weight without aliasing the chat draft", () => {
    const draft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00-04:00",
      exercises: [
        {
          name: "Pushups",
          catalog_id: "Pushups",
          sets: [
            { reps: 5, setType: "warmup" },
            { reps: 8, setType: "working", weight: 0 },
          ],
        },
      ],
    };
    const storage = createWorkoutDraftStorage(createMemoryStorage());
    const imported = saveAIWorkoutDraftToWorkoutForm(draft, "owner", storage);
    expect(imported?.date).toBe(draft.date);
    expect(imported?.exercises).toStrictEqual(draft.exercises);
    expect(storage.load("owner")?.exercises).toEqual(draft.exercises);
    const editable = imported?.exercises[0];
    const editableSet = editable?.sets[1];
    if (!editable || !editableSet) throw new Error("expected imported set");
    editable.name = "Edited pushups";
    editableSet.reps = 12;
    expect(draft.exercises[0]?.name).toBe("Pushups");
    expect(draft.exercises[0]?.sets[1]?.reps).toBe(8);
  });

  it("feeds the imported ai draft into the /workouts/new initialization path", () => {
    const storage = createMemoryStorage();
    const draftStorage = createWorkoutDraftStorage(storage);

    saveAIWorkoutDraftToWorkoutForm(
      {
        date: "2026-04-21T12:00:00Z",
        notes: "Use straps if grip fades",
        workoutFocus: "back",
        exercises: [
          {
            name: "Pushups",
            catalog_id: "Pushups",
            sets: [{ reps: 12, setType: "working", weight: 120 }],
          },
        ],
      },
      "user-123",
      draftStorage,
    );

    expect(getInitialValues("user-123", draftStorage)).toMatchObject({
      date: "2026-04-21T12:00:00Z",
      notes: "Use straps if grip fades",
      workoutFocus: "back",
      exercises: [
        {
          name: "Pushups",
          catalog_id: "Pushups",
          sets: [{ reps: 12, setType: "working", weight: 120 }],
        },
      ],
    });
  });
});
