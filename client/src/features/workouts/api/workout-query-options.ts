import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import {
  contributionDataQueryOptions,
  newWorkoutContextQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions,
  workoutsQueryOptions,
} from "@/features/workouts/api/workouts";
import {
  getDemoContributionDataQueryOptions,
  getDemoNewWorkoutContextQueryOptions,
  getDemoWorkoutsByIdQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
  getDemoWorkoutsQueryOptions,
} from "@/lib/demo-data/query-options";

type WorkoutQueryUser = CurrentUser | CurrentInternalUser | null;

export function getWorkoutListQueryOptions(user: WorkoutQueryUser) {
  return (
    user ? workoutsQueryOptions() : getDemoWorkoutsQueryOptions()
  ) as ReturnType<typeof workoutsQueryOptions>;
}

export function getWorkoutByIdQueryOptions(
  user: WorkoutQueryUser,
  workoutId: number,
) {
  return (
    user
      ? workoutQueryOptions(workoutId)
      : getDemoWorkoutsByIdQueryOptions(workoutId)
  ) as ReturnType<typeof workoutQueryOptions>;
}

export function getNewWorkoutContextQueryOptions(user: WorkoutQueryUser) {
  return (
    user
      ? newWorkoutContextQueryOptions()
      : getDemoNewWorkoutContextQueryOptions()
  ) as ReturnType<typeof newWorkoutContextQueryOptions>;
}

export function getWorkoutsFocusQueryOptions(user: WorkoutQueryUser) {
  return (
    user
      ? workoutsFocusValuesQueryOptions()
      : getDemoWorkoutsFocusValuesQueryOptions()
  ) as ReturnType<typeof workoutsFocusValuesQueryOptions>;
}

export function getWorkoutContributionQueryOptions(user: WorkoutQueryUser) {
  return (
    user
      ? contributionDataQueryOptions()
      : getDemoContributionDataQueryOptions()
  ) as ReturnType<typeof contributionDataQueryOptions>;
}
