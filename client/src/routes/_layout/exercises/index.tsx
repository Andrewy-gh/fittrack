import { noop } from "@tanstack/react-query";
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
      void context.queryClient
        .query({ ...getExerciseListQueryOptions(user), staleTime: "static" })
        .catch(noop);
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      void context.queryClient
        .query({ ...getExerciseListQueryOptions(user), staleTime: "static" })
        .catch(noop);
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <ExercisesPage user={user} />;
}
