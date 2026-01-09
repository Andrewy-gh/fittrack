import { Card } from '@/components/ui/card';
import { Calendar, Dumbbell } from 'lucide-react';
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
    <div className="grid grid-cols-2 gap-4">
      <Card className="p-4">
        <div className="flex items-center gap-2 mb-2">
          <Dumbbell className="w-5 h-5 text-primary" />
          <span className="text-sm font-semibold">Total Workouts</span>
        </div>
        <div className="text-2xl text-card-foreground font-bold">
          {totalWorkouts}
        </div>
      </Card>

      <Card className="p-4">
        <div className="flex items-center gap-2 mb-2">
          <Calendar className="w-5 h-5 text-primary" />
          <span className="text-sm font-semibold">This Week</span>
        </div>
        <div className="text-2xl text-card-foreground font-bold">
          {thisWeekWorkouts}
        </div>
      </Card>
    </div>
  );
}
