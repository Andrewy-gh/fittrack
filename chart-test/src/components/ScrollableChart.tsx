import { type ReactNode, useRef, useEffect, useState } from 'react';

interface ScrollableChartProps {
  children: ReactNode;
  dataLength: number;
  /** Minimum width per bar in pixels (default: 40) */
  barWidth?: number;
  /** Container height in pixels (default: 320 for h-80) */
  height?: number;
}

/**
 * Horizontal scrollable container for charts
 * - Shows fixed number of bars with overflow scroll
 * - Mounts at most recent data (right side)
 * - Includes left/right arrow buttons for non-touch navigation
 */
export function ScrollableChart({
  children,
  dataLength,
  barWidth = 40,
  height = 320
}: ScrollableChartProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);

  // Calculate minimum width needed to show all bars
  const minChartWidth = dataLength * barWidth;

  // Check scroll position and update button states
  const checkScrollPosition = () => {
    const element = scrollRef.current;
    if (!element) return;

    const { scrollLeft, scrollWidth, clientWidth } = element;
    setCanScrollLeft(scrollLeft > 0);
    setCanScrollRight(scrollLeft < scrollWidth - clientWidth - 1);
  };

  // Scroll to right (most recent data) on mount
  useEffect(() => {
    const element = scrollRef.current;
    if (!element) return;

    // Wait for layout to settle
    requestAnimationFrame(() => {
      element.scrollLeft = element.scrollWidth;
      checkScrollPosition();
    });
  }, []);

  // Update button states when data changes
  useEffect(() => {
    checkScrollPosition();
  }, [dataLength]);

  // Scroll by one "page" (container width)
  const scroll = (direction: 'left' | 'right') => {
    const element = scrollRef.current;
    if (!element) return;

    const scrollAmount = element.clientWidth * 0.8; // 80% of container width
    element.scrollBy({
      left: direction === 'left' ? -scrollAmount : scrollAmount,
      behavior: 'smooth',
    });

    // Update button states after scroll animation
    setTimeout(checkScrollPosition, 300);
  };

  return (
    <div className="relative">
      {/* Scroll Container */}
      <div
        ref={scrollRef}
        onScroll={checkScrollPosition}
        className="overflow-x-auto overflow-y-hidden scrollbar-thin scrollbar-thumb-[var(--color-muted)] scrollbar-track-transparent hover:scrollbar-thumb-[var(--color-muted-foreground)]"
        style={{ height: `${height}px` }}
      >
        <div style={{ minWidth: `${minChartWidth}px`, height: '100%' }}>
          {children}
        </div>
      </div>

      {/* Left Scroll Button */}
      {canScrollLeft && (
        <button
          onClick={() => scroll('left')}
          className="absolute left-2 top-1/2 -translate-y-1/2 bg-[var(--color-background)] border border-[var(--color-border)] rounded-full p-2 shadow-lg hover:bg-[var(--color-muted)] transition-colors"
          aria-label="Scroll left"
        >
          <svg
            width="16"
            height="16"
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

      {/* Right Scroll Button */}
      {canScrollRight && (
        <button
          onClick={() => scroll('right')}
          className="absolute right-2 top-1/2 -translate-y-1/2 bg-[var(--color-background)] border border-[var(--color-border)] rounded-full p-2 shadow-lg hover:bg-[var(--color-muted)] transition-colors"
          aria-label="Scroll right"
        >
          <svg
            width="16"
            height="16"
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
