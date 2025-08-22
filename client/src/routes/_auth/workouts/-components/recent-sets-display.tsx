import { useSuspenseQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { recentExerciseSetsQueryOptions } from '@/lib/api/exercises';
import { formatDate } from '@/lib/utils';
import type { User } from '@/lib/api/auth';
import type { exercise_RecentSetsResponse } from '@/generated';
import { sortByExerciseAndSetOrder } from '@/lib/utils';

interface RecentSetsDisplayProps {
  exerciseId: number;
  user: User;
}

export function RecentSetsDisplay({
  exerciseId,
  user,
}: RecentSetsDisplayProps) {
  const { data: recentSets } = useSuspenseQuery(
    recentExerciseSetsQueryOptions(exerciseId, user)
  );

  if (recentSets.length === 0) {
    return null;
  }

  const sortedRecentSets = sortByExerciseAndSetOrder(recentSets)

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
    {} as Record<string, { date: string; sets: exercise_RecentSetsResponse[] }>
  );

  return (
    <div className="space-y-4">
      <h2 className="font-semibold text-2xl tracking-tight text-foreground mb-4">
        Recent Sets
      </h2>
      {Object.entries(groupedSets).map(([dateKey, group]) => {
        return (
          <Card key={dateKey} className="border-0 shadow-sm backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="text-lg font-semibold">
                {formatDate(group.date)}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {group.sets.map((set, index) => {
                const volume = (set.weight || 0) * set.reps;
                return (
                  <div
                    key={set.set_id}
                    className="flex items-center justify-between py-2 px-3 rounded-lg bg-muted/50"
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
