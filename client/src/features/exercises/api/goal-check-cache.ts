import type { QueryClient } from "@tanstack/react-query";

// Workout edits may add/remove several exercises. Invalidate all observed checks.
export function invalidateGoalChecks(queryClient: QueryClient) {
  return queryClient.invalidateQueries({
    predicate: (query) => {
      const key = query.queryKey[0];
      return (
        typeof key === "object" &&
        key !== null &&
        "_id" in key &&
        key._id === "getExercisesByIdGoalCheck"
      );
    },
  });
}
