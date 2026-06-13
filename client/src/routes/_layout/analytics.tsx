import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { AnalyticsPage } from "@/features/analytics/pages/analytics-page";
import { getExerciseListQueryOptions } from "@/features/exercises/api/exercise-query-options";
import {
  getWorkoutContributionQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/features/workouts/api/workout-query-options";
import { clearDemoData, initializeDemoData } from "@/lib/demo-data/storage";

const analyticsSearchSchema = z.object({
  exerciseId: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute("/_layout/analytics")({
  validateSearch: analyticsSearchSchema,
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      clearDemoData();
      await Promise.all([
        context.queryClient.ensureQueryData(getExerciseListQueryOptions(user)),
        context.queryClient.ensureQueryData(
          getWorkoutContributionQueryOptions(user),
        ),
        context.queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(user)),
      ]);
    } else {
      initializeDemoData();
      await Promise.all([
        context.queryClient.ensureQueryData(getExerciseListQueryOptions(user)),
        context.queryClient.ensureQueryData(
          getWorkoutContributionQueryOptions(user),
        ),
        context.queryClient.ensureQueryData(getWorkoutsFocusQueryOptions(user)),
      ]);
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId } = Route.useSearch();
  const { user } = Route.useRouteContext();

  return (
    <AnalyticsPage
      exerciseId={exerciseId}
      user={user}
    />
  );
}
