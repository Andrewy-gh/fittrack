import { useState } from 'react';
import ReactApexChart from 'react-apexcharts';
import type { ApexOptions } from 'apexcharts';
import { format, parseISO } from 'date-fns';
import { ChartWrapper } from './ChartWrapper';
import { RangeSelector } from './RangeSelector';
import { mockVolumeData, filterDataByRange, getRangeLabel, getDateFormat, type RangeType } from '@/data/mockData';

export function ApexDemo() {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const filteredData = filterDataByRange(mockVolumeData, selectedRange);

  // Get CSS variable values
  const getComputedColor = (variable: string) => {
    if (typeof window === 'undefined') return '#000';
    return getComputedStyle(document.documentElement)
      .getPropertyValue(variable)
      .trim();
  };

  const series = [
    {
      name: 'Volume',
      data: filteredData.map((d) => ({
        x: format(parseISO(d.date), getDateFormat(selectedRange)),
        y: d.volume,
      })),
    },
  ];

  const options: ApexOptions = {
    chart: {
      type: 'bar',
      height: 320,
      toolbar: {
        show: false,
      },
      background: 'transparent',
      fontFamily: 'inherit',
    },
    plotOptions: {
      bar: {
        borderRadius: 4,
        borderRadiusApplication: 'end',
        columnWidth: '70%',
      },
    },
    colors: [getComputedColor('--color-primary')],
    dataLabels: {
      enabled: false,
    },
    stroke: {
      show: false,
    },
    grid: {
      borderColor: getComputedColor('--color-muted'),
      strokeDashArray: 3,
      xaxis: {
        lines: {
          show: false,
        },
      },
      yaxis: {
        lines: {
          show: true,
        },
      },
    },
    xaxis: {
      type: 'category',
      labels: {
        style: {
          colors: getComputedColor('--color-foreground'),
          fontSize: '12px',
        },
      },
      axisBorder: {
        show: false,
      },
      axisTicks: {
        show: false,
      },
    },
    yaxis: {
      labels: {
        style: {
          colors: getComputedColor('--color-foreground'),
          fontSize: '12px',
        },
        formatter: (value) => `${value.toLocaleString()}`,
      },
    },
    tooltip: {
      enabled: true,
      theme: 'dark',
      style: {
        fontSize: '14px',
      },
      fixed: {
        enabled: true,
        position: 'topLeft',
        offsetX: 120,
        offsetY: 0,
      },
      custom: function ({ series, seriesIndex, dataPointIndex, w }) {
        const value = series[seriesIndex][dataPointIndex];
        const category = w.globals.labels[dataPointIndex];
        return `
          <div style="
            background: ${getComputedColor('--color-background')};
            border: 1px solid ${getComputedColor('--color-border')};
            border-radius: var(--radius-md);
            padding: 0.75rem;
            box-shadow: var(--shadow);
          ">
            <div style="
              color: ${getComputedColor('--color-foreground')};
              font-size: 0.875rem;
              margin-bottom: 0.25rem;
            ">${category}</div>
            <div style="
              color: ${getComputedColor('--color-foreground')};
              font-size: 0.875rem;
              text-transform: capitalize;
            ">Volume: ${value.toLocaleString()} kg</div>
          </div>
        `;
      },
    },
  };

  return (
    <ChartWrapper
      title="5. ApexCharts"
      description="Feature-rich interactive charts with built-in zoom/pan and responsive design"
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
          <ReactApexChart
            options={options}
            series={series}
            type="bar"
            height={320}
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
