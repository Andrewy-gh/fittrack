import { useState, useEffect } from 'react';
import { BarChart } from '@tremor/react';
import { format, parseISO } from 'date-fns';
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { ScrollableChart } from './ScrollableChart';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';

export function TremorDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const [primaryColor, setPrimaryColor] = useState('#ea580c');
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);

  // Get the computed color value from CSS variable
  useEffect(() => {
    const computedColor = getComputedStyle(document.documentElement)
      .getPropertyValue('--color-primary')
      .trim();
    if (computedColor) {
      setPrimaryColor(computedColor);
    }
  }, []);

  // Transform data for Tremor (needs string values for display)
  const tremorData = filteredData.map((d) => ({
    date: format(parseISO(d.date), getDateFormat(selectedRange)),
    Volume: d.volume,
  }));

  return (
    <ChartWrapper
      title="2. Tremor BarChart"
      description="Built on Recharts with Tailwind CSS integration and pre-styled components"
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
          <BarChart
            data={tremorData}
            index="date"
            categories={['Volume']}
            colors={[primaryColor]}
            valueFormatter={(value) => `${value.toLocaleString()} kg`}
            showLegend={false}
            showGridLines={true}
            showXAxis={true}
            showYAxis={true}
            yAxisWidth={48}
            className="h-80"
            customTooltip={(props) => {
              if (!props.active || !props.payload || props.payload.length === 0) {
                return null;
              }
              const data = props.payload[0];
              return (
                <div
                  className="rounded-[var(--radius-md)] border border-[var(--color-border)] shadow-[var(--shadow)]"
                  style={{
                    backgroundColor: 'var(--color-background)',
                    padding: '0.75rem',
                  }}
                >
                  <div
                    className="text-sm mb-1"
                    style={{ color: 'var(--color-foreground)' }}
                  >
                    {data.payload.date}
                  </div>
                  <div
                    className="text-sm capitalize"
                    style={{ color: 'var(--color-foreground)' }}
                  >
                    Volume: {data.value?.toLocaleString()} kg
                  </div>
                </div>
              );
            }}
          />
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
