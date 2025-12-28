import { Suspense } from 'react';
import { withForm } from '@/hooks/form';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { MOCK_VALUES } from './form-options';
import {
  ErrorBoundary,
  InlineErrorFallback,
} from '@/components/error-boundary';
import { compose, required, minValue, maxValue } from '@/lib/validation';

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
    return (
      <Dialog
        open={true}
        onOpenChange={(open) => {
          if (!open) {
            onClose();
          }
        }}
      >
        <DialogContent className="w-[90vw] max-w-md sm:max-w-lg mx-auto my-8">
          <DialogHeader className="mb-4">
            <DialogTitle className="text-xl font-semibold">Add Set</DialogTitle>
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
                onBlur: ({ value }) => {
                  const error = compose(
                    required,
                    minValue(1),
                    maxValue(1000)
                  )(value, 'Reps');
                  return error;
                },
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
                const repsFieldState =
                  state.fieldMeta[
                    `exercises[${exerciseIndex}].sets[${setIndex}].reps`
                  ];
                return {
                  canSubmit: state.canSubmit,
                  isValid: state.isValid,
                  repsIsTouched: repsFieldState?.isTouched || false,
                };
              }}
              children={({ canSubmit, isValid, repsIsTouched }) => {
                const isDisabled = !canSubmit || !isValid || !repsIsTouched;
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
            <Button
              variant="outline"
              className="w-full mt-6 text-base font-semibold rounded-lg"
              onClick={onRemoveSet}
            >
              Remove Set
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    );
  },
});
