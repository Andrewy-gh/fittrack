import { describe, expect, it } from 'vitest';
import type { WorkoutWorkoutResponse } from '@/client';
import { buildDemoContributionData, getWorkoutSummary } from './analytics';

function workout(
  overrides: Partial<WorkoutWorkoutResponse>
): WorkoutWorkoutResponse {
  return {
    id: 0,
    date: '',
    workout_focus: null,
    ...overrides,
  } as WorkoutWorkoutResponse;
}

describe('analytics', () => {
  it('builds sorted contribution data from demo workouts', () => {
    const data = buildDemoContributionData([
      workout({
        id: 2,
        date: '2026-03-12T08:00:00.000Z',
        workout_focus: 'Upper',
      }),
      workout({
        id: 1,
        date: '2026-03-10T07:00:00.000Z',
        workout_focus: 'Lower',
      }),
      workout({
        id: 3,
        date: '2026-03-12T17:00:00.000Z',
        workout_focus: 'Conditioning',
      }),
    ]);

    expect(data).toEqual({
      days: [
        {
          date: '2026-03-10',
          count: 1,
          level: 1,
          workouts: [
            {
              id: 1,
              focus: 'Lower',
              time: '2026-03-10T07:00:00.000Z',
            },
          ],
        },
        {
          date: '2026-03-12',
          count: 2,
          level: 2,
          workouts: [
            {
              id: 2,
              focus: 'Upper',
              time: '2026-03-12T08:00:00.000Z',
            },
            {
              id: 3,
              focus: 'Conditioning',
              time: '2026-03-12T17:00:00.000Z',
            },
          ],
        },
      ],
    });
  });

  it('computes analytics workout summaries with streaks', () => {
    const summary = getWorkoutSummary(
      [
        { date: '2026-03-05', count: 1 },
        { date: '2026-03-08', count: 1 },
        { date: '2026-03-10', count: 1 },
        { date: '2026-03-11', count: 2 },
        { date: '2026-03-12', count: 1 },
      ],
      new Date('2026-03-12T12:00:00.000Z')
    );

    expect(summary).toEqual({
      totalWorkouts30d: 6,
      avgWorkoutsPerWeek: 1.4,
      currentStreak: 3,
      longestStreak: 3,
    });
  });
});
