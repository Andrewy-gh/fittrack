import { queryOptions } from '@tanstack/react-query';
import { ExercisesService, OpenAPI } from '@/generated';
import { getAccessToken, type User } from './auth';
import type {
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
} from '@/generated';

export type ExerciseOption = Pick<exercise_ExerciseResponse, 'id' | 'name'>;

// Legacy function for backward compatibility
export function exercisesQueryOptions(accessToken: string) {
  return queryOptions<exercise_ExerciseResponse[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises();
    },
  });
}

// New function that gets fresh tokens
export function exercisesQueryOptionsWithUser(user: User) {
  return queryOptions<exercise_ExerciseResponse[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: async () => {
      // Get a fresh token right before making the API call
      const accessToken = await getAccessToken(user);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises();
    },
  });
}

export function exerciseWithSetsQueryOptions(
  exerciseId: number,
  user: User
) {
  return queryOptions<exercise_ExerciseWithSetsResponse[], Error>({
    queryKey: ['exercises', 'details', exerciseId],
    queryFn: async () => {
      // Get a fresh token right before making the API call
      const accessToken = await getAccessToken(user);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises1(exerciseId);
    },
  });
}
