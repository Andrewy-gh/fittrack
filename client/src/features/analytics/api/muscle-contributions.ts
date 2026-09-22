import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

export type MuscleContributions = {
  period: {
    startDate: string;
    endDate: string;
    timezone: string;
    partial: boolean;
  };
  workingSets: number;
  invalidWorkingSets: number;
  unclassifiedSets: number;
  unreviewedSets: number;
  mappingVersion: string;
  catalogVersion: string;
  classificationBasis: string;
  muscles: {
    muscle: string;
    directSets: number;
    indirectSets: number;
    unresolvedSets: number;
    invalidMappedSets: number;
    invalidUnresolvedSets: number;
  }[];
  exercises: {
    exerciseId: number;
    name: string;
    catalogId: string;
    workingSets: number;
    invalidWorkingSets: number;
    roles: Record<string, string>;
  }[];
};
export function muscleContributionsQueryOptions(
  userId: string,
  startDate: string,
  timezone: string,
) {
  return queryOptions({
    queryKey: ["muscle-contributions", userId, startDate, timezone],
    queryFn: async ({ signal }) => {
      const response = await client.get<
        { 200: MuscleContributions },
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
