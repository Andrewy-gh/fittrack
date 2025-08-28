```typescript
// exercises-queries.ts
import { 
  getExercisesOptions,
  getExercises1Options, // or whatever the generated name is
  getExercisesRecentSetsOptions 
} from '@/generated/@tanstack/react-query';
import { ensureUser, getAccessToken, type User } from './auth';
import { OpenAPI } from '@/generated';

// Wrapper functions that maintain your current API but use generated options
export function exercisesQueryOptions(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base options from generated code
  const baseOptions = getExercisesOptions();
  
  // Override the queryFn to add your auth logic
  return {
    ...baseOptions,
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return baseOptions.queryFn();
    },
  };
}

export function exerciseQueryOptions(exerciseId: number, user: User) {
  const validatedUser = ensureUser(user);
  
  const baseOptions = getExercises1Options({ 
    path: { exerciseId } // Hey API uses structured parameters
  });
  
  return {
    ...baseOptions,
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return baseOptions.queryFn();
    },
  };
}

export function recentExerciseSetsQueryOptions(exerciseId: number, user: User) {
  const validatedUser = ensureUser(user);
  
  const baseOptions = getExercisesRecentSetsOptions({ 
    path: { exerciseId }
  });
  
  return {
    ...baseOptions,
    queryFn: async () => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return baseOptions.queryFn();
    },
  };
}
```