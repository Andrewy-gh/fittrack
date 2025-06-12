import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/workouts/$workoutId')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({ params }) => {
    const workoutId = params.workoutId;
    const res = await fetch(`/api/workouts/${workoutId}`);
    if (!res.ok) {
      throw new Error('Failed to fetch workout');
    }
    const workout = await res.json();
    return workout;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const workout = Route.useLoaderData();
  return <div>{JSON.stringify(workout)}</div>;
}
