import { Calendar, Dumbbell } from 'lucide-react';
import { StatsGrid } from '@/components/stats-grid';
import type { WorkoutWorkoutResponse } from '@/client';

export interface WorkoutSummaryCardsProps {
  workouts: Array<WorkoutWorkoutResponse>;
}

export function WorkoutSummaryCards({ workouts }: WorkoutSummaryCardsProps) {
  const totalWorkouts = workouts.length;
  const thisWeekWorkouts = workouts.filter((workout) => {
    const workoutDate = new Date(workout.date);
    const sixDaysAgo = new Date();
    sixDaysAgo.setDate(sixDaysAgo.getDate() - 6);
    sixDaysAgo.setHours(0, 0, 0, 0);
    return workoutDate >= sixDaysAgo;
  }).length;

  return (
    <StatsGrid
      items={[
        {
          label: 'Total Workouts',
          value: totalWorkouts,
          icon: Dumbbell,
        },
        {
          label: 'This Week',
          value: thisWeekWorkouts,
          icon: Calendar,
        },
      ]}
    />
  );
}
