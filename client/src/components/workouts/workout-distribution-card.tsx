import { Card, CardContent, CardTitle } from '@/components/ui/card';
import type { WorkoutWorkoutResponse } from '@/client';

export interface WorkoutDistributionCardProps {
  workouts: Array<WorkoutWorkoutResponse>;
}

export function WorkoutDistributionCard({
  workouts,
}: WorkoutDistributionCardProps) {
  const workoutFocusValues = workouts.reduce(
    (acc, workout) => {
      if (workout.workout_focus) {
        acc[workout.workout_focus] = (acc[workout.workout_focus] || 0) + 1;
      }
      return acc;
    },
    {} as Record<string, number>
  );

  return (
    <Card className="p-4">
      <CardTitle className="text-xl font-semibold">
        Workout Distribution
      </CardTitle>
      <CardContent className="px-0">
        <div className="flex flex-wrap gap-4">
          {Object.entries(workoutFocusValues).map(([type, count]) => (
            <div key={type} className="text-center p-4 rounded-xl">
              <p className="font-semibold text-lg">{count}</p>
              <p className="text-sm uppercase">{type}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
