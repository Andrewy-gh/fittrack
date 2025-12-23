import { useState } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { format, parseISO } from 'date-fns';
import { contributionDataQueryOptions, workoutsQueryOptions } from '@/lib/api/workouts';
import {
  ContributionGraph,
  ContributionGraphBlock,
  ContributionGraphCalendar,
  ContributionGraphFooter,
  ContributionGraphLegend,
  type Activity,
} from '@/components/kibo-ui/contribution-graph';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

export function WorkoutContributionGraph() {
  const [isOpen, setIsOpen] = useState(true);
  const [openPopover, setOpenPopover] = useState<string | null>(null);
  const navigate = useNavigate();
  const { data } = useSuspenseQuery(contributionDataQueryOptions());
  const { data: workouts } = useSuspenseQuery(workoutsQueryOptions());

  // Transform API response to Activity[] format and track workout IDs
  const activities: Activity[] =
    data.days?.map((day) => ({
      date: day.date || '',
      count: day.count || 0,
      level: day.level || 0,
    })) || [];

  // Create a map of date to workout IDs for navigation
  const workoutIdsByDate = new Map<string, number[]>(
    data.days?.map((day) => [day.date || '', day.workoutIds || []]) || []
  );

  // Create a map of workout ID to workout details
  const workoutDetailsById = new Map(
    workouts?.map((workout) => [
      workout.id,
      {
        id: workout.id,
        date: workout.date,
        focus: workout.workout_focus,
      },
    ]) || []
  );

  // Empty state: no workouts in 52-week period
  if (activities.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">
          Start your fitness journey! Log your first workout to see your
          progress here.
        </p>
      </div>
    );
  }

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen} className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold tracking-tight">Activity</h2>
        <CollapsibleTrigger className="p-2 hover:bg-muted rounded-md transition-colors">
          {isOpen ? (
            <ChevronUp className="h-5 w-5" />
          ) : (
            <ChevronDown className="h-5 w-5" />
          )}
        </CollapsibleTrigger>
      </div>
      <CollapsibleContent className="space-y-4">
        {isOpen && (
          <ContributionGraph data={activities}>
            <ContributionGraphCalendar>
              {({ activity, dayIndex, weekIndex }) => {
                const formattedDate = format(
                  parseISO(activity.date),
                  'EEEE, MMM d, yyyy'
                );
                const workingSets = activity.count;
                const workingSetsText =
                  workingSets === 1
                    ? '1 working set'
                    : `${workingSets} working sets`;

                const workoutIds = workoutIdsByDate.get(activity.date) || [];
                const hasSingleWorkout = workoutIds.length === 1;
                const hasMultipleWorkouts = workoutIds.length > 1;

                const handleClick = () => {
                  if (hasSingleWorkout) {
                    navigate({
                      to: '/workouts/$workoutId',
                      params: { workoutId: workoutIds[0] },
                    });
                  }
                };

                const block = (
                  <ContributionGraphBlock
                    activity={activity}
                    dayIndex={dayIndex}
                    weekIndex={weekIndex}
                    onClick={handleClick}
                    className="data-[level='0']:fill-muted data-[level='1']:fill-primary/20 data-[level='2']:fill-primary/40 data-[level='3']:fill-primary/60 data-[level='4']:fill-primary/80 cursor-pointer"
                  />
                );

                // For multiple workouts, wrap with Popover
                if (hasMultipleWorkouts) {
                  return (
                    <Popover
                      open={openPopover === activity.date}
                      onOpenChange={(open) =>
                        setOpenPopover(open ? activity.date : null)
                      }
                    >
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <PopoverTrigger asChild>{block}</PopoverTrigger>
                        </TooltipTrigger>
                        <TooltipContent>
                          <div className="text-center">
                            <div className="font-medium">{formattedDate}</div>
                            <div className="text-xs">{workingSetsText}</div>
                          </div>
                        </TooltipContent>
                      </Tooltip>
                      <PopoverContent className="w-64 p-2">
                        <div className="space-y-1">
                          <div className="px-2 py-1.5 text-sm font-semibold">
                            Select a workout
                          </div>
                          {workoutIds.map((workoutId) => {
                            const workout = workoutDetailsById.get(workoutId);
                            if (!workout) return null;

                            const time = format(parseISO(workout.date), 'h:mm a');
                            return (
                              <button
                                key={workoutId}
                                onClick={() => {
                                  navigate({
                                    to: '/workouts/$workoutId',
                                    params: { workoutId },
                                  });
                                }}
                                className="w-full rounded-md px-2 py-1.5 text-left text-sm hover:bg-muted transition-colors"
                              >
                                <div className="font-medium">{time}</div>
                                {workout.focus && (
                                  <div className="text-xs text-muted-foreground">
                                    {workout.focus}
                                  </div>
                                )}
                              </button>
                            );
                          })}
                        </div>
                      </PopoverContent>
                    </Popover>
                  );
                }

                // For single or no workouts, just show tooltip
                return (
                  <Tooltip>
                    <TooltipTrigger asChild>{block}</TooltipTrigger>
                    <TooltipContent>
                      <div className="text-center">
                        <div className="font-medium">{formattedDate}</div>
                        <div className="text-xs">{workingSetsText}</div>
                      </div>
                    </TooltipContent>
                  </Tooltip>
                );
              }}
            </ContributionGraphCalendar>
            <ContributionGraphFooter>
              <ContributionGraphLegend />
            </ContributionGraphFooter>
          </ContributionGraph>
        )}
      </CollapsibleContent>
    </Collapsible>
  );
}
