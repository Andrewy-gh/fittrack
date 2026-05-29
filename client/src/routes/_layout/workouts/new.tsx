import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import {
  NewWorkoutRouteComposition,
  preloadNewWorkoutRouteData,
} from "@/features/workouts/pages/workout-route-composition";

const workoutSearchSchema = z.object({
  addExercise: z.boolean().optional(),
  exerciseIndex: z.coerce.number().int().optional(),
  newExercise: z.boolean().optional(),
});

export const Route = createFileRoute("/_layout/workouts/new")({
  validateSearch: workoutSearchSchema,
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
