import { describe, expect, it } from 'vitest';

import type { ExerciseExerciseWithSetsResponse } from '@/client';

import { computeDemoMetricsHistory } from './metrics-history';

function set(
  overrides: Partial<ExerciseExerciseWithSetsResponse>
): ExerciseExerciseWithSetsResponse {
  return {
    workout_id: 0,
    workout_date: '2026-03-01T08:00:00.000Z',
    set_type: 'working',
    weight: 100,
    reps: 5,
    volume: 500,
    ...overrides,
  } as ExerciseExerciseWithSetsResponse;
}

describe('computeDemoMetricsHistory', () => {
  it('keeps same-day workouts as separate session points for yearly range', () => {
    const history = computeDemoMetricsHistory(
      [
        set({
          workout_id: 11,
          workout_date: '2026-03-24T08:00:00.000Z',
          weight: 100,
          reps: 5,
          volume: 500,
        }),
        set({
          workout_id: 22,
          workout_date: '2026-03-24T18:00:00.000Z',
          weight: 135,
          reps: 3,
          volume: 405,
        }),
      ],
      'Y'
    );

    expect(history.bucket).toBe('workout');
    expect(history.points).toHaveLength(2);
    expect(history.points.map((point) => point.workout_id)).toEqual([11, 22]);
    expect(history.points.map((point) => point.date)).toEqual([
      '2026-03-24',
      '2026-03-24',
    ]);
  });
});
