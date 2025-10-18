import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import type { ExerciseExerciseResponse } from '@/client';
import {
  getExercisesQueryOptions,
  getExercisesByIdQueryOptions,
  getExercisesByIdRecentSetsQueryOptions,
  deleteExercisesByIdMutation,
  patchExercisesByIdMutation,
  getExercisesQueryKey,
  getExercisesByIdQueryKey,
} from '@/client/@tanstack/react-query.gen';

/**
 * Exercise data as returned by the API - always has a database ID
 */
export type DbExercise = Pick<ExerciseExerciseResponse, 'id' | 'name'>;

/**
 * Exercise option for form components - may include manually created exercises without IDs
 * Used in dropdowns and forms where users can create new exercises on-the-fly
 */
export type ExerciseOption = {
  id: number | null; // null for manually created exercises, number for DB exercises
  name: string;
};

export function exercisesQueryOptions() {
  return getExercisesQueryOptions();
}

export function exerciseByIdQueryOptions(id: number) {
  return getExercisesByIdQueryOptions({ path: { id } });
}

export function recentExerciseSetsQueryOptions(id: number) {
  return getExercisesByIdRecentSetsQueryOptions({ path: { id } });
}

export function useDeleteExerciseMutation() {
  return useMutation({
    ...deleteExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getExercisesQueryKey(),
      });
      queryClient.removeQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id } }),
      });
    },
  });
}

export function useUpdateExerciseMutation() {
  return useMutation({
    ...patchExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getExercisesQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id } }),
      });
    },
  });
}