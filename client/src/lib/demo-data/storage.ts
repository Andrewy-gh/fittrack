import type {
  StoredExercise,
  StoredSet,
  StoredWorkout,
  WorkoutWorkoutResponse,
  WorkoutWorkoutWithSetsResponse,
  ExerciseExerciseResponse,
  ExerciseExerciseWithSetsResponse,
  ExerciseRecentSetsResponse,
} from './types';
import { STORAGE_KEYS, DEMO_USER_ID } from './types';
import {
  INITIAL_EXERCISES,
  INITIAL_SETS,
  INITIAL_WORKOUTS,
  getNextId,
} from './initial-data';
import {
  bootstrapDemoHistorical1Rm,
  clearDemoHistorical1Rm,
  handleDemoExerciseDeleted,
  handleDemoWorkoutCreated,
  handleDemoWorkoutDeleted,
  handleDemoWorkoutUpdated,
  resetDemoHistorical1Rm,
} from './historical-1rm';

// ===========================
// localStorage Utilities
// ===========================

function getFromStorage<T>(key: string, defaultValue: T): T {
  if (typeof window === 'undefined') return defaultValue;

  const stored = localStorage.getItem(key);
  if (!stored) return defaultValue;

  try {
    return JSON.parse(stored) as T;
  } catch {
    return defaultValue;
  }
}

function setInStorage<T>(key: string, value: T): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(key, JSON.stringify(value));
}

// ===========================
// Initialization
// ===========================

export function initializeDemoData(): void {
  if (typeof window === 'undefined') return;

  const hasData = localStorage.getItem(STORAGE_KEYS.WORKOUTS);

  if (!hasData) {
    setInStorage(STORAGE_KEYS.EXERCISES, INITIAL_EXERCISES);
    setInStorage(STORAGE_KEYS.WORKOUTS, INITIAL_WORKOUTS);
    setInStorage(STORAGE_KEYS.SETS, INITIAL_SETS);
    bootstrapDemoHistorical1Rm();
  }
}

export function resetDemoData(): void {
  setInStorage(STORAGE_KEYS.EXERCISES, INITIAL_EXERCISES);
  setInStorage(STORAGE_KEYS.WORKOUTS, INITIAL_WORKOUTS);
  setInStorage(STORAGE_KEYS.SETS, INITIAL_SETS);
  resetDemoHistorical1Rm();
}

export function clearDemoData(): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(STORAGE_KEYS.EXERCISES);
  localStorage.removeItem(STORAGE_KEYS.WORKOUTS);
  localStorage.removeItem(STORAGE_KEYS.SETS);
  clearDemoHistorical1Rm();
}

// ===========================
// Exercise CRUD
// ===========================

export function getAllExercises(): ExerciseExerciseResponse[] {
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);
  return exercises.map((ex) => ({
    id: ex.id,
    name: ex.name,
    user_id: ex.user_id,
    created_at: ex.created_at,
    updated_at: ex.updated_at,
  }));
}

export function getExerciseById(id: number): ExerciseExerciseResponse | null {
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);
  const exercise = exercises.find((ex) => ex.id === id);

  if (!exercise) return null;

  return {
    id: exercise.id,
    name: exercise.name,
    user_id: exercise.user_id,
    created_at: exercise.created_at,
    updated_at: exercise.updated_at,
  };
}

export function createExercise(name: string): ExerciseExerciseResponse {
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);

  const newExercise: StoredExercise = {
    id: getNextId(exercises),
    name,
    user_id: DEMO_USER_ID,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  exercises.push(newExercise);
  setInStorage(STORAGE_KEYS.EXERCISES, exercises);

  return {
    id: newExercise.id,
    name: newExercise.name,
    user_id: newExercise.user_id,
    created_at: newExercise.created_at,
    updated_at: newExercise.updated_at,
  };
}

export function updateExercise(id: number, name: string): boolean {
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);
  const exerciseIndex = exercises.findIndex((ex) => ex.id === id);

  if (exerciseIndex === -1) return false;

  // Check for duplicate name (case-insensitive, excluding the exercise being updated)
  const duplicate = exercises.find(
    (ex) => ex.id !== id && ex.name.toLowerCase() === name.toLowerCase()
  );
  if (duplicate) {
    throw new Error(`Exercise name "${name}" already exists`);
  }

  // Update the exercise
  exercises[exerciseIndex].name = name;
  exercises[exerciseIndex].updated_at = new Date().toISOString();

  setInStorage(STORAGE_KEYS.EXERCISES, exercises);
  return true;
}

export function deleteExercise(id: number): boolean {
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);
  const filtered = exercises.filter((ex) => ex.id !== id);

  if (filtered.length === exercises.length) return false;

  // Also delete all sets referencing this exercise
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const filteredSets = sets.filter((set) => set.exercise_id !== id);
  setInStorage(STORAGE_KEYS.SETS, filteredSets);
  handleDemoExerciseDeleted(id);

  setInStorage(STORAGE_KEYS.EXERCISES, filtered);
  return true;
}

// Get all sets for a specific exercise (with workout context)
export function getExerciseWithSets(exerciseId: number): ExerciseExerciseWithSetsResponse[] {
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);

  const exerciseSets = sets.filter((set) => set.exercise_id === exerciseId);

  return exerciseSets.map((set) => {
    const workout = workouts.find((w) => w.id === set.workout_id);
    const exercise = exercises.find((e) => e.id === set.exercise_id);

    return {
      set_id: set.id,
      exercise_id: set.exercise_id,
      exercise_name: exercise?.name || 'Unknown',
      workout_id: set.workout_id,
      workout_date: workout?.date || '',
      workout_notes: workout?.notes,
      reps: set.reps,
      weight: set.weight,
      set_type: set.set_type,
      set_order: set.set_order,
      exercise_order: set.exercise_order,
      volume: (set.weight || 0) * set.reps,
    };
  });
}

// Get recent sets for an exercise
export function getExerciseRecentSets(exerciseId: number, limit = 10): ExerciseRecentSetsResponse[] {
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);

  const exerciseSets = sets
    .filter((set) => set.exercise_id === exerciseId)
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, limit);

  return exerciseSets.map((set) => {
    const workout = workouts.find((w) => w.id === set.workout_id);

    return {
      set_id: set.id,
      workout_id: set.workout_id,
      workout_date: workout?.date || '',
      reps: set.reps,
      weight: set.weight,
      set_order: set.set_order,
      exercise_order: set.exercise_order,
      created_at: set.created_at,
    };
  });
}

// ===========================
// Workout CRUD
// ===========================

export function getAllWorkouts(): WorkoutWorkoutResponse[] {
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);

  return workouts
    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
    .map((workout) => ({
      id: workout.id,
      date: workout.date,
      notes: workout.notes,
      workout_focus: workout.workout_focus,
      user_id: workout.user_id,
      created_at: workout.created_at,
      updated_at: workout.updated_at,
    }));
}

export function getWorkoutById(id: number): WorkoutWorkoutWithSetsResponse[] {
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);

  const workout = workouts.find((w) => w.id === id);
  if (!workout) return [];

  const workoutSets = sets
    .filter((set) => set.workout_id === id)
    .sort((a, b) => {
      if (a.exercise_order !== b.exercise_order) {
        return (a.exercise_order || 0) - (b.exercise_order || 0);
      }
      return (a.set_order || 0) - (b.set_order || 0);
    });

  return workoutSets.map((set) => {
    const exercise = exercises.find((e) => e.id === set.exercise_id);

    return {
      workout_id: workout.id,
      workout_date: workout.date,
      workout_notes: workout.notes,
      workout_focus: workout.workout_focus,
      set_id: set.id,
      exercise_id: set.exercise_id,
      exercise_name: exercise?.name || 'Unknown',
      exercise_order: set.exercise_order,
      set_order: set.set_order,
      reps: set.reps,
      weight: set.weight,
      set_type: set.set_type,
      volume: (set.weight || 0) * set.reps,
    };
  });
}

interface CreateWorkoutInput {
  date: string;
  notes?: string;
  workoutFocus?: string;
  exercises: Array<{
    name: string;
    sets: Array<{
      reps: number;
      weight?: number;
      setType: 'warmup' | 'working';
    }>;
  }>;
}

export function createWorkout(input: CreateWorkoutInput): { success: boolean } {
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);

  // Create new workout
  const newWorkout: StoredWorkout = {
    id: getNextId(workouts),
    date: input.date,
    notes: input.notes,
    workout_focus: input.workoutFocus,
    user_id: DEMO_USER_ID,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  workouts.push(newWorkout);
  setInStorage(STORAGE_KEYS.WORKOUTS, workouts);

  // Create sets for each exercise
  input.exercises.forEach((exerciseInput, exerciseIndex) => {
    // Find or create exercise
    let exercise = exercises.find((e) => e.name === exerciseInput.name);

    if (!exercise) {
      exercise = {
        id: getNextId(exercises),
        name: exerciseInput.name,
        user_id: DEMO_USER_ID,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      exercises.push(exercise);
    }

    // Create sets for this exercise
    exerciseInput.sets.forEach((setInput, setIndex) => {
      const newSet: StoredSet = {
        id: getNextId(sets),
        exercise_id: exercise.id,
        workout_id: newWorkout.id,
        reps: setInput.reps,
        weight: setInput.weight,
        set_type: setInput.setType,
        exercise_order: exerciseIndex,
        set_order: setIndex,
        user_id: DEMO_USER_ID,
        created_at: new Date().toISOString(),
      };
      sets.push(newSet);
    });
  });

  setInStorage(STORAGE_KEYS.EXERCISES, exercises);
  setInStorage(STORAGE_KEYS.SETS, sets);
  handleDemoWorkoutCreated(newWorkout.id);

  return { success: true };
}

export function updateWorkout(
  id: number,
  input: CreateWorkoutInput
): { success: boolean } {
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
  const exercises = getFromStorage<StoredExercise[]>(STORAGE_KEYS.EXERCISES, []);

  const workoutIndex = workouts.findIndex((w) => w.id === id);
  if (workoutIndex === -1) return { success: false };

  // Update workout metadata
  workouts[workoutIndex] = {
    ...workouts[workoutIndex],
    date: input.date,
    notes: input.notes,
    workout_focus: input.workoutFocus,
    updated_at: new Date().toISOString(),
  };

  // Delete old sets for this workout
  const filteredSets = sets.filter((set) => set.workout_id !== id);

  // Create new sets
  const newSets: StoredSet[] = [];
  input.exercises.forEach((exerciseInput, exerciseIndex) => {
    let exercise = exercises.find((e) => e.name === exerciseInput.name);

    if (!exercise) {
      exercise = {
        id: getNextId([...exercises, ...newSets.map((s) => ({ id: s.exercise_id }))]),
        name: exerciseInput.name,
        user_id: DEMO_USER_ID,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      exercises.push(exercise);
    }

    exerciseInput.sets.forEach((setInput, setIndex) => {
      const newSet: StoredSet = {
        id: getNextId([...filteredSets, ...newSets]),
        exercise_id: exercise.id,
        workout_id: id,
        reps: setInput.reps,
        weight: setInput.weight,
        set_type: setInput.setType,
        exercise_order: exerciseIndex,
        set_order: setIndex,
        user_id: DEMO_USER_ID,
        created_at: new Date().toISOString(),
      };
      newSets.push(newSet);
    });
  });

  setInStorage(STORAGE_KEYS.WORKOUTS, workouts);
  setInStorage(STORAGE_KEYS.EXERCISES, exercises);
  setInStorage(STORAGE_KEYS.SETS, [...filteredSets, ...newSets]);
  handleDemoWorkoutUpdated(id);

  return { success: true };
}

export function deleteWorkout(id: number): boolean {
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const sets = getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);

  const filtered = workouts.filter((w) => w.id !== id);
  if (filtered.length === workouts.length) return false;

  // Delete all sets for this workout
  const filteredSets = sets.filter((set) => set.workout_id !== id);

  setInStorage(STORAGE_KEYS.WORKOUTS, filtered);
  setInStorage(STORAGE_KEYS.SETS, filteredSets);
  handleDemoWorkoutDeleted(id);

  return true;
}

// ===========================
// Focus Values
// ===========================

export function getWorkoutFocusValues(): string[] {
  const workouts = getFromStorage<StoredWorkout[]>(STORAGE_KEYS.WORKOUTS, []);
  const focusValues = workouts
    .map((w) => w.workout_focus)
    .filter((focus): focus is string => !!focus);

  return Array.from(new Set(focusValues)).sort();
}
