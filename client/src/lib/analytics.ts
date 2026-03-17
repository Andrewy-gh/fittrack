import { addDays, differenceInCalendarDays, format } from 'date-fns';
import type {
  WorkoutContributionDataResponse,
  WorkoutWorkoutResponse,
} from '@/client';

export interface AnalyticsWorkoutSummary {
  totalWorkouts30d: number;
  avgWorkoutsPerWeek: number;
  currentStreak: number;
  longestStreak: number;
}

export function buildDemoContributionData(
  workouts: WorkoutWorkoutResponse[]
): WorkoutContributionDataResponse {
  const byDate = new Map<string, WorkoutWorkoutResponse[]>();

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
      const count = items.length;
      const level =
        count === 0 ? 0 : count < 2 ? 1 : count < 3 ? 2 : count < 4 ? 3 : 4;

      return {
        date,
        count,
        level,
        workouts: items.map((workout) => ({
          id: workout.id,
          focus: workout.workout_focus,
          time: workout.date,
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
      count: day.count ?? 0,
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
