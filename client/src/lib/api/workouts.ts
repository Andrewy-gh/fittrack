import { queryOptions } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '../../generated';
import type { 
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse
} from '../../generated';

// Re-export generated types directly
export type { 
  workout_CreateWorkoutRequest as WorkoutFormValues,
  workout_WorkoutResponse as WorkoutData,
  workout_WorkoutWithSetsResponse as WorkoutWithSets
} from '../../generated';

export const getWorkouts = () => WorkoutsService.getWorkouts();
export const createWorkout = (data: workout_CreateWorkoutRequest) =>
  WorkoutsService.postWorkouts(data);
export const getWorkoutWithSets = (id: number) =>
  WorkoutsService.getWorkouts1(id);

export function workoutsQueryOptions(accessToken: string) {
  return queryOptions<workout_WorkoutResponse[], Error>({
    queryKey: ['workouts', 'list'],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts();
    },
  });
}

export function workoutByIdQueryOptions(workoutId: number, accessToken: string) {
  return queryOptions<workout_WorkoutWithSetsResponse[], Error>({
    queryKey: ['workouts', 'details', workoutId],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts1(workoutId);
    },
  });
}
