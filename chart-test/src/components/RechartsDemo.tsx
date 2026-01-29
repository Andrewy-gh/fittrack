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
import { useBreakpoint } from '../hooks/useBreakpoint';
import { responsiveConfig, getResponsiveValue } from '../utils/responsiveConfig';

export function RechartsDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const breakpoint = useBreakpoint();
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);
  const yAxisWidth = getResponsiveValue(responsiveConfig.yAxisWidth, breakpoint);
  const chartMargins = getResponsiveValue(responsiveConfig.chartMargins, breakpoint);
  const axisMargins = { ...chartMargins, left: 0, right: 0 };
  const plotMargins = { ...chartMargins, left: 0 };
  const chartHeight = 320;

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
        <div
          className="chart-axis-layout"
          style={{ gridTemplateColumns: `${yAxisWidth}px 1fr` }}
        >
          <div className="chart-axis" style={{ width: yAxisWidth, height: chartHeight }}>
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
                  fontSize={getResponsiveValue(responsiveConfig.fontSize, breakpoint)}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(value) => `${value}`}
                  tick={{
                    fill: 'var(--color-foreground)',
                  }}
                  domain={[0, 'dataMax']}
                  tickCount={5}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="chart-plot">
            <ScrollableChart
              dataLength={filteredData.length}
              barWidth={getResponsiveValue(responsiveConfig.barWidth, breakpoint)}
              height={chartHeight}
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
                    fontSize={getResponsiveValue(responsiveConfig.fontSize, breakpoint)}
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={(str) => format(parseISO(str), getDateFormat(selectedRange))}
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
          </div>
        </div>

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
