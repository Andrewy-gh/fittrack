import {
  type ReactNode,
  useLayoutEffect,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
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

interface ChartBarVolProps {
  data: Array<ExerciseExerciseWithSetsResponse>;
}

interface VolumeData {
  date: string;
  volume: number;
}

type RangeType = 'W' | 'M' | '6M' | 'Y';

type Breakpoint = 'mobile' | 'tablet' | 'desktop';

interface ResponsiveValue<T> {
  mobile: T;
  tablet: T;
  desktop: T;
}

const responsiveConfig = {
  barWidth: {
    mobile: 30,
    tablet: 40,
    desktop: 50,
  },
  fontSize: {
    mobile: 10,
    tablet: 11,
    desktop: 12,
  },
  chartMargins: {
    mobile: { top: 10, right: 10, bottom: 20, left: 40 },
    tablet: { top: 15, right: 15, bottom: 25, left: 45 },
    desktop: { top: 20, right: 20, bottom: 30, left: 48 },
  },
  buttonPadding: {
    mobile: 'px-3 py-1.5 text-xs',
    tablet: 'px-3.5 py-2 text-sm',
    desktop: 'px-4 py-2 text-sm',
  },
  containerGap: {
    mobile: 'gap-0.5 p-0.5',
    tablet: 'gap-1 p-1',
    desktop: 'gap-1 p-1',
  },
  scrollButton: {
    mobile: {
      padding: 'p-1',
      iconSize: 12,
    },
    tablet: {
      padding: 'p-1.5',
      iconSize: 14,
    },
    desktop: {
      padding: 'p-1.5',
      iconSize: 14,
    },
  },
  yAxisWidth: {
    mobile: 40,
    tablet: 45,
    desktop: 48,
  },
};

const ranges: Array<{ value: RangeType; label: string }> = [
  { value: 'W', label: 'W' },
  { value: 'M', label: 'M' },
  { value: '6M', label: '6M' },
  { value: 'Y', label: 'Y' },
];

function getResponsiveValue<T>(
  values: ResponsiveValue<T>,
  breakpoint: Breakpoint
): T {
  return values[breakpoint];
}

function useBreakpoint(): Breakpoint {
  const [breakpoint, setBreakpoint] = useState<Breakpoint>(() => {
    if (typeof window === 'undefined') return 'desktop';
    const width = window.innerWidth;
    if (width < 640) return 'mobile';
    if (width < 1024) return 'tablet';
    return 'desktop';
  });

  useEffect(() => {
    const mobileQuery = window.matchMedia('(max-width: 639px)');
    const tabletQuery = window.matchMedia('(min-width: 640px) and (max-width: 1023px)');

    const updateBreakpoint = () => {
      if (mobileQuery.matches) {
        setBreakpoint('mobile');
      } else if (tabletQuery.matches) {
        setBreakpoint('tablet');
      } else {
        setBreakpoint('desktop');
      }
    };

    mobileQuery.addEventListener('change', updateBreakpoint);
    tabletQuery.addEventListener('change', updateBreakpoint);

    return () => {
      mobileQuery.removeEventListener('change', updateBreakpoint);
      tabletQuery.removeEventListener('change', updateBreakpoint);
    };
  }, []);

  return breakpoint;
}

function filterDataByDays(data: VolumeData[], days: number): VolumeData[] {
  if (days >= data.length) return data;
  return data.slice(-days);
}

function aggregateToWeekly(data: VolumeData[]): VolumeData[] {
  if (data.length === 0) return [];

  const weekMap = new Map<string, { volumes: number[]; weekStart: Date }>();

  data.forEach((item) => {
    const date = new Date(item.date);
    const dayOfWeek = date.getDay();
    const diff = dayOfWeek === 0 ? -6 : 1 - dayOfWeek;
    const monday = new Date(date);
    monday.setDate(date.getDate() + diff);

    const weekKey = monday.toISOString().split('T')[0];

    if (!weekMap.has(weekKey)) {
      weekMap.set(weekKey, { volumes: [], weekStart: monday });
    }

    weekMap.get(weekKey)!.volumes.push(item.volume);
  });

  return Array.from(weekMap.entries())
    .map(([weekKey, { volumes }]) => ({
      date: weekKey,
      volume: Math.round(
        volumes.reduce((sum, v) => sum + v, 0) / volumes.length
      ),
    }))
    .sort((a, b) => a.date.localeCompare(b.date));
}

function aggregateToMonthly(data: VolumeData[]): VolumeData[] {
  if (data.length === 0) return [];

  const monthlyData: VolumeData[] = [];
  const sortedData = [...data].sort((a, b) => a.date.localeCompare(b.date));
  const endDate = new Date(sortedData[sortedData.length - 1].date);

  for (let i = 0; i < 12; i++) {
    const windowEnd = new Date(endDate);
    windowEnd.setDate(endDate.getDate() - i * 30);

    const windowStart = new Date(windowEnd);
    windowStart.setDate(windowEnd.getDate() - 30);

    const windowData = sortedData.filter((item) => {
      const itemDate = new Date(item.date);
      return itemDate >= windowStart && itemDate <= windowEnd;
    });

    if (windowData.length > 0) {
      const avgVolume = Math.round(
        windowData.reduce((sum, d) => sum + d.volume, 0) / windowData.length
      );

      monthlyData.unshift({
        date: windowEnd.toISOString().split('T')[0],
        volume: avgVolume,
      });
    }
  }

  return monthlyData;
}

function filterDataByRange(data: VolumeData[], range: RangeType): VolumeData[] {
  switch (range) {
    case 'W':
      return filterDataByDays(data, 7);
    case 'M':
      return filterDataByDays(data, 30);
    case '6M': {
      const sixMonthData = filterDataByDays(data, 180);
      const weeklyData = aggregateToWeekly(sixMonthData);
      return weeklyData.slice(-26);
    }
    case 'Y':
      return aggregateToMonthly(data);
    default:
      return data;
  }
}

function getRangeLabel(range: RangeType, count: number): string {
  switch (range) {
    case 'W':
      return `Showing ${count} day${count !== 1 ? 's' : ''}`;
    case 'M':
      return `Showing ${count} day${count !== 1 ? 's' : ''}`;
    case '6M':
      return `Showing ${count} week${count !== 1 ? 's' : ''}`;
    case 'Y':
      return `Showing ${count} month${count !== 1 ? 's' : ''}`;
    default:
      return `Showing ${count} points`;
  }
}

function getDateFormat(range: RangeType): string {
  switch (range) {
    case 'W':
    case 'M':
      return 'MMM d';
    case '6M':
      return 'MMM d';
    case 'Y':
      return 'MMM yyyy';
    default:
      return 'MMM d';
  }
}

interface RangeSelectorProps {
  selectedRange: RangeType;
  onRangeChange: (range: RangeType) => void;
}

function RangeSelector({ selectedRange, onRangeChange }: RangeSelectorProps) {
  const breakpoint = useBreakpoint();
  const containerClasses = getResponsiveValue(
    responsiveConfig.containerGap,
    breakpoint
  );
  const buttonPadding = getResponsiveValue(
    responsiveConfig.buttonPadding,
    breakpoint
  );

  return (
    <div
      className={`inline-flex bg-[var(--color-secondary)] rounded-[var(--radius-md)] ${containerClasses}`}
    >
      {ranges.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => onRangeChange(value)}
          className={`
            ${buttonPadding} rounded-[var(--radius-sm)] border-0 font-medium
            transition-all duration-200 ease-in-out cursor-pointer
            ${
              selectedRange === value
                ? 'bg-[var(--color-primary)] text-[var(--color-primary-foreground)] font-semibold'
                : 'bg-transparent text-[var(--color-foreground)]'
            }
          `}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

interface ScrollableChartProps {
  children: ReactNode;
  dataLength: number;
  barWidth?: number;
  height?: number;
  resetKey?: string | number;
}

function ScrollableChart({
  children,
  dataLength,
  barWidth,
  height = 320,
  resetKey,
}: ScrollableChartProps) {
  const breakpoint = useBreakpoint();
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);
  const [isTouchDevice, setIsTouchDevice] = useState(false);
  const [containerWidth, setContainerWidth] = useState(0);

  const effectiveBarWidth =
    barWidth ?? getResponsiveValue(responsiveConfig.barWidth, breakpoint);
  const minChartWidth = dataLength * effectiveBarWidth;
  const chartWidth = Math.max(minChartWidth, containerWidth || 0);
  const buttonConfig = responsiveConfig.scrollButton[breakpoint];

  const checkScrollPosition = () => {
    const element = scrollRef.current;
    if (!element) return;

    const { scrollLeft, scrollWidth, clientWidth } = element;
    const expectedWidth = Math.max(minChartWidth, clientWidth);
    const maxScrollLeft = Math.max(0, expectedWidth - clientWidth);
    setCanScrollLeft(scrollLeft > 0);
    setCanScrollRight(scrollLeft < maxScrollLeft - 1);
  };

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const touchQuery = window.matchMedia('(pointer: coarse)');

    const updateTouchState = () => {
      setIsTouchDevice(touchQuery.matches);
    };

    updateTouchState();
    touchQuery.addEventListener('change', updateTouchState);

    return () => {
      touchQuery.removeEventListener('change', updateTouchState);
    };
  }, []);

  useEffect(() => {
    const element = scrollRef.current;
    if (!element) return;
    const updateScroll = () => {
      const expectedWidth = Math.max(minChartWidth, element.clientWidth);
      const maxScrollLeft = Math.max(0, expectedWidth - element.clientWidth);
      if (element.scrollLeft > maxScrollLeft) {
        element.scrollLeft = maxScrollLeft;
      }
      checkScrollPosition();
    };
    const raf = requestAnimationFrame(updateScroll);
    const timeout = window.setTimeout(updateScroll, 60);
    return () => {
      cancelAnimationFrame(raf);
      window.clearTimeout(timeout);
    };
  }, [dataLength, barWidth, height]);

  useLayoutEffect(() => {
    const element = scrollRef.current;
    if (!element) return;
    const expectedWidth = Math.max(minChartWidth, element.clientWidth);
    const maxScrollLeft = Math.max(0, expectedWidth - element.clientWidth);
    element.scrollLeft = maxScrollLeft;
    checkScrollPosition();
  }, [resetKey, dataLength, barWidth, height, minChartWidth]);

  useEffect(() => {
    const element = scrollRef.current;
    if (!element || typeof ResizeObserver === 'undefined') return;
    const clampScroll = () => {
      const expectedWidth = Math.max(minChartWidth, element.clientWidth);
      const maxScrollLeft = Math.max(0, expectedWidth - element.clientWidth);
      if (element.scrollLeft > maxScrollLeft) {
        element.scrollLeft = maxScrollLeft;
      }
      checkScrollPosition();
    };
    const observer = new ResizeObserver(() => {
      requestAnimationFrame(clampScroll);
    });
    observer.observe(element);
    const inner = element.firstElementChild as HTMLElement | null;
    if (inner) observer.observe(inner);
    return () => observer.disconnect();
  }, []);

  useLayoutEffect(() => {
    const element = scrollRef.current;
    if (!element || typeof ResizeObserver === 'undefined') return;
    const updateWidth = () => {
      setContainerWidth(element.clientWidth);
    };
    updateWidth();
    const observer = new ResizeObserver(updateWidth);
    observer.observe(element);
    return () => observer.disconnect();
  }, []);

  const scroll = (direction: 'left' | 'right') => {
    const element = scrollRef.current;
    if (!element) return;

    const scrollAmount = element.clientWidth * 0.8;
    element.scrollBy({
      left: direction === 'left' ? -scrollAmount : scrollAmount,
      behavior: 'smooth',
    });

    setTimeout(checkScrollPosition, 300);
  };

  return (
    <div className="relative">
      <div
        ref={scrollRef}
        onScroll={checkScrollPosition}
        className="overflow-x-auto overflow-y-hidden touch-pan-x"
        style={{ height: `${height}px` }}
      >
        <div style={{ width: `${chartWidth}px`, height: '100%' }}>
          {children}
        </div>
      </div>

      {!isTouchDevice && canScrollLeft && (
        <button
          onClick={() => scroll('left')}
          className={`absolute ${breakpoint === 'mobile' ? 'left-1' : 'left-2'} top-1/2 -translate-y-1/2 bg-[var(--color-background)] border border-[var(--color-border)] rounded-full ${buttonConfig.padding} shadow-lg hover:bg-[var(--color-muted)] transition-colors`}
          aria-label="Scroll left"
        >
          <svg
            width={buttonConfig.iconSize}
            height={buttonConfig.iconSize}
            viewBox="0 0 16 16"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M10 12L6 8l4-4" />
          </svg>
        </button>
      )}

      {!isTouchDevice && canScrollRight && (
        <button
          onClick={() => scroll('right')}
          className={`absolute ${breakpoint === 'mobile' ? 'right-1' : 'right-2'} top-1/2 -translate-y-1/2 bg-[var(--color-background)] border border-[var(--color-border)] rounded-full ${buttonConfig.padding} shadow-lg hover:bg-[var(--color-muted)] transition-colors`}
          aria-label="Scroll right"
        >
          <svg
            width={buttonConfig.iconSize}
            height={buttonConfig.iconSize}
            viewBox="0 0 16 16"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M6 12l4-4-4-4" />
          </svg>
        </button>
      )}

    </div>
  );
}

export function ChartBarVol({ data }: ChartBarVolProps) {
  const [selectedRange, setSelectedRange] = useState<RangeType>('M');
  const breakpoint = useBreakpoint();
  const useMockData = true;

  const mockVolume = useMemo<VolumeData[]>(() => {
    const today = new Date();
    return Array.from({ length: 100 }, (_, index) => {
      const offset = 99 - index;
      const date = new Date(today);
      date.setDate(date.getDate() - offset);
      const base = 7000;
      const weeklyWave = Math.sin(index / 2.2) * 1400;
      const monthlyWave = Math.sin(index / 7.5) * 2200;
      const noiseSeed = Math.abs(Math.sin(index * 12.9898) * 43758.5453);
      const noise = Math.floor((noiseSeed % 1) * 3000);
      const spike = index % 17 === 0 ? 2800 : 0;
      const dip = index % 13 === 0 ? -1600 : 0;
      return {
        date: date.toISOString().split('T')[0],
        volume: Math.max(3000, Math.round(base + weeklyWave + monthlyWave + noise + spike + dip)),
      };
    });
  }, []);

  const dailyVolume = useMemo<VolumeData[]>(() => {
    if (useMockData) return mockVolume;
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
  }, [data, mockVolume, useMockData]);

  const filteredData = useMemo(
    () => filterDataByRange(dailyVolume, selectedRange),
    [dailyVolume, selectedRange]
  );

  const barWidth = getResponsiveValue(
    responsiveConfig.barWidth,
    breakpoint
  );
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
