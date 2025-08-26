import { 
  postWorkoutsMutation,
  putWorkoutsMutation,
  deleteWorkoutsByIdMutation,
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey
} from '@/client/@tanstack/react-query.gen';
import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import { WorkoutsService } from '@/client/sdk.gen';
import { ensureUser, type User } from './auth';
import type {
  workout_CreateWorkoutRequest,
  workout_UpdateWorkoutRequest,
} from '@/client/types.gen';

// Export generated query keys for easier cache invalidation
export { 
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey
};

// Wrapper for save workout mutation
export function useSaveWorkoutMutation(user: User) {
  ensureUser(user);
  
  // Get the base mutation options from generated code
  const baseOptions = postWorkoutsMutation();
  
  return useMutation({
    ...baseOptions,
    onSuccess: () => {
      // Use generated query key for consistency
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
    },
  });
}

// Wrapper for update workout mutation
export function useUpdateWorkoutMutation(user: User) {
  ensureUser(user);
  
  // Get the base mutation options from generated code
  const baseOptions = putWorkoutsMutation();
  
  return useMutation({
    ...baseOptions,
    onSuccess: (_, { id }) => {
      // Use generated query keys for consistency
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getWorkoutsByIdQueryKey({ path: { id } }),
      });
    },
  });
}

// Wrapper for delete workout
export async function deleteWorkout(
  workoutId: number,
  user: User
): Promise<void> {
  ensureUser(user);
  // The interceptor will automatically add the auth token
  return WorkoutsService.deleteWorkoutsById({ path: { id: workoutId } });
}