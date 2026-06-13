import { useQuery, useSuspenseQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";

import { AnalyticsDashboard } from "@/features/analytics/components/analytics-dashboard";
import {
  getExerciseDetailQueryOptions,
  getExerciseListQueryOptions,
} from "@/features/exercises/api/exercise-query-options";
import {
  getWorkoutContributionQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/features/workouts/api/workout-query-options";

type AnalyticsPageProps = {
  exerciseId?: number;
  user: CurrentUser | CurrentInternalUser | null;
};

export function AnalyticsPage({ exerciseId, user }: AnalyticsPageProps) {
  const navigate = useNavigate({ from: "/analytics" });
  const isDemoMode = !user;

  const { data: exercises } = useSuspenseQuery(
    getExerciseListQueryOptions(user),
  );

  const selectedExerciseId =
    exerciseId && exercises.some((exercise) => exercise.id === exerciseId)
      ? exerciseId
      : exercises[0]?.id;

  const exerciseDetailQuery = useQuery({
    ...getExerciseDetailQueryOptions(user, selectedExerciseId ?? 0),
    enabled: Boolean(selectedExerciseId),
  });

  const workoutContributionQuery = useQuery(
    getWorkoutContributionQueryOptions(user),
  );
  const workoutFocusValuesQuery = useQuery(getWorkoutsFocusQueryOptions(user));

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
      workoutContributionData={workoutContributionQuery.data}
      workoutFocusValues={workoutFocusValuesQuery.data ?? []}
    />
  );
}
