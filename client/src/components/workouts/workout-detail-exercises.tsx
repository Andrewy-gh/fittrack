import { Link } from '@tanstack/react-router';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { formatWeight, sortByExerciseAndSetOrder } from '@/lib/utils';
import type { WorkoutWorkoutWithSetsResponse } from '@/client';

export interface WorkoutDetailExercisesProps {
  workout: WorkoutWorkoutWithSetsResponse[];
}

export function WorkoutDetailExercises({ workout }: WorkoutDetailExercisesProps) {
  const sortedWorkouts = sortByExerciseAndSetOrder(workout);

  const exerciseGroups = sortedWorkouts.reduce(
    (acc, w) => {
      const exerciseId = w.exercise_id || 0;
      const exerciseOrder = w.exercise_order ?? w.exercise_id ?? 0;

      if (!acc[exerciseId]) {
        acc[exerciseId] = {
          name: w.exercise_name || 'Unknown Exercise',
          sets: [],
          order: exerciseOrder,
        };
      }
      acc[exerciseId].sets.push(w);
      return acc;
    },
    {} as Record<
      number,
      { name: string; sets: WorkoutWorkoutWithSetsResponse[]; order: number }
    >
  );

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-semibold">Exercises</h2>
      {Object.entries(exerciseGroups)
        .sort(([, a], [, b]) => a.order - b.order)
        .map(([exerciseId, exercise]) => {
          const exerciseReps = exercise.sets.reduce(
            (sum, set) => sum + (set.reps || 0),
            0
          );
          const exerciseVolume = exercise.sets.reduce(
            (sum, set) => sum + (set.volume || 0),
            0
          );

          return (
            <Card
              key={exerciseId}
              className="border-0 shadow-sm backdrop-blur-sm"
              data-testid="workout-detail-exercise-card"
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg font-semibold">
                    <Link
                      to="/exercises/$exerciseId"
                      params={{ exerciseId: Number(exerciseId) }}
                    >
                      {exercise.name}
                    </Link>
                  </CardTitle>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span>{exerciseReps} reps</span>
                    <span className="text-primary">
                      {formatWeight(exerciseVolume)} vol
                    </span>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-2">
                {exercise.sets.map((set, index) => (
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
                          {formatWeight(set.weight || 0)} lbs
                        </span>
                        <span>&times;</span>
                        <span className="font-medium">
                          {set.reps || 0} reps
                        </span>
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {formatWeight(set.volume || 0)} vol
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          );
        })}
    </div>
  );
}
