import { useMutation } from "@tanstack/react-query";
import type { ApplicationUser } from "@/lib/application-user";
import { queryClient } from "@/lib/api/api";
import {
  getExercisesQueryKey,
  getWorkoutsByIdQueryKey,
  getWorkoutsByIdQueryOptions,
  getWorkoutsNewWorkoutContextQueryKey,
  getWorkoutsNewWorkoutContextQueryOptions,
  getWorkoutsFocusValuesQueryKey,
  getWorkoutsQueryKey,
  getWorkoutsQueryOptions,
  getWorkoutsFocusValuesQueryOptions,
  getWorkoutsContributionDataQueryKey,
  getWorkoutsContributionDataQueryOptions,
  postWorkoutsMutation,
  putWorkoutsByIdMutation,
  deleteWorkoutsByIdMutation,
} from "@/client/@tanstack/react-query.gen";
import type {
  WorkoutUpdateExercise,
  WorkoutUpdateWorkoutRequest,
  WorkoutWorkoutWithSetsResponse,
} from "@/client";
import { sortByExerciseAndSetOrder } from "@/lib/utils";
import {
  deleteDemoWorkoutsByIdMutationWithMeta,
  postDemoWorkoutsMutation,
  putDemoWorkoutsByIdMutation,
} from "@/lib/demo-data/query-options";

export type WorkoutFocus = {
  name: string;
};

type WorkoutMutationUser = ApplicationUser | null;

// MARK: Get all
export function workoutsQueryOptions() {
  return getWorkoutsQueryOptions();
}

// MARK: Get one
export function workoutQueryOptions(id: number) {
  return getWorkoutsByIdQueryOptions({ path: { id } });
}

export function newWorkoutContextQueryOptions() {
  return getWorkoutsNewWorkoutContextQueryOptions();
}

// MARK: Get focus values
export function workoutsFocusValuesQueryOptions() {
  return getWorkoutsFocusValuesQueryOptions();
}

// MARK: Get contribution data
export function contributionDataQueryOptions() {
  return getWorkoutsContributionDataQueryOptions();
}

function invalidateWorkoutAnalyticsQueries() {
  queryClient.invalidateQueries({
    queryKey: getWorkoutsContributionDataQueryKey(),
  });
  queryClient.invalidateQueries({
    queryKey: getWorkoutsFocusValuesQueryKey(),
  });
  queryClient.invalidateQueries({
    queryKey: getWorkoutsNewWorkoutContextQueryKey(),
  });
}

// MARK: Create
// ! TODO: Return data from server to invalidate recent sets for exercise
// ! TODO: if I want to be granular, but stale time is so low not a priority
export function useSaveWorkoutMutation() {
  return useMutation({
    ...postWorkoutsMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getExercisesQueryKey(),
      });
      invalidateWorkoutAnalyticsQueries();
    },
  });
}

export function useSaveWorkoutForUserMutation(user: WorkoutMutationUser) {
  const apiMutation = useSaveWorkoutMutation();
  const demoMutation = useMutation(postDemoWorkoutsMutation());

  return user ? apiMutation : demoMutation;
}

// MARK: Update
export function useUpdateWorkoutMutation() {
  return useMutation({
    ...putWorkoutsByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getWorkoutsByIdQueryKey({ path: { id } }),
      });
      invalidateWorkoutAnalyticsQueries();
    },
  });
}

export function useUpdateWorkoutForUserMutation(user: WorkoutMutationUser) {
  const apiMutation = useUpdateWorkoutMutation();
  const demoMutation = useMutation(putDemoWorkoutsByIdMutation());

  return user ? apiMutation : demoMutation;
}

// MARK: Delete
// Delete mutation without automatic error toasts (for manual error handling)
export function useDeleteWorkoutMutation() {
  return useMutation({
    ...deleteWorkoutsByIdMutation(),
    meta: { skipGlobalErrorHandler: true },
    onSuccess: (_, { path: { id } }) => {
      queryClient.removeQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.removeQueries({
        queryKey: getWorkoutsByIdQueryKey({ path: { id } }),
      });
      invalidateWorkoutAnalyticsQueries();
    },
  });
}

export function useDeleteWorkoutForUserMutation(user: WorkoutMutationUser) {
  const apiMutation = useDeleteWorkoutMutation();
  const demoMutation = useMutation(deleteDemoWorkoutsByIdMutationWithMeta());

  return user ? apiMutation : demoMutation;
}

// MARK: Utils
function groupSetsByExercise(
  sortedWorkouts: WorkoutWorkoutWithSetsResponse[],
): Map<number, { exercise: WorkoutUpdateExercise; order: number }> {
  const exercisesMap = new Map<
    number,
    { exercise: WorkoutUpdateExercise; order: number }
  >();

  for (const workout of sortedWorkouts) {
    const exerciseId = workout.exercise_id || 0;
    const exerciseOrder = workout.exercise_order ?? workout.exercise_id ?? 0;

    let exerciseEntry = exercisesMap.get(exerciseId);
    if (!exerciseEntry) {
      exerciseEntry = {
        exercise: {
          name: workout.exercise_name || "",
          sets: [],
        },
        order: exerciseOrder,
      };
      exercisesMap.set(exerciseId, exerciseEntry);
    }

    const setType = workout.set_type;
    if (setType !== "warmup" && setType !== "working") {
      throw new Error(`Unexpected workout set type: ${setType}`);
    }

    exerciseEntry.exercise.sets.push({
      weight: workout.weight || 0,
      reps: workout.reps || 0,
      setType,
    });
  }

  return exercisesMap;
}

function extractOrderedExercises(
  exercisesMap: Map<number, { exercise: WorkoutUpdateExercise; order: number }>,
): WorkoutUpdateExercise[] {
  return Array.from(exercisesMap.values())
    .sort((a, b) => a.order - b.order)
    .map((entry) => entry.exercise);
}

export function transformToWorkoutFormValues(
  workouts: WorkoutWorkoutWithSetsResponse[],
): WorkoutUpdateWorkoutRequest {
  if (workouts.length === 0) {
    return {
      date: new Date().toISOString(),
      notes: "",
      exercises: [],
    };
  }

  const sortedWorkouts = sortByExerciseAndSetOrder(workouts);
  const exercisesMap = groupSetsByExercise(sortedWorkouts);
  const orderedExercises = extractOrderedExercises(exercisesMap);

  return {
    date: workouts[0].workout_date || new Date().toISOString(),
    notes: workouts[0].workout_notes || "",
    workoutFocus: workouts[0].workout_focus || undefined,
    exercises: orderedExercises,
  };
}
