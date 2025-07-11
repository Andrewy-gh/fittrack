import { createFileRoute } from '@tanstack/react-router'
import { WorkoutEntryForm } from '@/components/workout-entry-form'
import type { ExerciseOption } from '@/lib/types';
import { useUser } from '@stackframe/react';
import { fetchExerciseOptions } from '@/lib/api/exercises';

export const Route = createFileRoute('/workouts/new')({
  loader: async (): Promise<ExerciseOption[]> => {
    const user = useUser();
    if (!user) {
        throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
        throw new Error('Access token not found');
    }
    const exercises = await fetchExerciseOptions(accessToken);
    return exercises;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const exercises = Route.useLoaderData();
  return (
    <div className="container mx-auto space-y-4 p-4 md:p-12">
      <WorkoutEntryForm exercises={exercises} />
    </div>
  );
}
