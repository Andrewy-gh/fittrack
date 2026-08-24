import { addDays, differenceInCalendarDays, format } from "date-fns";
import type { WorkoutContributionDataResponse } from "@/client";
import type { RangeType } from "@/components/charts/chart-bar-vol.utils";
import type { MetricPoint } from "@/components/charts/chart-bar-metric";

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

function toIsoDate(date: Date) {
  return format(date, "yyyy-MM-dd");
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
  day: NonNullable<WorkoutContributionDataResponse["days"]>[number],
  focus?: string,
) {
  return (day.workouts ?? []).reduce((sum, workout) => {
    if (focus && workout.focus !== focus) {
      return sum;
    }
    return sum + (workout.volume ?? 0);
  }, 0);
}

function getFocusTypesForDay(
  day: NonNullable<WorkoutContributionDataResponse["days"]>[number],
) {
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

function buildDailyVolumeMap(
  days: WorkoutContributionDataResponse["days"] = [],
  focus?: string,
) {
  const volumeByDate = new Map<
    string,
    { focusType?: string; focusTypes: string[]; volume: number }
  >();

  for (const day of days ?? []) {
    if (!day?.date) continue;
    const volume = sumVolumeForDay(day, focus);
    const focusTypes = focus ? [focus] : getFocusTypesForDay(day);
    volumeByDate.set(day.date, {
      focusType: formatFocusTypes(new Set(focusTypes)),
      focusTypes,
      volume,
    });
  }

  return volumeByDate;
}

export function buildWorkoutVolumeChartData(
  days: WorkoutContributionDataResponse["days"] = [],
  range: RangeType,
  focus?: string,
  today: Date = new Date(),
): MetricPoint[] {
  const volumeByDate = buildDailyVolumeMap(days, focus);
  const firstWorkoutDate = (days ?? [])
    .flatMap((day) =>
      day?.date && (day.workouts?.length ?? 0) > 0 ? [day.date] : [],
    )
    .sort()[0];

  if (range === "W" || range === "M") {
    const span = range === "W" ? 7 : 30;
    const defaultStart = addDays(today, -(span - 1));
    const historyStart = firstWorkoutDate
      ? new Date(`${firstWorkoutDate}T00:00:00`)
      : defaultStart;
    const start = historyStart < defaultStart ? historyStart : defaultStart;
    const dayCount = differenceInCalendarDays(today, start) + 1;

    const points = Array.from({ length: dayCount }, (_, index) => {
      const date = addDays(start, index);
      const isoDate = toIsoDate(date);
      const dayVolume = volumeByDate.get(isoDate);
      return {
        x: isoDate,
        date: isoDate,
        focusType: focus ? undefined : dayVolume?.focusType,
        value: Math.round(dayVolume?.volume ?? 0),
      };
    });

    return points;
  }

  if (range === "6M") {
    const currentWeekStart = startOfWeek(today);
    const defaultFirstWeekStart = addDays(currentWeekStart, -(25 * 7));
    const historyWeekStart = firstWorkoutDate
      ? startOfWeek(new Date(`${firstWorkoutDate}T00:00:00`))
      : defaultFirstWeekStart;
    const firstWeekStart =
      historyWeekStart < defaultFirstWeekStart
        ? historyWeekStart
        : defaultFirstWeekStart;
    const weekCount =
      Math.floor(
        differenceInCalendarDays(currentWeekStart, firstWeekStart) / 7,
      ) + 1;

    return Array.from({ length: weekCount }, (_, index) => {
      const weekStart = addDays(firstWeekStart, index * 7);
      let total = 0;

      for (let dayOffset = 0; dayOffset < 7; dayOffset += 1) {
        const day = addDays(weekStart, dayOffset);
        const dayVolume = volumeByDate.get(toIsoDate(day));
        total += dayVolume?.volume ?? 0;
      }

      const isoDate = toIsoDate(weekStart);
      return {
        x: isoDate,
        date: isoDate,
        focusType: undefined,
        value: Math.round(total),
      };
    });
  }

  const currentMonthStart = startOfMonth(today);
  const firstMonthStart = firstWorkoutDate
    ? startOfMonth(new Date(`${firstWorkoutDate}T00:00:00`))
    : currentMonthStart;
  const monthCount =
    (currentMonthStart.getFullYear() - firstMonthStart.getFullYear()) * 12 +
    currentMonthStart.getMonth() -
    firstMonthStart.getMonth() +
    1;

  return Array.from({ length: monthCount }, (_, index) => {
    const monthStart = addMonths(firstMonthStart, index);
    const nextMonthStart = addMonths(monthStart, 1);
    let total = 0;

    for (
      let cursor = new Date(monthStart);
      cursor < nextMonthStart;
      cursor = addDays(cursor, 1)
    ) {
      const dayVolume = volumeByDate.get(toIsoDate(cursor));
      total += dayVolume?.volume ?? 0;
    }

    const isoDate = toIsoDate(monthStart);
    return {
      x: isoDate,
      date: isoDate,
      focusType: undefined,
      value: Math.round(total),
    };
  });
}

export function getWorkoutVolumeBucketLabel(range: RangeType): string {
  switch (range) {
    case "W":
      return "Daily bars from your first workout, 7 visible at a time";
    case "M":
      return "Daily bars from your first workout, 30 visible at a time";
    case "6M":
      return "Weekly bars from your first workout, 26 visible at a time";
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
