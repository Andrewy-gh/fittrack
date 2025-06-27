import { createFileRoute } from '@tanstack/react-router';
import { WorkoutEntryForm } from '@/components/workout-entry-form';
import type { ExerciseOption } from '@/lib/types';

export const Route = createFileRoute('/')({
  loader: async () => {
    const res = await fetch('/api/exercises');
    if (!res.ok) {
      throw new Error('Failed to fetch data');
    }
    const data = await res.json();
    return data as ExerciseOption[];
  },
  component: App,
});

function App() {
  const exercises = Route.useLoaderData();
  return (
    <main className="container mx-auto space-y-4 p-4 md:p-12">
      <WorkoutEntryForm exercises={exercises} />
    </main>
  );
}
