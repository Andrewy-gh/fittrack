import { Activity, BarChart3, Calendar, Hash, TrendingUp, Weight } from 'lucide-react';
import { StatsGrid, weightFormatter } from '@/components/stats-grid';

export interface ExerciseSummaryCardsProps {
  totalSets: number;
  uniqueWorkouts: number;
  averageWeight: number;
  maxWeight: number;
  averageVolume: number;
  maxVolume: number;
}

export function ExerciseSummaryCards({
  totalSets,
  uniqueWorkouts,
  averageWeight,
  maxWeight,
  averageVolume,
  maxVolume,
}: ExerciseSummaryCardsProps) {
  return (
    <StatsGrid
      items={[
        {
          label: 'Total Sets',
          value: totalSets,
          icon: Hash,
        },
        {
          label: 'Workouts',
          value: uniqueWorkouts,
          icon: Calendar,
        },
        {
          label: 'Average Weight',
          value: averageWeight,
          icon: Weight,
          valueSuffix: 'lbs',
          valueFormatter: weightFormatter,
        },
        {
          label: 'Max Weight',
          value: maxWeight,
          icon: TrendingUp,
          valueSuffix: 'lbs',
          valueFormatter: weightFormatter,
        },
        {
          label: 'Average Volume',
          labelShort: 'Avg. Volume',
          value: averageVolume.toLocaleString(),
          icon: BarChart3,
          hideLabelOnMobile: true,
        },
        {
          label: 'Max Volume',
          value: maxVolume.toLocaleString(),
          icon: Activity,
        },
      ]}
    />
  );
}
