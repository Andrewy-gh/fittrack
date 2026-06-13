import { queryClient } from "@/lib/api/api";
import { useMutation } from "@tanstack/react-query";
import type { ExerciseExerciseResponse } from "@/client";
import {
  getExercisesQueryOptions,
  getExercisesByIdQueryOptions,
  getExercisesByIdRecentSetsQueryOptions,
  getExercisesByIdMetricsHistoryQueryOptions,
  deleteExercisesByIdMutation,
  patchExercisesByIdMutation,
  patchExercisesByIdHistorical1RmMutation,
  getExercisesQueryKey,
  getExercisesByIdQueryKey,
} from "@/client/@tanstack/react-query.gen";
import {
  deleteDemoExercisesByIdMutationWithMeta,
  patchDemoExercisesByIdMutation,
} from "@/lib/demo-data/query-options";

export type DbExercise = Pick<ExerciseExerciseResponse, "id" | "name">;

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

export type MetricsHistoryRange = "W" | "M" | "6M" | "Y";

export function exerciseMetricsHistoryQueryOptions(
  id: number,
  range: MetricsHistoryRange,
) {
  return getExercisesByIdMetricsHistoryQueryOptions({
    path: { id },
    query: { range },
  });
}

function invalidateExerciseDetail(id: number) {
  queryClient.invalidateQueries({
    queryKey: getExercisesQueryKey(),
  });
  queryClient.invalidateQueries({
    queryKey: getExercisesByIdQueryKey({ path: { id } }),
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

export function useDeleteExerciseForModeMutation(isDemoMode: boolean) {
  const apiMutation = useDeleteExerciseMutation();
  const demoMutation = useMutation(deleteDemoExercisesByIdMutationWithMeta());

  return isDemoMode ? demoMutation : apiMutation;
}

export function useUpdateExerciseMutation() {
  return useMutation({
    ...patchExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      invalidateExerciseDetail(id);
    },
  });
}

export function useRenameExerciseMutation(isDemoMode: boolean) {
  const apiMutation = useMutation({
    ...patchExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      invalidateExerciseDetail(id);
    },
    onError: () => {
      // The edit dialog renders duplicate-name errors inline and other errors as a toast.
    },
  });
  const demoMutation = useMutation({
    ...patchDemoExercisesByIdMutation(),
    onError: () => {
      // The edit dialog owns the user-facing error message.
    },
  });

  return isDemoMode ? demoMutation : apiMutation;
}

export function useUpdateExerciseHistorical1RmMutation() {
  return useMutation({
    ...patchExercisesByIdHistorical1RmMutation(),
    onSuccess: (_, { path: { id } }) => {
      // GET /exercises/{id} now carries historical_1rm fields; keep it fresh.
      queryClient.invalidateQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id } }),
      });

      // historical 1RM affects metrics-history intensity calculations; invalidate all ranges.
      queryClient.invalidateQueries({
        predicate: (q) => {
          const key0 = q.queryKey?.[0] as any;
          return (
            key0?._id === "getExercisesByIdMetricsHistory" &&
            key0?.path?.id === id
          );
        },
      });
    },
  });
}
