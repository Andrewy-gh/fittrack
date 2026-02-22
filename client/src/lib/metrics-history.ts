import { parseISO, subDays, startOfISOWeek, startOfMonth } from 'date-fns';

import type { ExerciseExerciseWithSetsResponse } from '@/client';
import type { MetricsHistoryRange } from '@/lib/api/exercises';

type Bucket = 'workout' | 'week' | 'month';

export type MetricsHistoryPoint = {
  x: string;
  date: string; // ISO yyyy-MM-dd
  workout_id?: number;
  session_best_e1rm: number;
  session_avg_e1rm: number;
  session_avg_intensity: number;
  session_best_intensity: number;
  total_volume_working: number;
};

export function computeDemoMetricsHistory(
  exerciseSets: ExerciseExerciseWithSetsResponse[],
  range: MetricsHistoryRange
): { range: MetricsHistoryRange; bucket: Bucket; points: MetricsHistoryPoint[] } {
  const byWorkout = new Map<
    number,
    { date: string; sets: ExerciseExerciseWithSetsResponse[] }
  >();

  for (const s of exerciseSets) {
    const group = byWorkout.get(s.workout_id) ?? {
      date: s.workout_date.split('T')[0],
      sets: [],
    };
    group.sets.push(s);
    byWorkout.set(s.workout_id, group);
  }

  const workoutPoints: MetricsHistoryPoint[] = Array.from(byWorkout.entries())
    .map(([workout_id, g]) => {
      const working = g.sets.filter((s) => s.set_type === 'working');
      const e1rms = working.map((s) => {
        const w = s.weight ?? 0;
        return w * (1 + s.reps / 30);
      });
      const bestE1rm = e1rms.length ? Math.max(...e1rms) : 0;
      const avgE1rm =
        e1rms.length ? e1rms.reduce((a, b) => a + b, 0) / e1rms.length : 0;

      const intensities = working.map((s) => {
        const w = s.weight ?? 0;
        return bestE1rm > 0 ? (w / bestE1rm) * 100 : 0;
      });
      const avgIntensity =
        intensities.length
          ? intensities.reduce((a, b) => a + b, 0) / intensities.length
          : 0;
      const bestIntensity = intensities.length ? Math.max(...intensities) : 0;

      const totalVol = working.reduce((sum, s) => sum + (s.volume ?? 0), 0);

      return {
        x: String(workout_id),
        date: g.date,
        workout_id,
        session_best_e1rm: bestE1rm,
        session_avg_e1rm: avgE1rm,
        session_avg_intensity: avgIntensity,
        session_best_intensity: bestIntensity,
        total_volume_working: totalVol,
      };
    })
    .sort((a, b) => a.date.localeCompare(b.date));

  if (workoutPoints.length === 0) {
    return { range, bucket: 'workout', points: [] };
  }

  const endDate = parseISO(workoutPoints[workoutPoints.length - 1].date);

  const cutoffDays =
    range === 'W' ? 7 : range === 'M' ? 30 : range === '6M' ? 180 : 365;
  const startDate = subDays(endDate, cutoffDays);

  const filtered = workoutPoints.filter((p) => parseISO(p.date) >= startDate);

  if (range === 'W' || range === 'M') {
    return { range, bucket: 'workout', points: filtered };
  }

  if (range === '6M') {
    const buckets = new Map<string, MetricsHistoryPoint[]>();
    for (const p of filtered) {
      const key = startOfISOWeek(parseISO(p.date)).toISOString().split('T')[0];
      const arr = buckets.get(key) ?? [];
      arr.push(p);
      buckets.set(key, arr);
    }

    const points = Array.from(buckets.entries())
      .map(([date, pts]) => reduceBucket(date, pts))
      .sort((a, b) => a.date.localeCompare(b.date))
      .slice(-26);

    return { range, bucket: 'week', points };
  }

  // range === 'Y'
  const buckets = new Map<string, MetricsHistoryPoint[]>();
  for (const p of filtered) {
    const key = startOfMonth(parseISO(p.date)).toISOString().split('T')[0];
    const arr = buckets.get(key) ?? [];
    arr.push(p);
    buckets.set(key, arr);
  }

  const points = Array.from(buckets.entries())
    .map(([date, pts]) => reduceBucket(date, pts))
    .sort((a, b) => a.date.localeCompare(b.date))
    .slice(-12);

  return { range, bucket: 'month', points };
}

function reduceBucket(date: string, pts: MetricsHistoryPoint[]): MetricsHistoryPoint {
  const max = (xs: number[]) => (xs.length ? Math.max(...xs) : 0);
  const avg = (xs: number[]) =>
    xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : 0;

  return {
    x: date,
    date,
    session_best_e1rm: max(pts.map((p) => p.session_best_e1rm)),
    session_avg_e1rm: avg(pts.map((p) => p.session_avg_e1rm)),
    session_avg_intensity: avg(pts.map((p) => p.session_avg_intensity)),
    session_best_intensity: max(pts.map((p) => p.session_best_intensity)),
    total_volume_working: avg(pts.map((p) => p.total_volume_working)),
  };
}

