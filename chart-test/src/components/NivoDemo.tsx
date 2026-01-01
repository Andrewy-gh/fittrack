import { useState, useEffect } from 'react';
import { ResponsiveBar } from '@nivo/bar';
import { format, parseISO } from 'date-fns';
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';

export function NivoDemo() {
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

  // Transform data for Nivo
  const nivoData = filteredData.map((d) => ({
    date: format(parseISO(d.date), getDateFormat(selectedRange)),
    volume: d.volume,
  }));

  return (
    <ChartWrapper
      title="3. Nivo ResponsiveBar"
      description="Beautiful out-of-box charts with excellent theming and touch-optimized for mobile"
    >
      <div className="flex flex-col gap-4">
        {/* Range Selector */}
        <div className="flex justify-center">
          <RangeSelector
            selectedRange={selectedRange}
            onRangeChange={setSelectedRange}
          />
        </div>

        {/* Chart */}
        <div className="h-80 w-full">
          <ResponsiveBar
            data={nivoData}
            keys={['volume']}
            indexBy="date"
            margin={{ top: 20, right: 30, bottom: 50, left: 60 }}
            padding={0.3}
            valueScale={{ type: 'linear' }}
            indexScale={{ type: 'band', round: true }}
            colors={[primaryColor]}
            borderRadius={4}
            borderColor={{
              from: 'color',
              modifiers: [['darker', 1.6]],
            }}
            axisTop={null}
            axisRight={null}
            axisBottom={{
              tickSize: 0,
              tickPadding: 5,
              tickRotation: 0,
              legendPosition: 'middle',
              legendOffset: 32,
            }}
            axisLeft={{
              tickSize: 0,
              tickPadding: 5,
              tickRotation: 0,
              legendPosition: 'middle',
              legendOffset: -40,
            }}
            enableGridY={true}
            gridYValues={5}
            enableLabel={false}
            labelSkipWidth={12}
            labelSkipHeight={12}
            labelTextColor={{
              from: 'color',
              modifiers: [['darker', 1.6]],
            }}
            theme={{
              background: 'transparent',
              text: {
                fill: 'var(--color-foreground)',
                fontSize: 12,
              },
              axis: {
                domain: {
                  line: {
                    stroke: 'transparent',
                  },
                },
                ticks: {
                  line: {
                    stroke: 'transparent',
                  },
                  text: {
                    fill: 'var(--color-foreground)',
                  },
                },
              },
              grid: {
                line: {
                  stroke: 'var(--color-muted)',
                  strokeDasharray: '3 3',
                },
              },
              tooltip: {
                container: {
                  background: 'var(--color-background)',
                  color: 'var(--color-foreground)',
                  fontSize: '0.875rem',
                  borderRadius: 'var(--radius-md)',
                  border: '1px solid var(--color-border)',
                  boxShadow: 'var(--shadow)',
                },
              },
            }}
            tooltip={({ id, value, indexValue }) => (
              <div
                style={{
                  padding: '0.75rem',
                  background: 'var(--color-background)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-md)',
                  boxShadow: 'var(--shadow)',
                }}
              >
                <div
                  style={{
                    color: 'var(--color-foreground)',
                    fontSize: '0.875rem',
                    marginBottom: '0.25rem',
                  }}
                >
                  {indexValue}
                </div>
                <div
                  style={{
                    color: 'var(--color-foreground)',
                    fontSize: '0.875rem',
                    textTransform: 'capitalize',
                  }}
                >
                  {id}: {value.toLocaleString()} kg
                </div>
              </div>
            )}
            role="application"
            ariaLabel="Nivo bar chart demo"
          />
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
