import { type RangeType } from '@/data/mockData';
import { useBreakpoint } from '../hooks/useBreakpoint';
import { responsiveConfig, getResponsiveValue } from '../utils/responsiveConfig';

interface RangeSelectorProps {
  selectedRange: RangeType;
  onRangeChange: (range: RangeType) => void;
}

const ranges: Array<{ value: RangeType; label: string }> = [
  { value: 'W', label: 'W' },
  { value: 'M', label: 'M' },
  { value: '6M', label: '6M' },
  { value: 'Y', label: 'Y' },
];

export function RangeSelector({ selectedRange, onRangeChange }: RangeSelectorProps) {
  const breakpoint = useBreakpoint();
  const containerClasses = getResponsiveValue(responsiveConfig.containerGap, breakpoint);
  const buttonPadding = getResponsiveValue(responsiveConfig.buttonPadding, breakpoint);

  return (
    <div className={`inline-flex bg-[var(--color-secondary)] rounded-[var(--radius-md)] ${containerClasses}`}>
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
