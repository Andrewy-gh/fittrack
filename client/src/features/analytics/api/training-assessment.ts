import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type {
  WorkoutStrengthAssessment,
  GetExercisesByIdAssessmentResponses,
} from "@/client/types.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

/** Server-generated strength assessment. */
export type StrengthAssessment = WorkoutStrengthAssessment;

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
        GetExercisesByIdAssessmentResponses,
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
