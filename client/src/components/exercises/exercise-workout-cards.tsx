import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { formatDate, formatTime, formatWeight } from '@/lib/utils';
import type { ExerciseExerciseWithSetsResponse } from '@/client';

export type ExerciseWorkoutEntry = {
  workoutId: number;
  date: string;
  notes: string | null;
  sets: ExerciseExerciseWithSetsResponse[];
};

export interface ExerciseWorkoutCardsProps {
  workouts: ExerciseWorkoutEntry[];
}

export function ExerciseWorkoutCards({ workouts }: ExerciseWorkoutCardsProps) {
  if (workouts.length === 0) {
    return (
      <Card>
        <CardContent className="py-6 text-sm text-muted-foreground">
          No workouts logged for this exercise yet.
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      {workouts.map((workout) => {
        const exerciseReps = workout.sets.reduce(
          (sum, set) => sum + set.reps,
          0
        );
        const exerciseVolume = workout.sets.reduce(
          (sum, set) => sum + set.volume,
          0
        );
        return (
          <Card
            key={workout.workoutId}
            className="border-0 shadow-sm backdrop-blur-sm"
          >
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-lg font-semibold">
                    {formatDate(workout.date)}
                  </CardTitle>
                  <div className="flex items-center gap-2 mt-1">
                    <p className="text-sm text-muted-foreground">
                      {formatTime(workout.date)}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>{exerciseReps} reps</span>
                  <span className="text-primary">
                    {exerciseVolume.toLocaleString()} vol
                  </span>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-2">
              {workout.sets.map((set, index) => (
                <div
                  key={set.set_id}
                  className="flex items-center justify-between py-2 px-3 rounded-lg bg-muted/50"
                >
                  <div className="flex items-center space-x-4">
                    <span className="text-sm font-medium text-muted-foreground w-8">
                      {set.set_order ?? index + 1}
                    </span>
                    <div className="flex items-center space-x-4 text-sm">
                      <span className="font-medium">
                        {formatWeight(set.weight)} lbs
                      </span>
                      <span>&times;</span>
                      <span className="font-medium">{set.reps} reps</span>
                    </div>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {set.volume.toLocaleString()} vol
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        );
      })}
    </>
  );
}
