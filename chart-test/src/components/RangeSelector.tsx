import { type RangeType } from '@/data/mockData';

interface RangeSelectorProps {
  selectedRange: RangeType;
  onRangeChange: (range: RangeType) => void;
}

const ranges: Array<{ value: RangeType; label: string }> = [
  { value: 'D', label: 'D' },
  { value: 'W', label: 'W' },
  { value: 'M', label: 'M' },
  { value: '6M', label: '6M' },
  { value: 'Y', label: 'Y' },
];

export function RangeSelector({ selectedRange, onRangeChange }: RangeSelectorProps) {
  return (
    <div className="inline-flex bg-[var(--color-secondary)] rounded-[var(--radius-md)] p-1 gap-1">
      {ranges.map(({ value, label }) => (
        <button
          key={value}
          onClick={() => onRangeChange(value)}
          className={`
            px-4 py-2 rounded-[var(--radius-sm)] border-0 text-sm font-medium
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
