import type { WorkoutCreateWorkoutRequest, WorkoutUpdateWorkoutRequest } from "@/client";
import type { workout_ExerciseInput, workout_CreateWorkoutRequest } from "@/generated";
import { loadFromLocalStorage } from "@/lib/local-storage";

// MARK: Init values
export const MOCK_VALUES: WorkoutCreateWorkoutRequest | WorkoutUpdateWorkoutRequest = {
  date: new Date().toISOString(), // API expects ISO string
  notes: '',
  exercises: [] as workout_ExerciseInput[],
};

export const getInitialValues = (userId: string): WorkoutCreateWorkoutRequest  => {
  const saved = loadFromLocalStorage(userId);
  return saved || MOCK_VALUES;
};
