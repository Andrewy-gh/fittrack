import { afterEach, describe, expect, it } from 'vitest';
import {
  clearExerciseGoals,
  formatExerciseGoalSummary,
  getExerciseGoal,
  parseExerciseGoalInput,
  saveExerciseGoal,
} from './exercise-goals';

describe('exercise-goals', () => {
  afterEach(() => {
    clearExerciseGoals();
  });

  it('stores and reads goals by exercise id', () => {
    saveExerciseGoal(
      { exerciseId: 12, exerciseName: 'Bench Press' },
      {
        targetWeight: 205,
        targetReps: 8,
      }
    );

    expect(getExerciseGoal({ exerciseId: 12 })).toEqual({
      targetWeight: 205,
      targetReps: 8,
    });
  });

  it('falls back to normalized exercise name when no id exists', () => {
    saveExerciseGoal(
      { exerciseName: ' Romanian Deadlift ' },
      {
        frequencyPerWeek: 2,
      }
    );

    expect(
      getExerciseGoal({ exerciseName: 'romanian deadlift' })
    ).toEqual({
      frequencyPerWeek: 2,
    });
  });

  it('removes goals when all values are cleared', () => {
    saveExerciseGoal({ exerciseId: 8 }, { targetWeight: 135 });
    saveExerciseGoal({ exerciseId: 8 }, {});

    expect(getExerciseGoal({ exerciseId: 8 })).toBeNull();
  });

  it('formats a readable goal summary', () => {
    expect(
      formatExerciseGoalSummary({
        targetWeight: 225,
        targetReps: 5,
        frequencyPerWeek: 2,
      })
    ).toBe('225 lb • 5 reps • 2x / week');
  });

  it('rejects non-numeric goal input', () => {
    expect(parseExerciseGoalInput('225lb', 'Target Weight')).toEqual({
      error: 'Target Weight must be a number',
    });
  });

  it('rejects negative goal input', () => {
    expect(parseExerciseGoalInput('-5', 'Target Reps', { integer: true })).toEqual({
      error: 'Target Reps must be 0 or greater',
    });
  });

  it('rejects non-integer values for integer-only fields', () => {
    expect(
      parseExerciseGoalInput('2.5', 'Weekly Frequency', { integer: true })
    ).toEqual({
      error: 'Weekly Frequency must be a whole number',
    });
  });
});
