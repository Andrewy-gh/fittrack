import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";
import { exerciseByIdQueryOptions } from "@/features/exercises/api/exercises";
import { ExerciseDetailPage } from "@/features/exercises/pages/exercise-detail-page";
import { getDemoExercisesByIdQueryOptions } from "@/lib/demo-data/query-options";
import { initializeDemoData, clearDemoData } from "@/lib/demo-data/storage";

const exerciseSearchSchema = z.object({
  sortOrder: z.enum(["asc", "desc"]).optional(),
  itemsPerPage: z.coerce.number().int().positive().optional(),
  page: z.coerce.number().int().positive().optional(),
});

export const Route = createFileRoute("/_layout/exercises/$exerciseId")({
  validateSearch: exerciseSearchSchema,
  params: {
    parse: (params) => {
      const exerciseId = parseInt(params.exerciseId, 10);
      if (isNaN(exerciseId) || !Number.isInteger(exerciseId)) {
        throw new Error("Invalid exerciseId");
      }
      return { exerciseId };
    },
  },
  loader: ({ context, params }) => {
    const exerciseId = params.exerciseId;
    const user = context.user;

    if (user) {
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(exerciseByIdQueryOptions(exerciseId));
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(
        getDemoExercisesByIdQueryOptions(exerciseId),
      );
    }

    return { exerciseId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { exerciseId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();
  const { sortOrder, itemsPerPage, page } = Route.useSearch();

  return (
    <ExerciseDetailPage
      exerciseId={exerciseId}
      isDemoMode={!user}
      sortOrder={sortOrder}
      itemsPerPage={itemsPerPage}
      page={page}
    />
  );
}
