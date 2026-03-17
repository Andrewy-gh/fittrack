import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useQuery, useSuspenseQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { AnalyticsPage } from '@/components/analytics/analytics-page';
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
} from '@/lib/api/exercises';
import { contributionDataQueryOptions } from '@/lib/api/workouts';
import { buildDemoContributionData } from '@/lib/analytics';
import {
  getDemoExercisesByIdQueryOptions,
  getDemoExercisesQueryOptions,
  getDemoWorkoutsQueryOptions,
} from '@/lib/demo-data/query-options';
import { clearDemoData, initializeDemoData } from '@/lib/demo-data/storage';

const analyticsSearchSchema = z.object({
  exerciseId: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute('/_layout/analytics')({
  validateSearch: analyticsSearchSchema,
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      clearDemoData();
      await Promise.all([
        context.queryClient.ensureQueryData(exercisesQueryOptions()),
        context.queryClient.ensureQueryData(contributionDataQueryOptions()),
      ]);
    } else {
      initializeDemoData();
      await Promise.all([
        context.queryClient.ensureQueryData(getDemoExercisesQueryOptions()),
        context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions()),
      ]);
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId } = Route.useSearch();
  const { user } = Route.useRouteContext();
  const navigate = useNavigate({ from: Route.fullPath });

  const { data: exercises } = user
    ? useSuspenseQuery(exercisesQueryOptions())
    : useSuspenseQuery(getDemoExercisesQueryOptions());

  const selectedExerciseId =
    exerciseId && exercises.some((exercise) => exercise.id === exerciseId)
      ? exerciseId
      : exercises[0]?.id;

  const exerciseDetailQuery = user
    ? useQuery({
        ...exerciseByIdQueryOptions(selectedExerciseId ?? 0),
        enabled: Boolean(selectedExerciseId),
      })
    : useQuery({
        ...getDemoExercisesByIdQueryOptions(selectedExerciseId ?? 0),
        enabled: Boolean(selectedExerciseId),
      });

  const authedContributionQuery = useQuery({
    ...contributionDataQueryOptions(),
    enabled: Boolean(user),
  });
  const demoWorkoutsQuery = useQuery({
    ...getDemoWorkoutsQueryOptions(),
    enabled: !user,
  });

  const workoutContributionData = user
    ? authedContributionQuery.data
    : buildDemoContributionData(demoWorkoutsQuery.data ?? []);

  return (
    <AnalyticsPage
      isLoadingExercises={false}
      exercises={exercises}
      selectedExerciseId={selectedExerciseId}
      onSelectExercise={(id) =>
        navigate({ search: (prev) => ({ ...prev, exerciseId: id }) })
      }
      isLoadingDetails={exerciseDetailQuery.isLoading}
      exerciseSets={exerciseDetailQuery.data?.sets}
      isDemoMode={!user}
      workoutContributionData={workoutContributionData}
    />
  );
}
