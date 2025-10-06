import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { getDemoWorkoutsQueryOptions } from '@/lib/demo-data/query-options';
import { WorkoutList } from '@/components/workouts/workout-list';

export const Route = createFileRoute('/demo/workouts/')({
  loader: async ({ context }) => {
    context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions());
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { data: workouts } = useSuspenseQuery(getDemoWorkoutsQueryOptions());
  return (
    <WorkoutList
      workouts={workouts}
      hasWorkoutInProgress={false}
      newWorkoutLink="/demo/workouts/new"
      workoutDetailBasePath="/demo/workouts"
    />
  );
}
