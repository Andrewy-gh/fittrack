import { useMemo, useState } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { format, parseISO } from 'date-fns';

import type { ExerciseExerciseWithSetsResponse } from '@/client';
import {
  filterDataByRange,
  getDateFormat,
  getRangeLabel,
  getResponsiveValue,
  responsiveConfig,
  useBreakpoint,
  type RangeType,
  type VolumeData,
} from './chart-bar-vol.utils';
import { RangeSelector, ScrollableChart } from './chart-bar-vol.components';

interface ChartBarVolProps {
  data: Array<ExerciseExerciseWithSetsResponse>;
}

export function ChartBarVol({ data }: ChartBarVolProps) {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const breakpoint = useBreakpoint();

  const dailyVolume = useMemo<VolumeData[]>(() => {
    const volumeByDay: { [key: string]: number } = {};

    data.forEach((set) => {
      const day = set.workout_date.split('T')[0];
      if (!volumeByDay[day]) {
        volumeByDay[day] = 0;
      }
      volumeByDay[day] += set.volume;
    });

    return Object.keys(volumeByDay)
      .map((date) => ({
        date,
        volume: volumeByDay[date],
      }))
      .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());
  }, [data]);

  const filteredData = useMemo(
    () => filterDataByRange(dailyVolume, selectedRange),
    [dailyVolume, selectedRange]
  );

  const barWidth = getResponsiveValue(responsiveConfig.barWidth, breakpoint);
  const yAxisWidth = getResponsiveValue(
    responsiveConfig.yAxisWidth,
    breakpoint
  );
  const chartMargins = getResponsiveValue(
    responsiveConfig.chartMargins,
    breakpoint
  );
  const axisMargins = { ...chartMargins, left: 0, right: 0 };
  const plotMargins = { ...chartMargins, left: 0 };
  const chartHeight = 320;

  const averageVolume = filteredData.length > 0
    ? Math.round(
        filteredData.reduce((sum, entry) => sum + entry.volume, 0) /
          filteredData.length
      ).toLocaleString()
    : '0';

  return (
    <section className="space-y-3">
      <div>
        <h2 className="text-xl font-semibold">Daily Volume</h2>
        <p className="text-sm text-muted-foreground">
          Total training volume per day.
        </p>
      </div>

      <div className="flex justify-center">
        <RangeSelector
          selectedRange={selectedRange}
          onRangeChange={setSelectedRange}
        />
      </div>

      <div
        className="grid gap-2"
        style={{ gridTemplateColumns: `${yAxisWidth}px 1fr` }}
      >
        <div className="min-w-0" style={{ width: yAxisWidth, height: chartHeight }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={filteredData}
              margin={axisMargins}
            >
              <XAxis
                dataKey="date"
                tick={false}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                dataKey="volume"
                width={yAxisWidth}
                stroke="var(--color-muted-foreground)"
                fontSize={getResponsiveValue(
                  responsiveConfig.fontSize,
                  breakpoint
                )}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `${value}`}
                tick={{
                  fill: 'var(--color-foreground)',
                }}
                domain={[0, 'dataMax']}
                tickCount={5}
              />
              <Bar
                dataKey="volume"
                fill="transparent"
                stroke="transparent"
                isAnimationActive={false}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="min-w-0">
          <ScrollableChart
            dataLength={filteredData.length}
            barWidth={barWidth}
            height={chartHeight}
            resetKey={selectedRange}
          >
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={filteredData}
                margin={plotMargins}
              >
                <CartesianGrid
                  strokeDasharray="3 3"
                  stroke="var(--color-muted)"
                />

                <XAxis
                  dataKey="date"
                  stroke="var(--color-muted-foreground)"
                  fontSize={getResponsiveValue(
                    responsiveConfig.fontSize,
                    breakpoint
                  )}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(str) =>
                    format(parseISO(str), getDateFormat(selectedRange))
                  }
                  tick={{
                    fill: 'var(--color-foreground)',
                  }}
                />

                <YAxis
                  dataKey="volume"
                  hide
                  width={0}
                  domain={[0, 'dataMax']}
                  tickCount={5}
                />

                <Tooltip
                  cursor={false}
                  position={{ y: 0 }}
                  contentStyle={{
                    backgroundColor: 'var(--color-background)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--radius-md)',
                    boxShadow: 'var(--shadow)',
                  }}
                  itemStyle={{
                    color: 'var(--color-foreground)',
                    fontSize: '0.875rem',
                    textTransform: 'capitalize',
                  }}
                  labelStyle={{
                    color: 'var(--color-foreground)',
                    fontSize: '0.875rem',
                    marginBottom: '0.25rem',
                  }}
                  labelFormatter={(label) => {
                    const dateFormat =
                      selectedRange === 'Y' ? 'MMM yyyy' : 'PPP';
                    return format(parseISO(label as string), dateFormat);
                  }}
                />

                <Bar
                  dataKey="volume"
                  fill="var(--color-primary)"
                  radius={4}
                />
              </BarChart>
            </ResponsiveContainer>
          </ScrollableChart>
        </div>
      </div>

      <div className="flex justify-between text-sm text-muted-foreground">
        <span>{getRangeLabel(selectedRange, filteredData.length)}</span>
        <span>Avg: {averageVolume} vol</span>
      </div>
    </section>
  );
}
