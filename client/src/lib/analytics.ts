import { addDays, differenceInCalendarDays, format } from 'date-fns';
import type {
  WorkoutContributionDataResponse,
} from '@/client';
import type { RangeType } from '@/components/charts/chart-bar-vol.utils';
import type { MetricPoint } from '@/components/charts/chart-bar-metric';

export interface AnalyticsWorkoutSummary {
  totalWorkouts30d: number;
  avgWorkoutsPerWeek: number;
  currentStreak: number;
  longestStreak: number;
}

export interface DemoContributionWorkout {
  id: number;
  date: string;
  workout_focus?: string | null;
  volume: number;
  workingSetCount: number;
}

export function buildDemoContributionData(
  workouts: DemoContributionWorkout[]
): WorkoutContributionDataResponse {
  const byDate = new Map<string, DemoContributionWorkout[]>();

  for (const workout of workouts) {
    const day = (workout.date || '').split('T')[0];
    if (!day) continue;

    const list = byDate.get(day) ?? [];
    list.push(workout);
    byDate.set(day, list);
  }

  const days = Array.from(byDate.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, items]) => {
      const count = items.reduce(
        (sum, workout) => sum + workout.workingSetCount,
        0
      );
      const level =
        count === 0 ? 0 : count < 6 ? 1 : count < 11 ? 2 : count < 16 ? 3 : 4;

      return {
        date,
        count,
        level,
        workouts: items.map((workout) => ({
          id: workout.id,
          focus: workout.workout_focus ?? undefined,
          time: workout.date,
          volume: workout.volume,
        })),
      };
    });

  return { days };
}

export function getWorkoutSummary(
  days: WorkoutContributionDataResponse['days'] = [],
  today: Date = new Date()
): AnalyticsWorkoutSummary {
  const start30 = addDays(today, -29);

  const normalized = (days ?? [])
    .filter((day): day is NonNullable<typeof day> => Boolean(day?.date))
    .map((day) => ({
      date: day.date!,
      count: day.workouts?.length ?? day.count ?? 0,
    }))
    .sort((a, b) => a.date.localeCompare(b.date));

  const countByDate = new Map(normalized.map((day) => [day.date, day.count]));

  let totalWorkouts30d = 0;
  for (let i = 0; i < 30; i++) {
    const day = format(addDays(start30, i), 'yyyy-MM-dd');
    totalWorkouts30d += countByDate.get(day) ?? 0;
  }

  const avgWorkoutsPerWeek = Number(((totalWorkouts30d / 30) * 7).toFixed(1));

  let currentStreak = 0;
  for (let i = 0; i < 3650; i++) {
    const day = format(addDays(today, -i), 'yyyy-MM-dd');
    if ((countByDate.get(day) ?? 0) > 0) {
      currentStreak += 1;
    } else {
      break;
    }
  }

  let longestStreak = 0;
  let runningStreak = 0;
  let previousDate: Date | null = null;

  for (const item of normalized) {
    if (item.count <= 0) continue;

    const currentDate = new Date(`${item.date}T00:00:00`);
    if (!previousDate) {
      runningStreak = 1;
    } else {
      const diff = differenceInCalendarDays(currentDate, previousDate);
      runningStreak = diff === 1 ? runningStreak + 1 : 1;
    }

    previousDate = currentDate;
    longestStreak = Math.max(longestStreak, runningStreak);
  }

  return {
    totalWorkouts30d,
    avgWorkoutsPerWeek,
    currentStreak,
    longestStreak,
  };
}

function toIsoDate(date: Date) {
  return format(date, 'yyyy-MM-dd');
}

function startOfWeek(date: Date) {
  const result = new Date(date);
  const day = result.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  result.setDate(result.getDate() + diff);
  result.setHours(0, 0, 0, 0);
  return result;
}

function startOfMonth(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function addMonths(date: Date, amount: number) {
  return new Date(date.getFullYear(), date.getMonth() + amount, 1);
}

function sumVolumeForDay(
  day: NonNullable<WorkoutContributionDataResponse['days']>[number],
  focus?: string
) {
  return (day.workouts ?? []).reduce((sum, workout) => {
    if (focus && workout.focus !== focus) {
      return sum;
    }
    return sum + (workout.volume ?? 0);
  }, 0);
}

function buildDailyVolumeMap(
  days: WorkoutContributionDataResponse['days'] = [],
  focus?: string
) {
  const volumeByDate = new Map<string, number>();

  for (const day of days ?? []) {
    if (!day?.date) continue;
    volumeByDate.set(day.date, sumVolumeForDay(day, focus));
  }

  return volumeByDate;
}

export function buildWorkoutVolumeChartData(
  days: WorkoutContributionDataResponse['days'] = [],
  range: RangeType,
  focus?: string,
  today: Date = new Date()
): MetricPoint[] {
  const volumeByDate = buildDailyVolumeMap(days, focus);

  if (range === 'W' || range === 'M') {
    const span = range === 'W' ? 7 : 30;
    const start = addDays(today, -(span - 1));

    return Array.from({ length: span }, (_, index) => {
      const date = addDays(start, index);
      const isoDate = toIsoDate(date);
      return {
        x: isoDate,
        date: isoDate,
        value: Math.round(volumeByDate.get(isoDate) ?? 0),
      };
    });
  }

  if (range === '6M') {
    const currentWeekStart = startOfWeek(today);
    const firstWeekStart = addDays(currentWeekStart, -(25 * 7));

    return Array.from({ length: 26 }, (_, index) => {
      const weekStart = addDays(firstWeekStart, index * 7);
      let total = 0;

      for (let dayOffset = 0; dayOffset < 7; dayOffset += 1) {
        const day = addDays(weekStart, dayOffset);
        total += volumeByDate.get(toIsoDate(day)) ?? 0;
      }

      const isoDate = toIsoDate(weekStart);
      return {
        x: isoDate,
        date: isoDate,
        value: Math.round(total),
      };
    });
  }

  const currentMonthStart = startOfMonth(today);
  const firstMonthStart = addMonths(currentMonthStart, -11);

  return Array.from({ length: 12 }, (_, index) => {
    const monthStart = addMonths(firstMonthStart, index);
    const nextMonthStart = addMonths(monthStart, 1);
    let total = 0;

    for (
      let cursor = new Date(monthStart);
      cursor < nextMonthStart;
      cursor = addDays(cursor, 1)
    ) {
      total += volumeByDate.get(toIsoDate(cursor)) ?? 0;
    }

    const isoDate = toIsoDate(monthStart);
    return {
      x: isoDate,
      date: isoDate,
      value: Math.round(total),
    };
  });
}

export function getWorkoutVolumeBucketLabel(range: RangeType): string {
  switch (range) {
    case 'W':
      return 'Daily bars for the last 7 days';
    case 'M':
      return 'Daily bars for the last 30 days';
    case '6M':
      return 'Weekly bars for the last 26 weeks';
    case 'Y':
      return 'Monthly bars for the last 12 months';
    default:
      return 'Volume by time period';
  }
}

export function getWorkoutVolumeTitle(
  range: RangeType,
  focus?: string
): string {
  const period =
    range === '6M' ? 'Weekly' : range === 'Y' ? 'Monthly' : 'Daily';

  if (focus) {
    return `${period} ${focus} Volume`;
  }

  return `${period} Volume`;
}
