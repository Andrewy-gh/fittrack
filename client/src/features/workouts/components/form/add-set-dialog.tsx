import { Suspense } from "react";
import { withForm } from "@/hooks/form";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { MOCK_VALUES } from "./form-options";
import {
  ErrorBoundary,
  InlineErrorFallback,
} from "@/components/error-boundary";
import {
  shouldDiscardSetOnDismiss,
  validateSetReps,
} from "./workout-form-helpers";

type AddSetDialogProps = {
  exerciseIndex: number;
  setIndex: number;
  onSaveSet: () => void;
  onClose: () => void;
  onRemoveSet: () => void;
};

export const AddSetDialog = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as AddSetDialogProps,
  render: function Render({
    form,
    exerciseIndex,
    setIndex,
    onSaveSet,
    onClose,
    onRemoveSet,
  }) {
    const shouldDiscardSet = (values: typeof form.state.values) => {
      const set = values.exercises?.[exerciseIndex]?.sets?.[setIndex];
      return shouldDiscardSetOnDismiss(set);
    };

    const handleDismiss = () => {
      if (shouldDiscardSet(form.state.values)) {
        onRemoveSet();
        return;
      }
      onClose();
    };

    return (
      <Dialog
        open={true}
        onOpenChange={(open) => {
          if (!open) {
            handleDismiss();
          }
        }}
      >
        <DialogContent className="w-[90vw] max-w-md sm:max-w-lg mx-auto my-8">
          <DialogHeader className="mb-4">
            <DialogTitle className="text-xl font-semibold">Add Set</DialogTitle>
            <DialogDescription>
              Enter the load, reps, and set type for this exercise set.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-6">
            <form.AppField
              name={`exercises[${exerciseIndex}].sets[${setIndex}].weight`}
              children={(field) => (
                <field.InputField
                  label="Weight"
                  type="number"
                  step="0.1"
                  min="0"
                  className="sm:text-center sm:h-9"
                />
              )}
            />
            <form.AppField
              name={`exercises[${exerciseIndex}].sets[${setIndex}].reps`}
              validators={{
                onMount: ({ value }) => validateSetReps(value),
                onChange: ({ value }) => {
                  return validateSetReps(value);
                },
                onSubmit: ({ value }) => validateSetReps(value),
              }}
              children={(field) => {
                return (
                  <field.InputField
                    label="Reps"
                    type="number"
                    className="sm:text-center sm:h-9"
                  />
                );
              }}
            />
            <ErrorBoundary
              fallback={
                <InlineErrorFallback message="Failed to load set type selector" />
              }
            >
              <Suspense fallback={<div>Loading...</div>}>
                <form.AppField
                  name={`exercises[${exerciseIndex}].sets[${setIndex}].setType`}
                  children={(field) => <field.SetTypeSelect />}
                />
              </Suspense>
            </ErrorBoundary>
            <form.Subscribe
              selector={(state) => {
                return {
                  canSubmit: state.canSubmit,
                  isValid: state.isValid,
                };
              }}
              children={({ canSubmit, isValid }) => {
                const isDisabled = !canSubmit || !isValid;
                return (
                  <Button
                    className="w-full mt-6 text-base font-semibold rounded-lg"
                    onClick={onSaveSet}
                    disabled={isDisabled}
                  >
                    Save Set
                  </Button>
                );
              }}
            />
            <form.Subscribe
              selector={(state) => shouldDiscardSet(state.values)}
              children={(shouldDiscard) => (
                <Button
                  variant="outline"
                  className="w-full mt-6 text-base font-semibold rounded-lg"
                  onClick={onRemoveSet}
                >
                  {shouldDiscard ? "Cancel" : "Remove Set"}
                </Button>
              )}
            />
          </div>
        </DialogContent>
      </Dialog>
    );
  },
});
