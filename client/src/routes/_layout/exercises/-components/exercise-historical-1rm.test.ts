import { describe, expect, it } from 'vitest';
import type { ExerciseExerciseWithSetsResponse } from '@/client';
import { resolveBestE1rmForDisplay } from './exercise-historical-1rm';

function makeSet(overrides: Partial<ExerciseExerciseWithSetsResponse>): ExerciseExerciseWithSetsResponse {
  return {
    set_id: 1,
    exercise_id: 1,
    exercise_name: 'Bench Press',
    workout_id: 10,
    workout_date: '2026-02-11T00:00:00Z',
    reps: 5,
    set_type: 'working',
    volume: 0,
    ...overrides,
  };
}

describe('resolveBestE1rmForDisplay', () => {
  it('uses API best_e1rm in authenticated mode without set recompute', () => {
    const sets: ExerciseExerciseWithSetsResponse[] = [
      makeSet({ weight: 100, reps: 5, workout_id: 11 }),
      makeSet({ weight: 200, reps: 5, workout_id: 12 }),
    ];

    const result = resolveBestE1rmForDisplay({
      isDemoMode: false,
      apiBestE1RM: 333.3,
      exerciseSets: sets,
    });

    expect(result.best).toBe(333.3);
    expect(result.workoutId).toBeNull();
  });

  it('falls back to set-derived best when API best_e1rm is missing', () => {
    const sets: ExerciseExerciseWithSetsResponse[] = [
      makeSet({ weight: 185, reps: 5, workout_id: 21 }),
      makeSet({ weight: 200, reps: 3, workout_id: 22 }),
    ];

    const result = resolveBestE1rmForDisplay({
      isDemoMode: false,
      apiBestE1RM: null,
      exerciseSets: sets,
    });

    expect(result.best).toBeCloseTo(220);
    expect(result.workoutId).toBe(22);
  });

  it('uses set-derived best in demo mode', () => {
    const sets: ExerciseExerciseWithSetsResponse[] = [
      makeSet({ weight: 0, reps: 8, workout_id: 31 }),
      makeSet({ weight: 95, reps: 8, workout_id: 32 }),
      makeSet({ set_type: 'warmup', weight: 225, reps: 1, workout_id: 33 }),
    ];

    const result = resolveBestE1rmForDisplay({
      isDemoMode: true,
      apiBestE1RM: undefined,
      exerciseSets: sets,
    });

    expect(result.best).toBeCloseTo(120.333333, 5);
    expect(result.workoutId).toBe(32);
  });
});
