import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import {
  getExercisesQueryKey,
  getWorkoutsByIdQueryKey,
  getWorkoutsByIdQueryOptions,
  getWorkoutsQueryKey,
  getWorkoutsQueryOptions,
  getWorkoutsFocusValuesQueryOptions,
  postWorkoutsMutation,
  putWorkoutsByIdMutation,
  deleteWorkoutsByIdMutation,
} from '@/client/@tanstack/react-query.gen';
import type {
  WorkoutUpdateExercise,
  WorkoutUpdateSet,
  WorkoutUpdateWorkoutRequest,
  WorkoutWorkoutWithSetsResponse,
} from '@/client';
import { sortByExerciseAndSetOrder } from '../utils';

export type WorkoutFocus = {
  name: string;
};

// MARK: Get all
export function workoutsQueryOptions() {
  return getWorkoutsQueryOptions();
}

// MARK: Get one
export function workoutQueryOptions(id: number) {
  return getWorkoutsByIdQueryOptions({ path: { id } });
}

// MARK: Get focus values
export function workoutsFocusValuesQueryOptions() {
  return getWorkoutsFocusValuesQueryOptions();
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
    },
  });
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
    },
  });
}

// MARK: Delete
export function useDeleteWorkoutMutation() {
  return useMutation({
    ...deleteWorkoutsByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.removeQueries({
        queryKey: getWorkoutsByIdQueryKey({ path: { id } }),
      });
    },
  });
}

// MARK: Utils
function groupSetsByExercise(
  sortedWorkouts: WorkoutWorkoutWithSetsResponse[]
): Map<number, { exercise: WorkoutUpdateExercise; order: number }> {
  const exercisesMap = new Map<
    number,
    { exercise: WorkoutUpdateExercise; order: number }
  >();

  for (const workout of sortedWorkouts) {
    const exerciseId = workout.exercise_id || 0;
    const exerciseOrder = workout.exercise_order ?? workout.exercise_id ?? 0;

    if (!exercisesMap.has(exerciseId)) {
      exercisesMap.set(exerciseId, {
        exercise: {
          name: workout.exercise_name || '',
          sets: [],
        },
        order: exerciseOrder,
      });
    }

    const exerciseEntry = exercisesMap.get(exerciseId)!;
    exerciseEntry.exercise.sets.push({
      weight: workout.weight || 0,
      reps: workout.reps || 0,
      setType: workout.set_type as WorkoutUpdateSet['setType'],
    });
  }

  return exercisesMap;
}

function extractOrderedExercises(
  exercisesMap: Map<number, { exercise: WorkoutUpdateExercise; order: number }>
): WorkoutUpdateExercise[] {
  return Array.from(exercisesMap.values())
    .sort((a, b) => a.order - b.order)
    .map((entry) => entry.exercise);
}

export function transformToWorkoutFormValues(
  workouts: WorkoutWorkoutWithSetsResponse[]
): WorkoutUpdateWorkoutRequest {
  if (workouts.length === 0) {
    return {
      date: new Date().toISOString(),
      notes: '',
      exercises: [],
    };
  }

  const sortedWorkouts = sortByExerciseAndSetOrder(workouts);
  const exercisesMap = groupSetsByExercise(sortedWorkouts);
  const orderedExercises = extractOrderedExercises(exercisesMap);

  return {
    date: workouts[0].workout_date || new Date().toISOString(),
    notes: workouts[0].workout_notes || '',
    exercises: orderedExercises,
  };
}
