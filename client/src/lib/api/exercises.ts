import { queryOptions } from '@tanstack/react-query';
import { ExercisesService, OpenAPI } from '@/generated';
import { ensureUser, getAccessToken, type User } from './auth';
import type {
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
  exercise_RecentSetsResponse,
} from '@/generated';

import { getExercisesQueryOptions } from '@/client/@tanstack/react-query.gen';

/**
 * Exercise data as returned by the API - always has a database ID
 */
export type DbExercise = Pick<exercise_ExerciseResponse, 'id' | 'name'>;

/**
 * Exercise option for form components - may include manually created exercises without IDs
 * Used in dropdowns and forms where users can create new exercises on-the-fly
 */
export type ExerciseOption = {
  id: number | null; // null for manually created exercises, number for DB exercises
  name: string;
};

export function exercisesQueryOptions(user: User) {
  ensureUser(user);
  console.log('user ensured');
  // Get the base options from generated code
  const baseOptions = getExercisesQueryOptions();
  // The interceptor will automatically add the auth token
  return baseOptions;
}

// export function exercisesQueryOptions(user: User) {
//   const validatedUser = ensureUser(user);
//   return queryOptions<exercise_ExerciseResponse[], Error>({
//     queryKey: ['exercises', 'list'],
//     queryFn: async () => {
//       const accessToken = await getAccessToken(validatedUser);
//       OpenAPI.HEADERS = {
//         'x-stack-access-token': accessToken,
//       };
//       return ExercisesService.getExercises();
//     },
//   });
// }

export function exerciseQueryOptions(exerciseId: number, user: User) {
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

export function recentExerciseSetsQueryOptions(exerciseId: number, user: User) {
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
