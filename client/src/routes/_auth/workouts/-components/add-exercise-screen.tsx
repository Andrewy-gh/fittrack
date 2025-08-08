import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';
import { Card } from '@/components/ui/card';
import { withForm } from '@/hooks/form';
import type { ExerciseOption } from '@/lib/types';
import { MOCK_VALUES } from '../-components/form-options';

type AddExerciseScreenProps = {
  exercises: ExerciseOption[];
  onAddExercise: (exerciseIndex: number) => void;
  onBack: () => void;
};

export const AddExerciseScreen = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as AddExerciseScreenProps,
  render: function Render({ form, exercises, onAddExercise, onBack }) {
    return (
      <div className="min-h-screen bg-background">
        <div className="px-4 pb-8">
          <div className="max-w-md mx-auto space-y-6">
            {/* Header */}
            <div className="flex items-center pt-6 pb-2">
              <Button
                variant="ghost"
                onClick={onBack}
                className="p-0 h-auto text-primary hover:text-primary/80"
              >
                <ArrowLeft className="w-6 h-6" />
              </Button>
              <h1 className="font-bold text-3xl tracking-tight text-foreground flex-1 text-center">
                Add Exercise
              </h1>
            </div>

            {/* Exercise Name Input */}
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => (
                <>
                  <field.AddExerciseField2
                    exercises={exercises}
                    onAddExercise={onAddExercise}
                  />
                  <div className="space-y-4">
                    <h3 className="font-semibold text-sm tracking-tight uppercase text-muted-foreground text-center">
                      OR CHOOSE FROM COMMON EXERCISES:
                    </h3>
                    <div className="grid grid-cols-2 gap-3">
                      {exercises
                        .filter(
                          (exercise) =>
                            !field.state.value.some(
                              (e) => e.name === exercise.name
                            )
                        )
                        .map((exercise) => (
                          <Card
                            key={exercise.id} // ! TODO: handle button hover
                            className="shadow-sm hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50"
                          >
                            <Button
                              variant="ghost"
                              className="h-auto w-full text-sm font-medium whitespace-normal text-card-foreground hover:bg-transparent"
                              onClick={() => {
                                field.pushValue({
                                  name: exercise.name,
                                  sets: [],
                                });
                                onAddExercise(field.state.value.length - 1);
                              }}
                            >
                              {exercise.name}
                            </Button>
                          </Card>
                        ))}
                    </div>
                  </div>
                </>
              )}
            />
          </div>
        </div>
      </div>
    );
  },
});