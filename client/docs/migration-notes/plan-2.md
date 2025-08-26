# Migration Plan v2: Minimal Component Changes

This plan focuses on migrating to Hey API's TanStack Query plugin with zero breaking changes to your existing components.

## 1. Configuration Overview

Your current configuration in `openapi-ts.config.ts` is already well-suited for minimal changes:

```ts
{
  name: '@tanstack/react-query',
  queryOptions: {
    name: '{{name}}Options', // Matches your current naming pattern
  },
  queryKeys: {
    enabled: true,
    name: '{{name}}QueryKey',
    tags: true, // You already have tags enabled!
  },
  mutationOptions: {
    enabled: true,
    name: '{{name}}Mutation', // Standard naming for mutations
  },
  // ... other options
}
```

This will generate functions like:
- `getWorkoutsOptions()` (for queries)
- `postWorkoutsMutation()` (for mutations)

### 1.1 Tags Benefit

Since you already have `tags: true` enabled in your configuration, your generated query keys will include operation tags:

```ts
const queryKey = [{ _id: 'getWorkouts', tags: ['workouts', 'get'] }]
```

This provides better cache invalidation capabilities. You can take advantage of this in the future by invalidating queries by tags:

```ts
// Invalidate all queries with 'workouts' tag
queryClient.invalidateQueries({ 
  predicate: (query) => query.queryKey[0]?.tags?.includes('workouts') 
});
```

## 2. Migration Strategy by Function Type

### 2.1 Queries - Wrapper Function Approach

Create wrapper functions that maintain your current API but use the generated Hey API functions internally.

**File: `src/lib/api/workouts-queries.ts`**

```typescript
import { 
  getWorkoutsOptions,
  getWorkouts1Options
} from '@/generated/@tanstack/react-query';
import { ensureUser, getAccessToken, type User } from './auth';
import { OpenAPI } from '@/generated';
import type {
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse,
} from '@/generated';

// Wrapper functions that maintain your current API but use generated options
export function workoutsQueryOptions(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base options from generated code
  const baseOptions = getWorkoutsOptions();
  
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

export function workoutQueryOptions(workoutId: number, user: User) {
  const validatedUser = ensureUser(user);
  
  const baseOptions = getWorkouts1Options({ 
    path: { id: workoutId } // Adjust parameter name as per generated function
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

### 2.2 Mutations - Wrapper Function Approach

Similarly for mutations, create wrapper functions that preserve your current API:

**File: `src/lib/api/workouts-mutations.ts`**

```typescript
import { 
  postWorkoutsMutation,
  putWorkoutsMutation,
  deleteWorkoutsMutation
} from '@/generated/@tanstack/react-query';
import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '@/generated';
import { ensureUser, getAccessToken, type User } from './auth';
import type {
  workout_CreateWorkoutRequest,
  workout_UpdateWorkoutRequest,
} from '@/generated';

// Wrapper for save workout mutation
export function useSaveWorkoutMutation(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base mutation options from generated code
  const baseOptions = postWorkoutsMutation();
  
  return useMutation({
    ...baseOptions,
    mutationFn: async (data: workout_CreateWorkoutRequest) => {
      const accessToken = await getAccessToken(validatedUser);
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return await WorkoutsService.postWorkouts(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['workouts', 'list'],
      });
    },
  });
}

// Wrapper for update workout mutation
export function useUpdateWorkoutMutation(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base mutation options from generated code
  const baseOptions = putWorkoutsMutation();
  
  return useMutation({
    ...baseOptions,
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

// Wrapper for delete workout
export async function deleteWorkout(
  workoutId: number,
  user: User
): Promise<void> {
  const validatedUser = ensureUser(user);
  const accessToken = await getAccessToken(validatedUser);
  OpenAPI.HEADERS = {
    'x-stack-access-token': accessToken,
  };

  // You could also use the generated deleteWorkoutsMutation() here if needed
  return WorkoutsService.deleteWorkouts(workoutId);
}
```

### 2.3 Handling Query and Validation on Success

When using the generated mutation functions, you have two approaches for handling query invalidation and validation on success:

1. **Spread Approach** (used in the examples above):
   ```typescript
   return useMutation({
     ...postWorkoutsMutation(), // Spread the generated options
     onSuccess: () => {
       // Your custom success logic
       queryClient.invalidateQueries({ queryKey: ['workouts', 'list'] });
     },
   });
   ```

2. **Direct SDK Call Approach**:
   ```typescript
   return useMutation({
     mutationFn: async (data: workout_CreateWorkoutRequest) => {
       // Your auth setup
       const accessToken = await getAccessToken(validatedUser);
       OpenAPI.HEADERS = {
         'x-stack-access-token': accessToken,
       };
       // Direct call to the generated SDK
       return await WorkoutsService.postWorkouts(data);
     },
     onSuccess: () => {
       queryClient.invalidateQueries({ queryKey: ['workouts', 'list'] });
     },
   });
   ```

The first approach is recommended as it leverages the generated mutation options while still allowing you to customize the `onSuccess` handler.

## 3. Implementation Steps

### Step 1: Verify Configuration
Your configuration already has tags enabled, which is great for future cache invalidation capabilities.

### Step 2: Set Up Client Interceptor for Authentication
Based on your interceptor plan, you should set up a client interceptor to handle the `x-stack-access-token` header. Since you're currently using `openapi-typescript-codegen`, you'll need to modify your approach slightly:

**File: `src/lib/api/client-config.ts`**
```typescript
import { stackClientApp } from '@/stack';
import { OpenAPI } from '@/generated';

// Helper function to get current access token
export const getAccessToken = async (): Promise<string | null> => {
  try {
    const user = await stackClientApp.getUser();
    if (!user) return null;
    
    const { accessToken } = await user.getAuthJson();
    return accessToken || null;
  } catch (error) {
    console.warn('Failed to get access token:', error);
    return null;
  }
};

// Configure the service client
OpenAPI.BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

// You can set default headers here, but we'll set the auth header dynamically in each request
// For now, we'll continue with your current approach of setting headers per request
```

Update your `main.tsx` to import this configuration:

**File: `src/main.tsx`**
```typescript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './lib/api/client-config'; // Import to initialize client configuration

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60000,
      retry: (failureCount, error) => {
        // Don't retry on auth errors
        if (error && typeof error === 'object' && 'status' in error && error.status === 401) {
          return false;
        }
        return failureCount < 3;
      },
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  </React.StrictMode>
);
```

### Step 3: Generate the Code
Run your OpenAPI generation command:
```bash
bun run openapi:generate  # or whatever command you use
```

Note: After you switch to Hey API, you can then implement the client interceptor approach described in the interceptor plan document.

### Step 4: Create Wrapper Files
Create the wrapper files as shown above:
- `src/lib/api/client-config.ts`
- `src/lib/api/workouts-queries.ts`
- `src/lib/api/workouts-mutations.ts`

### Step 5: Update Imports
In your components, you don't need to change anything because the wrapper functions maintain the same API as your current functions.

### Step 6: Verify and Test
Test all your components to ensure they work exactly as before.

## 4. Benefits of This Approach

1. **Zero Component Changes**: Your existing components require no modifications
2. **Type Safety**: You get full type safety from the generated code
3. **Future-Proof**: When your API changes, simply regenerate and the types will update automatically
4. **Better Query Keys**: Hey API generates normalized query keys which are more robust
5. **Enhanced Cache Invalidation**: With tags already enabled, you can invalidate queries by tags rather than specific keys
6. **Preparation for Centralized Authentication**: Sets up the foundation for implementing client interceptors when you switch to Hey API
7. **Consistency**: All query/mutation patterns will be consistent across your codebase

## 5. Future Enhancements (Optional)

Once you're comfortable with the migration, you can gradually:

1. Move away from wrapper functions and use generated functions directly
2. Use the generated query keys for more precise cache invalidation
3. Take advantage of tag-based cache invalidation that you already have enabled:
   ```ts
   // You can invalidate by tags (since tags are already enabled)
   queryClient.invalidateQueries({ predicate: (query) => 
     query.queryKey[0]?.tags?.includes('workouts') 
   });
   ```
4. Take advantage of other Hey API features like infinite queries if needed

## 6. Rollback Plan

If any issues arise:
1. Revert the generated files
2. Remove the wrapper files
3. Your original code will work as before

This approach ensures a safe, incremental migration with minimal risk.

## 7. Summary

This migration plan provides:

1. **Zero breaking changes** to your existing components
2. **Immediate benefits** of type safety and consistent query/mutation patterns
3. **Future flexibility** to leverage advanced features like tags-based cache invalidation
4. **Foundation for centralized authentication** through client interceptors when you switch to Hey API
5. **Safe rollback** capability if needed

The wrapper function approach allows you to enjoy the benefits of Hey API's generated code while maintaining complete control over your authentication logic and cache invalidation strategies. Once you switch to Hey API, you can then implement the client interceptor approach described in your interceptor plan document for even cleaner authentication handling.