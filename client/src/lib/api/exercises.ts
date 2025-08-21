import { queryOptions } from '@tanstack/react-query';
import { ExercisesService, OpenAPI } from '@/generated';
import { ensureUser, getAccessToken, type User } from './auth';
import type {
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
  exercise_RecentSetsResponse,
} from '@/generated';

export type ExerciseOption = Pick<exercise_ExerciseResponse, 'id' | 'name'>;

export function exercisesQueryOptions(user: User) {
  const validatedUser = ensureUser(user);
  return queryOptions<exercise_ExerciseResponse[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises();
    },
  });
}

export function exerciseQueryOptions(
  exerciseId: number,
  user: User
) {
  const validatedUser = ensureUser(user);
  return queryOptions<exercise_ExerciseWithSetsResponse[], Error>({
    queryKey: ['exercises', 'details', exerciseId],
    queryFn: async () => {
      // Get a fresh token right before making the API call
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises1(exerciseId);
    },
  });
}

export function recentExerciseSetsQueryOptions(
  exerciseId: number,
  user: User
) {
  const validatedUser = ensureUser(user);
  return queryOptions<exercise_RecentSetsResponse[], Error>({
    queryKey: ['exercises', 'recent-sets', exerciseId],
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercisesRecentSets(exerciseId);
    },
  });
}
