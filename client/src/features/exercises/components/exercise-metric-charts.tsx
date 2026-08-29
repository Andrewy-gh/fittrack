import { useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";

import type {
  ExerciseExerciseMetricsHistoryPoint,
  ExerciseExerciseWithSetsResponse,
} from "@/client";
import { exerciseMetricsHistoryQueryOptions } from "@/features/exercises/api/exercises";
import { computeDemoMetricsHistory } from "@/features/exercises/utils/metrics-history";
import {
  buildSessionMetricChartData,
  getSessionMetricBucket,
  getSessionMetricBucketLabel,
  getSessionMetricPeriodLabel,
  type SessionMetricAggregation,
} from "@/features/exercises/utils/session-metric-buckets";
import { RangeSelector } from "@/components/charts/chart-bar-vol.components";
import type { RangeType } from "@/components/charts/chart-bar-vol.utils";
import {
  ChartBarMetric,
  type MetricPoint,
} from "@/components/charts/chart-bar-metric";
import { Spinner } from "@/components/ui/spinner";

function hasWeightedMetrics(
  points: ExerciseExerciseMetricsHistoryPoint[],
): boolean {
  return points.some(
    (point) =>
      (point.session_best_e1rm ?? 0) > 0 ||
      (point.session_avg_e1rm ?? 0) > 0 ||
      (point.session_avg_intensity ?? 0) > 0 ||
      (point.session_best_intensity ?? 0) > 0 ||
      (point.total_volume_working ?? 0) > 0,
  );
}

export function ExerciseMetricCharts({
  exerciseId,
  exerciseSets,
  isDemoMode,
}: {
  exerciseId: number;
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  isDemoMode: boolean;
}) {
  const router = useRouter();
  const [selectedRange, setSelectedRange] = useState<RangeType>("M");

  const onWorkoutClick = (workoutId: number) =>
    router.navigate({ to: "/workouts/$workoutId", params: { workoutId } });

  return (
    <section className="space-y-6">
      <div className="space-y-3">
        <div>
          <h2 className="text-xl font-semibold">Session Metrics</h2>
          <p className="text-sm text-muted-foreground">
            Bars group exercise sessions by the selected time range. e1RM,
            intensity, and volume are computed from working sets. Intensity can
            exceed 100%.
          </p>
        </div>

        <div className="flex justify-center">
          <RangeSelector
            selectedRange={selectedRange}
            onRangeChange={setSelectedRange}
          />
        </div>
      </div>

      {isDemoMode ? (
        <DemoCharts
          exerciseSets={exerciseSets}
          range={selectedRange}
          onWorkoutClick={onWorkoutClick}
        />
      ) : (
        <AuthedCharts
          exerciseId={exerciseId}
          range={selectedRange}
          onWorkoutClick={onWorkoutClick}
        />
      )}
    </section>
  );
}

function MetricChartsBody({
  points,
  range,
  onWorkoutClick,
  statusMessage,
}: {
  points: ExerciseExerciseMetricsHistoryPoint[];
  range: RangeType;
  onWorkoutClick: (workoutId: number) => void;
  statusMessage?: string;
}) {
  const [activeChartIndex, setActiveChartIndex] = useState(0);

  const buildChart = (
    pick: (point: ExerciseExerciseMetricsHistoryPoint) => number,
    aggregation: SessionMetricAggregation,
  ) => buildSessionMetricChartData(points, range, pick, aggregation);

  const best1rm = useMemo(
    () => buildChart((p) => p.session_best_e1rm ?? 0, "maximum"),
    [points, range],
  );
  const avg1rm = useMemo(
    () => buildChart((p) => p.session_avg_e1rm ?? 0, "average"),
    [points, range],
  );
  const avgIntensity = useMemo(
    () => buildChart((p) => p.session_avg_intensity ?? 0, "average"),
    [points, range],
  );
  const bestIntensity = useMemo(
    () => buildChart((p) => p.session_best_intensity ?? 0, "maximum"),
    [points, range],
  );
  const volumeWorking = useMemo(
    () => buildChart((p) => p.total_volume_working ?? 0, "sum"),
    [points, range],
  );
  const bucketLabel = getSessionMetricBucketLabel(range);
  const periodLabel = getSessionMetricPeriodLabel(range);

  const charts: Array<{
    title: string;
    data: MetricPoint[];
    unit: "lb" | "%" | "vol";
    description: string;
  }> = [
    {
      title: `${periodLabel} Session Best 1RM`,
      description: `${bucketLabel}. Highest estimated 1RM from any working set in the period.`,
      data: best1rm,
      unit: "lb",
    },
    {
      title: `${periodLabel} Average Session 1RM`,
      description: `${bucketLabel}. Average estimated 1RM across working-set sessions in the period.`,
      data: avg1rm,
      unit: "lb",
    },
    {
      title: `${periodLabel} Average Session Intensity`,
      description: `${bucketLabel}. Average intensity of working sets versus your historical 1RM.`,
      data: avgIntensity,
      unit: "%",
    },
    {
      title: `${periodLabel} Session Best Intensity`,
      description: `${bucketLabel}. Highest single-set intensity versus your historical 1RM.`,
      data: bestIntensity,
      unit: "%",
    },
    {
      title: `${periodLabel} Working-Set Volume`,
      description: `${bucketLabel}. Total volume from working sets in the period.`,
      data: volumeWorking,
      unit: "vol",
    },
  ] as const;

  if (points.length === 0) {
    return (
      <p className="py-6 text-center text-sm text-muted-foreground">
        No working-set sessions for this exercise.
      </p>
    );
  }

  if (!hasWeightedMetrics(points)) {
    return (
      <p className="py-6 text-center text-sm text-muted-foreground">
        No weighted metrics for this exercise.
      </p>
    );
  }

  const safeIndex = Math.min(activeChartIndex, charts.length - 1);
  const activeChart = charts[safeIndex];

  return (
    <div className="space-y-4">
      {statusMessage ? (
        <p
          role="status"
          aria-live="polite"
          className="text-sm text-muted-foreground"
        >
          {statusMessage}
        </p>
      ) : null}

      <ChartBarMetric
        title={activeChart.title}
        description={activeChart.description}
        range={range}
        bucket={getSessionMetricBucket(range)}
        data={activeChart.data}
        unit={activeChart.unit}
        onWorkoutClick={onWorkoutClick}
      />

      <div className="flex items-center justify-center gap-4 pb-1">
        <button
          type="button"
          aria-label="Previous graph"
          className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border bg-background text-foreground disabled:opacity-40"
          onClick={() => setActiveChartIndex((prev) => Math.max(0, prev - 1))}
          disabled={safeIndex === 0}
        >
          <ChevronLeft className="h-5 w-5" />
        </button>

        <p className="text-xs text-muted-foreground">
          Graph {safeIndex + 1} of {charts.length}
        </p>

        <button
          type="button"
          aria-label="Next graph"
          className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border bg-background text-foreground disabled:opacity-40"
          onClick={() =>
            setActiveChartIndex((prev) => Math.min(charts.length - 1, prev + 1))
          }
          disabled={safeIndex === charts.length - 1}
        >
          <ChevronRight className="h-5 w-5" />
        </button>
      </div>
    </div>
  );
}

function AuthedCharts({
  exerciseId,
  range,
  onWorkoutClick,
}: {
  exerciseId: number;
  range: RangeType;
  onWorkoutClick: (workoutId: number) => void;
}) {
  const { data, error, isFetching, isPending } = useQuery(
    exerciseMetricsHistoryQueryOptions(exerciseId),
  );

  if (error && !data) {
    throw error;
  }
  if (isPending && !data) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground"
      >
        <Spinner
          size="small"
          className="text-primary"
        />
        <span>Loading session metrics...</span>
      </div>
    );
  }

  const points = data?.points ?? [];
  const statusMessage =
    error && data
      ? "Couldn't update chart. Showing previous data."
      : isFetching
        ? "Updating chart..."
        : undefined;
  return (
    <MetricChartsBody
      points={points}
      range={range}
      onWorkoutClick={onWorkoutClick}
      statusMessage={statusMessage}
    />
  );
}

function DemoCharts({
  exerciseSets,
  range,
  onWorkoutClick,
}: {
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  range: RangeType;
  onWorkoutClick: (workoutId: number) => void;
}) {
  const demo = useMemo(
    () => computeDemoMetricsHistory(exerciseSets),
    [exerciseSets],
  );
  const points = demo.points;
  return (
    <MetricChartsBody
      points={points}
      range={range}
      onWorkoutClick={onWorkoutClick}
    />
  );
}
