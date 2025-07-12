import { createFileRoute } from '@tanstack/react-router';
import type { ExerciseOption } from '@/lib/types';
import { fetchExerciseOptions } from '@/lib/api/exercises';
import { WorkoutEntryForm } from '@/components/workout-entry-form';

export const Route = createFileRoute('/_auth/workouts/new')({
  loader: async ({
    context,
  }): Promise<{ accessToken: string; exercises: ExerciseOption[] }> => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const exercises = await fetchExerciseOptions(accessToken);
    return { accessToken, exercises };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { accessToken, exercises } = Route.useLoaderData();
  return (
    <div className="container mx-auto space-y-4 p-4 md:p-12">
      <WorkoutEntryForm exercises={exercises} accessToken={accessToken} />
    </div>
  );
}
