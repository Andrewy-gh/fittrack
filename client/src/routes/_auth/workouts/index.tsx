import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { workoutsQueryOptions } from '@/lib/api/workouts';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import { loadFromLocalStorage } from '@/lib/local-storage';
import { WorkoutList } from '@/components/workouts/workout-list';

export const Route = createFileRoute('/_auth/workouts/')({
  loader: async ({
    context,
  }): Promise<{
    user: CurrentUser | CurrentInternalUser;
  }> => {
    const user = context.user;
    context.queryClient.ensureQueryData(workoutsQueryOptions());
    return { user };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();
  const { data: workouts } = useSuspenseQuery(workoutsQueryOptions());
  const hasWorkoutInProgress = loadFromLocalStorage(user.id) !== null;

  return (
    <WorkoutList
      workouts={workouts}
      hasWorkoutInProgress={hasWorkoutInProgress}
    />
  );
}
