import { useSuspenseQuery } from '@tanstack/react-query';
import { contributionDataQueryOptions } from '@/lib/api/workouts';
import {
  ContributionGraph,
  ContributionGraphBlock,
  ContributionGraphCalendar,
  ContributionGraphFooter,
  ContributionGraphLegend,
  type Activity,
} from '@/components/kibo-ui/contribution-graph';

export function WorkoutContributionGraph() {
  const { data } = useSuspenseQuery(contributionDataQueryOptions());

  // Transform API response to Activity[] format
  const activities: Activity[] =
    data.days?.map((day) => ({
      date: day.date || '',
      count: day.count || 0,
      level: day.level || 0,
    })) || [];

  // Empty state: no workouts in 52-week period
  if (activities.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">
          Start your fitness journey! Log your first workout to see your progress here.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <ContributionGraph data={activities}>
        <ContributionGraphCalendar>
          {({ activity, dayIndex, weekIndex }) => (
            <ContributionGraphBlock
              activity={activity}
              dayIndex={dayIndex}
              weekIndex={weekIndex}
            />
          )}
        </ContributionGraphCalendar>
        <ContributionGraphFooter>
          <ContributionGraphLegend />
        </ContributionGraphFooter>
      </ContributionGraph>
    </div>
  );
}
