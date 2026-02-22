import type { StoredSet } from './types';
import { STORAGE_KEYS } from './types';

type Historical1RmEntry = {
  historical_1rm: number;
  updated_at: string;
  source_workout_id: number | null;
};

type Historical1RmMap = Record<string, Historical1RmEntry>;

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

function removeFromStorage(key: string): void {
  if (typeof window === 'undefined') return;
  localStorage.removeItem(key);
}

function getAllSets(): StoredSet[] {
  return getFromStorage<StoredSet[]>(STORAGE_KEYS.SETS, []);
}

function getMap(): Historical1RmMap {
  return getFromStorage<Historical1RmMap>(STORAGE_KEYS.HISTORICAL_1RM, {});
}

function setMap(map: Historical1RmMap): void {
  setInStorage(STORAGE_KEYS.HISTORICAL_1RM, map);
}

function computeSetE1rm(s: StoredSet): number | null {
  if (s.set_type !== 'working') return null;
  const w = s.weight ?? 0;
  const reps = s.reps ?? 0;
  const e1rm = w * (1 + reps / 30);
  if (!Number.isFinite(e1rm) || e1rm <= 0) return null;
  return e1rm;
}

function computeBestE1rmForExercise(sets: StoredSet[], exerciseId: number): { best: number; workoutId: number } | null {
  let best: number | null = null;
  let workoutId: number | null = null;
  for (const s of sets) {
    if (s.exercise_id !== exerciseId) continue;
    const e1rm = computeSetE1rm(s);
    if (e1rm == null) continue;
    if (best == null || e1rm > best) {
      best = e1rm;
      workoutId = s.workout_id;
    }
  }
  if (best == null || workoutId == null) return null;
  return { best, workoutId };
}

function computeWorkoutBestE1rmByExercise(sets: StoredSet[], workoutId: number): Map<number, number> {
  const byExercise = new Map<number, number>();
  for (const s of sets) {
    if (s.workout_id !== workoutId) continue;
    const e1rm = computeSetE1rm(s);
    if (e1rm == null) continue;
    const prev = byExercise.get(s.exercise_id);
    if (prev == null || e1rm > prev) byExercise.set(s.exercise_id, e1rm);
  }
  return byExercise;
}

export function getDemoExerciseHistorical1Rm(exerciseId: number): Historical1RmEntry | null {
  const map = getMap();
  return map[String(exerciseId)] ?? null;
}

export function setDemoExerciseHistorical1RmManual(exerciseId: number, historical1rm: number | null): Historical1RmEntry | null {
  const map = getMap();
  const key = String(exerciseId);

  if (historical1rm == null) {
    delete map[key];
    setMap(map);
    return null;
  }

  const entry: Historical1RmEntry = {
    historical_1rm: historical1rm,
    updated_at: new Date().toISOString(),
    source_workout_id: null,
  };
  map[key] = entry;
  setMap(map);
  return entry;
}

export function recomputeDemoExerciseHistorical1Rm(exerciseId: number): Historical1RmEntry | null {
  const sets = getAllSets();
  const best = computeBestE1rmForExercise(sets, exerciseId);

  const map = getMap();
  const key = String(exerciseId);
  if (best == null) {
    delete map[key];
    setMap(map);
    return null;
  }

  const entry: Historical1RmEntry = {
    historical_1rm: best.best,
    updated_at: new Date().toISOString(),
    source_workout_id: best.workoutId,
  };
  map[key] = entry;
  setMap(map);
  return entry;
}

export function bootstrapDemoHistorical1Rm(): void {
  const sets = getAllSets();
  const map: Historical1RmMap = {};

  for (const s of sets) {
    const e1rm = computeSetE1rm(s);
    if (e1rm == null) continue;
    const key = String(s.exercise_id);
    const prev = map[key];
    if (!prev || e1rm > prev.historical_1rm) {
      map[key] = {
        historical_1rm: e1rm,
        updated_at: new Date().toISOString(),
        source_workout_id: s.workout_id,
      };
    }
  }

  setMap(map);
}

export function handleDemoWorkoutCreated(workoutId: number): void {
  const sets = getAllSets();
  const bestByExercise = computeWorkoutBestE1rmByExercise(sets, workoutId);
  if (bestByExercise.size === 0) return;

  const map = getMap();
  const now = new Date().toISOString();
  for (const [exerciseId, bestE1rm] of bestByExercise.entries()) {
    const key = String(exerciseId);
    const prev = map[key];
    if (!prev || bestE1rm > prev.historical_1rm) {
      map[key] = {
        historical_1rm: bestE1rm,
        updated_at: now,
        source_workout_id: workoutId,
      };
    }
  }
  setMap(map);
}

export function handleDemoWorkoutUpdated(workoutId: number): void {
  handleDemoWorkoutDeleted(workoutId);
  handleDemoWorkoutCreated(workoutId);
}

export function handleDemoWorkoutDeleted(workoutId: number): void {
  const map = getMap();
  const entries = Object.entries(map);
  const affectedExerciseIds: number[] = [];
  for (const [exerciseId, entry] of entries) {
    if (entry.source_workout_id === workoutId) affectedExerciseIds.push(Number(exerciseId));
  }
  if (affectedExerciseIds.length === 0) return;

  const sets = getAllSets();
  const now = new Date().toISOString();

  for (const exerciseId of affectedExerciseIds) {
    const best = computeBestE1rmForExercise(sets, exerciseId);
    const key = String(exerciseId);
    if (best == null) {
      delete map[key];
      continue;
    }
    map[key] = {
      historical_1rm: best.best,
      updated_at: now,
      source_workout_id: best.workoutId,
    };
  }

  setMap(map);
}

export function handleDemoExerciseDeleted(exerciseId: number): void {
  const map = getMap();
  delete map[String(exerciseId)];
  setMap(map);
}

export function resetDemoHistorical1Rm(): void {
  setMap({});
  bootstrapDemoHistorical1Rm();
}

export function clearDemoHistorical1Rm(): void {
  removeFromStorage(STORAGE_KEYS.HISTORICAL_1RM);
}

