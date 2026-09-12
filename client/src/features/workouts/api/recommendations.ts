import * as v from "valibot";
import {
  getExercisesByIdRecommendation,
  getWorkoutsByIdRecommendations,
  putExercisesByIdPrescription,
  type RecommendationRange,
  type RecommendationResult,
  type RecommendationSnapshot,
} from "@/client";
import "@/lib/api/client-config";
import {
  recommendationSchema,
  recommendationSnapshotSchema,
  type Readiness,
} from "@/features/workouts/utils/recommendation-context";

class RecommendationRequestFailed extends Error {
  readonly _tag = "RecommendationRequestFailed";
  constructor() {
    super("Could not load or save exercise guidance. Please try again.");
  }
}
type Outcome<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: RecommendationRequestFailed };

/** Network boundary used by exercise guidance and its behavioral tests. */
export type RecommendationApi = {
  get(
    exerciseId: number,
    readiness: Readiness,
    signal?: AbortSignal,
  ): Promise<Outcome<RecommendationResult>>;
  prescribe(
    exerciseId: number,
    baseline: RecommendationRange | null,
  ): Promise<Outcome<null>>;
  saved(
    workoutId: number,
    signal?: AbortSignal,
  ): Promise<Outcome<RecommendationSnapshot[]>>;
};

/** Generated-client adapter: expected transport and malformed-response failures remain values. */
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
  async prescribe(exerciseId, baseline) {
    try {
      const response = await putExercisesByIdPrescription({
        path: { id: exerciseId },
        body: { baseline },
      });
      return response.response?.ok
        ? { ok: true, value: null }
        : { ok: false, error: new RecommendationRequestFailed() };
    } catch {
      return { ok: false, error: new RecommendationRequestFailed() };
    }
  },
  async saved(workoutId, signal) {
    try {
      const response = await getWorkoutsByIdRecommendations({
        path: { id: workoutId },
        signal,
      });
      const parsed = v.safeParse(
        v.array(recommendationSnapshotSchema),
        response.data,
      );
      return parsed.success
        ? { ok: true, value: parsed.output }
        : { ok: false, error: new RecommendationRequestFailed() };
    } catch {
      return { ok: false, error: new RecommendationRequestFailed() };
    }
  },
};
