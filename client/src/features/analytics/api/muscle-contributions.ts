import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type {
  WorkoutMuscleContributions,
  GetAnalyticsMuscleContributionsResponses,
} from "@/client/types.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

/** Server-generated muscle contribution summary. */
export type MuscleContributions = WorkoutMuscleContributions;
export function muscleContributionsQueryOptions(
  userId: string,
  startDate: string,
  timezone: string,
) {
  return queryOptions({
    queryKey: ["muscle-contributions", userId, startDate, timezone],
    queryFn: async ({ signal }) => {
      const response = await client.get<
        GetAnalyticsMuscleContributionsResponses,
        ApiError,
        true
      >({
        url: "/analytics/muscle-contributions",
        query: { startDate, timezone },
        signal,
        throwOnError: true,
      });
      return response.data;
    },
  });
}
