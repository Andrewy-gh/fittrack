import { useState } from 'react';
import {
  Activity,
  BarChart3,
  Calendar,
  Edit,
  Hash,
  Trash,
  TrendingUp,
  Weight,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChartBarVol } from '@/components/charts/chart-bar-vol';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { formatDate, formatTime } from '@/lib/utils';
import { sortByExerciseAndSetOrder } from '@/lib/utils';
import type { ExerciseExerciseWithSetsResponse } from '@/client';
import { ExerciseDeleteDialog } from '@/routes/exercises/-components/exercise-delete-dialog';

export interface ExerciseDetailProps {
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  exerciseId: number;
}

export function ExerciseDetail({
  exerciseSets,
  exerciseId,
}: ExerciseDetailProps) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  // Calculate summary statistics
  const totalSets = exerciseSets.length;
  const uniqueWorkouts = new Set(exerciseSets.map((set) => set.workout_id))
    .size;
  const weights = exerciseSets.map((set) => set.weight || 0);
  const volumes = exerciseSets.map((set) => set.volume);

  const averageWeight = totalSets > 0
    ? Math.round(weights.reduce((sum, weight) => sum + weight, 0) / weights.length)
    : 0;
  const maxWeight = totalSets > 0 ? Math.max(...weights) : 0;
  const averageVolume = totalSets > 0
    ? Math.round(volumes.reduce((sum, volume) => sum + volume, 0) / volumes.length)
    : 0;
  const maxVolume = totalSets > 0 ? Math.max(...volumes) : 0;

  const sortedExerciseSets = sortByExerciseAndSetOrder(exerciseSets);

  // Group sets by workout while preserving order
  const workoutGroups = sortedExerciseSets.reduce(
    (acc, set) => {
      if (!acc[set.workout_id]) {
        acc[set.workout_id] = {
          date: set.workout_date,
          notes: set.workout_notes || null,
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

  const handleOpenDeleteDialog = () => {
    setIsDeleteDialogOpen(true);
  };

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
          <div className="flex flex-col items-center gap-3 md:flex-row">
            <Button size="sm">
              <Edit className="mr-2 hidden h-4 w-4 md:block" />
              Edit
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={handleOpenDeleteDialog}
              data-testid="delete-exercise-button"
            >
              <Trash className="mr-2 hidden h-4 w-4 md:block" />
              Delete
            </Button>
          </div>
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
        <ChartBarVol data={exerciseSets} />

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
                          {set.set_order ?? index + 1}
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
        {/* MARK: Dialog */}
        <ExerciseDeleteDialog
          isOpen={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          exerciseId={exerciseId}
        />
      </div>
    </main>
  );
}
