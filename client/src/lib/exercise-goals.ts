export type ExerciseGoal = {
  targetWeight?: number;
  targetReps?: number;
  frequencyPerWeek?: number;
};

type ExerciseGoalParseOptions = {
  integer?: boolean;
};

type ExerciseGoalLookup = {
  exerciseId?: number | null;
  exerciseName?: string | null;
};

const STORAGE_KEY = 'fittrack-exercise-goals-v1';

function normalizeExerciseName(exerciseName: string): string {
  return exerciseName.trim().toLowerCase();
}

function getExerciseGoalKey({
  exerciseId,
  exerciseName,
}: ExerciseGoalLookup): string | null {
  if (exerciseId != null) {
    return `id:${exerciseId}`;
  }

  if (exerciseName && normalizeExerciseName(exerciseName)) {
    return `name:${normalizeExerciseName(exerciseName)}`;
  }

  return null;
}

function readExerciseGoals(): Record<string, ExerciseGoal> {
  if (typeof window === 'undefined') {
    return {};
  }

  const storedValue = localStorage.getItem(STORAGE_KEY);
  if (!storedValue) {
    return {};
  }

  try {
    return JSON.parse(storedValue) as Record<string, ExerciseGoal>;
  } catch {
    return {};
  }
}

function writeExerciseGoals(goals: Record<string, ExerciseGoal>): void {
  if (typeof window === 'undefined') {
    return;
  }

  localStorage.setItem(STORAGE_KEY, JSON.stringify(goals));
}

export function getExerciseGoal(lookup: ExerciseGoalLookup): ExerciseGoal | null {
  const goalKey = getExerciseGoalKey(lookup);
  if (!goalKey) {
    return null;
  }

  const goals = readExerciseGoals();
  return goals[goalKey] ?? null;
}

export function saveExerciseGoal(
  lookup: ExerciseGoalLookup,
  goal: ExerciseGoal
): void {
  const goalKey = getExerciseGoalKey(lookup);
  if (!goalKey) {
    return;
  }

  const hasValues = Object.values(goal).some((value) => value != null);
  const goals = readExerciseGoals();

  if (!hasValues) {
    delete goals[goalKey];
  } else {
    goals[goalKey] = goal;
  }

  writeExerciseGoals(goals);
}

export function formatExerciseGoalSummary(goal?: ExerciseGoal | null): string | null {
  if (!goal) {
    return null;
  }

  const parts: string[] = [];

  if (goal.targetWeight != null) {
    parts.push(`${goal.targetWeight} lb`);
  }

  if (goal.targetReps != null) {
    parts.push(`${goal.targetReps} reps`);
  }

  if (goal.frequencyPerWeek != null) {
    parts.push(`${goal.frequencyPerWeek}x / week`);
  }

  return parts.length > 0 ? parts.join(' • ') : null;
}

export function parseExerciseGoalInput(
  value: string,
  label: string,
  options: ExerciseGoalParseOptions = {}
): { value?: number; error?: string } {
  const trimmedValue = value.trim();
  if (!trimmedValue) {
    return {};
  }

  const parsedValue = Number(trimmedValue);

  if (!Number.isFinite(parsedValue)) {
    return {
      error: `${label} must be a number`,
    };
  }

  if (parsedValue < 0) {
    return {
      error: `${label} must be 0 or greater`,
    };
  }

  if (options.integer && !Number.isInteger(parsedValue)) {
    return {
      error: `${label} must be a whole number`,
    };
  }

  return {
    value: parsedValue,
  };
}

export function clearExerciseGoals(): void {
  if (typeof window === 'undefined') {
    return;
  }

  localStorage.removeItem(STORAGE_KEY);
}
