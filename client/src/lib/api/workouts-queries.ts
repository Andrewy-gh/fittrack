import { 
  getWorkoutsQueryOptions,
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey,
  getWorkoutsByIdQueryOptions
} from '@/client/@tanstack/react-query.gen';
import { ensureUser, type User } from './auth';
import type {
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse,
} from '@/client/types.gen';

// Export generated query keys for easier cache invalidation
export { 
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey
};

// Wrapper functions that maintain your current API but use generated options
export function workoutsQueryOptions(user: User) {
  ensureUser(user);
  
  // Get the base options from generated code
  const baseOptions = getWorkoutsQueryOptions();
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}

export function workoutQueryOptions(workoutId: number, user: User) {
  ensureUser(user);
  
  const baseOptions = getWorkoutsByIdQueryOptions({ 
    path: { id: workoutId }
  });
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}