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
import { deleteWorkoutsByIdMutation, getWorkoutsQueryKey, getWorkoutsByIdQueryKey, getWorkoutsContributionDataQueryKey } from '@/client/@tanstack/react-query.gen';
import { deleteDemoWorkoutsByIdMutation } from '@/lib/demo-data/query-options';
import { getErrorMessage } from '@/lib/errors';
import { toast } from 'sonner';
import { queryClient } from '@/lib/api/api';

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
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { user } = useRouteContext({ from: '/_layout/workouts/$workoutId/' });

  // Override global error handler to prevent double toasts
  // We handle errors manually in the catch block below
  const authDeleteMutation = useMutation({
    ...deleteWorkoutsByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getWorkoutsQueryKey(),
      });
      queryClient.removeQueries({
        queryKey: getWorkoutsByIdQueryKey({ path: { id } }),
      });
      queryClient.invalidateQueries({
        queryKey: getWorkoutsContributionDataQueryKey(),
      });
    },
    onError: () => {
      // Don't show toast - we handle errors manually
    },
  });
  const demoDeleteMutation = useMutation({
    ...deleteDemoWorkoutsByIdMutation(),
    onError: () => {
      // Don't show toast - we handle errors manually
    },
  });
  const deleteMutation = user ? authDeleteMutation : demoDeleteMutation;

  const handleDelete = async () => {
    setError(null);
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
      // Show toast for delete failure
      const errorMessage = getErrorMessage(err, 'Failed to delete workout. Please try again.');
      toast.error(errorMessage);
      // Also show inline error in dialog
      setError(errorMessage);
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
          {error && <div className="text-sm text-red-600 mt-2">{error}</div>}
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
