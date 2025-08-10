import { createFileRoute } from '@tanstack/react-router';
import type { ExerciseOption } from '@/lib/api/exercises';
import { fetchExerciseOptions } from '@/lib/api/exercises';
import { WorkoutEntryForm } from '@/components/workout-entry-form';

export const Route = createFileRoute('/_auth/workouts/new')({
  loader: async ({
    context,
  }): Promise<{ accessToken: string; userId: string; exercises: ExerciseOption[] }> => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    if (!user.id || typeof user.id !== 'string') {
      throw new Error('User ID not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const exercises = await fetchExerciseOptions(accessToken);
    return { accessToken, userId: user.id, exercises };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { accessToken, exercises, userId } = Route.useLoaderData();
  return (
    <div className="container mx-auto space-y-4 p-4 md:p-12">
      <WorkoutEntryForm exercises={exercises} accessToken={accessToken} userId={userId} />
    </div>
  );
}
