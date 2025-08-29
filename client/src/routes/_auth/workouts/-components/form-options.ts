import type { WorkoutCreateWorkoutRequest, WorkoutExerciseInput, WorkoutUpdateWorkoutRequest } from "@/client";
import { loadFromLocalStorage } from "@/lib/local-storage";

export const MOCK_VALUES: WorkoutCreateWorkoutRequest | WorkoutUpdateWorkoutRequest = {
  date: new Date().toISOString(), // API expects ISO string
  notes: '',
  exercises: [] as Array<WorkoutExerciseInput>,
};

export const getInitialValues = (userId: string): WorkoutCreateWorkoutRequest  => {
  const saved = loadFromLocalStorage(userId);
  return saved || MOCK_VALUES;
};
