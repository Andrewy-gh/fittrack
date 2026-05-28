import { createFileRoute } from "@tanstack/react-router";
import { exercisesQueryOptions } from "@/features/exercises/api/exercises";
import { ExercisesPage } from "@/features/exercises/pages/exercises-page";
import { getDemoExercisesQueryOptions } from "@/lib/demo-data/query-options";
import { initializeDemoData, clearDemoData } from "@/lib/demo-data/storage";

export const Route = createFileRoute("/_layout/exercises/")({
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(exercisesQueryOptions());
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoExercisesQueryOptions());
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <ExercisesPage isDemoMode={!user} />;
}
