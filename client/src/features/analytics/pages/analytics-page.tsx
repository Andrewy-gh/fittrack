import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { AnalyticsDashboard } from "@/features/analytics/components/analytics-dashboard";
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
} from "@/features/exercises/api/exercises";
import {
  contributionDataQueryOptions,
  workoutsFocusValuesQueryOptions,
} from "@/features/workouts/api/workouts";
import {
  getDemoContributionDataQueryOptions,
  getDemoExercisesByIdQueryOptions,
  getDemoExercisesQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from "@/lib/demo-data/query-options";

type AnalyticsPageProps = {
  exerciseId?: number;
  isDemoMode: boolean;
};

export function AnalyticsPage({ exerciseId, isDemoMode }: AnalyticsPageProps) {
  const navigate = useNavigate({ from: "/analytics" });

  const { data: exercises } = isDemoMode
    ? useSuspenseQuery(getDemoExercisesQueryOptions())
    : useSuspenseQuery(exercisesQueryOptions());

  const selectedExerciseId =
    exerciseId && exercises.some((exercise) => exercise.id === exerciseId)
      ? exerciseId
      : exercises[0]?.id;

  const exerciseDetailQuery = isDemoMode
    ? useQuery({
        ...getDemoExercisesByIdQueryOptions(selectedExerciseId ?? 0),
        enabled: Boolean(selectedExerciseId),
      })
    : useQuery({
        ...exerciseByIdQueryOptions(selectedExerciseId ?? 0),
        enabled: Boolean(selectedExerciseId),
      });

  const authedContributionQuery = useQuery({
    ...contributionDataQueryOptions(),
    enabled: !isDemoMode,
  });
  const demoContributionQuery = useQuery({
    ...getDemoContributionDataQueryOptions(),
    enabled: isDemoMode,
  });
  const authedFocusValuesQuery = useQuery({
    ...workoutsFocusValuesQueryOptions(),
    enabled: !isDemoMode,
  });
  const demoFocusValuesQuery = useQuery({
    ...getDemoWorkoutsFocusValuesQueryOptions(),
    enabled: isDemoMode,
  });

  const workoutContributionData = isDemoMode
    ? demoContributionQuery.data
    : authedContributionQuery.data;
  const workoutFocusValues = isDemoMode
    ? (demoFocusValuesQuery.data ?? [])
    : (authedFocusValuesQuery.data ?? []);

  return (
    <AnalyticsDashboard
      isLoadingExercises={false}
      exercises={exercises}
      selectedExerciseId={selectedExerciseId}
      onSelectExercise={(id) =>
        navigate({ search: (prev) => ({ ...prev, exerciseId: id }) })
      }
      isLoadingDetails={exerciseDetailQuery.isLoading}
      exerciseSets={exerciseDetailQuery.data?.sets}
      isDemoMode={isDemoMode}
      workoutContributionData={workoutContributionData}
      workoutFocusValues={workoutFocusValues}
    />
  );
}
