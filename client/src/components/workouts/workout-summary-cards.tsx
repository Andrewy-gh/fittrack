import { Calendar, BarChart3, Dumbbell, Scale } from 'lucide-react';
import { StatsGrid } from '@/components/stats-grid';
import type { WorkoutWorkoutResponse } from '@/client';
import {
  formatWeekComparison,
  getWorkoutConsistencySummary,
} from '@/lib/workout-insights';

export interface WorkoutSummaryCardsProps {
  workouts: Array<WorkoutWorkoutResponse>;
}

export function WorkoutSummaryCards({ workouts }: WorkoutSummaryCardsProps) {
  const summary = getWorkoutConsistencySummary(workouts);

  return (
    <StatsGrid
      columns={4}
      items={[
        {
          label: 'Total Workouts',
          value: summary.totalWorkouts,
          icon: Dumbbell,
        },
        {
          label: 'This Week',
          value: summary.workoutsThisWeek,
          icon: Calendar,
          helperText: formatWeekComparison(
            summary.workoutsThisWeek,
            summary.workoutsLastWeek
          ),
        },
        {
          label: 'Active Days',
          value: summary.activeDaysThisMonth,
          icon: BarChart3,
          helperText: 'This month',
        },
        {
          label: 'Avg / Week',
          value: summary.averageWorkoutsPerWeek,
          icon: Scale,
          helperText: 'Rolling 8 weeks',
        },
      ]}
    />
  );
}
