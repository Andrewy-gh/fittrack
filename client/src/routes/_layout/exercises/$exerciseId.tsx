import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { exerciseByIdQueryOptions } from '@/lib/api/exercises';
import { getDemoExercisesByIdQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { ExerciseDetail } from '@/components/exercises/exercise-detail';

const exerciseSearchSchema = z.object({
  sortOrder: z.enum(['asc', 'desc']).optional(),
  itemsPerPage: z.coerce.number().int().positive().optional(),
  page: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute('/_layout/exercises/$exerciseId')({
  validateSearch: exerciseSearchSchema,
  params: {
    parse: (params) => {
      const exerciseId = parseInt(params.exerciseId, 10);
      if (isNaN(exerciseId) || !Number.isInteger(exerciseId)) {
        throw new Error('Invalid exerciseId');
      }
      return { exerciseId };
    },
  },
  loader: ({ context, params }) => {
    const exerciseId = params.exerciseId;
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(exerciseByIdQueryOptions(exerciseId));
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoExercisesByIdQueryOptions(exerciseId));
    }

    return { exerciseId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();
  const { sortOrder, itemsPerPage, page } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const { data: exerciseSets } = user
    ? useSuspenseQuery(exerciseByIdQueryOptions(exerciseId))
    : useSuspenseQuery(getDemoExercisesByIdQueryOptions(exerciseId));

  const normalizedSortOrder = sortOrder ?? 'desc';
  const normalizedItemsPerPage = [10, 20, 50].includes(itemsPerPage ?? 10)
    ? (itemsPerPage ?? 10)
    : 10;

  return (
    <ExerciseDetail
      exerciseSets={exerciseSets}
      exerciseId={exerciseId}
      isDemoMode={!user}
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
