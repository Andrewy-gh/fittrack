import { describe, it, expect } from 'vitest';
import { updateExercise, createExercise, getAllExercises } from './storage';

describe('updateExercise', () => {

  it('should successfully update exercise name', () => {
    // Create an exercise
    const exercise = createExercise('Bench Press');

    // Small delay to ensure timestamp will be different
    const startTime = new Date().getTime();
    while (new Date().getTime() - startTime < 5) {
      // Wait 5ms
    }

    // Update the exercise
    const success = updateExercise(exercise.id, 'Incline Bench Press');

    expect(success).toBe(true);

    // Verify the update
    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated).toBeDefined();
    expect(updated?.name).toBe('Incline Bench Press');
    expect(updated?.updated_at).not.toBe(exercise.updated_at);
  });

  it('should throw error when duplicate name exists', () => {
    // Create two exercises
    const exercise1 = createExercise('Squat');
    createExercise('Deadlift');

    // Try to update exercise1 to have the same name as exercise2
    expect(() => {
      updateExercise(exercise1.id, 'Deadlift');
    }).toThrow('Exercise name "Deadlift" already exists');
  });

  it('should throw error when duplicate name exists (case-insensitive)', () => {
    // Create two exercises
    const exercise1 = createExercise('Squat');
    createExercise('Deadlift');

    // Try to update exercise1 to have the same name (different case)
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

    // Update to the same name (should succeed)
    const success = updateExercise(exercise.id, 'Bench Press');
    expect(success).toBe(true);

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);
    expect(updated?.name).toBe('Bench Press');
  });

  it('should update updated_at timestamp', () => {
    const exercise = createExercise('Bench Press');
    const oldUpdatedAt = exercise.updated_at;

    // Small delay to ensure timestamp changes
    const startTime = new Date().getTime();
    while (new Date().getTime() - startTime < 10) {
      // Wait 10ms
    }

    updateExercise(exercise.id, 'New Name');

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated?.updated_at).not.toBe(oldUpdatedAt);
  });
});
