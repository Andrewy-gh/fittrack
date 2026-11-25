import * as v from "valibot";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
  WorkoutSetInput,
} from "@/client";

// New protocol fields require an explicit eligibility decision rather than
// silently treating an edited future field as unchanged.
const unhandledWorkoutFields: Record<
  Exclude<
    keyof WorkoutCreateWorkoutRequest,
    "date" | "notes" | "workoutFocus" | "exercises" | "recommendations"
  >,
  never
> = {};
const unhandledExerciseFields: Record<
  Exclude<keyof WorkoutExerciseInput, "name" | "catalog_id" | "sets">,
  never
> = {};
const unhandledSetFields: Record<
  Exclude<keyof WorkoutSetInput, "reps" | "setType" | "weight">,
  never
> = {};

/** Compare the editable inputs, not recommendation results. Undo restores eligibility. */
export function sameWorkoutInputs(
  left: WorkoutCreateWorkoutRequest,
  right: WorkoutCreateWorkoutRequest,
): boolean {
  void unhandledWorkoutFields;
  void unhandledExerciseFields;
  void unhandledSetFields;
  const inputs = (
    workout: WorkoutCreateWorkoutRequest,
    submitting: boolean,
  ) => ({
    date: workout.date,
    // The form trims text on POST. Compare the actual outgoing values to the
    // original, so that trimming itself cannot silently retain old context.
    notes: submitting ? (workout.notes?.trim() ?? "") : (workout.notes ?? ""),
    workoutFocus: submitting
      ? (workout.workoutFocus?.trim() ?? "")
      : (workout.workoutFocus ?? ""),
    exercises: workout.exercises.map((exercise) => ({
      name: exercise.name,
      catalog_id: exercise.catalog_id ?? null,
      sets: exercise.sets.map((set) => ({
        reps: set.reps,
        setType: set.setType,
        weight: set.weight ?? null,
      })),
    })),
  });
  return (
    JSON.stringify(inputs(left, true)) === JSON.stringify(inputs(right, false))
  );
}

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
