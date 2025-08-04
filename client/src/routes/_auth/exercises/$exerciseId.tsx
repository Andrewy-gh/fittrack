import { createFileRoute } from '@tanstack/react-router';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Edit,
  Hash,
  Calendar,
  Weight,
  TrendingUp,
  BarChart3,
  Activity,
} from 'lucide-react';
import { formatDate, formatTime } from '@/lib/utils';
import type { ExerciseWithSets } from '@/lib/types';
import { fetchExerciseWithSets } from '@/lib/api/exercises';
// ! TODO - Implement chart component
// import { ChartBarVol } from '@/components/charts/chart-bar-vol';

function ExerciseDisplay({
  exerciseSets,
}: {
  exerciseSets: ExerciseWithSets[];
}) {
  // Calculate summary statistics
  const totalSets = exerciseSets.length;
  const uniqueWorkouts = new Set(exerciseSets.map((set) => set.workout_id))
    .size;
  const weights = exerciseSets.map((set) => set.weight);
  const volumes = exerciseSets.map((set) => set.volume);

  const averageWeight = Math.round(
    weights.reduce((sum, weight) => sum + weight, 0) / weights.length
  );
  const maxWeight = Math.max(...weights);
  const averageVolume = Math.round(
    volumes.reduce((sum, volume) => sum + volume, 0) / volumes.length
  );
  const maxVolume = Math.max(...volumes);

  // Group sets by workout
  const workoutGroups = exerciseSets.reduce(
    (acc, set) => {
      if (!acc[set.workout_id]) {
        acc[set.workout_id] = {
          date: set.workout_date,
          notes: set.workout_notes,
          sets: [],
        };
      }
      acc[set.workout_id].sets.push(set);
      return acc;
    },
    {} as Record<
      number,
      { date: string; notes: string | null; sets: typeof exerciseSets }
    >
  );

  const exerciseName = exerciseSets[0]?.exercise_name || 'Exercise';

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {exerciseName}
            </h1>
          </div>
          <Button size="sm">
            <Edit className="w-4 h-4 mr-2" />
            Edit
          </Button>
        </div>

        {/* MARK: Summary Cards */}
        <div className="grid grid-cols-2 gap-4">
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
              <Calendar className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Workouts</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {uniqueWorkouts}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Weight className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Average Weight</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {averageWeight} lbs
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Max Weight</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {maxWeight} lbs
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold hidden md:inline">
                Average Volume
              </span>
              <span className="text-sm font-semibold md:hidden">
                Avg. Volume
              </span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {averageVolume.toLocaleString()}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Max Volume</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {maxVolume.toLocaleString()}
            </div>
          </Card>
        </div>

        {/* MARK: Workouts */}
        <div className="space-y-4">
          <h2 className="text-xl font-semibold">Workouts</h2>
          {Object.entries(workoutGroups).map(([workoutId, workout]) => {
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
                key={workoutId}
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
                        {workout.notes && (
                          <>
                            <span className="text-muted-foreground">•</span>
                            <Badge
                              variant="outline"
                              className="border-border bg-muted text-xs"
                            >
                              {workout.notes.toUpperCase()}
                            </Badge>
                          </>
                        )}
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
                          {index + 1}
                        </span>
                        <div className="flex items-center space-x-4 text-sm">
                          <span className="font-medium">{set.weight} lbs</span>
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
        </div>
      </div>
    </main>
  );
}

export const Route = createFileRoute('/_auth/exercises/$exerciseId')({
  params: {
    parse: (params) => {
      const exerciseId = parseInt(params.exerciseId, 10);
      if (isNaN(exerciseId) || !Number.isInteger(exerciseId)) {
        throw new Error('Invalid exerciseId');
      }
      return { exerciseId };
    },
  },
  loader: async ({ context, params }) => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const exerciseId = params.exerciseId;
    const exerciseData = await fetchExerciseWithSets(exerciseId, accessToken);
    return exerciseData;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const exerciseSets = Route.useLoaderData();
  return <ExerciseDisplay exerciseSets={exerciseSets} />;
}
