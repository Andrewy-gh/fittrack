import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useState } from 'react';
import { useRouter, useRouteContext } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { useDeleteWorkoutMutation } from '@/lib/api/workouts';
import { deleteDemoWorkoutsByIdMutationWithMeta } from '@/lib/demo-data/query-options';
import { getErrorMessage } from '@/lib/errors';
import { toast } from 'sonner';

interface DeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  workoutId: number;
}

export function DeleteDialog({
  isOpen,
  onOpenChange,
  workoutId,
}: DeleteDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const router = useRouter();
  const { user } = useRouteContext({ from: '/_layout/workouts/$workoutId/' });

  const authDeleteMutation = useDeleteWorkoutMutation();
  const demoDeleteMutation = useMutation(
    deleteDemoWorkoutsByIdMutationWithMeta()
  );
  const deleteMutation = user ? authDeleteMutation : demoDeleteMutation;

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      await deleteMutation.mutateAsync(
        { path: { id: workoutId } },
        {
          onSuccess: () => {
            router.navigate({ to: '/workouts' });
          },
        }
      );
    } catch (err) {
      const errorMessage = getErrorMessage(
        err,
        'Failed to delete workout. Please try again.'
      );
      toast.error(errorMessage);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <AlertDialog open={isOpen} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete this
            workout and all associated sets from your training history.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {isDeleting ? 'Deleting...' : 'Delete Workout'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
