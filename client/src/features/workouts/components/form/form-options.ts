import type {
  WorkoutCreateWorkoutRequest,
  WorkoutUpdateWorkoutRequest,
} from "@/client";
import {
  type WorkoutDraftStorage,
  type StoredWorkoutDraft,
  workoutDraftStorage,
} from "@/lib/local-storage";

export const MOCK_VALUES:
  | WorkoutCreateWorkoutRequest
  | WorkoutUpdateWorkoutRequest = {
  date: "", // Type witness for composed fields; new drafts use fresh defaults below.
  notes: "",
  exercises: [],
  workoutFocus: "",
};

/** Create a fresh, empty workout at the moment a draft is started. */
export const createNewWorkoutValues = (): WorkoutCreateWorkoutRequest => ({
  date: new Date().toISOString(),
  notes: "",
  exercises: [],
  workoutFocus: "",
});

/** Restore a saved draft without replacing its date, or start a new one. */
export const getInitialValues = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
): StoredWorkoutDraft => {
  const saved = draftStorage.load(userId);
  return saved ?? createNewWorkoutValues();
};
