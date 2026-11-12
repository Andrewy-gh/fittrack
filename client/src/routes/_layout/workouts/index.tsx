import { noop } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { contributionDataQueryOptions } from "@/features/workouts/api/workouts";
import { getWorkoutListQueryOptions } from "@/features/workouts/api/workout-query-options";
import { WorkoutsPage } from "@/features/workouts/pages/workouts-page";
import { workoutsSearchValidator } from "@/lib/route-search-validation";
import { clearDemoData, initializeDemoData } from "@/lib/demo-data/storage";

export const Route = createFileRoute("/_layout/workouts/")({
  validateSearch: workoutsSearchValidator,
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      clearDemoData();
      void context.queryClient
        .query({ ...getWorkoutListQueryOptions(user), staleTime: Infinity })
        .catch(noop);
      void context.queryClient
        .query({ ...contributionDataQueryOptions(), staleTime: Infinity })
        .catch(noop);
    } else {
      initializeDemoData();
      void context.queryClient
        .query({ ...getWorkoutListQueryOptions(user), staleTime: Infinity })
        .catch(noop);
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  return (
    <WorkoutsPage
      user={user}
      search={search}
    />
  );
}
