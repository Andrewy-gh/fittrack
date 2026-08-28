import { useEffect, useState } from "react";

export type RangeType = "W" | "M" | "6M" | "Y";

export type Breakpoint = "mobile" | "tablet" | "desktop";

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
  yearBarWidth: {
    mobile: 20,
    tablet: 28,
    desktop: 36,
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
    mobile: "px-3 py-1.5 text-xs",
    tablet: "px-3.5 py-2 text-sm",
    desktop: "px-4 py-2 text-sm",
  },
  containerGap: {
    mobile: "gap-0.5 p-0.5",
    tablet: "gap-1 p-1",
    desktop: "gap-1 p-1",
  },
  scrollButton: {
    mobile: {
      padding: "p-1",
      iconSize: 12,
    },
    tablet: {
      padding: "p-1.5",
      iconSize: 14,
    },
    desktop: {
      padding: "p-1.5",
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
  { value: "W", label: "W" },
  { value: "M", label: "M" },
  { value: "6M", label: "6M" },
  { value: "Y", label: "Y" },
];

/** Returns the shared number of bars shown for a chart range. */
export function getVisibleBarCount(range: RangeType): number | undefined {
  if (range === "W") return 7;
  if (range === "M" || range === "6M") return 26;
  return undefined;
}

export function getResponsiveValue<T>(
  values: ResponsiveValue<T>,
  breakpoint: Breakpoint,
): T {
  return values[breakpoint];
}

export function useBreakpoint(): Breakpoint {
  const [breakpoint, setBreakpoint] = useState<Breakpoint>(() => {
    if (typeof window === "undefined") return "desktop";
    const width = window.innerWidth;
    if (width < 640) return "mobile";
    if (width < 1024) return "tablet";
    return "desktop";
  });

  useEffect(() => {
    const mobileQuery = window.matchMedia("(max-width: 639px)");
    const tabletQuery = window.matchMedia(
      "(min-width: 640px) and (max-width: 1023px)",
    );

    const updateBreakpoint = () => {
      if (mobileQuery.matches) {
        setBreakpoint("mobile");
      } else if (tabletQuery.matches) {
        setBreakpoint("tablet");
      } else {
        setBreakpoint("desktop");
      }
    };

    mobileQuery.addEventListener("change", updateBreakpoint);
    tabletQuery.addEventListener("change", updateBreakpoint);

    return () => {
      mobileQuery.removeEventListener("change", updateBreakpoint);
      tabletQuery.removeEventListener("change", updateBreakpoint);
    };
  }, []);

  return breakpoint;
}
