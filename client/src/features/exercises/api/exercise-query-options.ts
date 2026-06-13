import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
  recentExerciseSetsQueryOptions,
} from "@/features/exercises/api/exercises";
import {
  getDemoExercisesByIdQueryOptions,
  getDemoExercisesByIdRecentSetsQueryOptions,
  getDemoExercisesQueryOptions,
} from "@/lib/demo-data/query-options";

type ExerciseQueryUser = CurrentUser | CurrentInternalUser | null;

export function getExerciseListQueryOptions(user: ExerciseQueryUser) {
  return (
    user ? exercisesQueryOptions() : getDemoExercisesQueryOptions()
  ) as ReturnType<typeof exercisesQueryOptions>;
}

export function getExerciseDetailQueryOptions(
  user: ExerciseQueryUser,
  exerciseId: number,
) {
  return (
    user
      ? exerciseByIdQueryOptions(exerciseId)
      : getDemoExercisesByIdQueryOptions(exerciseId)
  ) as ReturnType<typeof exerciseByIdQueryOptions>;
}

export function getRecentExerciseSetsQueryOptions(
  user: ExerciseQueryUser,
  exerciseId: number,
) {
  return (
    user
      ? recentExerciseSetsQueryOptions(exerciseId)
      : getDemoExercisesByIdRecentSetsQueryOptions(exerciseId)
  ) as ReturnType<typeof recentExerciseSetsQueryOptions>;
}
