import { Link } from "@tanstack/react-router";
import { Badge } from "@/components/ui/badge";
import { Clock, ChevronRight } from "lucide-react";
import { formatDate, formatTime } from "@/lib/utils";
import type { WorkoutWorkoutResponse } from "@/client";

export interface RecentWorkoutsListProps {
  workouts: Array<WorkoutWorkoutResponse>;
  hasWorkoutInProgress?: boolean;
  newWorkoutLink?: string;
}

export function RecentWorkoutsList({
  workouts,
  hasWorkoutInProgress = false,
  newWorkoutLink = "/workouts/new",
}: RecentWorkoutsListProps) {
  return (
    <div className="space-y-3">
      {hasWorkoutInProgress && (
        <Link
          to={newWorkoutLink}
          data-testid="workout-in-progress-card"
          className="flex cursor-pointer items-center justify-between rounded-xl border border-primary/20 bg-primary/10 px-3 py-2"
        >
          <div className="flex items-center space-x-4">
            <Badge
              variant="outline"
              className="border-primary/20 bg-primary/10 text-primary text-xs"
            >
              IN PROGRESS
            </Badge>
            <div className="flex items-center space-x-2 text-sm">
              <Clock className="w-4 h-4 text-primary" />
              <span className="text-primary">Continue workout</span>
            </div>
          </div>
          <ChevronRight className="w-5 h-5 text-primary" />
        </Link>
      )}
      {workouts.map((workout) => (
        <Link
          key={workout.id}
          to="/workouts/$workoutId"
          params={{
            workoutId: workout.id,
          }}
          data-testid="workout-card"
          className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-xl border bg-card px-3 py-2 shadow-sm"
        >
          <div className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3">
            {workout.workout_focus && (
              <Badge
                variant="outline"
                className="border-border bg-muted text-xs leading-snug whitespace-normal break-words text-left max-w-[220px] sm:max-w-[280px]"
              >
                {workout.workout_focus.toUpperCase()}
              </Badge>
            )}
            <div className="flex items-center gap-2 text-sm text-muted-foreground whitespace-nowrap mt-1">
              <span>{formatDate(workout.date)}</span>
              <span>•</span>
              <span>{formatTime(workout.created_at)}</span>
            </div>
          </div>
          <ChevronRight className="w-5 h-5 text-muted-foreground" />
        </Link>
      ))}
    </div>
  );
}
