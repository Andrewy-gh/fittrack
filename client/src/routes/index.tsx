import { createFileRoute } from '@tanstack/react-router';
import { WorkoutEntryForm } from '@/components/workout-entry-form';

export const Route = createFileRoute('/')({
  // loader: async () => {
  //   const res = await fetch('/api/hello');
  //   if (!res.ok) {
  //     throw new Error('Failed to fetch data');
  //   }
  //   const data = await res.json();
  //   return data;
  // },
  loader: () => {
    const exercises = [
      { id: 1, name: 'Squat' },
      { id: 2, name: 'Deadlift' },
      { id: 3, name: 'Bench Press' },
    ];
    return exercises;
  },
  component: App,
});

function App() {
  // const { message } = Route.useLoaderData();
  const exercises = Route.useLoaderData();
  return (
    <main className="container mx-auto space-y-4 p-4 md:p-12">
      <WorkoutEntryForm exercises={exercises} />
    </main>
  );
}
