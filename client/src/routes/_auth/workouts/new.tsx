import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth/workouts/new')({
  loader: async ({ context }) => {
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
  },
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="container mx-auto space-y-4 p-4 md:p-12">
      <h1 className="text-3xl font-bold tracking-tight">New Workout</h1>
    </div>
  );
}
