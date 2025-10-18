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
import { Label } from '@/components/ui/label';
import { useState } from 'react';
import { useRouteContext } from '@tanstack/react-router';
import { useUpdateExerciseMutation } from '@/lib/api/exercises';
import { useMutation } from '@tanstack/react-query';
import { patchDemoExercisesByIdMutation } from '@/lib/demo-data/query-options';

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
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { user } = useRouteContext({ from: '/_layout/exercises/$exerciseId' });

  const authUpdateMutation = useUpdateExerciseMutation();
  const demoUpdateMutation = useMutation(patchDemoExercisesByIdMutation());
  const updateMutation = user ? authUpdateMutation : demoUpdateMutation;

  // Reset form when dialog opens/closes
  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setName(exerciseName);
      setError(null);
    }
    onOpenChange(open);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    const trimmedName = name.trim();
    if (!trimmedName) {
      setError('Exercise name is required');
      return;
    }
    if (trimmedName.length > 100) {
      setError('Exercise name must be 100 characters or less');
      return;
    }

    setIsUpdating(true);
    try {
      await updateMutation.mutateAsync(
        {
          path: { id: exerciseId },
          body: { name: trimmedName },
        },
        {
          onSuccess: () => {
            onOpenChange(false);
          },
          onError: (error: any) => {
            // Handle 409 Conflict - duplicate name
            if (error?.response?.status === 409 || error?.message?.includes('already exists')) {
              setError(`You already have an exercise named '${trimmedName}'`);
            } else {
              setError('Failed to update exercise name. Please try again.');
            }
          },
        }
      );
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit Exercise Name</DialogTitle>
            <DialogDescription>
              Update the name of your exercise. This will apply to all your training history.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label htmlFor="exercise-name">Exercise Name</Label>
            <Input
              id="exercise-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter exercise name"
              disabled={isUpdating}
              className="mt-2"
              autoFocus
              maxLength={100}
            />
            {error && <div className="text-sm text-red-600 mt-2">{error}</div>}
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isUpdating}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isUpdating}>
              {isUpdating ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
