import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { exercisesQueryOptions } from '@/lib/api/exercises';
import { getDemoExercisesQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { ExerciseList } from '@/components/exercises/exercise-list';

export const Route = createFileRoute('/exercises/')({
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      await context.queryClient.ensureQueryData(exercisesQueryOptions());
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      await context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  // Use separate query calls to avoid type conflicts
  const { data: exercises } = user
    ? useSuspenseQuery(exercisesQueryOptions())
    : useSuspenseQuery(getDemoExercisesQueryOptions());

  return <ExerciseList exercises={exercises} />;
}
