import type { CurrentUser, CurrentInternalUser } from "@stackframe/react";
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
  recentExerciseSetsQueryOptions,
} from "@/features/exercises/api/exercises";
import {
  newWorkoutContextQueryOptions,
  workoutsQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions,
} from "@/features/workouts/api/workouts";
import {
  getDemoExercisesQueryOptions,
  getDemoExercisesByIdQueryOptions,
  getDemoExercisesByIdRecentSetsQueryOptions,
  getDemoWorkoutsQueryOptions,
  getDemoWorkoutsByIdQueryOptions,
  getDemoNewWorkoutContextQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from "@/lib/demo-data/query-options";

export function getExercisesQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
) {
  return (
    user ? exercisesQueryOptions() : getDemoExercisesQueryOptions()
  ) as ReturnType<typeof exercisesQueryOptions>;
}

export function getRecentSetsQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
  exerciseId: number,
) {
  return (
    user
      ? recentExerciseSetsQueryOptions(exerciseId)
      : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId)
  ) as ReturnType<typeof recentExerciseSetsQueryOptions>;
}

export function getExerciseByIdQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
  exerciseId: number,
) {
  return (
    user
      ? exerciseByIdQueryOptions(exerciseId)
      : getDemoExercisesByIdQueryOptions(exerciseId)
  ) as ReturnType<typeof exerciseByIdQueryOptions>;
}

export function getWorkoutsQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
) {
  return (
    user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions()
  ) as ReturnType<typeof workoutsQueryOptions>;
}

export function getWorkoutByIdQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
  id: number,
) {
  return (
    user ? workoutQueryOptions(id) : getDemoWorkoutsByIdQueryOptions(id)
  ) as ReturnType<typeof workoutQueryOptions>;
}

export function getNewWorkoutContextQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
) {
  return (
    user
      ? newWorkoutContextQueryOptions()
      : getDemoNewWorkoutContextQueryOptions()
  ) as ReturnType<typeof newWorkoutContextQueryOptions>;
}

export function getWorkoutsFocusQueryOptions(
  user: CurrentUser | CurrentInternalUser | null,
) {
  return (
    user
      ? workoutsFocusValuesQueryOptions()
      : getDemoWorkoutsFocusValuesQueryOptions()
  ) as ReturnType<typeof workoutsFocusValuesQueryOptions>;
}
