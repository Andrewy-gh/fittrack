import { createFileRoute } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { z } from "zod";
import {
  transformToWorkoutFormValues,
  type WorkoutFocus,
} from "@/features/workouts/api/workouts";
import { EditWorkoutPage } from "@/features/workouts/pages/edit-workout-page";
import type { WorkoutUpdateWorkoutRequest } from "@/client";
import {
  getExercisesQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/lib/api/unified-query-options";
import { initializeDemoData } from "@/lib/demo-data/storage";

const workoutSearchSchema = z.object({
  addExercise: z.boolean().optional(),
  exerciseIndex: z.coerce.number().int().optional(),
  newExercise: z.boolean().optional(),
});

export const Route = createFileRoute("/_layout/workouts/$workoutId/edit")({
  validateSearch: workoutSearchSchema,
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
    const workoutId = params.workoutId;
    if (!context.user) initializeDemoData();

    context.queryClient.ensureQueryData(
      getWorkoutByIdQueryOptions(context.user, workoutId),
    );
    context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
    context.queryClient.ensureQueryData(
      getWorkoutsFocusQueryOptions(context.user),
    );

    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions(user));
  const { data: workout } = useSuspenseQuery(
    getWorkoutByIdQueryOptions(user, workoutId),
  );
  const { data: workoutsFocusValues } = useSuspenseQuery(
    getWorkoutsFocusQueryOptions(user),
  );

  const workoutFormValues: WorkoutUpdateWorkoutRequest =
    transformToWorkoutFormValues(workout);

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((name) => ({
    name,
  }));

  return (
    <EditWorkoutPage
      user={user}
      exercises={exercises}
      workout={workoutFormValues}
      workoutId={workoutId}
      workoutsFocus={workoutsFocus}
      search={search}
    />
  );
}
