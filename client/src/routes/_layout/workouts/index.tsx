import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { workoutsQueryOptions } from '@/lib/api/workouts';
import { getDemoWorkoutsQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { loadFromLocalStorage } from '@/lib/local-storage';
import { WorkoutList } from '@/components/workouts/workout-list';

export const Route = createFileRoute('/_layout/workouts/')({
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(workoutsQueryOptions());
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions());
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  const { data: workouts } = user
    ? useSuspenseQuery(workoutsQueryOptions())
    : useSuspenseQuery(getDemoWorkoutsQueryOptions());

  // Check for workout in progress (pass user.id if authenticated, undefined for demo)
  const hasWorkoutInProgress = loadFromLocalStorage(user?.id) !== null;

  return (
    <WorkoutList
      workouts={workouts}
      hasWorkoutInProgress={hasWorkoutInProgress}
    />
  );
}
