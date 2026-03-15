import type { ComponentType, ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { Plus, Save, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import type { WorkoutExerciseInput } from '@/client';
import type { WorkoutFocus } from '@/lib/api/workouts';

type WorkoutExerciseCard = Pick<WorkoutExerciseInput, 'name' | 'sets'>;
type WorkoutFormSectionApi = {
  AppField: ComponentType<any>;
  Subscribe: ComponentType<any>;
};

type WorkoutMetadataFieldsProps = {
  form: WorkoutFormSectionApi;
  workoutsFocus: WorkoutFocus[];
};

type WorkoutExerciseCardsProps = {
  exercises: WorkoutExerciseCard[];
  dataTestId: string;
  onRemoveExercise: (index: number) => void;
  formatVolume: (volume: number) => string;
  renderNameSupplement?: (exercise: WorkoutExerciseCard) => ReactNode;
  renderMetrics?: (exercise: WorkoutExerciseCard) => ReactNode;
};

type WorkoutFormActionsProps = {
  form: WorkoutFormSectionApi;
};

function getExerciseVolume(exercise: WorkoutExerciseCard): number {
  return exercise.sets.reduce(
    (total, set) => total + (set.reps || 0) * (set.weight || 0),
    0
  );
}

export function WorkoutMetadataFields({
  form,
  workoutsFocus,
}: WorkoutMetadataFieldsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 mb-4">
      <form.AppField
        name="date"
        children={(field: any) => <field.DatePicker />}
      />
      <form.AppField
        name="workoutFocus"
        children={(field: any) => (
          <field.WorkoutFocusCombobox workoutsFocus={workoutsFocus} />
        )}
      />
      <div className="col-span-2">
        <form.AppField
          name="notes"
          children={(field: any) => <field.NotesTextarea />}
        />
      </div>
    </div>
  );
}

export function WorkoutExerciseCards({
  exercises,
  dataTestId,
  onRemoveExercise,
  formatVolume,
  renderNameSupplement,
  renderMetrics,
}: WorkoutExerciseCardsProps) {
  return (
    <div className="space-y-3">
      {exercises.map((exercise, exerciseIndex) => {
        const volume = getExerciseVolume(exercise);

        return (
          <Link
            key={`exercise-${exerciseIndex}`}
            to="."
            search={{ exerciseIndex }}
            className="block"
          >
            <Card
              className="p-4 cursor-pointer hover:shadow-md transition-all duration-200"
              data-testid={dataTestId}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 mb-2">
                      <div className="w-2 h-2 bg-primary rounded-full"></div>
                      <div>
                        <span className="text-primary font-medium text-sm">
                          {exercise.name}
                        </span>
                        {renderNameSupplement?.(exercise)}
                      </div>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-primary hover:text-primary/80 hover:bg-primary/10"
                      aria-label={`Delete ${exercise.name}`}
                      onClick={(event) => {
                        event.stopPropagation();
                        event.preventDefault();
                        onRemoveExercise(exerciseIndex);
                      }}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>

                  <div className="flex items-end justify-between">
                    <div>
                      <div className="font-bold text-lg text-card-foreground">
                        {exercise.sets.length}
                      </div>
                      <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground">
                        sets
                      </div>
                    </div>

                    <div className="flex items-end gap-4">
                      <div className="text-right">
                        <div className="text-card-foreground font-bold text-lg">
                          {formatVolume(volume)}
                        </div>
                        <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground">
                          volume
                        </div>
                      </div>
                      {renderMetrics?.(exercise)}
                    </div>
                  </div>
                </div>
              </div>
            </Card>
          </Link>
        );
      })}
    </div>
  );
}

export function WorkoutFormActions({ form }: WorkoutFormActionsProps) {
  return (
    <>
      <div className="py-6">
        <Link to="." search={{ addExercise: true }}>
          <Button
            type="button"
            variant="outline"
            className="w-full text-base font-semibold rounded-lg"
          >
            <Plus className="w-5 h-5 mr-2" />
            Add Exercise
          </Button>
        </Link>
      </div>
      <div className="mt-8">
        <form.Subscribe
          selector={(state: any) => [state.canSubmit, state.isSubmitting]}
          children={([canSubmit, isSubmitting]: [boolean, boolean]) => (
            <Button
              type="submit"
              disabled={!canSubmit}
              className="w-full text-base font-semibold rounded-lg"
            >
              <Save className="w-3.5 h-3.5 mr-1.5" />
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          )}
        />
      </div>
    </>
  );
}
