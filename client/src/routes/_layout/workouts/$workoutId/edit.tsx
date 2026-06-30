import { createFileRoute } from "@tanstack/react-router";
import {
  EditWorkoutRouteComposition,
  preloadEditWorkoutRouteData,
} from "@/features/workouts/pages/workout-route-composition";
import { workoutEditorSearchValidator } from "@/lib/route-search-validation";

export const Route = createFileRoute("/_layout/workouts/$workoutId/edit")({
  validateSearch: workoutEditorSearchValidator,
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error("Invalid workoutId");
      }
      return { workoutId };
    },
  },
  loader: async ({ context, params }): Promise<{ workoutId: number }> => {
    return preloadEditWorkoutRouteData({
      ...context,
      workoutId: params.workoutId,
    });
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  return (
    <EditWorkoutRouteComposition
      user={user}
      workoutId={workoutId}
      search={search}
    />
  );
}
