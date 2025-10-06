import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { exerciseByIdQueryOptions } from '@/lib/api/exercises';
import { ExerciseDetail } from '@/components/exercises/exercise-detail';

export const Route = createFileRoute('/_auth/exercises/$exerciseId')({
  params: {
    parse: (params) => {
      const exerciseId = parseInt(params.exerciseId, 10);
      if (isNaN(exerciseId) || !Number.isInteger(exerciseId)) {
        throw new Error('Invalid exerciseId');
      }
      return { exerciseId };
    },
  },
  loader: async ({ context, params }) => {
    const exerciseId = params.exerciseId;
    context.queryClient.ensureQueryData(exerciseByIdQueryOptions(exerciseId));
    return { exerciseId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId } = Route.useLoaderData();
  const { data: exerciseSets } = useSuspenseQuery(
    exerciseByIdQueryOptions(exerciseId)
  );
  return <ExerciseDetail exerciseSets={exerciseSets} exerciseId={exerciseId} />;
}
