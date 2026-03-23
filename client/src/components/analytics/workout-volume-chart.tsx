import { useState } from 'react';
import type { WorkoutContributionDataResponse } from '@/client';
import { ChartBarMetric } from '@/components/charts/chart-bar-metric';
import { RangeSelector } from '@/components/charts/chart-bar-vol.components';
import type { RangeType } from '@/components/charts/chart-bar-vol.utils';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  buildWorkoutVolumeChartData,
  getWorkoutVolumeBucketLabel,
  getWorkoutVolumeTitle,
} from '@/lib/analytics';

const ALL_FOCUS_VALUE = 'all';

export function WorkoutVolumeChart({
  data,
  focusValues,
}: {
  data: WorkoutContributionDataResponse;
  focusValues: string[];
}) {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const [selectedFocus, setSelectedFocus] = useState(ALL_FOCUS_VALUE);

  const activeFocus =
    selectedFocus === ALL_FOCUS_VALUE ? undefined : selectedFocus;
  const chartData = buildWorkoutVolumeChartData(
    data.days,
    selectedRange,
    activeFocus
  );
  const bucketLabel = getWorkoutVolumeBucketLabel(selectedRange);
  const title = getWorkoutVolumeTitle(selectedRange, activeFocus);
  const description = activeFocus
    ? `${bucketLabel}. Working-set volume for ${activeFocus} workouts.`
    : `${bucketLabel}. Working-set volume across all workouts.`;

  return (
    <Card>
      <CardHeader className="space-y-4">
        <div>
          <CardTitle>Workout Volume</CardTitle>
          <CardDescription>
            Separate from exercise metrics: this tracks total working-set volume
            by time range and focus.
          </CardDescription>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <RangeSelector
            selectedRange={selectedRange}
            onRangeChange={setSelectedRange}
          />

          <Select value={selectedFocus} onValueChange={setSelectedFocus}>
            <SelectTrigger
              aria-label="Workout focus filter"
              className="w-full sm:w-[200px]"
            >
              <SelectValue placeholder="All focus types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_FOCUS_VALUE}>All focus types</SelectItem>
              {focusValues.map((focus) => (
                <SelectItem key={focus} value={focus}>
                  {focus}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <p className="text-xs text-muted-foreground">{bucketLabel}</p>
      </CardHeader>

      <CardContent className="pt-0">
        {chartData.every((point) => point.value === 0) ? (
          <p className="py-6 text-sm text-muted-foreground">
            No workout volume for the selected focus in this range yet.
          </p>
        ) : (
          <ChartBarMetric
            title={title}
            description={description}
            range={selectedRange}
            data={chartData}
            unit="vol"
          />
        )}
      </CardContent>
    </Card>
  );
}
