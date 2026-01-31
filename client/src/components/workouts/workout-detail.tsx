import { Link } from '@tanstack/react-router';
import { type ReactNode, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Edit, Trash } from 'lucide-react';
import { DeleteDialog } from '@/routes/_layout/workouts/-components/delete-dialog';
import type { WorkoutWorkoutWithSetsResponse } from '@/client';
import { WorkoutDetailExercises } from '@/components/workouts/workout-detail-exercises';
import { WorkoutDetailHeader } from '@/components/workouts/workout-detail-header';
import { WorkoutDetailSummaryCards } from '@/components/workouts/workout-detail-summary-cards';

export interface WorkoutDetailProps {
  workout: WorkoutWorkoutWithSetsResponse[];
}

type WorkoutDetailBaseProps = WorkoutDetailProps & {
  headerActions?: ReactNode;
  dialogSlot?: ReactNode;
};

function WorkoutDetailBase({
  workout,
  headerActions,
  dialogSlot,
}: WorkoutDetailBaseProps) {
  if (workout.length === 0) {
    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          <Card>
            <CardContent className="py-6 text-sm text-muted-foreground">
              No workout data available.
            </CardContent>
          </Card>
        </div>
      </main>
    );
  }

  const uniqueExercises = new Set(workout.map((w) => w.exercise_id)).size;
  const totalSets = workout.length;
  const totalReps = workout.reduce((sum, w) => sum + (w.reps || 0), 0);
  const totalVolume = workout.reduce((sum, w) => sum + (w.volume || 0), 0);

  const workoutDate = workout[0]?.workout_date;
  const workoutFocus = workout[0]?.workout_focus;

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        <WorkoutDetailHeader
          workoutDate={workoutDate}
          workoutFocus={workoutFocus}
          actions={headerActions}
        />
        <WorkoutDetailSummaryCards
          uniqueExercises={uniqueExercises}
          totalSets={totalSets}
          totalReps={totalReps}
          totalVolume={totalVolume}
        />
        <WorkoutDetailExercises workout={workout} />
        {dialogSlot}
      </div>
    </main>
  );
}

export function WorkoutDetail({ workout }: WorkoutDetailProps) {
  return <WorkoutDetailBase workout={workout} />;
}

export function WorkoutDetailEditable({ workout }: WorkoutDetailProps) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  if (workout.length === 0) {
    return <WorkoutDetailBase workout={workout} />;
  }
  const workoutId = workout[0]?.workout_id ?? 0;

  const handleOpenDeleteDialog = () => {
    setIsDeleteDialogOpen(true);
  };

  return (
    <WorkoutDetailBase
      workout={workout}
      headerActions={
        <>
          <Button size="sm" variant="outline" asChild>
            <Link
              to="/workouts/$workoutId/edit"
              params={{ workoutId }}
            >
              <Edit className="mr-2 hidden h-4 w-4 md:block" />
              Edit
            </Link>
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={handleOpenDeleteDialog}
          >
            <Trash className="mr-2 hidden h-4 w-4 md:block" />
            Delete
          </Button>
        </>
      }
      dialogSlot={
        <DeleteDialog
          isOpen={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          workoutId={workoutId}
        />
      }
    />
  );
}
