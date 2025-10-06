import { withForm } from '@/hooks/form';
import { useState } from 'react';
import { AddSetDialog } from '../-components/add-set-dialog';
import { ChevronLeft, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { MOCK_VALUES } from '../-components/form-options';

type ExerciseScreenProps = {
  exerciseIndex: number;
  onBack: () => void;
};

// MARK: Header
export const ExerciseHeader = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as ExerciseScreenProps,
  render: function Render({ form, exerciseIndex, onBack }) {
    return (
      <div className="flex items-center justify-between pt-6 pb-2">
        <button onClick={onBack}>
          <ChevronLeft className="text-primary" />
        </button>
        <form.AppField
          name={`exercises[${exerciseIndex}].name`}
          children={(field) => (
            <div>
              <h1 className="font-bold text-2xl tracking-tight text-foreground flex-1 ">
                {field.state.value}
              </h1>
            </div>
          )}
        />
      </div>
    );
  },
});

export const ExerciseSets = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as Pick<ExerciseScreenProps, 'exerciseIndex'>,
  render: function Render({ form, exerciseIndex }) {
    const [dialogOpenIndex, setDialogOpenIndex] = useState<number | null>(null);

    return (
      <form.AppField
        name={`exercises[${exerciseIndex}].sets`}
        mode="array"
        children={(setsField) => {
          const sets = setsField.state.value || [];
          const totalSets = sets.length;
          const totalVolume = sets.reduce(
            (acc, set) => acc + (set?.weight || 0) * (set?.reps || 0),
            0
          );
          return (
            // MARK: Stats
            <>
              <h2 className="font-semibold text-xl tracking-tight text-foreground mb-4">
                Today's Sets
              </h2>
              <div className="grid grid-cols-2 gap-4">
                <Card className="bg-card border border-border shadow-sm p-6">
                  <div className="text-center">
                    <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground mb-2">
                      TOTAL SETS
                    </div>
                    <div className="font-bold text-lg text-primary">
                      {totalSets}
                    </div>
                  </div>
                </Card>
                <Card className="bg-card border border-border shadow-sm p-6">
                  <div className="text-center">
                    <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground mb-2">
                      TOTAL VOLUME
                    </div>
                    <div className="font-bold text-lg text-primary">
                      {totalVolume}
                    </div>
                  </div>
                </Card>
              </div>

              {/* Sets List */}
              <div>
                <div className="space-y-3">
                  {sets.map((set, setIndex) => {
                    // MARK: Dialog
                    const isDialogOpen = dialogOpenIndex === setIndex;
                    if (isDialogOpen) {
                      return (
                        <AddSetDialog
                          key={`exercises[${exerciseIndex}].sets[${setIndex}]`}
                          form={form}
                          exerciseIndex={exerciseIndex}
                          setIndex={setIndex}
                          onSaveSet={() => {
                            setDialogOpenIndex(null);
                          }}
                          onClose={() => {
                            setDialogOpenIndex(null);
                          }}
                          onRemoveSet={() => {
                            setsField.removeValue(setIndex);
                            setDialogOpenIndex(null);
                          }}
                        />
                      );
                    }
                    return (
                      // MARK: Set Cards
                      <Card
                        key={`exercises[${exerciseIndex}].sets[${setIndex}]`}
                        className="bg-card border border-border shadow-sm p-4"
                        onClick={() => {
                          setDialogOpenIndex(setIndex);
                        }}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4">
                            <div className="font-bold text-lg">
                              #{setIndex + 1}
                            </div>
                            <div>
                              <span
                                className={`px-3 py-1 rounded-full text-xs font-medium ${
                                  set.setType === 'working'
                                    ? 'bg-primary/20 text-primary'
                                    : 'bg-muted text-muted-foreground'
                                }`}
                              >
                                {set.setType}
                              </span>
                            </div>
                          </div>
                          <div className="text-right">
                            <div className="text-card-foreground font-bold text-lg">
                              {set.weight}lb &#215; {set.reps}
                            </div>
                            <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground">
                              {set.weight && set.reps && set.weight * set.reps}{' '}
                              volume
                            </div>
                          </div>
                        </div>
                      </Card>
                    );
                  })}
                </div>
              </div>
              {/* Add Set Button */}
              <div className="pt-4">
                <Button
                  className="hover:bg-primary/90 w-full py-4 text-base font-semibold"
                  onClick={() => {
                    setsField.pushValue({
                      weight: 0,
                      reps: 0,
                      setType: 'working',
                    });
                    const updatedSets = setsField.state.value || [];
                    setDialogOpenIndex(updatedSets.length - 1);
                  }}
                >
                  <Plus className="w-5 h-5 mr-2" />
                  Add Set
                </Button>
              </div>
            </>
          );
        }}
      />
    );
  },
});



// MARK: Exercise Screen
export function ExerciseScreen({
  header,
  recentSets,
  sets,
}: {
  header: React.ReactNode;
  recentSets?: React.ReactNode;
  sets: React.ReactNode;
}) {
  return (
    <div className="min-h-screen">
      <div className="px-4 pb-8">
        <div className="max-w-md mx-auto space-y-6">
          {header}
          {recentSets}
          {sets}
        </div>
      </div>
    </div>
  );
}
