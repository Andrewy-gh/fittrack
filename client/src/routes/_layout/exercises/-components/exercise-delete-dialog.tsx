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
import { deleteExercisesByIdMutation, getExercisesQueryKey, getExercisesByIdQueryKey } from '@/client/@tanstack/react-query.gen';
import { deleteDemoExercisesByIdMutation } from '@/lib/demo-data/query-options';
import { showErrorToast } from '@/lib/errors';
import { queryClient } from '@/lib/api/api';

interface ExerciseDeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  exerciseId: number;
}

export function ExerciseDeleteDialog({
  isOpen,
  onOpenChange,
  exerciseId,
}: ExerciseDeleteDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  const { user } = useRouteContext({ from: "/_layout/exercises/$exerciseId" });

  // Override global error handler to prevent double toasts
  // We handle errors manually in the catch block below
  const authDeleteMutation = useMutation({
    ...deleteExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getExercisesQueryKey(),
      });
      queryClient.removeQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id } }),
      });
    },
    onError: () => {
      // Don't show toast - we handle errors manually
    },
  });

  const demoDeleteMutation = useMutation({
    ...deleteDemoExercisesByIdMutation(),
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
        { path: { id: exerciseId } },
        {
          onSuccess: () => {
            router.navigate({ to: '/exercises' });
          },
        }
      );
    } catch (error) {
      // Show toast for delete failure
      showErrorToast(error, 'Failed to delete exercise. Please try again.');
      // Also show inline error in dialog
      setError('Failed to delete exercise. Please try again.');
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
            exercise and all associated sets from your training history.
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
            {isDeleting ? 'Deleting...' : 'Delete Exercise'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
