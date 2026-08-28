import {
  addDays,
  addMonths,
  differenceInCalendarDays,
  format,
  startOfMonth,
  startOfWeek,
} from "date-fns";

import type { RangeType } from "@/components/charts/chart-bar-vol.utils";
import type { MetricPoint } from "@/components/charts/chart-bar-metric";

export type SessionMetricAggregation = "average" | "maximum" | "sum";

type SessionMetricPoint = {
  date?: string | null;
  workout_id?: number;
};

const toIsoDate = (date: Date) => format(date, "yyyy-MM-dd");
const monday = (date: Date) => startOfWeek(date, { weekStartsOn: 1 });

export function getSessionMetricBucket(range: RangeType) {
  return range === "Y" ? "month" : range === "6M" ? "week" : "day";
}

export function getSessionMetricPeriodLabel(range: RangeType): string {
  return range === "Y" ? "Monthly" : range === "6M" ? "Weekly" : "Daily";
}

export function getSessionMetricBucketLabel(range: RangeType): string {
  if (range === "Y") return "Monthly bars for the last 12 months";
  if (range === "6M") return "Weekly bars for the last 26 weeks";
  return range === "M"
    ? "Daily bars for the last 30 days, including days without a session"
    : "Daily bars for the last 7 days, including days without a session";
}

export function buildSessionMetricChartData<T extends SessionMetricPoint>(
  points: T[],
  range: RangeType,
  pick: (point: T) => number,
  aggregation: SessionMetricAggregation,
  today: Date = new Date(),
): MetricPoint[] {
  const valuesByBucket = new Map<
    string,
    Array<{ value: number; workout_id?: number }>
  >();
  const todayStart = new Date(
    today.getFullYear(),
    today.getMonth(),
    today.getDate(),
  );

  for (const point of points) {
    if (!point.date) continue;
    const date = new Date(`${point.date.split("T")[0]}T00:00:00`);
    if (date > todayStart) continue;
    const bucketDate =
      range === "Y" ? startOfMonth(date) : range === "6M" ? monday(date) : date;
    const key = toIsoDate(bucketDate);
    const values = valuesByBucket.get(key) ?? [];
    values.push({ value: pick(point) ?? 0, workout_id: point.workout_id });
    valuesByBucket.set(key, values);
  }

  const end =
    range === "Y"
      ? startOfMonth(todayStart)
      : range === "6M"
        ? monday(todayStart)
        : todayStart;
  const start =
    range === "Y"
      ? addMonths(end, -11)
      : range === "6M"
        ? addDays(end, -25 * 7)
        : addDays(end, range === "M" ? -29 : -6);
  const count =
    range === "Y"
      ? 12
      : range === "6M"
        ? 26
        : differenceInCalendarDays(end, start) + 1;

  return Array.from({ length: count }, (_, index) => {
    const date =
      range === "Y"
        ? addMonths(start, index)
        : addDays(start, index * (range === "6M" ? 7 : 1));
    const isoDate = toIsoDate(date);
    const entries = valuesByBucket.get(isoDate) ?? [];
    const rawValues = entries.map((entry) => entry.value);
    const value =
      rawValues.length === 0
        ? 0
        : aggregation === "sum"
          ? rawValues.reduce((sum, current) => sum + current, 0)
          : aggregation === "maximum"
            ? Math.max(...rawValues)
            : rawValues.reduce((sum, current) => sum + current, 0) /
              rawValues.length;

    return {
      x: isoDate,
      date: isoDate,
      value,
      workout_id: entries.length === 1 ? entries[0].workout_id : undefined,
    };
  });
}
