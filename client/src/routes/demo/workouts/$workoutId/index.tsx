import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { getDemoWorkoutsByIdQueryOptions } from '@/lib/demo-data/query-options';
import { WorkoutDetail } from '@/components/workouts/workout-detail';

export const Route = createFileRoute('/demo/workouts/$workoutId/')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({ context, params }) => {
    const workoutId = params.workoutId;
    context.queryClient.ensureQueryData(
      getDemoWorkoutsByIdQueryOptions(workoutId)
    );
    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { data: workout } = useSuspenseQuery(
    getDemoWorkoutsByIdQueryOptions(workoutId)
  );
  return (
    <WorkoutDetail
      workout={workout}
      exerciseDetailBasePath="/demo/exercises"
      showEditDelete={false}
    />
  );
}
