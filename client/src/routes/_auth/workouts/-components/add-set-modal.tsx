import { Button } from '@/components/ui/button';
import { formOptsMock } from './form-options';
import { withForm } from '@/hooks/form';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Suspense } from 'react';

type AddSetModalProps = {
  exerciseIndex: number;
  setIndex: number;
  onSaveSet: () => void;
  onClose: () => void;
  onRemoveSet: () => void;
};

export const AddSetModal2 = withForm({
  ...formOptsMock,
  props: {} as AddSetModalProps,
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
          <DialogHeader>
            <DialogTitle>Add Set</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <form.AppField
              name={`exercises[${exerciseIndex}].sets[${setIndex}].weight`}
              children={(field) => (
                <field.InputField
                  label="Weight"
                  type="number"
                  className="sm:text-center sm:h-9"
                />
              )}
            />
            <form.AppField
              name={`exercises[${exerciseIndex}].sets[${setIndex}].reps`}
              children={(field) => (
                <field.InputField
                  label="Reps"
                  type="number"
                  className="sm:text-center sm:h-9"
                />
              )}
            />
            <Suspense fallback={<div>Loading...</div>}>
              <form.AppField
                name={`exercises[${exerciseIndex}].sets[${setIndex}].setType`}
                children={(field) => <field.SetTypeSelect />}
              />
            </Suspense>
            <Button
              className="w-full mt-6 text-base font-semibold rounded-lg"
              onClick={onSaveSet}
            >
              Save Set
            </Button>
            <Button
              variant="outline"
              className="w-full py-4 mt-6 text-base font-semibold rounded-lg"
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
