import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { workoutQueryOptions } from '@/lib/api/workouts';
import { getDemoWorkoutsByIdQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { WorkoutDetail } from '@/components/workouts/workout-detail';

export const Route = createFileRoute('/_layout/workouts/$workoutId/')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: ({ context, params }) => {
    const workoutId = params.workoutId;
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(workoutQueryOptions(workoutId));
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoWorkoutsByIdQueryOptions(workoutId));
    }

    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();

  const { data: workout } = user
    ? useSuspenseQuery(workoutQueryOptions(workoutId))
    : useSuspenseQuery(getDemoWorkoutsByIdQueryOptions(workoutId));

  // Show edit/delete in both auth and demo modes (demo has mutations)
  return <WorkoutDetail workout={workout} showEditDelete={true} />;
}
