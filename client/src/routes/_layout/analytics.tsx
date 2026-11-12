import { createFileRoute } from "@tanstack/react-router";
import { AnalyticsPage } from "@/features/analytics/pages/analytics-page";
import { getExerciseListQueryOptions } from "@/features/exercises/api/exercise-query-options";
import {
  getWorkoutContributionQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/features/workouts/api/workout-query-options";
import { analyticsSearchValidator } from "@/lib/route-search-validation";
import { clearDemoData, initializeDemoData } from "@/lib/demo-data/storage";

export const Route = createFileRoute("/_layout/analytics")({
  validateSearch: analyticsSearchValidator,
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      clearDemoData();
      await Promise.all([
        context.queryClient.query({
          ...getExerciseListQueryOptions(user),
          staleTime: Infinity,
        }),
        context.queryClient.query({
          ...getWorkoutContributionQueryOptions(user),
          staleTime: Infinity,
        }),
        context.queryClient.query({
          ...getWorkoutsFocusQueryOptions(user),
          staleTime: Infinity,
        }),
      ]);
    } else {
      initializeDemoData();
      await Promise.all([
        context.queryClient.query({
          ...getExerciseListQueryOptions(user),
          staleTime: Infinity,
        }),
        context.queryClient.query({
          ...getWorkoutContributionQueryOptions(user),
          staleTime: Infinity,
        }),
        context.queryClient.query({
          ...getWorkoutsFocusQueryOptions(user),
          staleTime: Infinity,
        }),
      ]);
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId, assessmentWeek } = Route.useSearch();
  const { user } = Route.useRouteContext();

  return (
    <AnalyticsPage
      exerciseId={exerciseId}
      assessmentWeek={assessmentWeek}
      user={user}
    />
  );
}
