import { useMemo, useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { useRouter } from '@tanstack/react-router';

import type {
  ExerciseExerciseMetricsHistoryPoint,
  ExerciseExerciseWithSetsResponse,
} from '@/client';
import {
  exerciseMetricsHistoryQueryOptions,
  type MetricsHistoryRange,
} from '@/lib/api/exercises';
import { computeDemoMetricsHistory } from '@/lib/metrics-history';
import { RangeSelector } from '@/components/charts/chart-bar-vol.components';
import type { RangeType } from '@/components/charts/chart-bar-vol.utils';
import { ChartBarMetric, type MetricPoint } from '@/components/charts/chart-bar-metric';

function toMetricPoints(
  points: ExerciseExerciseMetricsHistoryPoint[],
  pick: (p: ExerciseExerciseMetricsHistoryPoint) => number
): MetricPoint[] {
  return points
    .map((p) => ({
      x: p.x ?? '',
      date: (p.date ?? '').split('T')[0],
      workout_id: p.workout_id,
      value: pick(p) ?? 0,
    }))
    .filter((p) => p.x && p.date);
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
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');

  const onWorkoutClick = (workoutId: number) =>
    router.navigate({ to: '/workouts/$workoutId', params: { workoutId } });

  return (
    <section className="space-y-6">
      <div className="space-y-3">
        <div>
          <h2 className="text-xl font-semibold">Session Metrics</h2>
          <p className="text-sm text-muted-foreground">
            e1RM and intensity are computed from working sets. Intensity can exceed 100%.
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
          range={selectedRange as MetricsHistoryRange}
          onWorkoutClick={onWorkoutClick}
        />
      ) : (
        <AuthedCharts
          exerciseId={exerciseId}
          range={selectedRange as MetricsHistoryRange}
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
}: {
  points: ExerciseExerciseMetricsHistoryPoint[];
  range: RangeType;
  onWorkoutClick: (workoutId: number) => void;
}) {
  const [activeChartIndex, setActiveChartIndex] = useState(0);

  const best1rm = useMemo(
    () => toMetricPoints(points, (p) => p.session_best_e1rm ?? 0),
    [points]
  );
  const avg1rm = useMemo(
    () => toMetricPoints(points, (p) => p.session_avg_e1rm ?? 0),
    [points]
  );
  const avgIntensity = useMemo(
    () => toMetricPoints(points, (p) => p.session_avg_intensity ?? 0),
    [points]
  );
  const bestIntensity = useMemo(
    () => toMetricPoints(points, (p) => p.session_best_intensity ?? 0),
    [points]
  );
  const volumeWorking = useMemo(
    () => toMetricPoints(points, (p) => p.total_volume_working ?? 0),
    [points]
  );

  const charts: Array<{
    title: string;
    data: MetricPoint[];
    unit: 'lb' | '%' | 'vol';
    description?: string;
  }> = [
    {
      title: 'Session Best 1RM',
      data: best1rm,
      unit: 'lb',
    },
    {
      title: 'Session Average 1RM',
      data: avg1rm,
      unit: 'lb',
    },
    {
      title: 'Session Average Intensity',
      data: avgIntensity,
      unit: '%',
    },
    {
      title: 'Session Best Intensity',
      data: bestIntensity,
      unit: '%',
    },
    {
      title: 'Working-Set Volume',
      description: 'Total volume from working sets.',
      data: volumeWorking,
      unit: 'vol',
    },
  ] as const;

  if (points.length === 0) return null;

  const safeIndex = Math.min(activeChartIndex, charts.length - 1);
  const activeChart = charts[safeIndex];

  return (
    <div className="space-y-4">
      <ChartBarMetric
        title={activeChart.title}
        description={activeChart.description}
        range={range}
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
          onClick={() => setActiveChartIndex((prev) => Math.min(charts.length - 1, prev + 1))}
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
  range: MetricsHistoryRange;
  onWorkoutClick: (workoutId: number) => void;
}) {
  const { data } = useSuspenseQuery(exerciseMetricsHistoryQueryOptions(exerciseId, range));
  const points = (data.points ?? []) as ExerciseExerciseMetricsHistoryPoint[];
  return <MetricChartsBody points={points} range={range} onWorkoutClick={onWorkoutClick} />;
}

function DemoCharts({
  exerciseSets,
  range,
  onWorkoutClick,
}: {
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  range: MetricsHistoryRange;
  onWorkoutClick: (workoutId: number) => void;
}) {
  const demo = useMemo(() => computeDemoMetricsHistory(exerciseSets, range), [exerciseSets, range]);
  const points = (demo.points ?? []) as any as ExerciseExerciseMetricsHistoryPoint[];
  return <MetricChartsBody points={points} range={range} onWorkoutClick={onWorkoutClick} />;
}
