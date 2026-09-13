import * as v from "valibot";
import {
  getExercisesByIdRecommendation,
  type RecommendationResult,
} from "@/client";
import "@/lib/api/client-config";
import {
  recommendationSchema,
  type Readiness,
} from "@/features/workouts/utils/recommendation-context";

class RecommendationRequestFailed extends Error {
  readonly _tag = "RecommendationRequestFailed" as const;

  constructor() {
    super("We couldn’t load your exercise suggestion. Please try again.");
  }
}

type Outcome<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: RecommendationRequestFailed };

/** Network boundary used by exercise guidance and its behavioral tests. */
export type RecommendationApi = {
  /** Loads the current suggestion for an exercise and readiness level. */
  get(
    exerciseId: number,
    readiness: Readiness,
    signal?: AbortSignal,
  ): Promise<Outcome<RecommendationResult>>;
};

/** Generated-client adapter that parses malformed responses into an expected failure. */
export const recommendationApi: RecommendationApi = {
  async get(exerciseId, readiness, signal) {
    try {
      const response = await getExercisesByIdRecommendation({
        path: { id: exerciseId },
        query: { readiness },
        signal,
      });
      const parsed = v.safeParse(recommendationSchema, response.data);
      return parsed.success
        ? { ok: true, value: parsed.output }
        : { ok: false, error: new RecommendationRequestFailed() };
    } catch {
      return { ok: false, error: new RecommendationRequestFailed() };
    }
  },
};
