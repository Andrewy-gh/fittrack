import type { WorkoutCreateWorkoutRequest } from "../client/types.gen";
import * as v from "valibot";

const STORAGE_KEY = "workout-entry-form-data";

export type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;

const getStorageKey = (userId?: string): string => {
  return userId ? `${STORAGE_KEY}-${userId}` : STORAGE_KEY;
};

// Type for form data that may contain Date objects
export type FormDataType = Omit<WorkoutCreateWorkoutRequest, "date"> & {
  date: Date | string;
};

export type WorkoutDraftStorage = {
  save: (data: FormDataType, userId?: string) => void;
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
  date: v.string(),
  notes: v.optional(v.string()),
  workoutFocus: v.optional(v.string()),
  exercises: v.array(WorkoutExerciseSchema),
});

function serializeDraft(data: FormDataType): WorkoutCreateWorkoutRequest {
  return {
    ...data,
    date: data.date instanceof Date ? data.date.toISOString() : data.date,
  };
}

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
    // SAFETY: Existing draft consumers expect storage to rehydrate the saved
    // ISO string into a Date for the date picker, while the generated request
    // type still represents the submitted JSON payload.
    date: date as unknown as string,
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
        storage.setItem(
          getStorageKey(userId),
          JSON.stringify(serializeDraft(data)),
        );
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

export const saveToLocalStorage = (
  data: FormDataType,
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
) => {
  draftStorage.save(data, userId);
};

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
