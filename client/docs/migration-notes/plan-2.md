# Migration Plan v2: Minimal Component Changes

This plan focuses on migrating to Hey API's TanStack Query plugin with zero breaking changes to your existing components. You've already generated the Hey API code, so now we'll integrate it properly with your authentication flow.

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

This generates functions like:
- `getWorkoutsQueryOptions()` (for queries)
- `getWorkoutsQueryKey()` (for query keys)
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

### 1.2 Using Generated Query Keys

The Hey API plugin generates both query options and query keys. You can access query keys in two ways:

1. From query options:
   ```ts
   const { queryKey } = getWorkoutsQueryOptions();
   ```

2. Directly from query key functions:
   ```ts
   const queryKey = getWorkoutsQueryKey();
   ```

Using generated query keys ensures consistency and type safety. They contain normalized parameters and metadata that make cache management more robust.

## 2. Migration Strategy by Function Type

### 2.1 Queries - Wrapper Function Approach

Create wrapper functions that maintain your current API but use the generated Hey API functions internally. Since you've implemented a client interceptor in `client-config.ts`, the auth headers will be automatically added to each request. See `src/lib/api/exercises.ts` for a working example.

For consistency with your existing pattern, you can also export the generated query keys alongside query options. This makes it easier to use them for cache invalidation.

**File: `src/lib/api/workouts-queries.ts`**

```typescript
import { 
  getWorkoutsOptions,
  getWorkoutsQueryKey,
  getWorkouts1Options,
  getWorkoutsByIdQueryKey
} from '@/generated/@tanstack/react-query';
import { ensureUser, type User } from './auth';
import type {
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse,
} from '@/generated';

// Export generated query keys for easier cache invalidation
export { 
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey
};

// Wrapper functions that maintain your current API but use generated options
export function workoutsQueryOptions(user: User) {
  ensureUser(user);
  
  // Get the base options from generated code
  const baseOptions = getWorkoutsOptions();
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}

export function workoutQueryOptions(workoutId: number, user: User) {
  ensureUser(user);
  
  const baseOptions = getWorkouts1Options({ 
    path: { id: workoutId } // Adjust parameter name as per generated function
  });
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}
```

### 2.2 Mutations - Wrapper Function Approach

Similarly for mutations, create wrapper functions that preserve your current API. See `src/lib/api/exercises.ts` for examples of how to structure these functions.

When handling cache invalidation in `onSuccess`, you can use the generated query keys for more precise invalidation. This ensures consistency with the query key structure used by the generated code.

**File: `src/lib/api/workouts-mutations.ts`**

```typescript
import { 
  postWorkoutsMutation,
  putWorkoutsMutation,
  deleteWorkoutsMutation,
  getWorkoutsQueryKey,
  getWorkoutsByIdQueryKey
} from '@/generated/@tanstack/react-query';
import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import { WorkoutsService } from '@/generated';
import { ensureUser, type User } from './auth';
import type {
  workout_CreateWorkoutRequest,
  workout_UpdateWorkoutRequest,
} from '@/generated';

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
  return WorkoutsService.deleteWorkouts(workoutId);
}
```

### 2.3 Handling Query and Validation on Success

When using the generated mutation functions, you can leverage the generated mutation options while still customizing the `onSuccess` handler. The client interceptor will automatically handle authentication for all requests.

**Recommended Approach** (used in the examples above):
   ```typescript
   return useMutation({
     ...postWorkoutsMutation(), // Spread the generated options
     onSuccess: () => {
       // Your custom success logic
       queryClient.invalidateQueries({ queryKey: ['workouts', 'list'] });
     },
   });
   ```

This approach is recommended because it:
- Leverages the generated mutation options for type safety
- Allows you to customize the `onSuccess` handler for cache invalidation
- Automatically uses the client interceptor for authentication
- Maintains consistency with the pattern shown in `src/lib/api/exercises.ts`

### Step 3: Update Main.tsx
Update your `main.tsx` to import the client configuration:

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

### Step 4: Generate the Code
Run your OpenAPI generation command:
```bash
bun run openapi-ts  # or whatever command you use
```

### Step 5: Create Wrapper Files
Create the wrapper files as shown above:
- `src/lib/api/workouts-queries.ts` (demonstrates query options and key usage)
- `src/lib/api/workouts-mutations.ts` (demonstrates mutation usage with proper cache invalidation)
- `src/lib/api/exercises.ts` (already updated to show proper patterns)

### Step 6: Update Imports
In your components, you don't need to change anything because the wrapper functions maintain the same API as your current functions.

### Step 7: Verify and Test
Test all your components to ensure they work exactly as before.

## 4. Benefits of This Approach

1. **Zero Component Changes**: Your existing components require no modifications
2. **Type Safety**: You get full type safety from the generated code
3. **Future-Proof**: When your API changes, simply regenerate and the types will update automatically
4. **Better Query Keys**: Hey API generates normalized query keys which are more robust
5. **Enhanced Cache Invalidation**: With tags already enabled, you can invalidate queries by tags rather than specific keys
6. **Centralized Authentication**: Client interceptors handle authentication automatically (see `src/lib/api/client-config.ts`)
7. **Consistency**: All query/mutation patterns will be consistent across your codebase

### 4.1 Benefits of Generated Query Keys

Using the generated query keys provides several advantages over manually crafted keys:

1. **Consistency**: Generated keys follow the same structure as used in the queries themselves
2. **Type Safety**: Full TypeScript support with proper typing of parameters
3. **Automatic Normalization**: Parameters are automatically normalized, making cache hits more reliable
4. **Future-Proof**: When API changes occur, query keys update automatically with regeneration
5. **Metadata Support**: Generated keys include operation metadata like tags for advanced invalidation

For example, a generated query key includes all relevant information:
```ts
const queryKey = [{
  _id: 'getWorkouts',
  baseUrl: '/api',
  tags: ['workouts', 'get']
}];
```

This is more robust than manually crafted keys like `['workouts', 'list']` because it includes metadata that can be used for more sophisticated cache management.

### 4.2 Exporting Query Keys

As demonstrated in `src/lib/api/workouts-queries.ts` and `src/lib/api/exercises.ts`, exporting the generated query keys from your API modules provides several benefits:

1. **Easy Cache Invalidation**: Consumers can use the exact same query keys that the queries use
2. **Consistency**: Ensures the same key structure is used for both queries and invalidation
3. **Type Safety**: Full TypeScript support when invalidating queries
4. **Maintainability**: When query signatures change, both queries and invalidation update together

Example usage in mutations:
```ts
// In a mutation's onSuccess handler
onSuccess: () => {
  // Use generated query keys for consistent cache invalidation
  queryClient.invalidateQueries({
    queryKey: getWorkoutsQueryKey(),
  });
}
```

## 5. Future Enhancements (Optional)

Once you're comfortable with the migration, you can gradually:

1. Move away from wrapper functions and use generated functions directly
2. Use the generated query keys for more precise cache invalidation:
   ```ts
   // Instead of manually crafted keys like ['workouts', 'list']
   queryClient.invalidateQueries({ queryKey: getWorkoutsQueryKey() });
   
   // For specific items
   queryClient.invalidateQueries({ queryKey: getWorkoutsByIdQueryKey({ path: { id: workoutId } }) });
   ```
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
4. **Centralized authentication** through client interceptors
5. **Robust cache management** through generated query keys
6. **Safe rollback** capability if needed

The wrapper function approach allows you to enjoy the benefits of Hey API's generated code while maintaining the same API surface for your components. The client interceptor approach handles authentication automatically, making your code cleaner and more maintainable. See `src/lib/api/exercises.ts`, `src/lib/api/workouts-queries.ts`, and `src/lib/api/workouts-mutations.ts` for working implementations of this pattern that demonstrate proper usage of generated query keys and cache invalidation.