import { afterEach, describe, expect, it, vi } from 'vitest';
import { updateExercise, createExercise, getAllExercises } from './storage';

afterEach(() => {
  vi.useRealTimers();
});

describe('updateExercise', () => {
  it('should successfully update exercise name', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-15T09:00:00.000Z'));
    const exercise = createExercise('Bench Press');
    vi.setSystemTime(new Date('2026-03-15T09:00:01.000Z'));
    const success = updateExercise(exercise.id, 'Incline Bench Press');

    expect(success).toBe(true);

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated).toBeDefined();
    expect(updated?.name).toBe('Incline Bench Press');
    expect(updated?.updated_at).not.toBe(exercise.updated_at);
  });

  it('should throw error when duplicate name exists', () => {
    const exercise1 = createExercise('Squat');
    createExercise('Deadlift');

    expect(() => {
      updateExercise(exercise1.id, 'Deadlift');
    }).toThrow('Exercise name "Deadlift" already exists');
  });

  it('should throw error when duplicate name exists (case-insensitive)', () => {
    const exercise1 = createExercise('Squat');
    createExercise('Deadlift');

    expect(() => {
      updateExercise(exercise1.id, 'deadlift');
    }).toThrow('Exercise name "deadlift" already exists');
  });

  it('should return false when exercise not found', () => {
    const success = updateExercise(999, 'Non-existent Exercise');
    expect(success).toBe(false);
  });

  it('should allow updating to the same name', () => {
    const exercise = createExercise('Bench Press');

    const success = updateExercise(exercise.id, 'Bench Press');
    expect(success).toBe(true);

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);
    expect(updated?.name).toBe('Bench Press');
  });

  it('should update updated_at timestamp', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-15T10:00:00.000Z'));
    const exercise = createExercise('Bench Press');
    const oldUpdatedAt = exercise.updated_at;
    vi.setSystemTime(new Date('2026-03-15T10:00:01.000Z'));
    updateExercise(exercise.id, 'New Name');

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated?.updated_at).not.toBe(oldUpdatedAt);
  });
});
