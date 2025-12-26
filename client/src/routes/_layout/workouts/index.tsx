import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery, useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { Clock, Plus } from 'lucide-react';
import { workoutsQueryOptions, contributionDataQueryOptions } from '@/lib/api/workouts';
import { getDemoWorkoutsQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { loadFromLocalStorage } from '@/lib/local-storage';
import { WorkoutSummaryCards } from '@/components/workouts/workout-summary-cards';
import { WorkoutContributionGraph } from '@/components/workouts/workout-contribution-graph';
import { ContributionGraphError } from '@/components/workouts/contribution-graph-error';
import { RecentWorkoutsCard } from '@/components/workouts/recent-workouts-card';
import { WorkoutDistributionCard } from '@/components/workouts/workout-distribution-card';

export const Route = createFileRoute('/_layout/workouts/')({
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
  const defaultContributionGraphOpen = typeof window !== 'undefined' && window.innerWidth >= 768;

  const newWorkoutLink = '/workouts/new';

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
              {hasWorkoutInProgress ? 'In Progress' : 'New Workout'}
            </Link>
          </Button>
        </div>

        {/* Summary Cards */}
        <WorkoutSummaryCards workouts={workouts} />

        {/* Contribution Graph (authenticated users only) */}
        {user && contributionQuery.isError && <ContributionGraphError />}
        {user && contributionQuery.isSuccess && (
          <WorkoutContributionGraph
            data={contributionQuery.data}
            defaultOpen={defaultContributionGraphOpen}
          />
        )}

        {/* Recent Workouts */}
        <RecentWorkoutsCard
          workouts={workouts}
          hasWorkoutInProgress={hasWorkoutInProgress}
          newWorkoutLink={newWorkoutLink}
        />

        {/* Workout Distribution */}
        <WorkoutDistributionCard workouts={workouts} />
      </div>
    </main>
  );
}
