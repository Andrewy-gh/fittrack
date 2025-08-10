import { queryOptions } from '@tanstack/react-query';
import { WorkoutsService } from '../../generated';
import type { 
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse
} from '../../generated';

// Type aliases for better compatibility with existing code
export type WorkoutFormValues = workout_CreateWorkoutRequest;
export type WorkoutData = workout_WorkoutResponse;
export type WorkoutWithSets = workout_WorkoutWithSetsResponse;

// Delegated functions using generated service
export const getWorkouts = () => WorkoutsService.getWorkouts();
export const createWorkout = (data: WorkoutFormValues) =>
  WorkoutsService.postWorkouts(data);
export const getWorkoutWithSets = (id: number) =>
  WorkoutsService.getWorkouts1(id);

export async function fetchWorkouts(
  accessToken: string
): Promise<WorkoutData[]> {
  const res = await fetch('/api/workouts', {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!res.ok) {
    throw new Error('Failed to fetch workouts');
  }
  return res.json();
}

export function workoutsQueryOptions(accessToken: string) {
  return queryOptions<WorkoutData[], Error>({
    queryKey: ['workouts', 'list'],
    queryFn: () => fetchWorkouts(accessToken),
  });
}

export async function fetchWorkoutById(
  workoutId: number,
  accessToken: string
): Promise<WorkoutWithSets[]> {
  const res = await fetch(`/api/workouts/${workoutId}`, {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch workout: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

export function workoutByIdQueryOptions(workoutId: number, accessToken: string) {
  return queryOptions<WorkoutWithSets[], Error>({
    queryKey: ['workouts', 'details', workoutId],
    queryFn: () => fetchWorkoutById(workoutId, accessToken),
  });
}
