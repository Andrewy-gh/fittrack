import type { WorkoutContributionDataResponse } from "@/client";

export interface DemoContributionWorkout {
  id: number;
  date: string;
  workout_focus?: string | null;
  volume: number;
  workingSetCount: number;
}

export function buildDemoContributionData(
  workouts: DemoContributionWorkout[],
): WorkoutContributionDataResponse {
  const byDate = new Map<string, DemoContributionWorkout[]>();

  for (const workout of workouts) {
    const day = (workout.date || "").split("T")[0];
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
        0,
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
