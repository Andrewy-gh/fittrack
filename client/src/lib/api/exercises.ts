import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import type { ExerciseExerciseResponse } from '@/client';
import {
  getExercisesQueryOptions,
  getExercisesByIdQueryOptions,
  getExercisesByIdHistorical1RmQueryOptions,
  getExercisesByIdRecentSetsQueryOptions,
  getExercisesByIdMetricsHistoryQueryOptions,
  deleteExercisesByIdMutation,
  patchExercisesByIdMutation,
  patchExercisesByIdHistorical1RmMutation,
  getExercisesQueryKey,
  getExercisesByIdQueryKey,
  getExercisesByIdHistorical1RmQueryKey,
} from '@/client/@tanstack/react-query.gen';

export type DbExercise = Pick<ExerciseExerciseResponse, 'id' | 'name'>;

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

export function exerciseHistorical1RmQueryOptions(id: number) {
  return getExercisesByIdHistorical1RmQueryOptions({ path: { id } });
}

export function recentExerciseSetsQueryOptions(id: number) {
  return getExercisesByIdRecentSetsQueryOptions({ path: { id } });
}

export type MetricsHistoryRange = 'W' | 'M' | '6M' | 'Y';

export function exerciseMetricsHistoryQueryOptions(id: number, range: MetricsHistoryRange) {
  return getExercisesByIdMetricsHistoryQueryOptions({
    path: { id },
    query: { range },
  });
}

export function useDeleteExerciseMutation() {
  return useMutation({
    ...deleteExercisesByIdMutation(),
    meta: { skipGlobalErrorHandler: true },
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

export function useUpdateExerciseHistorical1RmMutation() {
  return useMutation({
    ...patchExercisesByIdHistorical1RmMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getExercisesByIdHistorical1RmQueryKey({ path: { id } }),
      });

      // historical 1RM affects metrics-history intensity calculations; invalidate all ranges.
      queryClient.invalidateQueries({
        predicate: (q) => {
          const key0 = q.queryKey?.[0] as any;
          return (
            key0?._id === 'getExercisesByIdMetricsHistory' &&
            key0?.path?.id === id
          );
        },
      });
    },
  });
}
