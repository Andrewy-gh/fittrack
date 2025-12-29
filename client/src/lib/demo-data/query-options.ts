import { queryOptions, type UseMutationOptions } from '@tanstack/react-query';
import { queryClient } from '../api/api';
import type {
  WorkoutWorkoutResponse,
  WorkoutWorkoutWithSetsResponse,
  ExerciseExerciseResponse,
  ExerciseExerciseWithSetsResponse,
  ExerciseRecentSetsResponse,
  WorkoutCreateWorkoutRequest,
  WorkoutUpdateWorkoutRequest,
  ResponseSuccessResponse,
} from './types';
import {
  getAllExercises,
  createExercise,
  updateExercise,
  deleteExercise,
  getExerciseWithSets,
  getExerciseRecentSets,
  getAllWorkouts,
  getWorkoutById,
  createWorkout,
  updateWorkout,
  deleteWorkout,
  getWorkoutFocusValues,
  initializeDemoData,
} from './storage';

// Initialize demo data when module loads
if (typeof window !== 'undefined') {
  initializeDemoData();
}

// ===========================
// Query Keys
// ===========================

export const getDemoExercisesQueryKey = () => [{ _id: 'demo_getExercises' }] as const;

export const getDemoExercisesByIdQueryKey = (id: number) =>
  [{ _id: 'demo_getExercisesById', path: { id } }] as const;

export const getDemoExercisesByIdRecentSetsQueryKey = (id: number) =>
  [{ _id: 'demo_getExercisesByIdRecentSets', path: { id } }] as const;

export const getDemoWorkoutsQueryKey = () => [{ _id: 'demo_getWorkouts' }] as const;

export const getDemoWorkoutsByIdQueryKey = (id: number) =>
  [{ _id: 'demo_getWorkoutsById', path: { id } }] as const;

export const getDemoWorkoutsFocusValuesQueryKey = () =>
  [{ _id: 'demo_getWorkoutsFocusValues' }] as const;

// ===========================
// Exercise Query Options
// ===========================

export const getDemoExercisesQueryOptions = () => {
  return queryOptions({
    queryKey: getDemoExercisesQueryKey(),
    queryFn: async (): Promise<ExerciseExerciseResponse[]> => {
      // Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 100));
      return getAllExercises();
    },
  });
};

export const getDemoExercisesByIdQueryOptions = (id: number) => {
  return queryOptions({
    queryKey: getDemoExercisesByIdQueryKey(id),
    queryFn: async (): Promise<ExerciseExerciseWithSetsResponse[]> => {
      await new Promise((resolve) => setTimeout(resolve, 100));
      return getExerciseWithSets(id);
    },
  });
};

export const getDemoExercisesByIdRecentSetsQueryOptions = (id: number) => {
  return queryOptions({
    queryKey: getDemoExercisesByIdRecentSetsQueryKey(id),
    queryFn: async (): Promise<ExerciseRecentSetsResponse[]> => {
      await new Promise((resolve) => setTimeout(resolve, 100));
      return getExerciseRecentSets(id);
    },
  });
};

// ===========================
// Workout Query Options
// ===========================

export const getDemoWorkoutsQueryOptions = () => {
  return queryOptions({
    queryKey: getDemoWorkoutsQueryKey(),
    queryFn: async (): Promise<WorkoutWorkoutResponse[]> => {
      await new Promise((resolve) => setTimeout(resolve, 100));
      return getAllWorkouts();
    },
  });
};

export const getDemoWorkoutsByIdQueryOptions = (id: number) => {
  return queryOptions({
    queryKey: getDemoWorkoutsByIdQueryKey(id),
    queryFn: async (): Promise<WorkoutWorkoutWithSetsResponse[]> => {
      await new Promise((resolve) => setTimeout(resolve, 100));
      return getWorkoutById(id);
    },
  });
};

export const getDemoWorkoutsFocusValuesQueryOptions = () => {
  return queryOptions({
    queryKey: getDemoWorkoutsFocusValuesQueryKey(),
    queryFn: async (): Promise<string[]> => {
      await new Promise((resolve) => setTimeout(resolve, 50));
      return getWorkoutFocusValues();
    },
  });
};

// ===========================
// Exercise Mutations
// ===========================

export const postDemoExercisesMutation = (): UseMutationOptions<
  ExerciseExerciseResponse,
  Error,
  { body: { name: string } }
> => ({
  mutationFn: async ({ body }) => {
    await new Promise((resolve) => setTimeout(resolve, 100));
    return createExercise(body.name);
  },
});

export const patchDemoExercisesByIdMutation = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number }; body: { name: string } }
> => ({
  mutationFn: async ({ path: { id }, body }) => {
    await new Promise((resolve) => setTimeout(resolve, 100));
    const success = updateExercise(id, body.name);
    if (!success) {
      throw new Error('Exercise not found');
    }
  },
  onSuccess: (_, { path: { id } }) => {
    queryClient.invalidateQueries({
      queryKey: getDemoExercisesQueryKey(),
    });
    queryClient.invalidateQueries({
      queryKey: getDemoExercisesByIdQueryKey(id),
    });
  },
});

export const deleteDemoExercisesByIdMutation = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number } }
> => ({
  mutationFn: async ({ path: { id } }) => {
    await new Promise((resolve) => setTimeout(resolve, 100));
    const success = deleteExercise(id);
    if (!success) {
      throw new Error('Exercise not found');
    }
  },
  onSuccess: (_, { path: { id } }) => {
    queryClient.invalidateQueries({
      queryKey: getDemoExercisesQueryKey(),
    });
    queryClient.removeQueries({
      queryKey: getDemoExercisesByIdQueryKey(id),
    });
  },
});

// Delete mutation without automatic error toasts (for manual error handling)
export const deleteDemoExercisesByIdMutationWithMeta = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number } }
> => ({
  ...deleteDemoExercisesByIdMutation(),
  meta: { skipGlobalErrorHandler: true },
});

// ===========================
// Workout Mutations
// ===========================

export const postDemoWorkoutsMutation = (): UseMutationOptions<
  ResponseSuccessResponse,
  Error,
  { body: WorkoutCreateWorkoutRequest }
> => ({
  mutationFn: async ({ body }) => {
    await new Promise((resolve) => setTimeout(resolve, 200));
    return createWorkout(body);
  },
  onSuccess: () => {
    queryClient.invalidateQueries({
      queryKey: getDemoWorkoutsQueryKey(),
    });
    queryClient.invalidateQueries({
      queryKey: getDemoExercisesQueryKey(),
    });
  },
});

export const putDemoWorkoutsByIdMutation = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number }; body: WorkoutUpdateWorkoutRequest }
> => ({
  mutationFn: async ({ path: { id }, body }) => {
    await new Promise((resolve) => setTimeout(resolve, 200));
    const result = updateWorkout(id, body);
    if (!result.success) {
      throw new Error('Workout not found');
    }
  },
  onSuccess: (_, { path: { id } }) => {
    queryClient.invalidateQueries({
      queryKey: getDemoWorkoutsQueryKey(),
    });
    queryClient.invalidateQueries({
      queryKey: getDemoWorkoutsByIdQueryKey(id),
    });
  },
});

export const deleteDemoWorkoutsByIdMutation = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number } }
> => ({
  mutationFn: async ({ path: { id } }) => {
    await new Promise((resolve) => setTimeout(resolve, 100));
    const success = deleteWorkout(id);
    if (!success) {
      throw new Error('Workout not found');
    }
  },
  onSuccess: (_, { path: { id } }) => {
    queryClient.invalidateQueries({
      queryKey: getDemoWorkoutsQueryKey(),
    });
    queryClient.removeQueries({
      queryKey: getDemoWorkoutsByIdQueryKey(id),
    });
  },
});

// Delete mutation without automatic error toasts (for manual error handling)
export const deleteDemoWorkoutsByIdMutationWithMeta = (): UseMutationOptions<
  void,
  Error,
  { path: { id: number } }
> => ({
  ...deleteDemoWorkoutsByIdMutation(),
  meta: { skipGlobalErrorHandler: true },
});
