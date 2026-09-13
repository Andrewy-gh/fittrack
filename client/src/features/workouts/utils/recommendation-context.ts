import * as v from "valibot";

/** How the user feels for this exercise today. */
export type Readiness = "sluggish" | "normal" | "great";

/** Parses exercise guidance at API and draft-storage boundaries. */
export const recommendationSchema = v.object({
  plan: v.optional(
    v.object({
      sets: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(20)),
      reps: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(100)),
      weight: v.nullish(v.pipe(v.number(), v.minValue(0), v.maxValue(10000))),
    }),
  ),
  exerciseId: v.pipe(v.number(), v.integer(), v.minValue(1)),
  readiness: v.picklist(["sluggish", "normal", "great"]),
  explanation: v.string(),
  policyVersion: v.literal("exercise-plan-v2"),
});

/** Parses guidance saved with a workout separately from its actual sets. */
export const recommendationSnapshotSchema = v.object({
  exerciseName: v.string(),
  recommendation: recommendationSchema,
});
