import { useMemo } from 'react';
import {
  Bar,
  BarChart,
  Brush,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { format, parseISO } from 'date-fns';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import type { ExerciseWithSets } from '@/lib/api/exercises';

interface ChartBarVolProps {
  data: ExerciseWithSets[];
}

export function ChartBarVol({ data }: ChartBarVolProps) {
  const dailyVolume = useMemo(() => {
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
  }, [data]);

  return (
    <Card>
      <CardHeader>
        <CardTitle>Daily Volume</CardTitle>
        <CardDescription>Total training volume per day.</CardDescription>
      </CardHeader>
      <CardContent className="px-0">
        {/* Responsive container ensures the chart fits its parent */}
        <div className="h-80 w-full">
          {/* ResponsiveContainer makes the chart responsive */}
          <ResponsiveContainer width="100%" height="100%">
            {/* BarChart is the main container for bar charts */}
            <BarChart
              // Data array containing objects with date and volume properties
              data={dailyVolume}
              // Margin around the chart (not padding, affects axis labels and legends)
              margin={{
                top: 5,
                right: 30, // Extra space for right-side labels
                // left: 20, // Space for Y-axis labels
                bottom: 5, // Space for X-axis labels
              }}
            >
              {/* Grid lines for better readability */}
              <CartesianGrid
                strokeDasharray="3 3" // Dashed lines
                stroke="var(--color-muted)" // Muted color from theme
              />

              {/* X-Axis Configuration */}
              <XAxis
                dataKey="date" // Key in data object for X values
                stroke="hsl(var(--muted-foreground))" // Axis line color
                fontSize={12}
                tickLine={false} // Hide tick lines
                axisLine={false} // Hide axis line
                // Format date ticks (e.g., "Jul 8")
                tickFormatter={(str) => format(parseISO(str), 'MMM d')}
                // Style for axis ticks
                tick={{
                  fill: 'currentColor',
                  // className: 'text-neutral-300',
                }}
              />

              {/* Y-Axis Configuration */}
              <YAxis
                stroke="hsl(var(--muted-foreground))"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                // Format number values (empty string means show as is)
                tickFormatter={(value) => `${value}`}
                tick={{
                  fill: 'currentColor',
                  // className: 'text-neutral-300',
                }}
              />

              {/* Tooltip Configuration */}
              <Tooltip
                cursor={false}
                // Styling for the tooltip container
                contentStyle={{
                  backgroundColor: 'var(--color-background)', // Dark background
                  border: 'var(--border-popover)', // Border color
                  borderRadius: '0.5rem',
                  // color: 'hsl(0 0% 98%)', // Light text
                  boxShadow:
                    '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
                }}
                // Style for individual items in the tooltip
                itemStyle={{
                  color: 'var(--color-foreground)',
                  fontSize: '0.875rem',
                  textTransform: 'capitalize',
                }}
                // Style for the label at the top of the tooltip
                labelStyle={{
                  color: 'var(--color-foreground)',
                  fontSize: '0.875rem',
                  marginBottom: '0.25rem',
                }}
                // Format the date in the tooltip (e.g., "July 8, 2025")
                labelFormatter={(label) => format(parseISO(label), 'PPP')}
              />

              {/* Bar Series */}
              <Bar
                dataKey="volume" // Key in data object for Y values
                fill="var(--color-primary)"
                radius={4} // Rounded top corners
              />

              {/* Brush/Slider for navigating the chart */}
              <Brush
                dataKey="date" // Key to brush on
                height={30} // Height of the brush area
                stroke="var(--color-foreground)" // Color of the brush handles
                fill="var(--color-background)" // Background of the brush area
                // Format dates in the brush
                tickFormatter={(str) => format(parseISO(str), 'MMM d')}
                role="slider"
                className="text-xs"
                // x={50} // move brush 50px from left edge
                // width={380}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
