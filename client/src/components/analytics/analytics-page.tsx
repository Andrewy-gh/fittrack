import type {
  ExerciseExerciseResponse,
  ExerciseExerciseWithSetsResponse,
  WorkoutContributionDataResponse,
} from '@/client';
import { ExerciseMetricCharts } from '@/components/exercises/exercise-metric-charts';
import { GenericCombobox } from '@/components/generic-combobox';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { WorkoutContributionGraph } from '@/components/workouts/workout-contribution-graph';
import { getWorkoutSummary } from '@/lib/analytics';
import { AnalyticsSummaryCards } from './analytics-summary-cards';
import { WorkoutVolumeChart } from './workout-volume-chart';

export interface AnalyticsPageProps {
  isLoadingExercises: boolean;
  exercises: ExerciseExerciseResponse[];
  selectedExerciseId?: number;
  onSelectExercise: (id: number) => void;
  isLoadingDetails: boolean;
  exerciseSets?: ExerciseExerciseWithSetsResponse[];
  isDemoMode: boolean;
  workoutContributionData?: WorkoutContributionDataResponse;
  workoutFocusValues?: string[];
}

export function AnalyticsPage({
  isLoadingExercises,
  exercises,
  selectedExerciseId,
  onSelectExercise,
  isLoadingDetails,
  exerciseSets = [],
  isDemoMode,
  workoutContributionData,
  workoutFocusValues = [],
}: AnalyticsPageProps) {
  const summary = getWorkoutSummary(workoutContributionData?.days);

  if (isLoadingExercises) {
    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          <div className="pt-4">
            <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
          </div>

          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Loading analytics...
            </CardContent>
          </Card>
        </div>
      </main>
    );
  }

  if (exercises.length === 0) {
    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          <div className="space-y-2 pt-4">
            <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
            <p className="text-sm text-muted-foreground">
              Review workout consistency and exercise progress in one place.
            </p>
          </div>

          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Add an exercise first to see analytics.
            </CardContent>
          </Card>
        </div>
      </main>
    );
  }

  const selectedExerciseName =
    exercises.find((exercise) => exercise.id === selectedExerciseId)?.name ??
    '';

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        <div className="space-y-2 pt-4">
          <h1 className="text-3xl font-bold tracking-tight">Analytics</h1>
          <p className="text-sm text-muted-foreground">
            Review workout consistency and exercise progress in one place.
          </p>
        </div>

        <AnalyticsSummaryCards summary={summary} />

        <Card>
          <CardHeader>
            <CardTitle>Exercise Progress</CardTitle>
            <CardDescription>
              Pick an exercise to inspect how session metrics move over time.
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            <GenericCombobox
              options={exercises}
              selected={selectedExerciseName}
              ariaLabel="Exercise options"
              inputAriaLabel="Search exercises"
              placeholder="Select an exercise"
              onChange={(exercise) => onSelectExercise(exercise.id)}
            />
          </CardContent>
        </Card>

        {isLoadingDetails || !selectedExerciseId ? (
          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Loading exercise metrics...
            </CardContent>
          </Card>
        ) : (
          <ExerciseMetricCharts
            exerciseId={selectedExerciseId}
            exerciseSets={exerciseSets}
            isDemoMode={isDemoMode}
          />
        )}

        {workoutContributionData ? (
          <WorkoutVolumeChart
            data={workoutContributionData}
            focusValues={workoutFocusValues}
          />
        ) : (
          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Loading workout volume...
            </CardContent>
          </Card>
        )}

        {workoutContributionData ? (
          <WorkoutContributionGraph data={workoutContributionData} defaultOpen />
        ) : (
          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Loading workout trends...
            </CardContent>
          </Card>
        )}
      </div>
    </main>
  );
}
