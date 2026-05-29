import { useSuspenseQuery, type QueryClient } from "@tanstack/react-query";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import type { WorkoutUpdateWorkoutRequest } from "@/client";
import type { DbExercise } from "@/features/exercises/api/exercises";
import {
  transformToWorkoutFormValues,
  type WorkoutFocus,
} from "@/features/workouts/api/workouts";
import {
  getExercisesQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsFocusQueryOptions,
  getWorkoutsQueryOptions,
} from "@/lib/api/unified-query-options";
import { initializeDemoData } from "@/lib/demo-data/storage";
import { EditWorkoutPage } from "@/features/workouts/pages/edit-workout-page";
import {
  NewWorkoutPage,
  type WorkoutFormSearch,
} from "@/features/workouts/pages/new-workout-page";

type WorkoutRouteUser = CurrentUser | CurrentInternalUser | null;

type WorkoutRouteLoaderContext = {
  queryClient: QueryClient;
  user: WorkoutRouteUser;
};

export function preloadNewWorkoutRouteData({
  queryClient,
  user,
}: WorkoutRouteLoaderContext) {
  if (!user) initializeDemoData();

  queryClient.ensureQueryData(getExercisesQueryOptions(user));
  queryClient.ensureQueryData(getWorkoutsQueryOptions(user));
  queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(user));
}

export function preloadEditWorkoutRouteData({
  queryClient,
  user,
  workoutId,
}: WorkoutRouteLoaderContext & { workoutId: number }) {
  if (!user) initializeDemoData();

  queryClient.ensureQueryData(getWorkoutByIdQueryOptions(user, workoutId));
  queryClient.ensureQueryData(getExercisesQueryOptions(user));
  queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(user));

  return { workoutId };
}

function toWorkoutFocus(workoutsFocusValues: string[]): WorkoutFocus[] {
  return workoutsFocusValues.map((name) => ({ name }));
}

export function NewWorkoutRouteComposition({
  user,
  search,
}: {
  user: WorkoutRouteUser;
  search: WorkoutFormSearch;
}) {
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

  return (
    <NewWorkoutPage
      user={user}
      exercises={exercises}
      workouts={workouts}
      workoutsFocus={toWorkoutFocus(workoutsFocusValues)}
      search={search}
    />
  );
}

export function EditWorkoutRouteComposition({
  user,
  workoutId,
  search,
}: {
  user: WorkoutRouteUser;
  workoutId: number;
  search: WorkoutFormSearch;
}) {
  const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions(user));
  const { data: workout } = useSuspenseQuery(
    getWorkoutByIdQueryOptions(user, workoutId),
  );
  const { data: workoutsFocusValues } = useSuspenseQuery(
    getWorkoutsFocusQueryOptions(user),
  );

  const workoutFormValues: WorkoutUpdateWorkoutRequest =
    transformToWorkoutFormValues(workout);

  return (
    <EditWorkoutPage
      user={user}
      exercises={exercises}
      workout={workoutFormValues}
      workoutId={workoutId}
      workoutsFocus={toWorkoutFocus(workoutsFocusValues)}
      search={search}
    />
  );
}
