import * as v from "valibot";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
  WorkoutSetInput,
} from "@/client";
import type { ImportedRecommendationContext } from "@/lib/local-storage";

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

/** Build the request the form will actually save, including eligible historical context. */
export function buildNewWorkoutRequest(
  value: WorkoutCreateWorkoutRequest,
  importedContext?: ImportedRecommendationContext,
): WorkoutCreateWorkoutRequest {
  const baseline = importedContext?.baseline;
  const payload = {
    ...value,
    notes: workoutTextForSave(value.notes, baseline?.notes),
    workoutFocus: workoutTextForSave(
      value.workoutFocus,
      baseline?.workoutFocus,
    ),
  };
  return {
    ...payload,
    recommendations:
      importedContext?.recommendations !== undefined
        ? sameWorkoutInputs(payload, importedContext.baseline)
          ? importedContext.recommendations
          : undefined
        : payload.recommendations?.filter((snapshot) =>
            payload.exercises.some(
              (exercise) => exercise.name === snapshot.exerciseName,
            ),
          ),
  };
}

function workoutTextForSave(
  current: string | null | undefined,
  original: string | null | undefined,
): string | undefined {
  // Reuse untouched chat text verbatim. JS trim removes some characters that
  // Go's chat normalization retains (e.g. U+FEFF). Trim only newly entered text.
  return (current === original ? current : current?.trim()) || undefined;
}

function sameWorkoutInputs(
  left: WorkoutCreateWorkoutRequest,
  right: WorkoutCreateWorkoutRequest,
): boolean {
  void unhandledWorkoutFields;
  void unhandledExerciseFields;
  void unhandledSetFields;
  const inputs = (workout: WorkoutCreateWorkoutRequest) => ({
    date: workout.date,
    notes: workout.notes ?? "",
    workoutFocus: workout.workoutFocus ?? "",
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
  return JSON.stringify(inputs(left)) === JSON.stringify(inputs(right));
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
