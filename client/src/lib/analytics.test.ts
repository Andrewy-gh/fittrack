import { describe, expect, it } from 'vitest';
import {
  buildDemoContributionData,
  buildWorkoutVolumeChartData,
  getWorkoutVolumeBucketLabel,
  getWorkoutVolumeTitle,
  getWorkoutSummary,
  type DemoContributionWorkout,
} from './analytics';

function workout(
  overrides: Partial<DemoContributionWorkout>
): DemoContributionWorkout {
  return {
    id: 0,
    date: '',
    workout_focus: null,
    volume: 0,
    workingSetCount: 0,
    ...overrides,
  };
}

describe('analytics', () => {
  it('builds sorted contribution data from demo workouts', () => {
    const data = buildDemoContributionData([
      workout({
        id: 2,
        date: '2026-03-12T08:00:00.000Z',
        workout_focus: 'Upper',
        volume: 3600,
        workingSetCount: 3,
      }),
      workout({
        id: 1,
        date: '2026-03-10T07:00:00.000Z',
        workout_focus: 'Lower',
        volume: 2800,
        workingSetCount: 2,
      }),
      workout({
        id: 3,
        date: '2026-03-12T17:00:00.000Z',
        workout_focus: 'Conditioning',
        volume: 1800,
        workingSetCount: 4,
      }),
    ]);

    expect(data).toEqual({
      days: [
        {
          date: '2026-03-10',
          count: 2,
          level: 1,
          workouts: [
            {
              id: 1,
              focus: 'Lower',
              time: '2026-03-10T07:00:00.000Z',
              volume: 2800,
            },
          ],
        },
        {
          date: '2026-03-12',
          count: 7,
          level: 2,
          workouts: [
            {
              id: 2,
              focus: 'Upper',
              time: '2026-03-12T08:00:00.000Z',
              volume: 3600,
            },
            {
              id: 3,
              focus: 'Conditioning',
              time: '2026-03-12T17:00:00.000Z',
              volume: 1800,
            },
          ],
        },
      ],
    });
  });

  it('builds workout volume buckets for focus-filtered chart ranges', () => {
    const data = buildDemoContributionData([
      workout({
        id: 1,
        date: '2026-03-21T08:00:00.000Z',
        workout_focus: 'Push',
        volume: 1800,
        workingSetCount: 3,
      }),
      workout({
        id: 2,
        date: '2026-03-22T08:00:00.000Z',
        workout_focus: 'Pull',
        volume: 2200,
        workingSetCount: 4,
      }),
      workout({
        id: 3,
        date: '2026-03-23T08:00:00.000Z',
        workout_focus: 'Push',
        volume: 2600,
        workingSetCount: 5,
      }),
    ]);

    expect(
      buildWorkoutVolumeChartData(
        data.days,
        'W',
        'Push',
        new Date('2026-03-23T12:00:00.000Z')
      )
    ).toEqual([
      { x: '2026-03-17', date: '2026-03-17', value: 0 },
      { x: '2026-03-18', date: '2026-03-18', value: 0 },
      { x: '2026-03-19', date: '2026-03-19', value: 0 },
      { x: '2026-03-20', date: '2026-03-20', value: 0 },
      { x: '2026-03-21', date: '2026-03-21', value: 1800 },
      { x: '2026-03-22', date: '2026-03-22', value: 0 },
      { x: '2026-03-23', date: '2026-03-23', value: 2600 },
    ]);
  });

  it('returns explicit bucket labels for each workout volume range', () => {
    expect(getWorkoutVolumeBucketLabel('W')).toBe(
      'Daily bars for the last 7 days'
    );
    expect(getWorkoutVolumeBucketLabel('M')).toBe(
      'Daily bars for the last 30 days'
    );
    expect(getWorkoutVolumeBucketLabel('6M')).toBe(
      'Weekly bars for the last 26 weeks'
    );
    expect(getWorkoutVolumeBucketLabel('Y')).toBe(
      'Monthly bars for the last 12 months'
    );
  });

  it('returns range-aware workout volume titles', () => {
    expect(getWorkoutVolumeTitle('W')).toBe('Daily Volume');
    expect(getWorkoutVolumeTitle('M', 'Push')).toBe('Daily Push Volume');
    expect(getWorkoutVolumeTitle('6M')).toBe('Weekly Volume');
    expect(getWorkoutVolumeTitle('Y', 'Pull')).toBe('Monthly Pull Volume');
  });

  it('computes analytics workout summaries from workout counts when workout metadata is present', () => {
    const summary = getWorkoutSummary(
      [
        {
          date: '2026-03-10',
          count: 9,
          level: 2,
          workouts: [
            {
              id: 1,
              time: '2026-03-10T08:00:00.000Z',
              volume: 1200,
            },
            {
              id: 2,
              time: '2026-03-10T17:00:00.000Z',
              volume: 1400,
            },
          ],
        },
        {
          date: '2026-03-11',
          count: 4,
          level: 1,
          workouts: [{ id: 3, time: '2026-03-11T08:00:00.000Z', volume: 1800 }],
        },
      ],
      new Date('2026-03-11T12:00:00.000Z')
    );

    expect(summary).toEqual({
      totalWorkouts30d: 3,
      avgWorkoutsPerWeek: 0.7,
      currentStreak: 2,
      longestStreak: 2,
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
