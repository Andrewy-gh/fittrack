import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/workouts')({
  loader: async () => {
    const res = await fetch('/api/workouts');
    if (!res.ok) {
      throw new Error('Failed to fetch workouts');
    }
    const data = await res.json();
    return data;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const workouts = Route.useLoaderData();
  return <div>{JSON.stringify(workouts)}</div>;
}
