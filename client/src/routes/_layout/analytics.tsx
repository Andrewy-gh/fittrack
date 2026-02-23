import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';

import { ExerciseMetricCharts } from '@/components/exercises/exercise-metric-charts';
import { WorkoutContributionGraph } from '@/components/workouts/workout-contribution-graph';
import { Card, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
} from '@/lib/api/exercises';
import { contributionDataQueryOptions } from '@/lib/api/workouts';
import {
  getDemoExercisesByIdQueryOptions,
  getDemoExercisesQueryOptions,
  getDemoWorkoutsQueryOptions,
} from '@/lib/demo-data/query-options';
import type { WorkoutContributionDataResponse, WorkoutWorkoutResponse } from '@/client';

const analyticsSearchSchema = z.object({
  exerciseId: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute('/_layout/analytics')({
  validateSearch: analyticsSearchSchema,
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  return user ? <AuthedAnalytics /> : <DemoAnalytics />;
}

function AuthedAnalytics() {
  const { exerciseId } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const exercisesQuery = useQuery(exercisesQueryOptions());
  const exercises = exercisesQuery.data ?? [];

  const selectedExerciseId =
    exerciseId && exercises.some((e) => e.id === exerciseId)
      ? exerciseId
      : exercises[0]?.id;

  const exerciseDetailQuery = useQuery({
    ...exerciseByIdQueryOptions(selectedExerciseId ?? 0),
    enabled: Boolean(selectedExerciseId),
  });
  const contributionQuery = useQuery(contributionDataQueryOptions());

  return (
    <AnalyticsLayout
      isLoadingExercises={exercisesQuery.isLoading}
      exercises={exercises}
      selectedExerciseId={selectedExerciseId}
      onSelectExercise={(id) =>
        navigate({ search: (prev) => ({ ...prev, exerciseId: id }) })
      }
      isLoadingDetails={exerciseDetailQuery.isLoading}
      exerciseSets={exerciseDetailQuery.data?.sets}
      isDemoMode={false}
      workoutContributionData={contributionQuery.data}
    />
  );
}

function DemoAnalytics() {
  const { exerciseId } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const exercisesQuery = useQuery(getDemoExercisesQueryOptions());
  const exercises = exercisesQuery.data ?? [];

  const selectedExerciseId =
    exerciseId && exercises.some((e) => e.id === exerciseId)
      ? exerciseId
      : exercises[0]?.id;

  const exerciseDetailQuery = useQuery({
    ...getDemoExercisesByIdQueryOptions(selectedExerciseId ?? 0),
    enabled: Boolean(selectedExerciseId),
  });
  const demoWorkoutsQuery = useQuery(getDemoWorkoutsQueryOptions());

  return (
    <AnalyticsLayout
      isLoadingExercises={exercisesQuery.isLoading}
      exercises={exercises}
      selectedExerciseId={selectedExerciseId}
      onSelectExercise={(id) =>
        navigate({ search: (prev) => ({ ...prev, exerciseId: id }) })
      }
      isLoadingDetails={exerciseDetailQuery.isLoading}
      exerciseSets={exerciseDetailQuery.data?.sets}
      isDemoMode
      workoutContributionData={buildDemoContributionData(demoWorkoutsQuery.data ?? [])}
    />
  );
}

function buildDemoContributionData(workouts: WorkoutWorkoutResponse[]): WorkoutContributionDataResponse {
  const byDate = new Map<string, WorkoutWorkoutResponse[]>();

  for (const workout of workouts) {
    const day = (workout.date || '').split('T')[0];
    if (!day) continue;
    const list = byDate.get(day) ?? [];
    list.push(workout);
    byDate.set(day, list);
  }

  const days = Array.from(byDate.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([date, items]) => {
      const count = items.length;
      const level = count === 0 ? 0 : count < 2 ? 1 : count < 3 ? 2 : count < 4 ? 3 : 4;

      return {
        date,
        count,
        level,
        workouts: items.map((w) => ({
          id: w.id,
          focus: w.workout_focus,
          time: w.date,
        })),
      };
    });

  return { days };
}

function AnalyticsLayout({
  isLoadingExercises,
  exercises,
  selectedExerciseId,
  onSelectExercise,
  isLoadingDetails,
  exerciseSets,
  isDemoMode,
  workoutContributionData,
}: {
  isLoadingExercises: boolean;
  exercises: Array<{ id: number; name: string }>;
  selectedExerciseId?: number;
  onSelectExercise: (id: number) => void;
  isLoadingDetails: boolean;
  exerciseSets?: unknown;
  isDemoMode: boolean;
  workoutContributionData?: WorkoutContributionDataResponse;
}) {
  const safeSets = Array.isArray(exerciseSets) ? exerciseSets : [];

  if (isLoadingExercises) {
    return (
      <main className="max-w-lg mx-auto px-4 py-6">
        <p className="text-sm text-muted-foreground">Loading analytics…</p>
      </main>
    );
  }

  if (exercises.length === 0) {
    return (
      <main className="max-w-lg mx-auto px-4 py-6">
        <h1 className="text-2xl font-semibold">Analytics</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Add an exercise first to see analytics.
        </p>
      </main>
    );
  }

  return (
    <main className="max-w-lg mx-auto space-y-6 px-4 pb-8">
      <section className="space-y-2 pt-4">
        <h1 className="text-2xl font-semibold">Analytics</h1>
        <p className="text-sm text-muted-foreground">
          Track performance trends per exercise.
        </p>
      </section>

      <Card>
        <CardContent className="pt-6 space-y-2">
          <Label htmlFor="exercise-select">Exercise</Label>
          <Select
            value={String(selectedExerciseId)}
            onValueChange={(value) => onSelectExercise(Number(value))}
          >
            <SelectTrigger id="exercise-select">
              <SelectValue placeholder="Select an exercise" />
            </SelectTrigger>
            <SelectContent>
              {exercises.map((exercise) => (
                <SelectItem key={exercise.id} value={String(exercise.id)}>
                  {exercise.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      {isLoadingDetails || !selectedExerciseId ? (
        <p className="text-sm text-muted-foreground">Loading graph data…</p>
      ) : (
        <ExerciseMetricCharts
          exerciseId={selectedExerciseId}
          exerciseSets={safeSets as any}
          isDemoMode={isDemoMode}
        />
      )}

      <section className="space-y-3">
        <div>
          <h2 className="text-xl font-semibold">Workout Trends</h2>
          <p className="text-sm text-muted-foreground">
            Weekly activity heatmap from your logged workouts.
          </p>
        </div>

        {workoutContributionData ? (
          <WorkoutContributionGraph data={workoutContributionData} defaultOpen />
        ) : (
          <p className="text-sm text-muted-foreground">Loading workout trends…</p>
        )}
      </section>
    </main>
  );
}
