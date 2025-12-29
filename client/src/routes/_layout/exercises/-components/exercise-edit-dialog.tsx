import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useState } from 'react';
import { useRouteContext } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { patchExercisesByIdMutation, getExercisesQueryKey, getExercisesByIdQueryKey } from '@/client/@tanstack/react-query.gen';
import { patchDemoExercisesByIdMutation } from '@/lib/demo-data/query-options';
import { isApiError, getErrorMessage } from '@/lib/errors';
import { toast } from 'sonner';
import { queryClient } from '@/lib/api/api';

interface ExerciseEditDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  exerciseId: number;
  exerciseName: string;
}

export function ExerciseEditDialog({
  isOpen,
  onOpenChange,
  exerciseId,
  exerciseName,
}: ExerciseEditDialogProps) {
  const [name, setName] = useState(exerciseName);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { user } = useRouteContext({ from: '/_layout/exercises/$exerciseId' });

  // Override global error handler to prevent double toasts
  // We handle errors manually in the catch block below
  const authUpdateMutation = useMutation({
    ...patchExercisesByIdMutation(),
    onSuccess: (_, { path: { id } }) => {
      queryClient.invalidateQueries({
        queryKey: getExercisesQueryKey(),
      });
      queryClient.invalidateQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id } }),
      });
    },
    onError: () => {
      // Don't show toast - we handle errors manually
    },
  });

  const demoUpdateMutation = useMutation({
    ...patchDemoExercisesByIdMutation(),
    onError: () => {
      // Don't show toast - we handle errors manually
    },
  });

  const updateMutation = user ? authUpdateMutation : demoUpdateMutation;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmedName = name.trim();

    // Validation
    if (!trimmedName) {
      setError('Exercise name is required');
      return;
    }

    if (trimmedName.length > 100) {
      setError('Exercise name must be 100 characters or less');
      return;
    }

    setIsSubmitting(true);
    try {
      await updateMutation.mutateAsync({
        path: { id: exerciseId },
        body: { name: trimmedName },
      });
      toast.success('Exercise updated successfully');
      onOpenChange(false);
    } catch (err) {
      // Check if it's a duplicate name error (409 conflict)
      if (isApiError(err) && err.message.toLowerCase().includes('already exists')) {
        // Show inline error for duplicate name so user can fix it
        setError(`You already have an exercise named '${trimmedName}'`);
      } else {
        // For other errors (network, server errors, etc.), show toast
        // The global mutation error handler will also show a toast,
        // but we override it here to prevent double toasts
        const errorMessage = getErrorMessage(err, 'Failed to update exercise. Please try again.');
        toast.error(errorMessage);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        onOpenChange(open);
        if (!open) {
          setName(exerciseName);
          setError(null);
        }
      }}
    >
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Exercise Name</DialogTitle>
            <DialogDescription>
              Update the name of your exercise. The name must be unique.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Exercise name"
              maxLength={100}
              disabled={isSubmitting}
              autoFocus
            />
            {error && <div className="text-sm text-red-600 mt-2">{error}</div>}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
