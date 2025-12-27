import { Suspense } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChevronRight } from 'lucide-react';
import { Link } from '@tanstack/react-router';
import type { ExerciseRecentSetsResponse } from '@/client';
import { formatDate } from '@/lib/utils';
import { sortByExerciseAndSetOrder } from '@/lib/utils';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import { getRecentSetsQueryOptions } from '@/lib/api/unified-query-options';
import { ErrorBoundary, InlineErrorFallback } from '@/components/error-boundary';

interface RecentSetsDisplayProps {
  exerciseId: number;
  user: CurrentUser | CurrentInternalUser | null;
}

function RecentSetsDisplay({ exerciseId, user }: RecentSetsDisplayProps) {
  const { data: recentSets } = useSuspenseQuery(
    getRecentSetsQueryOptions(user, exerciseId)
  );

  if (recentSets.length === 0) {
    return null;
  }

  const sortedRecentSets = sortByExerciseAndSetOrder(recentSets);

  // Group sets by workout_date, preserving order
  const groupedSets = sortedRecentSets.reduce(
    (acc, set) => {
      const dateKey = set.workout_date;
      if (!acc[dateKey]) {
        acc[dateKey] = {
          date: set.workout_date,
          sets: [],
        };
      }
      acc[dateKey].sets.push(set);
      return acc;
    },
    {} as Record<string, { date: string; sets: ExerciseRecentSetsResponse[] }>
  );

  return (
    <div className="space-y-4">
      <h2 className="font-semibold text-xl tracking-tight text-foreground mb-4">
        Recent Sets
      </h2>
      {Object.entries(groupedSets).map(([dateKey, group]) => {
        return (
          <Card key={dateKey} className="border-0 shadow-sm backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="text-medium font-semibold">
                <Link
                  key={group.sets[0].workout_id}
                  to="/workouts/$workoutId"
                  params={{ workoutId: group.sets[0].workout_id }}
                  className="flex cursor-pointer items-center justify-between"
                >
                  {formatDate(group.date)}
                  <ChevronRight className="w-5 h-5 text-muted-foreground" />
                </Link>
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {group.sets.map((set, index) => {
                const volume = (set.weight || 0) * set.reps;
                return (
                  <div
                    key={set.set_id}
                    className="flex items-center justify-between py-2 px-3 rounded-lg bg-muted/50 hover:bg-muted/80 transition-colors"
                  >
                    <div className="flex items-center space-x-4">
                      <span className="text-sm font-medium text-muted-foreground w-8">
                        {set.set_order ?? index + 1}
                      </span>
                      <div className="flex items-center space-x-4 text-sm">
                        <span className="font-medium">
                          {set.weight || 0} lbs
                        </span>
                        <span>&times;</span>
                        <span className="font-medium">{set.reps} reps</span>
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {volume.toLocaleString()} vol
                    </div>
                  </div>
                );
              })}
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

// MARK: Recent Susp.
// Helper component to wrap recent sets with proper error boundaries
export function RecentSets({
  exerciseId,
  user
}: {
  exerciseId: number | null;
  user: CurrentUser | CurrentInternalUser | null;
}) {
  if (!exerciseId) {
    return null;
  }

  return (
    <ErrorBoundary
      fallback={
        <div className="space-y-4">
          <h2 className="font-semibold text-xl tracking-tight text-foreground mb-4">
            Recent Sets
          </h2>
          <InlineErrorFallback message="Failed to load recent sets" />
        </div>
      }
    >
      <Suspense
        fallback={
          <div className="space-y-4">
            <h2 className="font-semibold text-xl tracking-tight text-foreground mb-4">
              Recent Sets
            </h2>
            <div className="text-center py-8">
              <p className="text-muted-foreground text-sm">
                Loading recent sets...
              </p>
            </div>
          </div>
        }
      >
        <RecentSetsDisplay exerciseId={exerciseId} user={user} />
      </Suspense>
    </ErrorBoundary>
  );
}
