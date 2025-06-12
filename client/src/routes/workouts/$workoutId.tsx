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
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useParams();
  return <div>{workoutId}</div>;
}
