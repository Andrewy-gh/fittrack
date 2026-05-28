import { createFileRoute } from "@tanstack/react-router";
import { useSuspenseQuery } from "@tanstack/react-query";
import { z } from "zod";
import type { WorkoutFocus } from "@/features/workouts/api/workouts";
import { NewWorkoutPage } from "@/features/workouts/pages/new-workout-page";
import type { DbExercise } from "@/lib/api/exercises";
import {
  getExercisesQueryOptions,
  getWorkoutsFocusQueryOptions,
  getWorkoutsQueryOptions,
} from "@/lib/api/unified-query-options";
import { initializeDemoData } from "@/lib/demo-data/storage";

const workoutSearchSchema = z.object({
  addExercise: z.boolean().optional(),
  exerciseIndex: z.coerce.number().int().optional(),
  newExercise: z.boolean().optional(),
});

export const Route = createFileRoute("/_layout/workouts/new")({
  validateSearch: workoutSearchSchema,
  loader: ({ context }) => {
    if (!context.user) initializeDemoData();
    context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
    context.queryClient.ensureQueryData(getWorkoutsQueryOptions(context.user));
    context.queryClient.ensureQueryData(
      getWorkoutsFocusQueryOptions(context.user),
    );
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  const { data: exercisesResponse } = useSuspenseQuery(
    getExercisesQueryOptions(user),
  );
  const { data: workouts } = useSuspenseQuery(getWorkoutsQueryOptions(user));
  const { data: workoutsFocusValues } = useSuspenseQuery(
    getWorkoutsFocusQueryOptions(user),
  );

  const exercises: DbExercise[] = exercisesResponse.map((exercise) => ({
    id: exercise.id,
    name: exercise.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((name) => ({
    name,
  }));

  return (
    <NewWorkoutPage
      user={user}
      exercises={exercises}
      workouts={workouts}
      workoutsFocus={workoutsFocus}
      search={search}
    />
  );
}
