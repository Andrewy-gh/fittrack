import { Dumbbell, Hash, RotateCcw, Weight } from 'lucide-react';
import { StatsGrid, weightFormatter } from '@/components/stats-grid';

export interface WorkoutDetailSummaryCardsProps {
  uniqueExercises: number;
  totalSets: number;
  totalReps: number;
  totalVolume: number;
}

export function WorkoutDetailSummaryCards({
  uniqueExercises,
  totalSets,
  totalReps,
  totalVolume,
}: WorkoutDetailSummaryCardsProps) {
  return (
    <StatsGrid
      items={[
        {
          label: 'Exercises',
          value: uniqueExercises,
          icon: Dumbbell,
        },
        {
          label: 'Total Sets',
          value: totalSets,
          icon: Hash,
        },
        {
          label: 'Total Reps',
          value: totalReps,
          icon: RotateCcw,
        },
        {
          label: 'Volume',
          value: totalVolume,
          icon: Weight,
          valueFormatter: weightFormatter,
        },
      ]}
    />
  );
}
