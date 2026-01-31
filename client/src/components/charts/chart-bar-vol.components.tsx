import {
  type ReactNode,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from 'react';

import {
  getResponsiveValue,
  ranges,
  responsiveConfig,
  useBreakpoint,
  type RangeType,
} from './chart-bar-vol.utils';

interface RangeSelectorProps {
  selectedRange: RangeType;
  onRangeChange: (range: RangeType) => void;
}

export function RangeSelector({
  selectedRange,
  onRangeChange,
}: RangeSelectorProps) {
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

export function ScrollableChart({
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

    const { scrollLeft, clientWidth } = element;
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
  }, [dataLength, barWidth, height, minChartWidth]);

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
  }, [minChartWidth]);

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
