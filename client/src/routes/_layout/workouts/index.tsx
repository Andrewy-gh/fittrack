import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery, useQuery } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { Clock, Plus } from 'lucide-react';
import {
  workoutsQueryOptions,
  contributionDataQueryOptions,
} from '@/lib/api/workouts';
import { getDemoWorkoutsQueryOptions } from '@/lib/demo-data/query-options';
import { initializeDemoData, clearDemoData } from '@/lib/demo-data/storage';
import { loadFromLocalStorage } from '@/lib/local-storage';
import { WorkoutSummaryCards } from '@/components/workouts/workout-summary-cards';
import { WorkoutContributionGraph } from '@/components/workouts/workout-contribution-graph';
import { ContributionGraphError } from '@/components/workouts/contribution-graph-error';
import { RecentWorkoutsCard } from '@/components/workouts/recent-workouts-card';
import { WorkoutDistributionCard } from '@/components/workouts/workout-distribution-card';
import { useState } from 'react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import { Toggle } from '@/components/ui/toggle';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { ArrowDownAz, ArrowUpAz, Filter } from 'lucide-react';
import Calendar04 from '@/components/calendar-04';
import Calendar05 from '@/components/calendar-05';
import { Label } from '@/components/ui/label';
import { workoutsFocusValuesQueryOptions } from '@/lib/api/workouts';

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
  const defaultContributionGraphOpen =
    typeof window !== 'undefined' && window.innerWidth >= 768;

  const newWorkoutLink = '/workouts/new';

  // Filter state
  const [focusArea, setFocusArea] = useState<string>('all');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [itemsPerPage, setItemsPerPage] = useState<string>('10');
  const [currentPage, setCurrentPage] = useState(1);
  const [calendarVariant, setCalendarVariant] = useState<'v4' | 'v5'>('v4');

  const { data: focusAreas } = useQuery(workoutsFocusValuesQueryOptions());

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

        {/* Filtering and Sorting (Now below activity) */}
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="filters" className="border-none">
            <AccordionTrigger className="flex items-center gap-2 hover:no-underline py-2 px-4 bg-muted/50 rounded-lg">
              <div className="flex items-center gap-2">
                <Filter className="w-4 h-4" />
                <span className="font-semibold text-sm">Filters & Sorting</span>
              </div>
            </AccordionTrigger>
            <AccordionContent className="pt-4 pb-2 px-4 space-y-6">
              {/* Calendar Variant Selection (for demo) */}
              <div className="space-y-2">
                <Label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Calendar Variant
                </Label>
                <div className="flex gap-2">
                  <Button
                    variant={calendarVariant === 'v4' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setCalendarVariant('v4')}
                    className="flex-1"
                  >
                    Variant 4 (1 Month)
                  </Button>
                  <Button
                    variant={calendarVariant === 'v5' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setCalendarVariant('v5')}
                    className="flex-1"
                  >
                    Variant 5 (2 Months)
                  </Button>
                </div>
              </div>

              {/* Date Range Picker */}
              <div className="space-y-2">
                <Label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Date Range
                </Label>
                <div className="flex justify-center border rounded-lg p-2 bg-card">
                  {calendarVariant === 'v4' ? <Calendar04 /> : <Calendar05 />}
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {/* Focus Area Dropdown */}
                <div className="space-y-2">
                  <Label
                    htmlFor="focus-area"
                    className="text-xs font-medium text-muted-foreground uppercase tracking-wider"
                  >
                    Focus Area
                  </Label>
                  <Select value={focusArea} onValueChange={setFocusArea}>
                    <SelectTrigger id="focus-area">
                      <SelectValue placeholder="Select focus area" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Focus Areas</SelectItem>
                      {focusAreas?.map((focus) => (
                        <SelectItem key={focus} value={focus}>
                          {focus}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Sort Order */}
                <div className="space-y-2">
                  <Label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Sort Order
                  </Label>
                  <div className="flex items-center gap-2">
                    <Toggle
                      pressed={sortOrder === 'asc'}
                      onPressedChange={(pressed) =>
                        setSortOrder(pressed ? 'asc' : 'desc')
                      }
                      aria-label="Toggle sort order"
                      className="w-full flex justify-between px-3"
                    >
                      <span className="text-sm">
                        {sortOrder === 'asc' ? 'Ascending' : 'Descending'}
                      </span>
                      {sortOrder === 'asc' ? (
                        <ArrowUpAz className="w-4 h-4 ml-2" />
                      ) : (
                        <ArrowDownAz className="w-4 h-4 ml-2" />
                      )}
                    </Toggle>
                  </div>
                </div>
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>

        {/* Items Per Page (Now above recent workouts) */}
        <div className="flex items-center gap-2 justify-center px-1">
          <Label htmlFor="items-per-page" className="text-xs whitespace-nowrap">
            Show
          </Label>
          <Select value={itemsPerPage} onValueChange={setItemsPerPage}>
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
          workouts={workouts}
          hasWorkoutInProgress={hasWorkoutInProgress}
          newWorkoutLink={newWorkoutLink}
        />

        {/* Pagination (Now below recent workouts) */}
        <Pagination className="flex justify-center py-2">
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  setCurrentPage((p) => Math.max(1, p - 1));
                }}
              />
            </PaginationItem>
            <PaginationItem className="hidden sm:inline-block">
              <PaginationLink href="#" isActive>
                {currentPage}
              </PaginationLink>
            </PaginationItem>
            <PaginationItem>
              <PaginationEllipsis />
            </PaginationItem>
            <PaginationItem>
              <PaginationNext
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  setCurrentPage((p) => p + 1);
                }}
              />
            </PaginationItem>
          </PaginationContent>
        </Pagination>

        {/* Workout Distribution */}
        <WorkoutDistributionCard workouts={workouts} />
      </div>
    </main>
  );
}
