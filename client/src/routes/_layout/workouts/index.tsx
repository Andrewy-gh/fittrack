import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useSuspenseQuery, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Toggle } from "@/components/ui/toggle";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PaginationControl } from "@/components/ui/pagination-control";
import { ArrowDownAz, ArrowUpAz, Clock, Plus } from "lucide-react";
import {
  workoutsQueryOptions,
  contributionDataQueryOptions,
} from "@/lib/api/workouts";
import { getDemoWorkoutsQueryOptions } from "@/lib/demo-data/query-options";
import { initializeDemoData, clearDemoData } from "@/lib/demo-data/storage";
import { loadFromLocalStorage } from "@/lib/local-storage";
import { WorkoutSummaryCards } from "@/components/workouts/workout-summary-cards";
import { WorkoutContributionGraph } from "@/components/workouts/workout-contribution-graph";
import { ContributionGraphError } from "@/components/workouts/contribution-graph-error";
import { RecentWorkoutsCard } from "@/components/workouts/recent-workouts-card";
import { WorkoutDistributionCard } from "@/components/workouts/workout-distribution-card";
import {
  filterWorkoutsByFocus,
  getFocusAreas,
  paginateWorkouts,
  sortWorkoutsByCreatedAt,
} from "@/lib/workouts-filters";

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
      // Authenticated: use API data
      clearDemoData();
      context.queryClient.ensureQueryData(workoutsQueryOptions());
      context.queryClient.ensureQueryData(contributionDataQueryOptions());
    } else {
      // Demo mode: use localStorage
      initializeDemoData();
      context.queryClient.ensureQueryData(getDemoWorkoutsQueryOptions());
    }
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const { focusArea, sortOrder, itemsPerPage, page } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const { data: workouts } = user
    ? useSuspenseQuery(workoutsQueryOptions())
    : useSuspenseQuery(getDemoWorkoutsQueryOptions());

  const contributionQuery = useQuery({
    ...contributionDataQueryOptions(),
    enabled: !!user,
  });

  // Check for workout in progress (pass user.id if authenticated, undefined for demo)
  const hasWorkoutInProgress = loadFromLocalStorage(user?.id) !== null;

  // Determine default open state for contribution graph (desktop vs mobile)
  const defaultContributionGraphOpen =
    typeof window !== "undefined" && window.innerWidth >= 768;

  const newWorkoutLink = "/workouts/new";
  const normalizedFocusArea = focusArea ?? "all";
  const normalizedSortOrder = sortOrder ?? "desc";
  const normalizedItemsPerPage = [10, 20, 50].includes(itemsPerPage ?? 10)
    ? (itemsPerPage ?? 10)
    : 10;

  const focusAreas = useMemo(() => getFocusAreas(workouts), [workouts]);
  const filteredWorkouts = useMemo(
    () => filterWorkoutsByFocus(workouts, normalizedFocusArea),
    [normalizedFocusArea, workouts],
  );
  const sortedWorkouts = useMemo(
    () => sortWorkoutsByCreatedAt(filteredWorkouts, normalizedSortOrder),
    [filteredWorkouts, normalizedSortOrder],
  );
  const { pagedWorkouts, totalPages, currentPage } = useMemo(
    () =>
      paginateWorkouts(sortedWorkouts, normalizedItemsPerPage, page),
    [sortedWorkouts, normalizedItemsPerPage, page],
  );

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Workouts</h1>
          </div>
          <Button size="sm" asChild>
            <Link to={newWorkoutLink}>
              {hasWorkoutInProgress ? (
                <Clock className="w-4 h-4 mr-2" />
              ) : (
                <Plus className="w-4 h-4 mr-2" />
              )}
              {hasWorkoutInProgress ? "In Progress" : "New Workout"}
            </Link>
          </Button>
        </div>

        {/* Summary Cards */}
        <WorkoutSummaryCards workouts={workouts} />

        {/* Contribution Graph (authenticated users only) */}
        {user && contributionQuery.isLoading && (
          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              Loading contribution graph...
            </CardContent>
          </Card>
        )}
        {user && contributionQuery.isError && <ContributionGraphError />}
        {user && contributionQuery.isSuccess && (
          <WorkoutContributionGraph
            data={contributionQuery.data}
            defaultOpen={defaultContributionGraphOpen}
          />
        )}

        <Card>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label
                  htmlFor="focus-area-compact"
                  className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide"
                >
                  Focus Area
                </Label>
                <Select
                  value={normalizedFocusArea}
                  onValueChange={(value) =>
                    navigate({
                      search: (prev) => ({
                        ...prev,
                        focusArea: value,
                        page: 1,
                      }),
                    })
                  }
                >
                  <SelectTrigger id="focus-area-compact" className="h-9">
                    <SelectValue placeholder="Select focus area" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Focus Areas</SelectItem>
                    {focusAreas.map((focus) => (
                      <SelectItem key={focus} value={focus}>
                        {focus}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-1.5">
                <Label className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide">
                  Sort Order
                </Label>
                <Toggle
                  pressed={normalizedSortOrder === "asc"}
                  onPressedChange={(pressed) =>
                    navigate({
                      search: (prev) => ({
                        ...prev,
                        sortOrder: pressed ? "asc" : "desc",
                        page: 1,
                      }),
                    })
                  }
                  aria-label="Toggle sort order"
                  className="px-3 py-1.5 inline-flex items-center justify-start gap-2"
                >
                  <span className="text-sm">
                    {normalizedSortOrder === "asc" ? "Ascending" : "Descending"}
                  </span>
                  {normalizedSortOrder === "asc" ? (
                    <ArrowUpAz className="w-4 h-4 opacity-70" />
                  ) : (
                    <ArrowDownAz className="w-4 h-4 opacity-70" />
                  )}
                </Toggle>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="flex items-center gap-2 justify-center px-1">
          <Label htmlFor="items-per-page" className="text-xs whitespace-nowrap">
            Show
          </Label>
          <Select
            value={String(normalizedItemsPerPage)}
            onValueChange={(value) =>
              navigate({
                search: (prev) => ({
                  ...prev,
                  itemsPerPage: Number(value),
                  page: 1,
                }),
              })
            }
          >
            <SelectTrigger id="items-per-page" className="h-8 w-[70px]">
              <SelectValue placeholder="10" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="10">10</SelectItem>
              <SelectItem value="20">20</SelectItem>
              <SelectItem value="50">50</SelectItem>
            </SelectContent>
          </Select>
          <span className="text-xs text-muted-foreground">per page</span>
        </div>

        {/* Recent Workouts */}
        <RecentWorkoutsCard
          workouts={pagedWorkouts}
          hasWorkoutInProgress={hasWorkoutInProgress}
          newWorkoutLink={newWorkoutLink}
        />

        <PaginationControl
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={(nextPage) =>
            navigate({
              search: (prev) => ({
                ...prev,
                page: nextPage,
              }),
            })
          }
        />

        {/* Workout Distribution */}
        <WorkoutDistributionCard workouts={workouts} />
      </div>
    </main>
  );
}
