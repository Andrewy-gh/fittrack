import { createFileRoute } from "@tanstack/react-router";
import {
  NewWorkoutRouteComposition,
  preloadNewWorkoutRouteData,
} from "@/features/workouts/pages/workout-route-composition";
import { workoutEditorSearchValidator } from "@/lib/route-search-validation";

export const Route = createFileRoute("/_layout/workouts/new")({
  validateSearch: workoutEditorSearchValidator,
  loader: ({ context }) => {
    preloadNewWorkoutRouteData(context);
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  return (
    <NewWorkoutRouteComposition
      user={user}
      search={search}
    />
  );
}
