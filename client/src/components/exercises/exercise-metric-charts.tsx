import { useMemo, useState } from 'react';
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

  if (points.length === 0) return null;

  return (
    <div className="space-y-8">
      <ChartBarMetric
        title="Session Best 1RM"
        range={range}
        data={best1rm}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />

      <ChartBarMetric
        title="Session Average 1RM"
        range={range}
        data={avg1rm}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />

      <ChartBarMetric
        title="Session Average Intensity"
        range={range}
        data={avgIntensity}
        unit="%"
        onWorkoutClick={onWorkoutClick}
      />

      <ChartBarMetric
        title="Session Best Intensity"
        range={range}
        data={bestIntensity}
        unit="%"
        onWorkoutClick={onWorkoutClick}
      />

      <ChartBarMetric
        title="Working-Set Volume"
        description="Total volume from working sets."
        range={range}
        data={volumeWorking}
        unit="vol"
        onWorkoutClick={onWorkoutClick}
      />
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
