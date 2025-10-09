import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { exerciseByIdQueryOptions } from '@/lib/api/exercises';
import { getDemoExercisesByIdQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { ExerciseDetail } from '@/components/exercises/exercise-detail';

export const Route = createFileRoute('/_layout/exercises/$exerciseId')({
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

  const { data: exerciseSets } = user
    ? useSuspenseQuery(exerciseByIdQueryOptions(exerciseId))
    : useSuspenseQuery(getDemoExercisesByIdQueryOptions(exerciseId));

  return <ExerciseDetail exerciseSets={exerciseSets} exerciseId={exerciseId} />;
}
