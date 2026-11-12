import { queryClient } from "@/lib/api/api";
import {
  getExercisesQueryKey,
  getWorkoutsQueryKey,
} from "@/client/@tanstack/react-query.gen";

/**
 * Workout writes and exercise deletion change the training graph, not just a
 * list: workout details/recommendations, exercise sets/metrics, activity, focus
 * templates, assessments and muscle contributions all depend on that graph.
 * Generated tags match every cached parameter variant without guessing IDs
 * from a write response (updates can also remove exercises).
 *
 * Only mounted consumers refetch now. Inactive queries remain invalidated for
 * their next read; loaders must use Infinity, not invalidation-proof "static".
 * No response data is copied between query keys or authenticated/demo caches.
 * Custom analytics keys are scoped to the mutation's starting user, captured
 * by onMutate so an in-flight completion cannot refresh another user's key.
 * Generated authenticated keys retain their existing session/transport boundary.
 */
export async function invalidateTrainingData(userId: string) {
  const exercises = getExercisesQueryKey()[0];
  const workouts = getWorkoutsQueryKey()[0];
  await Promise.all([
    queryClient.invalidateQueries({
      queryKey: [{ baseUrl: exercises.baseUrl, tags: exercises.tags }],
    }),
    queryClient.invalidateQueries({
      queryKey: [{ baseUrl: workouts.baseUrl, tags: workouts.tags }],
    }),
    queryClient.invalidateQueries({
      queryKey: ["training-assessment", userId],
    }),
    queryClient.invalidateQueries({
      queryKey: ["muscle-contributions", userId],
    }),
  ]);
}
