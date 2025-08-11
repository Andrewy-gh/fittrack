import { queryClient } from './api';
import { queryOptions , useMutation } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '@/generated';
import type {
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse,
} from '@/generated';


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

export function workoutByIdQueryOptions(
  workoutId: number,
  accessToken: string
) {
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

export function useSaveWorkoutMutation(accessToken: string) {
  return useMutation({
    mutationFn: async (data: workout_CreateWorkoutRequest) => {
            OpenAPI.HEADERS = {
              'x-stack-access-token': accessToken,
            };
      WorkoutsService.postWorkouts(data)
    },
    onSuccess: () => {
       queryClient.invalidateQueries({
         queryKey: ['workouts', 'list'],
       });
    },
  });
};