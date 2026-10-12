/** Server workout text and exercise-name limit, measured in Unicode code points. */
export const WORKOUT_TEXT_LIMIT = 256;

/** Count code points like the Go validator, rather than UTF-16 code units. */
export function workoutTextLength(value: string): number {
  return Array.from(value).length;
}

/** Validate the submitted text without truncating the user's draft. */
export function validateWorkoutText(
  value: string,
  label: string,
): string | undefined {
  return workoutTextLength(value) > WORKOUT_TEXT_LIMIT
    ? `${label} must be ${WORKOUT_TEXT_LIMIT} characters or less.`
    : undefined;
}

/** Check every bounded text value before either workout mutation. */
export function workoutTextError(value: {
  notes?: string;
  workoutFocus?: string;
  exercises: readonly { name: string }[];
}): string | undefined {
  return (
    validateWorkoutText(value.notes?.trim() ?? "", "Notes") ??
    validateWorkoutText(value.workoutFocus?.trim() ?? "", "Workout focus") ??
    value.exercises
      .map((exercise) => validateWorkoutText(exercise.name, "Exercise name"))
      .find(Boolean)
  );
}
