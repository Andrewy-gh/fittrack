import { Card } from '@/components/ui/card';
import { type LucideIcon } from 'lucide-react';
import { formatWeight } from '@/lib/utils';

export type StatsValue = string | number;

export type StatsCardItem = {
  label: string;
  value: StatsValue;
  icon: LucideIcon;
  valueSuffix?: string;
  hideLabelOnMobile?: boolean;
  labelShort?: string;
  valueFormatter?: (value: StatsValue) => string;
};

export interface StatsGridProps {
  items: StatsCardItem[];
  columns?: 2 | 3 | 4;
}

const defaultFormatter = (value: StatsValue) => String(value);

export function StatsGrid({ items, columns = 2 }: StatsGridProps) {
  const columnClass = columns === 3
    ? 'grid-cols-3'
    : columns === 4
      ? 'grid-cols-4'
      : 'grid-cols-2';

  return (
    <div className={`grid ${columnClass} gap-4`}>
      {items.map((item) => {
        const Icon = item.icon;
        const formattedValue = item.valueFormatter
          ? item.valueFormatter(item.value)
          : defaultFormatter(item.value);
        return (
          <Card className="p-4" key={item.label}>
            <div className="flex items-center gap-2 mb-2">
              <Icon className="w-5 h-5 text-primary" />
              {item.hideLabelOnMobile ? (
                <>
                  <span className="text-sm font-semibold hidden md:inline">
                    {item.label}
                  </span>
                  {item.labelShort && (
                    <span className="text-sm font-semibold md:hidden">
                      {item.labelShort}
                    </span>
                  )}
                </>
              ) : (
                <span className="text-sm font-semibold">
                  {item.label}
                </span>
              )}
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {formattedValue}
              {item.valueSuffix ? ` ${item.valueSuffix}` : ''}
            </div>
          </Card>
        );
      })}
    </div>
  );
}

export const weightFormatter = (value: StatsValue) => {
  if (typeof value === 'number') {
    return formatWeight(value);
  }
  return defaultFormatter(value);
};
