import { createFileRoute } from '@tanstack/react-router';
import { WorkoutEntryForm } from '@/components/workout-entry-form';
import { fetchExerciseOptions } from '@/lib/api/exercises';

export const Route = createFileRoute('/')({
  loader: () => fetchExerciseOptions(),
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
