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
import type { ExerciseWithSets } from '@/lib/types';

interface ChartDataPoint {
  date: string;
  volume: number;
}

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
      <CardContent>
        <div className="h-80 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={dailyVolume}
              margin={{
                top: 5,
                right: 30,
                left: 20,
                bottom: 5,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--muted))" />
              <XAxis
                dataKey="date"
                stroke="hsl(var(--muted-foreground))"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(str) => format(parseISO(str), 'MMM d')}
              />
              <YAxis
                stroke="hsl(var(--muted-foreground))"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                tickFormatter={(value) => `${value}`}
              />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'hsl(var(--background))',
                  borderColor: 'hsl(var(--border))',
                }}
                labelStyle={{ color: 'hsl(var(--foreground))' }}
                labelFormatter={(label) => format(parseISO(label), 'PPP')}
              />
              <Bar dataKey="volume" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
              <Brush
                dataKey="date"
                height={30}
                stroke="hsl(var(--primary))"
                fill="hsl(var(--secondary))"
                tickFormatter={(str) => format(parseISO(str), 'MMM d')}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}
