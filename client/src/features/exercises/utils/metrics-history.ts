import type { ExerciseExerciseWithSetsResponse } from "@/client";
type Bucket = "workout";

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

type MetricsHistoryResult = {
  bucket: Bucket;
  points: MetricsHistoryPoint[];
};

export function computeDemoMetricsHistory(
  exerciseSets: ExerciseExerciseWithSetsResponse[],
): MetricsHistoryResult {
  const byWorkout = new Map<
    number,
    { date: string; sets: ExerciseExerciseWithSetsResponse[] }
  >();

  for (const s of exerciseSets) {
    const group = byWorkout.get(s.workout_id) ?? {
      date: s.workout_date.split("T")[0],
      sets: [],
    };
    group.sets.push(s);
    byWorkout.set(s.workout_id, group);
  }

  const workoutPoints: MetricsHistoryPoint[] = Array.from(byWorkout.entries())
    .map(([workout_id, g]) => {
      const working = g.sets.filter((s) => s.set_type === "working");
      const e1rms = working.map((s) => {
        const w = s.weight ?? 0;
        return w * (1 + s.reps / 30);
      });
      const bestE1rm = e1rms.length ? Math.max(...e1rms) : 0;
      const avgE1rm = e1rms.length
        ? e1rms.reduce((a, b) => a + b, 0) / e1rms.length
        : 0;

      const intensities = working.map((s) => {
        const w = s.weight ?? 0;
        return bestE1rm > 0 ? (w / bestE1rm) * 100 : 0;
      });
      const avgIntensity = intensities.length
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
    return { bucket: "workout", points: [] };
  }

  return { bucket: "workout", points: workoutPoints };
}
