import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { validateWorkoutText } from "@/lib/workout-text";
import { useState } from "react";
import { useRenameExerciseMutation } from "@/features/exercises/api/exercises";
import { isApiError, getErrorMessage } from "@/lib/errors";
import { toast } from "sonner";

interface ExerciseEditDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  exerciseId: number;
  exerciseName: string;
  isDemoMode: boolean;
}

export function ExerciseEditDialog({
  isOpen,
  onOpenChange,
  exerciseId,
  exerciseName,
  isDemoMode,
}: ExerciseEditDialogProps) {
  const [name, setName] = useState(exerciseName);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const nameError = validateWorkoutText(name.trim(), "Exercise name");
  const updateMutation = useRenameExerciseMutation(isDemoMode);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const trimmedName = name.trim();

    // Validation
    if (!trimmedName) {
      setError("Exercise name is required");
      return;
    }

    if (nameError) {
      setError(nameError);
      return;
    }

    setIsSubmitting(true);
    try {
      await updateMutation.mutateAsync({
        path: { id: exerciseId },
        body: { name: trimmedName },
      });
      toast.success("Exercise updated successfully");
      onOpenChange(false);
    } catch (err) {
      // Check if it's a duplicate name error (409 conflict)
      if (
        isApiError(err) &&
        err.message.toLowerCase().includes("already exists")
      ) {
        // Show inline error for duplicate name so user can fix it
        setError(`You already have an exercise named '${trimmedName}'`);
      } else {
        // For other errors (network, server errors, etc.), show toast
        // The global mutation error handler will also show a toast,
        // but we override it here to prevent double toasts
        const errorMessage = getErrorMessage(
          err,
          "Failed to update exercise. Please try again.",
        );
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
              aria-label="Exercise name"
              aria-invalid={Boolean(nameError || error)}
              aria-describedby="exercise-name-error"
              disabled={isSubmitting}
              autoFocus
            />
            {(nameError || error) && (
              <div
                id="exercise-name-error"
                role="alert"
                className="text-sm text-destructive mt-2"
              >
                {nameError || error}
              </div>
            )}
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
            <Button
              type="submit"
              disabled={isSubmitting || Boolean(nameError)}
            >
              {isSubmitting ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
