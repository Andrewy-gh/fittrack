import { useState, useEffect } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  TooltipPositionerFunction,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';
import { format, parseISO } from 'date-fns';
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { ScrollableChart } from './ScrollableChart';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';
import { useBreakpoint } from '../hooks/useBreakpoint';
import { responsiveConfig, getResponsiveValue } from '../utils/responsiveConfig';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

// Custom tooltip positioner - anchors above bar, fixed Y position
const customPositioner: TooltipPositionerFunction<'bar'> = function(items) {
  if (!items.length) {
    return { x: 0, y: 0 };
  }

  const pos = Tooltip.positioners.average.call(this, items);
  if (pos === false) {
    return { x: 0, y: 0 };
  }

  // Fixed Y position (top of chart area, under buttons)
  const fixedY = this.chart.chartArea.top;

  return {
    x: pos.x,
    y: fixedY,
    xAlign: 'center',
    yAlign: 'bottom',
  };
};

// Register custom positioner
Tooltip.positioners.fixedTop = customPositioner;

export function ChartJsDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const breakpoint = useBreakpoint();
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);
  const yAxisWidth = getResponsiveValue(responsiveConfig.yAxisWidth, breakpoint);
  const chartHeight = 320;

  // Get CSS variable values
  const getComputedColor = (variable: string) => {
    if (typeof window === 'undefined') return '#000';
    return getComputedStyle(document.documentElement)
      .getPropertyValue(variable)
      .trim();
  };

  const maxValue = filteredData.reduce((max, d) => Math.max(max, d.volume), 0);
  const yAxisMax = maxValue > 0 ? maxValue : 10;
  const formatYAxisTick = (value: number) => {
    if (value >= 1000) {
      const rounded = value / 1000;
      const formatted = rounded >= 10
        ? Math.round(rounded).toString()
        : rounded.toFixed(1).replace(/\.0$/, '');
      return `${formatted}k`;
    }
    return `${value}`;
  };

  const chartData = {
    labels: filteredData.map((d) => format(parseISO(d.date), getDateFormat(selectedRange))),
    datasets: [
      {
        label: 'Volume',
        data: filteredData.map((d) => d.volume),
        backgroundColor: getComputedColor('--color-primary'),
        borderRadius: 4,
        borderSkipped: false,
      },
    ],
  };

  const baseOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
    },
    scales: {
      x: {
        grid: {
          display: false,
        },
        border: {
          display: false,
        },
        ticks: {
          color: getComputedColor('--color-foreground'),
          font: {
            size: getResponsiveValue(responsiveConfig.fontSize, breakpoint),
          },
        },
      },
      y: {
        beginAtZero: true,
        max: yAxisMax,
        grid: {
          color: getComputedColor('--color-muted'),
          lineWidth: 1,
          drawBorder: false,
        },
        border: {
          display: false,
          dash: [3, 3],
        },
        ticks: {
          color: getComputedColor('--color-foreground'),
          font: {
            size: getResponsiveValue(responsiveConfig.fontSize, breakpoint),
          },
          maxTicksLimit: 5,
          callback: (value: string | number) => formatYAxisTick(Number(value)),
        },
      },
    },
  };

  const axisData = {
    labels: chartData.labels,
    datasets: [
      {
        ...chartData.datasets[0],
        backgroundColor: 'transparent',
        borderColor: 'transparent',
      },
    ],
  };

  const axisOptions = {
    ...baseOptions,
    plugins: {
      legend: { display: false },
      tooltip: { enabled: false },
    },
    scales: {
      x: {
        display: false,
      },
      y: {
        ...baseOptions.scales.y,
        ticks: {
          ...baseOptions.scales.y.ticks,
        },
      },
    },
  };

  const plotOptions = {
    ...baseOptions,
    plugins: {
      legend: { display: false },
      tooltip: {
        position: 'fixedTop' as any, // Custom positioner
        backgroundColor: getComputedColor('--color-background'),
        titleColor: getComputedColor('--color-foreground'),
        bodyColor: getComputedColor('--color-foreground'),
        borderColor: getComputedColor('--color-border'),
        borderWidth: 1,
        padding: 12,
        displayColors: false,
        callbacks: {
          title: function (context: any) {
            return context[0].label;
          },
          label: function (context: any) {
            return `Volume: ${context.parsed.y.toLocaleString()} kg`;
          },
        },
      },
    },
    scales: {
      x: baseOptions.scales.x,
      y: {
        ...baseOptions.scales.y,
        ticks: {
          ...baseOptions.scales.y.ticks,
          callback: () => '',
        },
      },
    },
  };

  return (
    <ChartWrapper
      title="4. Chart.js with react-chartjs-2"
      description="Canvas-based rendering with excellent mobile performance and smooth animations"
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
            <Bar data={axisData} options={axisOptions} />
          </div>
          <div className="chart-plot">
            <ScrollableChart
              dataLength={filteredData.length}
              barWidth={getResponsiveValue(responsiveConfig.barWidth, breakpoint)}
              height={chartHeight}
            >
              <Bar data={chartData} options={plotOptions} />
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
