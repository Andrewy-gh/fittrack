import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Edit, Dumbbell, Hash, RotateCcw, Weight } from 'lucide-react';
import { getAccessToken } from '@/lib/api/auth';
import { formatDate, formatTime } from '@/lib/utils';
import { workoutByIdQueryOptions } from '@/lib/api/workouts';
import type { workout_WorkoutWithSetsResponse } from '@/generated';

function IndividualWorkoutPage({
  workout,
}: {
  workout: workout_WorkoutWithSetsResponse[];
}) {
  // Calculate summary statistics
  const uniqueExercises = new Set(workout.map((w) => w.exercise_id)).size;
  const totalSets = workout.length;
  const totalReps = workout.reduce((sum, w) => sum + (w.reps || 0), 0);
  const totalVolume = workout.reduce((sum, w) => sum + (w.volume || 0), 0);

  // Group exercises
  const exerciseGroups = workout.reduce(
    (acc, w) => {
      const exerciseId = w.exercise_id || 0;
      if (!acc[exerciseId]) {
        acc[exerciseId] = {
          name: w.exercise_name || 'Unknown Exercise',
          sets: [],
        };
      }
      acc[exerciseId].sets.push(w);
      return acc;
    },
    {} as Record<number, { name: string; sets: typeof workout }>
  );

  const workoutDate = workout[0]?.workout_date;
  const workoutNotes = workout[0]?.workout_notes;

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {formatDate(workoutDate)}
            </h1>
            <div className="flex items-center gap-2 mt-1">
              <p className="text-muted-foreground">{formatTime(workoutDate)}</p>
              {workoutNotes && (
                <>
                  <span className="text-muted-foreground">•</span>
                  <Badge
                    variant="outline"
                    className="border-border bg-muted text-xs"
                  >
                    {workoutNotes.toUpperCase()}
                  </Badge>
                </>
              )}
            </div>
          </div>
          <Button size="sm" asChild>
            <Link
              to="/workouts/$workoutId/edit"
              params={{ workoutId: workout[0]?.workout_id }}
            >
              <Edit className="w-4 h-4 mr-2" />
              Edit
            </Link>
          </Button>
        </div>

        {/* MARK: Summary Cards */}
        <div className="grid grid-cols-2 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Dumbbell className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Exercises</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {uniqueExercises}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Hash className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Total Sets</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalSets}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <RotateCcw className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Total Reps</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalReps}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Weight className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Volume</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalVolume.toLocaleString()}
            </div>
          </Card>
        </div>

        {/* MARK: Exercises */}
        <div className="space-y-4">
          <h2 className="text-2xl font-semibold">Exercises</h2>
          {Object.entries(exerciseGroups).map(([exerciseId, exercise]) => {
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
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg font-semibold">
                      <Link
                        to={`/exercises/$exerciseId`}
                        params={{ exerciseId: Number(exerciseId) }}
                      >
                        {exercise.name}
                      </Link>
                    </CardTitle>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>{exerciseReps} reps</span>
                      <span className="text-primary">
                        {exerciseVolume.toLocaleString()} vol
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
                          {index + 1}
                        </span>
                        <div className="flex items-center space-x-4 text-sm">
                          <span className="font-medium">{set.weight || 0} lbs</span>
                          <span>&times;</span>
                          <span className="font-medium">{set.reps || 0} reps</span>
                        </div>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {(set.volume || 0).toLocaleString()} vol
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </main>
  );
}

export const Route = createFileRoute('/_auth/workouts/$workoutId/')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({ context, params }) => {
    const accessToken = await getAccessToken(context.user);
    const workoutId = params.workoutId;
    context.queryClient.ensureQueryData(
      workoutByIdQueryOptions(workoutId, accessToken)
    );
    return { accessToken, workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { accessToken, workoutId } = Route.useLoaderData();
  const { data: workout } = useSuspenseQuery(
    workoutByIdQueryOptions(workoutId, accessToken)
  );
  return <IndividualWorkoutPage workout={workout} />;
}
