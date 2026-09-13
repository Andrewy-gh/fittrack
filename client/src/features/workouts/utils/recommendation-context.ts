import * as v from "valibot";

/** How the user feels for this exercise today. */
export type Readiness = "sluggish" | "normal" | "great";
/** Optional retrospective feedback saved with the displayed recommendation. */
export type ExerciseFeedback = "" | "too_little" | "about_right" | "too_much";

const rangeSchema = v.pipe(
  v.object({
    min: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(20)),
    max: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(20)),
  }),
  v.check((range) => range.min <= range.max),
);

/** Parses guidance at API and draft-storage boundaries. */
export const recommendationSchema = v.object({
  plan: v.optional(
    v.object({
      sets: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(20)),
      reps: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(100)),
      weight: v.nullish(v.pipe(v.number(), v.minValue(0), v.maxValue(10000))),
      goal: v.optional(v.string()),
      experience: v.optional(v.string()),
    }),
  ),
  exerciseId: v.pipe(v.number(), v.integer(), v.minValue(1)),
  readiness: v.picklist(["sluggish", "normal", "great"]),
  range: v.nullish(rangeSchema),
  baseline: v.nullish(rangeSchema),
  source: v.picklist(["none", "history", "prescription"]),
  previous: v.nullish(
    v.object({
      workoutId: v.number(),
      date: v.string(),
      workingSets: v.number(),
    }),
  ),
  explanation: v.string(),
  policyVersion: v.string(),
});

/** Stored guidance is a snapshot, never a source of actual workout sets. */
export const recommendationSnapshotSchema = v.object({
  exerciseName: v.string(),
  recommendation: recommendationSchema,
  feedback: v.optional(
    v.picklist(["", "too_little", "about_right", "too_much"]),
  ),
});
