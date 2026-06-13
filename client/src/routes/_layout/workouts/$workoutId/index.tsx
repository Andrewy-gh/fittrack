import { createFileRoute } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { getWorkoutByIdQueryOptions } from "@/features/workouts/api/workout-query-options";
import { WorkoutDetailEditable } from "@/features/workouts/components/workout-detail";
import { clearDemoData, initializeDemoData } from "@/lib/demo-data/storage";

export const Route = createFileRoute("/_layout/workouts/$workoutId/")({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error("Invalid workoutId");
      }
      return { workoutId };
    },
  },
  loader: ({ context, params }) => {
    const workoutId = params.workoutId;
    const user = context.user;

    if (user) {
      clearDemoData();
      context.queryClient.ensureQueryData(
        getWorkoutByIdQueryOptions(user, workoutId),
      );
    } else {
      initializeDemoData();
      context.queryClient.ensureQueryData(
        getWorkoutByIdQueryOptions(user, workoutId),
      );
    }

    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();

  const { data: workout } = useSuspenseQuery(
    getWorkoutByIdQueryOptions(user, workoutId),
  );

  return <WorkoutDetailEditable workout={workout} />;
}
