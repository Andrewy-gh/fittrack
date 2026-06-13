import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useState } from "react";
import { useRouter } from "@tanstack/react-router";
import { useDeleteExerciseForModeMutation } from "@/features/exercises/api/exercises";
import { showErrorToast } from "@/lib/errors";
import { toast } from "sonner";

interface ExerciseDeleteDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  exerciseId: number;
  isDemoMode: boolean;
}

export function ExerciseDeleteDialog({
  isOpen,
  onOpenChange,
  exerciseId,
  isDemoMode,
}: ExerciseDeleteDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const router = useRouter();

  const deleteMutation = useDeleteExerciseForModeMutation(isDemoMode);

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      await deleteMutation.mutateAsync(
        { path: { id: exerciseId } },
        {
          onSuccess: () => {
            toast.success("Exercise deleted successfully");
            router.navigate({ to: "/exercises" });
          },
        },
      );
    } catch (error) {
      showErrorToast(error, "Failed to delete exercise. Please try again.");
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <AlertDialog
      open={isOpen}
      onOpenChange={onOpenChange}
    >
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
            {isDeleting ? "Deleting..." : "Delete Exercise"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
