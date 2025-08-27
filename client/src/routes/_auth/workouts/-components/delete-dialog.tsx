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
import { useRouter } from '@tanstack/react-router';
// import { useMutation, useQueryClient } from '@tanstack/react-query';
// import { deleteWorkout } from '@/lib/api/workouts';
// import type { User } from '@/lib/api/auth';
import { useDeleteWorkoutMutation } from '@/lib/api/workouts';

interface DeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  workoutId: number;
  // user: Exclude<User, null>;
}

export function DeleteDialog({
  isOpen,
  onOpenChange,
  workoutId,
}: DeleteDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();
  // const queryClient = useQueryClient();

  // const deleteMutation = useMutation({
  //   mutationFn: async () => {
  //     return deleteWorkout(workoutId, user);
  //   },
  //   onSuccess: () => {
  //     queryClient.invalidateQueries({ queryKey: ['workouts'] });
  //     onOpenChange(false);
  //     router.navigate({ to: '/workouts' });
  //   },
  //   onError: (error: any) => {
  //     setError(error?.message || 'Failed to delete workout');
  //   },
  // });

  const deleteMutation = useDeleteWorkoutMutation();

  const handleDelete = async () => {
    setError(null);
    setIsDeleting(true);
    try {
      const res = await deleteMutation.mutateAsync(
        { path: { id: workoutId } },
        {
          onSuccess: () => {
            router.navigate({ to: '/workouts' });
          },
        }
      );
      console.log('res', res);
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
