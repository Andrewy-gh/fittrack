import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { workoutQueryOptions } from '@/lib/api/workouts';
import { WorkoutDetail } from '@/components/workouts/workout-detail';

export const Route = createFileRoute('/_auth/workouts/$workoutId/')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({
    context,
    params,
  }): Promise<{
    workoutId: number;
  }> => {
    const workoutId = params.workoutId;
    context.queryClient.ensureQueryData(workoutQueryOptions(workoutId));
    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { data: workout } = useSuspenseQuery(workoutQueryOptions(workoutId));
  return <WorkoutDetail workout={workout} />;
}
