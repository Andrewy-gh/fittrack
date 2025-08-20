import { queryClient } from './api';
import { queryOptions, useMutation } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '@/generated';
import { ensureUser, getAccessToken, type User } from './auth';
import type {
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse,
  workout_UpdateWorkoutRequest,
  workout_UpdateExercise,
  workout_UpdateSet,
} from '@/generated';

export function workoutsQueryOptions(user: User) {
  const validatedUser = ensureUser(user);
  return queryOptions<workout_WorkoutResponse[], Error>({
    queryKey: ['workouts', 'list'],
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts();
    },
  });
}

export function workoutQueryOptions(
  workoutId: number,
  user: User
) {
  const validatedUser = ensureUser(user);
  return queryOptions<workout_WorkoutWithSetsResponse[], Error>({
    queryKey: ['workouts', 'details', workoutId],
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts1(workoutId);
    },
  });
}

export function useSaveWorkoutMutation(user: User) {
  const validatedUser = ensureUser(user);
  return useMutation({
    mutationFn: async (data: workout_CreateWorkoutRequest) => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return await WorkoutsService.postWorkouts(data); // await for form.Subscribe to update
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'list'],
      });
    },
  });
}

export function useUpdateWorkoutMutation(user: User) {
  const validatedUser = ensureUser(user);
  return useMutation({
    mutationFn: async ({
      id,
      data,
    }: {
      id: number;
      data: workout_UpdateWorkoutRequest;
    }) => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return await WorkoutsService.putWorkouts(id, data);
    },
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'list'],
      });
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'details', id],
      });
    },
  });
}

export async function deleteWorkout(
  workoutId: number,
  user: User
): Promise<void> {
  const validatedUser = ensureUser(user);
  const accessToken = await getAccessToken(validatedUser);
  OpenAPI.HEADERS = {
    'x-stack-access-token': accessToken,
  };
  
  return WorkoutsService.deleteWorkouts(workoutId);
}

export function transformToWorkoutFormValues(
  workouts: workout_WorkoutWithSetsResponse[]
): workout_UpdateWorkoutRequest {
  if (workouts.length === 0) {
    return {
      date: new Date().toISOString(),
      notes: '',
      exercises: [],
    };
  }

  // Group sets by exercise
  const exercisesMap = new Map<number, workout_UpdateExercise>();

  // Sort all workouts by set_id first to ensure consistent ordering
  const sortedWorkouts = [...workouts].sort(
    (a, b) => (a.set_id || 0) - (b.set_id || 0)
  );

  for (const workout of sortedWorkouts) {
    const exerciseId = workout.exercise_id || 0;
    if (!exercisesMap.has(exerciseId)) {
      exercisesMap.set(exerciseId, {
        name: workout.exercise_name || '',
        sets: [],
      });
    }

    const exercise = exercisesMap.get(exerciseId)!;
    exercise.sets.push({
      weight: workout.weight || 0,
      reps: workout.reps || 0,
      setType: workout.set_type as workout_UpdateSet.setType,
    });
  }

  return {
    date: workouts[0].workout_date || new Date().toISOString(),
    notes: workouts[0].workout_notes || '',
    exercises: Array.from(exercisesMap.values()),
  };
}


