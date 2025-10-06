import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  exercisesQueryOptions,
  recentExerciseSetsQueryOptions
} from './exercises';
import {
  workoutsQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions
} from './workouts';
import {
  getDemoExercisesQueryOptions,
  getDemoExercisesByIdRecentSetsQueryOptions,
  getDemoWorkoutsQueryOptions,
  getDemoWorkoutsByIdQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from '@/lib/demo-data/query-options';

export function getExercisesQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? exercisesQueryOptions() : getDemoExercisesQueryOptions();
}

export function getRecentSetsQueryOptions(user: CurrentUser | CurrentInternalUser | null, exerciseId: number) {
  return user
    ? recentExerciseSetsQueryOptions(exerciseId)
    : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId);
}

export function getWorkoutsQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions();
}

export function getWorkoutByIdQueryOptions(user: CurrentUser | CurrentInternalUser | null, id: number) {
  return user ? workoutQueryOptions(id) : getDemoWorkoutsByIdQueryOptions(id);
}

export function getWorkoutsFocusQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return user ? workoutsFocusValuesQueryOptions() : getDemoWorkoutsFocusValuesQueryOptions();
}
