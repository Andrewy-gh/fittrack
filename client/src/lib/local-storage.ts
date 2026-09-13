import type { WorkoutCreateWorkoutRequest } from "../client/types.gen";
import * as v from "valibot";
import { recommendationSnapshotSchema } from "@/features/workouts/utils/recommendation-context";

const STORAGE_KEY = "workout-entry-form-data";

export type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;

const getStorageKey = (userId?: string): string => {
  return userId ? `${STORAGE_KEY}-${userId}` : STORAGE_KEY;
};

export type WorkoutDraftStorage = {
  save: (data: WorkoutCreateWorkoutRequest, userId?: string) => void;
  load: (userId?: string) => WorkoutCreateWorkoutRequest | null;
  clear: (userId?: string) => void;
};

const WorkoutSetSchema = v.object({
  reps: v.number(),
  weight: v.optional(v.number()),
  setType: v.picklist(["warmup", "working"]),
});

const WorkoutExerciseSchema = v.object({
  name: v.string(),
  sets: v.array(WorkoutSetSchema),
});

const WorkoutDraftSchema = v.object({
  recommendations: v.optional(
    v.fallback(v.array(recommendationSnapshotSchema), []),
  ),
  date: v.string(),
  notes: v.optional(v.string()),
  workoutFocus: v.optional(v.string()),
  exercises: v.array(WorkoutExerciseSchema),
});

function parseWorkoutDraft(input: unknown): WorkoutCreateWorkoutRequest | null {
  const result = v.safeParse(WorkoutDraftSchema, input);

  if (!result.success) {
    return null;
  }

  const date = new Date(result.output.date);
  if (Number.isNaN(date.getTime())) {
    return null;
  }

  return {
    recommendations: result.output.recommendations,
    date: result.output.date,
    notes: result.output.notes,
    workoutFocus: result.output.workoutFocus,
    exercises: result.output.exercises,
  };
}

function deserializeDraft(
  rawValue: string,
): WorkoutCreateWorkoutRequest | null {
  try {
    const parsed: unknown = JSON.parse(rawValue);
    return parseWorkoutDraft(parsed);
  } catch {
    return null;
  }
}

export function createWorkoutDraftStorage(
  storage?: StorageLike,
): WorkoutDraftStorage {
  return {
    save(data, userId) {
      if (!storage) {
        return;
      }

      try {
        storage.setItem(getStorageKey(userId), JSON.stringify(data));
      } catch (error) {
        console.warn("Failed to save to localStorage:", error);
      }
    },
    load(userId) {
      if (!storage) {
        return null;
      }

      try {
        const saved = storage.getItem(getStorageKey(userId));
        return saved ? deserializeDraft(saved) : null;
      } catch (error) {
        console.warn("Failed to load from localStorage:", error);
        return null;
      }
    },
    clear(userId) {
      if (!storage) {
        return;
      }

      try {
        storage.removeItem(getStorageKey(userId));
      } catch (error) {
        console.warn("Failed to clear localStorage:", error);
      }
    },
  };
}

function getBrowserStorage(): StorageLike | undefined {
  if (typeof window === "undefined") {
    return undefined;
  }

  try {
    return window.localStorage;
  } catch (error) {
    console.warn("Failed to access localStorage:", error);
    return undefined;
  }
}

export const workoutDraftStorage =
  createWorkoutDraftStorage(getBrowserStorage());

export const loadFromLocalStorage = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
): WorkoutCreateWorkoutRequest | null => {
  return draftStorage.load(userId);
};

export const clearLocalStorage = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
) => {
  draftStorage.clear(userId);
};
