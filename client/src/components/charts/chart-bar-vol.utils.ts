import { useEffect, useState } from 'react';

export interface VolumeData {
  date: string;
  volume: number;
}

export type RangeType = 'W' | 'M' | '6M' | 'Y';

export type Breakpoint = 'mobile' | 'tablet' | 'desktop';

export interface ResponsiveValue<T> {
  mobile: T;
  tablet: T;
  desktop: T;
}

export const responsiveConfig = {
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

export const ranges: Array<{ value: RangeType; label: string }> = [
  { value: 'W', label: 'W' },
  { value: 'M', label: 'M' },
  { value: '6M', label: '6M' },
  { value: 'Y', label: 'Y' },
];

export function getResponsiveValue<T>(
  values: ResponsiveValue<T>,
  breakpoint: Breakpoint
): T {
  return values[breakpoint];
}

export function useBreakpoint(): Breakpoint {
  const [breakpoint, setBreakpoint] = useState<Breakpoint>(() => {
    if (typeof window === 'undefined') return 'desktop';
    const width = window.innerWidth;
    if (width < 640) return 'mobile';
    if (width < 1024) return 'tablet';
    return 'desktop';
  });

  useEffect(() => {
    const mobileQuery = window.matchMedia('(max-width: 639px)');
    const tabletQuery = window.matchMedia(
      '(min-width: 640px) and (max-width: 1023px)'
    );

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

export function filterDataByDays(
  data: VolumeData[],
  days: number
): VolumeData[] {
  if (days >= data.length) return data;
  return data.slice(-days);
}

export function aggregateToWeekly(data: VolumeData[]): VolumeData[] {
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

export function aggregateToMonthly(data: VolumeData[]): VolumeData[] {
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

export function filterDataByRange(
  data: VolumeData[],
  range: RangeType
): VolumeData[] {
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

export function getRangeLabel(range: RangeType, count: number): string {
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

export function getDateFormat(range: RangeType): string {
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
