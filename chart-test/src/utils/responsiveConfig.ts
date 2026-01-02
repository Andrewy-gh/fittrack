import type { Breakpoint } from '../hooks/useBreakpoint';

interface ResponsiveValue<T> {
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
  nivoMargins: {
    mobile: { top: 10, right: 10, bottom: 40, left: 40 },
    tablet: { top: 15, right: 15, bottom: 45, left: 45 },
    desktop: { top: 20, right: 20, bottom: 50, left: 48 },
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
      padding: 'p-1.5',
      iconSize: 14,
      position: 'left-1 right-1',
    },
    tablet: {
      padding: 'p-2',
      iconSize: 16,
      position: 'left-2 right-2',
    },
    desktop: {
      padding: 'p-2',
      iconSize: 16,
      position: 'left-2 right-2',
    },
  },
  yAxisWidth: {
    mobile: 40,
    tablet: 45,
    desktop: 48,
  },
};

export function getResponsiveValue<T>(
  values: ResponsiveValue<T>,
  breakpoint: Breakpoint
): T {
  return values[breakpoint];
}
