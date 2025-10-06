import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { exercisesQueryOptions } from '@/lib/api/exercises';
import { ExerciseList } from '@/components/exercises/exercise-list';

export const Route = createFileRoute('/_auth/exercises/')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(exercisesQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { data: exercises } = useSuspenseQuery(exercisesQueryOptions());
  return <ExerciseList exercises={exercises} />;
}
