import { useState } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';
import { format, parseISO } from 'date-fns';
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
);

export function ChartJsDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);

  // Get CSS variable values
  const getComputedColor = (variable: string) => {
    if (typeof window === 'undefined') return '#000';
    return getComputedStyle(document.documentElement)
      .getPropertyValue(variable)
      .trim();
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

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      tooltip: {
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
            size: 12,
          },
        },
      },
      y: {
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
            size: 12,
          },
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

        {/* Chart */}
        <div className="h-80 w-full">
          <Bar data={chartData} options={options} />
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
