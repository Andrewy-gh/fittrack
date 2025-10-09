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
  return (user ? exercisesQueryOptions() : getDemoExercisesQueryOptions()) as ReturnType<typeof exercisesQueryOptions>;
}

export function getRecentSetsQueryOptions(user: CurrentUser | CurrentInternalUser | null, exerciseId: number) {
  return (user
    ? recentExerciseSetsQueryOptions(exerciseId)
    : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId)) as ReturnType<typeof recentExerciseSetsQueryOptions>;
}

export function getWorkoutsQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return (user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions()) as ReturnType<typeof workoutsQueryOptions>;
}

export function getWorkoutByIdQueryOptions(user: CurrentUser | CurrentInternalUser | null, id: number) {
  return (user ? workoutQueryOptions(id) : getDemoWorkoutsByIdQueryOptions(id)) as ReturnType<typeof workoutQueryOptions>;
}

export function getWorkoutsFocusQueryOptions(user: CurrentUser | CurrentInternalUser | null) {
  return (user ? workoutsFocusValuesQueryOptions() : getDemoWorkoutsFocusValuesQueryOptions()) as ReturnType<typeof workoutsFocusValuesQueryOptions>;
}
