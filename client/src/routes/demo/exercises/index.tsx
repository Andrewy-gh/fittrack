import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { getDemoExercisesQueryOptions } from '@/lib/demo-data/query-options';
import { ExerciseList } from '@/components/exercises/exercise-list';

export const Route = createFileRoute('/demo/exercises/')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { data: exercises } = useSuspenseQuery(getDemoExercisesQueryOptions());
  return <ExerciseList exercises={exercises} />;
}
