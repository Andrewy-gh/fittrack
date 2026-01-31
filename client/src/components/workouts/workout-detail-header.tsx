import { type ReactNode } from 'react';
import { Badge } from '@/components/ui/badge';
import { formatDate, formatTime } from '@/lib/utils';

export interface WorkoutDetailHeaderProps {
  workoutDate?: string;
  workoutFocus?: string | null;
  actions?: ReactNode;
}

export function WorkoutDetailHeader({
  workoutDate,
  workoutFocus,
  actions,
}: WorkoutDetailHeaderProps) {
  return (
    <div className="flex items-center justify-between pt-4">
      <div>
        <div className="mb-2">
          <h1 className="text-2xl md:text-3xl font-bold tracking-tight">
            {workoutDate ? formatDate(workoutDate) : ''}
          </h1>
        </div>
        <div className="flex items-center gap-2 mt-1">
          <p className="text-sm md:text-base text-muted-foreground">
            {workoutDate ? formatTime(workoutDate) : ''}
          </p>
          {workoutFocus && (
            <>
              <span className="text-muted-foreground">&bull;</span>
              <Badge
                variant="outline"
                className="border-border bg-muted text-xs"
              >
                {workoutFocus.toUpperCase()}
              </Badge>
            </>
          )}
        </div>
      </div>
      {actions && (
        <div className="flex flex-col items-center gap-3 md:flex-row">
          {actions}
        </div>
      )}
    </div>
  );
}
