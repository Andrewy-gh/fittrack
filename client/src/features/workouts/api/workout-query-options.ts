import type { ApplicationUser } from "@/lib/application-user";
import { normalizeQueryOptions } from "@/lib/query-option-adapter";
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

type WorkoutQueryUser = ApplicationUser | null;

export function getWorkoutListQueryOptions(user: WorkoutQueryUser) {
  return user
    ? normalizeQueryOptions(workoutsQueryOptions())
    : normalizeQueryOptions(getDemoWorkoutsQueryOptions());
}

export function getWorkoutByIdQueryOptions(
  user: WorkoutQueryUser,
  workoutId: number,
) {
  return user
    ? normalizeQueryOptions(workoutQueryOptions(workoutId))
    : normalizeQueryOptions(getDemoWorkoutsByIdQueryOptions(workoutId));
}

export function getNewWorkoutContextQueryOptions(user: WorkoutQueryUser) {
  return user
    ? normalizeQueryOptions(newWorkoutContextQueryOptions())
    : normalizeQueryOptions(getDemoNewWorkoutContextQueryOptions());
}

export function getWorkoutsFocusQueryOptions(user: WorkoutQueryUser) {
  return user
    ? normalizeQueryOptions(workoutsFocusValuesQueryOptions())
    : normalizeQueryOptions(getDemoWorkoutsFocusValuesQueryOptions());
}

export function getWorkoutContributionQueryOptions(user: WorkoutQueryUser) {
  return user
    ? normalizeQueryOptions(contributionDataQueryOptions())
    : normalizeQueryOptions(getDemoContributionDataQueryOptions());
}
