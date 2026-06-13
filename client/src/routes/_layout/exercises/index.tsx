import { createFileRoute } from "@tanstack/react-router";
import { getExerciseListQueryOptions } from "@/features/exercises/api/exercise-query-options";
import { ExercisesPage } from "@/features/exercises/pages/exercises-page";
import { initializeDemoData, clearDemoData } from "@/lib/demo-data/storage";

export const Route = createFileRoute("/_layout/exercises/")({
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(getExerciseListQueryOptions(user));
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getExerciseListQueryOptions(user));
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <ExercisesPage user={user} />;
}
