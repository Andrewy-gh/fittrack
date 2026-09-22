import { useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ExercisecatalogEntry } from "@/client";
import {
  getExercisesByIdQueryKey,
  putExercisesByIdCatalogMutation,
} from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ExerciseCatalogPicker } from "./exercise-catalog-picker";

export function ExerciseClassification({
  exerciseId,
  catalog,
}: {
  exerciseId: number;
  catalog?: ExercisecatalogEntry;
}) {
  const trigger = useRef<HTMLButtonElement>(null);
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();
  const mutation = useMutation({
    ...putExercisesByIdCatalogMutation(),
    meta: { skipGlobalErrorHandler: true },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["muscle-contributions"],
      });
      await queryClient.invalidateQueries({
        queryKey: getExercisesByIdQueryKey({ path: { id: exerciseId } }),
      });
      setOpen(false);
    },
  });
  const save = (catalogId: string) =>
    mutation.mutate({
      path: { id: exerciseId },
      body: { catalog_id: catalogId },
    });

  return (
    <Card>
      <CardContent className="space-y-3">
        <h2 className="font-semibold">Classification</h2>
        {catalog ? (
          <div className="text-sm space-y-1">
            <p>{catalog.name}</p>
            <p>Equipment: {catalog.equipment}</p>
            <p>Primary: {catalog.primaryMuscles.join(", ")}</p>
            <p>
              Secondary: {catalog.secondaryMuscles.join(", ") || "None listed"}
            </p>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Unclassified</p>
        )}
        <div className="flex gap-2">
          <Button
            ref={trigger}
            type="button"
            variant="outline"
            disabled={mutation.isPending}
            onClick={() => {
              mutation.reset();
              setOpen(true);
            }}
          >
            {catalog ? "Change classification" : "Classify exercise"}
          </Button>
          {catalog && (
            <Button
              type="button"
              variant="ghost"
              disabled={mutation.isPending}
              onClick={() => save("")}
            >
              Clear
            </Button>
          )}
        </div>
        {mutation.isError && !open && (
          <p role="alert">Could not save classification. Please try again.</p>
        )}
        <Dialog
          open={open}
          onOpenChange={(value) => {
            if (!mutation.isPending) setOpen(value);
          }}
        >
          <DialogContent
            className="max-h-[85dvh] overflow-y-auto"
            onCloseAutoFocus={(event) => {
              event.preventDefault();
              trigger.current?.focus();
            }}
          >
            <DialogHeader>
              <DialogTitle>Classify exercise</DialogTitle>
              <DialogDescription>
                Choose a match. Your name and history stay the same.
              </DialogDescription>
            </DialogHeader>
            <ExerciseCatalogPicker
              disabled={mutation.isPending}
              onSelect={(entry) => save(entry.id)}
            />
            {mutation.isPending && (
              <p role="status">Saving classification...</p>
            )}
            {mutation.isError && (
              <p role="alert">
                Could not save classification. Please try again.
              </p>
            )}
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}
