import { useState } from 'react';
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
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { ScrollableChart } from './ScrollableChart';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';

export function RechartsDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);

  return (
    <ChartWrapper
      title="1. Recharts + Custom Buttons"
      description="Baseline implementation using existing Recharts library with button-based range selection"
    >
      <div className="flex flex-col gap-4">
        {/* Range Selector */}
        <div className="flex justify-center">
          <RangeSelector
            selectedRange={selectedRange}
            onRangeChange={setSelectedRange}
          />
        </div>

        {/* Chart with Horizontal Scroll */}
        <ScrollableChart dataLength={filteredData.length} barWidth={50}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={filteredData}
              margin={{
                top: 5,
                right: 30,
                left: 60,
                bottom: 5,
              }}
            >
              <CartesianGrid
                strokeDasharray="3 3"
                stroke="var(--color-muted)"
              />

              <XAxis
                dataKey="date"
                stroke="var(--color-muted-foreground)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(str) => format(parseISO(str), getDateFormat(selectedRange))}
                tick={{
                  fill: 'var(--color-foreground)',
                }}
              />

              <YAxis
                stroke="var(--color-muted-foreground)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `${value}`}
                tick={{
                  fill: 'var(--color-foreground)',
                }}
              />

              <Tooltip
                cursor={false}
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
                  const dateFormat = selectedRange === 'Y' ? 'MMM yyyy' : 'PPP';
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

        {/* Stats */}
        <div className="flex justify-between text-sm text-[var(--color-muted-foreground)]">
          <span>{getRangeLabel(selectedRange, filteredData.length)}</span>
          <span>
            Avg:{' '}
            {filteredData.length > 0
              ? Math.round(
                  filteredData.reduce((sum, d) => sum + d.volume, 0) /
                    filteredData.length
                ).toLocaleString()
              : 0}{' '}
            kg
          </span>
        </div>
      </div>
    </ChartWrapper>
  );
}
