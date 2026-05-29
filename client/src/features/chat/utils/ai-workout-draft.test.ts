import { describe, expect, it } from "vitest";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import { createWorkoutDraftStorage } from "@/lib/local-storage";
import {
  saveAIWorkoutDraftToWorkoutForm,
  toWorkoutCreateRequest,
} from "@/features/chat/utils/ai-workout-draft";
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

    expect(savedDraft).toEqual({
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working", weight: undefined }],
        },
      ],
    });
    expect(draftStorage.load("user-123")).toEqual({
      date: new Date("2026-04-21T12:00:00Z"),
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
            name: "Lat Pulldown",
            sets: [{ reps: 12, setType: "working", weight: 120 }],
          },
        ],
      },
      "user-123",
      draftStorage,
    );

    expect(getInitialValues("user-123", draftStorage)).toEqual({
      date: new Date("2026-04-21T12:00:00Z"),
      notes: "Use straps if grip fades",
      workoutFocus: "back",
      exercises: [
        {
          name: "Lat Pulldown",
          sets: [{ reps: 12, setType: "working", weight: 120 }],
        },
      ],
    });
  });

  it("normalizes optional text fields for direct workout creation", () => {
    expect(
      toWorkoutCreateRequest({
        date: "2026-04-21T12:00:00Z",
        notes: "  ",
        workoutFocus: "  legs  ",
        exercises: [
          {
            name: "Back Squat",
            sets: [{ reps: 5, setType: "working", weight: 225 }],
          },
        ],
      }),
    ).toEqual({
      date: "2026-04-21T12:00:00Z",
      notes: undefined,
      workoutFocus: "legs",
      exercises: [
        {
          name: "Back Squat",
          sets: [{ reps: 5, setType: "working", weight: 225 }],
        },
      ],
    });
  });
});
