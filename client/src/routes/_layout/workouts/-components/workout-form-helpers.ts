import { compose, maxValue, minValue } from '@/lib/validation';
import type { WorkoutCreateWorkoutRequest } from '@/client';

type WorkoutSetLike = {
  weight?: number | null;
  reps?: number | null;
  setType?: string | null;
};

const DEFAULT_SET_TYPE = 'working';

export const validateSetReps = (value: unknown) =>
  compose(minValue(1), maxValue(1000))(value, 'Reps');

export const isSetEmptyForDismiss = (set?: WorkoutSetLike): boolean => {
  const weight = Number(set?.weight ?? 0);
  const reps = Number(set?.reps ?? 0);
  const setType = set?.setType ?? DEFAULT_SET_TYPE;

  return weight <= 0 && reps <= 0 && setType === DEFAULT_SET_TYPE;
};

export const shouldDiscardNewExerciseAfterSetRemoval = (
  isNewExercise: boolean | undefined,
  setCount: number
): boolean => Boolean(isNewExercise) && setCount <= 1;

export const hasWorkoutDraftContent = (
  value: Pick<WorkoutCreateWorkoutRequest, 'notes' | 'workoutFocus' | 'exercises'>
): boolean => {
  const hasNotes = Boolean(value.notes?.trim());
  const hasWorkoutFocus = Boolean(value.workoutFocus?.trim());
  const hasExercises = value.exercises.length > 0;

  return hasNotes || hasWorkoutFocus || hasExercises;
};
