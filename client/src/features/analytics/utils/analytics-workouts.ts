import { addDays, differenceInCalendarDays, format } from "date-fns";
import type { WorkoutContributionDataResponse } from "@/client";
import {
  getVisibleBarCount,
  type RangeType,
} from "@/components/charts/chart-bar-vol.utils";
import type { MetricPoint } from "@/components/charts/chart-bar-metric";
import { buildHistoricalRangeTimeline } from "@/components/charts/historical-range-timeline";

export interface AnalyticsWorkoutSummary {
  totalWorkouts30d: number;
  avgWorkoutsPerWeek: number;
  currentStreak: number;
  longestStreak: number;
}

export function getWorkoutSummary(
  days: WorkoutContributionDataResponse["days"] = [],
  today: Date = new Date(),
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
    const day = format(addDays(start30, i), "yyyy-MM-dd");
    totalWorkouts30d += countByDate.get(day) ?? 0;
  }

  const avgWorkoutsPerWeek = Number(((totalWorkouts30d / 30) * 7).toFixed(1));

  let currentStreak = 0;
  for (let i = 0; i < 3650; i++) {
    const day = format(addDays(today, -i), "yyyy-MM-dd");
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

type ContributionDay = NonNullable<
  WorkoutContributionDataResponse["days"]
>[number];

function sumVolumeForDay(day: ContributionDay, focus?: string) {
  return (day.workouts ?? []).reduce((sum, workout) => {
    if (focus && workout.focus !== focus) {
      return sum;
    }
    return sum + (workout.volume ?? 0);
  }, 0);
}

function getFocusTypesForDay(day: ContributionDay) {
  const focusTypes = new Set<string>();

  for (const workout of day.workouts ?? []) {
    if (workout.focus) {
      focusTypes.add(workout.focus);
    }
  }

  return Array.from(focusTypes);
}

function formatFocusTypes(focusTypes: Set<string>) {
  return Array.from(focusTypes)
    .sort((a, b) => a.localeCompare(b))
    .join(", ");
}

export function buildWorkoutVolumeChartData(
  days: WorkoutContributionDataResponse["days"] = [],
  range: RangeType,
  focus?: string,
  today: Date = new Date(),
): MetricPoint[] {
  const workoutDays = (days ?? []).filter(
    (day) => day.date && (day.workouts?.length ?? 0) > 0,
  );
  const timeline = buildHistoricalRangeTimeline({
    range,
    items: workoutDays,
    dateOf: (day) => day.date,
    today,
  });

  return timeline.buckets.map(({ date, items }) => {
    const volume = items.reduce(
      (total, day) => total + sumVolumeForDay(day, focus),
      0,
    );
    let focusType: string | undefined;

    if (!focus && timeline.bucket === "day" && items.length > 0) {
      const focusTypes = new Set<string>();
      for (const day of items) {
        for (const dayFocus of getFocusTypesForDay(day)) {
          focusTypes.add(dayFocus);
        }
      }
      focusType = formatFocusTypes(focusTypes);
    }

    return {
      x: date,
      date,
      focusType,
      value: Math.round(volume),
    };
  });
}

export function getWorkoutVolumeBucketLabel(range: RangeType): string {
  const visibleBarCount = getVisibleBarCount(range);

  switch (range) {
    case "W":
    case "M":
      return `Daily bars from your first workout, ${visibleBarCount} visible at a time`;
    case "6M":
      return `Weekly bars from your first workout, ${visibleBarCount} visible at a time`;
    case "Y":
      return "Monthly bars from your first workout";
    default:
      return "Volume by time period";
  }
}

export function getWorkoutVolumeTitle(
  range: RangeType,
  focus?: string,
): string {
  const period =
    range === "6M" ? "Weekly" : range === "Y" ? "Monthly" : "Daily";

  if (focus) {
    return `${period} ${focus} Volume`;
  }

  return `${period} Volume`;
}
