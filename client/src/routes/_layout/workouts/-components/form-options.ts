import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
  WorkoutUpdateWorkoutRequest,
} from '@/client';
import {
  loadFromLocalStorage,
  type WorkoutDraftStorage,
  workoutDraftStorage,
} from '@/lib/local-storage';

export const MOCK_VALUES: WorkoutCreateWorkoutRequest | WorkoutUpdateWorkoutRequest = {
  date: new Date().toISOString(), // API expects ISO string
  notes: '',
  exercises: [] as Array<WorkoutExerciseInput>,
  workoutFocus: '',
};

export const getInitialValues = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage
): WorkoutCreateWorkoutRequest => {
  const saved = loadFromLocalStorage(userId, draftStorage);
  return saved || MOCK_VALUES;
};
