import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

export type StrengthAssessment = {
  exerciseId: number;
  period: {
    startDate: string;
    endDate: string;
    timezone: string;
    partial: boolean;
  };
  sessions: {
    workoutId: number;
    date: string;
    workingSets: number;
    invalidWorkingSets: number;
    reasonCode: string;
    state: "below_reference" | "reference_met" | "cannot_assess";
    reason: string;
  }[];
  workingSets: number;
  trainingDays: number;
  sessionsMeetingReference: number;
  reference: string;
  sourceUrl: string;
  applicability: string;
  limitations: string[];
  policyVersion: string;
};

export function trainingAssessmentQueryOptions(
  userId: string,
  exerciseId: number,
  startDate: string,
  timezone: string,
) {
  return queryOptions({
    queryKey: ["training-assessment", userId, exerciseId, startDate, timezone],
    queryFn: async ({ signal }) => {
      const response = await client.get<
        { 200: StrengthAssessment },
        ApiError,
        true
      >({
        url: `/exercises/${exerciseId}/assessment`,
        query: { startDate, timezone },
        signal,
        throwOnError: true,
      });
      return response.data;
    },
  });
}
