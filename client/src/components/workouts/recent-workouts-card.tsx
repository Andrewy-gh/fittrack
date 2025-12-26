import { Link } from '@tanstack/react-router';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Clock, ChevronRight } from 'lucide-react';
import { formatDate, formatTime } from '@/lib/utils';
import type { WorkoutWorkoutResponse } from '@/client';

export interface RecentWorkoutsCardProps {
  workouts: Array<WorkoutWorkoutResponse>;
  hasWorkoutInProgress?: boolean;
  newWorkoutLink?: string;
}

export function RecentWorkoutsCard({
  workouts,
  hasWorkoutInProgress = false,
  newWorkoutLink = '/workouts/new',
}: RecentWorkoutsCardProps) {
  return (
    <Card className="border-0 shadow-sm backdrop-blur-sm">
      <CardHeader>
        <CardTitle className="text-xl font-semibold">
          Recent Workouts
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {hasWorkoutInProgress && (
          <Link
            to={newWorkoutLink}
            data-testid="workout-in-progress-card"
            className="flex cursor-pointer items-center justify-between rounded-lg bg-primary/10 px-3 py-2"
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
            className="flex cursor-pointer items-center justify-between rounded-lg bg-muted/50 px-3 py-2"
          >
            <div className="flex items-center space-x-4">
              {workout.workout_focus && (
                <div className="flex items-center space-x-2">
                  <Badge
                    variant="outline"
                    className="border-border bg-muted text-xs"
                  >
                    {workout.workout_focus.toUpperCase()}
                  </Badge>
                </div>
              )}
              <div className="flex items-center space-x-2 text-sm mt-1">
                <span>{formatDate(workout.date)}</span>
                <span>•</span>
                <span className="text-muted-foreground">
                  {formatTime(workout.created_at)}
                </span>
              </div>
            </div>
            <ChevronRight className="w-5 h-5 text-muted-foreground" />
          </Link>
        ))}
      </CardContent>
    </Card>
  );
}
