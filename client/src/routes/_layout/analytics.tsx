import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { AnalyticsPage } from "@/features/analytics/pages/analytics-page";
import { exercisesQueryOptions } from "@/features/exercises/api/exercises";
import {
  contributionDataQueryOptions,
  workoutsFocusValuesQueryOptions,
} from "@/features/workouts/api/workouts";
import {
  getDemoContributionDataQueryOptions,
  getDemoExercisesQueryOptions,
  getDemoWorkoutsFocusValuesQueryOptions,
} from "@/lib/demo-data/query-options";
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
        context.queryClient.ensureQueryData(exercisesQueryOptions()),
        context.queryClient.ensureQueryData(contributionDataQueryOptions()),
        context.queryClient.ensureQueryData(workoutsFocusValuesQueryOptions()),
      ]);
    } else {
      initializeDemoData();
      await Promise.all([
        context.queryClient.ensureQueryData(getDemoExercisesQueryOptions()),
        context.queryClient.ensureQueryData(
          getDemoContributionDataQueryOptions(),
        ),
        context.queryClient.ensureQueryData(
          getDemoWorkoutsFocusValuesQueryOptions(),
        ),
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
      isDemoMode={!user}
    />
  );
}
