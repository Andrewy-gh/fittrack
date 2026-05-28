import { useSuspenseQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { ExerciseDetail } from "@/features/exercises/components/exercise-detail";
import { exerciseByIdQueryOptions } from "@/features/exercises/api/exercises";
import { getDemoExercisesByIdQueryOptions } from "@/lib/demo-data/query-options";

type ExerciseDetailPageProps = {
  exerciseId: number;
  isDemoMode: boolean;
  sortOrder?: "asc" | "desc";
  itemsPerPage?: number;
  page?: number;
};

export function ExerciseDetailPage({
  exerciseId,
  isDemoMode,
  sortOrder,
  itemsPerPage,
  page,
}: ExerciseDetailPageProps) {
  const navigate = useNavigate({ from: "/exercises/$exerciseId" });

  const { data: exerciseDetail } = isDemoMode
    ? useSuspenseQuery(getDemoExercisesByIdQueryOptions(exerciseId))
    : useSuspenseQuery(exerciseByIdQueryOptions(exerciseId));

  const normalizedSortOrder = sortOrder ?? "desc";
  const normalizedItemsPerPage = [10, 20, 50].includes(itemsPerPage ?? 10)
    ? (itemsPerPage ?? 10)
    : 10;

  const safeExerciseSets = Array.isArray(exerciseDetail?.sets)
    ? exerciseDetail.sets
    : [];

  return (
    <ExerciseDetail
      exercise={exerciseDetail.exercise}
      exerciseSets={safeExerciseSets}
      exerciseId={exerciseId}
      isDemoMode={isDemoMode}
      sortOrder={normalizedSortOrder}
      itemsPerPage={normalizedItemsPerPage}
      page={page}
      onSortOrderChange={(nextSortOrder) =>
        navigate({
          search: (prev) => ({
            ...prev,
            sortOrder: nextSortOrder,
            page: 1,
          }),
        })
      }
      onItemsPerPageChange={(nextItemsPerPage) =>
        navigate({
          search: (prev) => ({
            ...prev,
            itemsPerPage: nextItemsPerPage,
            page: 1,
          }),
        })
      }
      onPageChange={(nextPage) =>
        navigate({
          search: (prev) => ({
            ...prev,
            page: nextPage,
          }),
        })
      }
    />
  );
}
