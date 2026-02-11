import { useMemo } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipProps,
} from 'recharts';
import { format, parseISO } from 'date-fns';

import {
  getDateFormat,
  getResponsiveValue,
  responsiveConfig,
  useBreakpoint,
  type RangeType,
} from './chart-bar-vol.utils';
import { ScrollableChart } from './chart-bar-vol.components';

export type MetricPoint = {
  x: string;
  date: string; // ISO date (bucket/workout day)
  workout_id?: number;
  value: number;
};

type Unit = 'lb' | '%' | 'vol';

function formatValue(unit: Unit, value: number) {
  if (unit === '%') return `${value.toFixed(1)}%`;
  if (unit === 'vol') return `${Math.round(value).toLocaleString()} vol`;
  return `${Math.round(value).toLocaleString()} lb`;
}

export function ChartBarMetric({
  title,
  description,
  range,
  data,
  unit,
  barColorVar = 'var(--color-primary)',
  onWorkoutClick,
}: {
  title: string;
  description?: string;
  range: RangeType;
  data: MetricPoint[];
  unit: Unit;
  barColorVar?: string;
  onWorkoutClick?: (workoutId: number) => void;
}) {
  const breakpoint = useBreakpoint();

  const barWidth = getResponsiveValue(responsiveConfig.barWidth, breakpoint);
  const yAxisWidth = getResponsiveValue(responsiveConfig.yAxisWidth, breakpoint);
  const chartMargins = getResponsiveValue(responsiveConfig.chartMargins, breakpoint);
  const axisMargins = { ...chartMargins, left: 0, right: 0 };
  const plotMargins = { ...chartMargins, left: 0 };
  const chartHeight = 280;

  const dateByX = useMemo(() => {
    const m = new Map<string, string>();
    for (const p of data) m.set(p.x, p.date);
    return m;
  }, [data]);

  const tooltipLabelFormatter: TooltipProps<number, string>['labelFormatter'] = (x) => {
    const date = dateByX.get(String(x));
    if (!date) return '';
    const dateFormat = range === 'Y' ? 'MMM yyyy' : 'PPP';
    return format(parseISO(date), dateFormat);
  };

  return (
    <section className="space-y-3">
      <div>
        <h3 className="text-lg font-semibold">{title}</h3>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </div>

      <div
        className="grid gap-2"
        style={{ gridTemplateColumns: `${yAxisWidth}px 1fr` }}
      >
        <div className="min-w-0" style={{ width: yAxisWidth, height: chartHeight }}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data} margin={axisMargins}>
              <XAxis dataKey="x" tick={false} tickLine={false} axisLine={false} />
              <YAxis
                dataKey="value"
                width={yAxisWidth}
                stroke="var(--color-muted-foreground)"
                fontSize={getResponsiveValue(responsiveConfig.fontSize, breakpoint)}
                tickLine={false}
                axisLine={false}
                tickFormatter={(v) => (unit === '%' ? `${v}%` : `${v}`)}
                tick={{ fill: 'var(--color-foreground)' }}
                domain={[0, 'dataMax']}
                tickCount={5}
              />
              <Bar dataKey="value" fill="transparent" stroke="transparent" isAnimationActive={false} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="min-w-0">
          <ScrollableChart
            dataLength={data.length}
            barWidth={barWidth}
            height={chartHeight}
            resetKey={range}
          >
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data} margin={plotMargins}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-muted)" />

                <XAxis
                  dataKey="x"
                  stroke="var(--color-muted-foreground)"
                  fontSize={getResponsiveValue(responsiveConfig.fontSize, breakpoint)}
                  tickLine={false}
                  axisLine={false}
                  tickFormatter={(x) => {
                    const date = dateByX.get(String(x));
                    if (!date) return '';
                    return format(parseISO(date), getDateFormat(range));
                  }}
                  tick={{ fill: 'var(--color-foreground)' }}
                />

                <YAxis dataKey="value" hide width={0} domain={[0, 'dataMax']} tickCount={5} />

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
                  labelFormatter={tooltipLabelFormatter}
                  formatter={(v) => formatValue(unit, Number(v))}
                />

                <Bar
                  dataKey="value"
                  fill={barColorVar}
                  radius={4}
                  isAnimationActive={false}
                  onClick={(p) => {
                    const workoutId = (p?.payload as MetricPoint | undefined)?.workout_id;
                    if (workoutId && onWorkoutClick) onWorkoutClick(workoutId);
                  }}
                />
              </BarChart>
            </ResponsiveContainer>
          </ScrollableChart>
        </div>
      </div>
    </section>
  );
}
