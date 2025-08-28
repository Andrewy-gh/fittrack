# Hey API Implementation Tutorial

This tutorial will guide you through implementing the new Hey API generated code in your existing components, using the exercises route as an example.

## Prerequisites

Before starting, ensure you have:
1. Generated the Hey API code using `bun run openapi-ts`
2. Set up the client interceptor as described in `client-interceptor-plan.md`
3. Created wrapper functions as described in `plan-2.md`

## Updating Your Exercises Route

Let's walk through updating `@client/src/routes/_auth/exercises/index.tsx` to use the new Hey API generated code.

### 1. Update Imports

First, update your imports to use the new generated functions:

**Before:**
```typescript
import type { exercise_ExerciseResponse } from '@/generated';
import { exercisesQueryOptions } from '@/lib/api/exercises';
```

**After:**
```typescript
import type { ExerciseExerciseResponse } from '@/client/types.gen';
import { exercisesQueryOptions } from '@/lib/api/exercises';
```

Note that you're still using the same wrapper function (`exercisesQueryOptions`), so your component code doesn't need to change.

### 2. Update Your Exercises API Wrapper Functions

Update your `src/lib/api/exercises.ts` file to use the new Hey API generated functions:

```typescript
import { 
  getExercisesQueryOptions,
  getExercisesByIdQueryOptions,
  getExercisesByIdRecentSetsQueryOptions,
  postExercisesMutation
} from '@/client/@tanstack/react-query.gen';
import { queryClient } from './api';
import { useMutation } from '@tanstack/react-query';
import { 
  getExercises,
  postExercises,
  getExercisesById,
  getExercisesByIdRecentSets
} from '@/client/sdk.gen';
import { ensureUser, type User } from './auth';
import type {
  ExerciseExerciseResponse,
  ExerciseExerciseWithSetsResponse,
  ExerciseRecentSetsResponse,
  ExerciseCreateExerciseRequest,
} from '@/client/types.gen';

/**
 * Exercise data as returned by the API - always has a database ID
 */
export type DbExercise = Pick<ExerciseExerciseResponse, 'id' | 'name'>;

/**
 * Exercise option for form components - may include manually created exercises without IDs
 * Used in dropdowns and forms where users can create new exercises on-the-fly
 */
export type ExerciseOption = {
  id: number | null; // null for manually created exercises, number for DB exercises
  name: string;
};

// Wrapper function that maintains your current API but uses generated options
export function exercisesQueryOptions(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base options from generated code
  const baseOptions = getExercisesQueryOptions();
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}

export function exerciseQueryOptions(
  exerciseId: number,
  user: User
) {
  const validatedUser = ensureUser(user);
  
  const baseOptions = getExercisesByIdQueryOptions({
    path: { id: exerciseId }
  });
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}

export function recentExerciseSetsQueryOptions(
  exerciseId: number,
  user: User
) {
  const validatedUser = ensureUser(user);
  
  const baseOptions = getExercisesByIdRecentSetsQueryOptions({
    path: { id: exerciseId }
  });
  
  // The interceptor will automatically add the auth token
  return baseOptions;
}

// Mutation wrapper functions
export function useCreateExerciseMutation(user: User) {
  const validatedUser = ensureUser(user);
  
  // Get the base mutation options from generated code
  const baseOptions = postExercisesMutation();
  
  return useMutation({
    ...baseOptions,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['exercises', 'list'],
      });
    },
  });
}
```

### 3. Component Code Remains Unchanged

The beauty of this approach is that your component code doesn't need to change at all:

```typescript
export const Route = createFileRoute('/_auth/exercises/')({
  loader: async ({ context }): Promise<{
    user: Exclude<User, null>;
  }> => {
    const user = context.user;
    checkUser(user); 
    context.queryClient.ensureQueryData(exercisesQueryOptions(user));
    return { user };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();
  const { data: exercises } = useSuspenseQuery(
    exercisesQueryOptions(user)
  );
  return <ExercisesDisplay exercises={exercises} />;
}
```

## How Authentication Works

With the client interceptor you set up, authentication is handled automatically:

1. The interceptor fetches the access token using your Stackframe auth
2. It automatically adds the `x-stack-access-token` header to all requests
3. Your components don't need to know about authentication at all

```typescript
// This is in your client-config.ts
client.interceptors.request.use(async (request, options) => {
  const accessToken = await getAccessToken();
  
  if (accessToken) {
    request.headers.set('x-stack-access-token', accessToken);
  }
  
  return request;
});
```

## Benefits of This Approach

1. **Zero Component Changes**: Your existing components work without modification
2. **Type Safety**: Full TypeScript support from the OpenAPI spec
3. **Automatic Authentication**: Handled by the interceptor
4. **Better Query Keys**: Hey API generates normalized query keys
5. **Future-Proof**: When your API changes, just regenerate the code

## Next Steps

1. Apply the same pattern to your other routes
2. Gradually migrate away from wrapper functions if desired
3. Take advantage of advanced features like tag-based cache invalidation

### Direct Usage (Optional)

Once you're comfortable, you can use the generated functions directly:

```typescript
// For queries
import { getExercisesQueryOptions } from '@/client/@tanstack/react-query.gen';
const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions());

// For mutations
import { postExercisesMutation } from '@/client/@tanstack/react-query.gen';
const createExercise = useMutation(postExercisesMutation());
```

But the wrapper approach allows for a gradual, safe migration.