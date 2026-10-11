import type { RangeType } from "@/components/charts/chart-bar-vol.utils";
import type { MetricPoint } from "@/components/charts/chart-bar-metric";
import { buildHistoricalRangeTimeline } from "@/components/charts/historical-range-timeline";

export type SessionMetricAggregation = "average" | "maximum" | "sum";

type SessionMetricPoint = {
  date?: string | null;
  workout_id?: number;
};

export function getSessionMetricBucket(range: RangeType) {
  return range === "Y" ? "month" : range === "6M" ? "week" : "day";
}

export function buildSessionMetricChartData<T extends SessionMetricPoint>(
  points: T[],
  range: RangeType,
  pick: (point: T) => number,
  aggregation: SessionMetricAggregation,
  today: Date = new Date(),
): MetricPoint[] {
  const timeline = buildHistoricalRangeTimeline({
    range,
    items: points,
    dateOf: (point) => point.date,
    today,
  });

  return timeline.buckets.map(({ date, items }) => {
    const values = items.map(pick);
    const value =
      values.length === 0
        ? 0
        : aggregation === "sum"
          ? values.reduce((sum, current) => sum + current, 0)
          : aggregation === "maximum"
            ? Math.max(...values)
            : values.reduce((sum, current) => sum + current, 0) / values.length;

    return {
      x: date,
      date,
      value,
      workout_id: items.length === 1 ? items[0].workout_id : undefined,
    };
  });
}
