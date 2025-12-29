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
import { useDeleteExerciseMutation } from '@/lib/api/exercises';
import { deleteDemoExercisesByIdMutationWithMeta } from '@/lib/demo-data/query-options';
import { showErrorToast } from '@/lib/errors';

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
  const router = useRouter();
  const { user } = useRouteContext({ from: "/_layout/exercises/$exerciseId" });

  const authDeleteMutation = useDeleteExerciseMutation();
  const demoDeleteMutation = useMutation(deleteDemoExercisesByIdMutationWithMeta());
  const deleteMutation = user ? authDeleteMutation : demoDeleteMutation;

  const handleDelete = async () => {
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
      showErrorToast(error, 'Failed to delete exercise. Please try again.');
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
