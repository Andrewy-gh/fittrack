import type { ApplicationUser } from "@/lib/application-user";
import { normalizeQueryOptions } from "@/lib/query-option-adapter";
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

type ExerciseQueryUser = ApplicationUser | null;

export function getExerciseListQueryOptions(user: ExerciseQueryUser) {
  return user
    ? normalizeQueryOptions(exercisesQueryOptions())
    : normalizeQueryOptions(getDemoExercisesQueryOptions());
}

export function getExerciseDetailQueryOptions(
  user: ExerciseQueryUser,
  exerciseId: number,
) {
  return user
    ? normalizeQueryOptions(exerciseByIdQueryOptions(exerciseId))
    : normalizeQueryOptions(getDemoExercisesByIdQueryOptions(exerciseId));
}

export function getRecentExerciseSetsQueryOptions(
  user: ExerciseQueryUser,
  exerciseId: number,
) {
  return user
    ? normalizeQueryOptions(recentExerciseSetsQueryOptions(exerciseId))
    : normalizeQueryOptions(
        getDemoExercisesByIdRecentSetsQueryOptions(exerciseId),
      );
}
