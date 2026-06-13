import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { contributionDataQueryOptions } from "@/features/workouts/api/workouts";
import { getWorkoutListQueryOptions } from "@/features/workouts/api/workout-query-options";
import { WorkoutsPage } from "@/features/workouts/pages/workouts-page";
import { clearDemoData, initializeDemoData } from "@/lib/demo-data/storage";

const workoutsSearchSchema = z.object({
  focusArea: z.string().optional(),
  sortOrder: z.enum(["asc", "desc"]).optional(),
  itemsPerPage: z.coerce.number().int().positive().optional(),
  page: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute("/_layout/workouts/")({
  validateSearch: workoutsSearchSchema,
  loader: async ({ context }) => {
    const user = context.user;

    if (user) {
      clearDemoData();
      context.queryClient.ensureQueryData(getWorkoutListQueryOptions(user));
      context.queryClient.ensureQueryData(contributionDataQueryOptions());
    } else {
      initializeDemoData();
      context.queryClient.ensureQueryData(getWorkoutListQueryOptions(user));
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  return (
    <WorkoutsPage
      user={user}
      search={search}
    />
  );
}
