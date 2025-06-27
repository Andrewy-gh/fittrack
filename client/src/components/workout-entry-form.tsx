import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { clearLocalStorage, loadFromLocalStorage, saveToLocalStorage } from '@/lib/local-storage';
import type { Exercise, ExerciseOption } from '@/lib/types';
import { ExerciseCombobox } from '@/components/exercise-combobox';
import { Button } from '@/components/ui/button';
import { DatePicker } from '@/components/date-picker';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { SetTypeSelect } from '@/components/set-type-select';
import { Textarea } from '@/components/ui/textarea';
import { Trash2 } from 'lucide-react';
import type { WorkoutFormValues } from '@/lib/types';


export function WorkoutEntryForm({
  exercises,
}: {
  exercises: ExerciseOption[];
}) {
  // Load initial values from localStorage
  const getInitialValues = (): WorkoutFormValues => {
    const saved = loadFromLocalStorage();
    return (
      saved || {
        date: new Date(),
        notes: '',
        exercises: [] as Exercise[],
      }
    );
  };

  const form = useForm({
    defaultValues: getInitialValues(),
    listeners: {
      onChange: ({ formApi }) => {
        console.log('Saving form data to localStorage');
        saveToLocalStorage(formApi.state.values);
      },
      onChangeDebounceMs: 500,
    },
    onSubmit: async ({ value }) => {
      try {
        const response = await fetch('/api/workouts', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(value),
        });

        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(errorText ?? 'Failed to submit workout');
        }

        const result = await response.json();
        console.log(
          'Workout submitted! Server says: ' + JSON.stringify(result)
        );

        // Clear localStorage after successful submission
        clearLocalStorage();

        // Reset form to default values
        form.reset();
      } catch (error) {
        alert(error);
      }
    },
  });

  const [selectedExercise, setSelectedExercise] = useState<ExerciseOption>();

  function handleSelect(option: ExerciseOption) {
    console.log('handleSelect');
    console.log(option);
    setSelectedExercise(option);
  }

  function handleAppendGroup(name: ExerciseOption['name']) {
    const newExercise = {
      id: exercises.length + 1,
      name,
    };
    exercises.push(newExercise);
    console.log('handleAppendGroup');
    console.log(newExercise);
    handleSelect(newExercise);
  }

  // Add a function to manually clear the form and localStorage
  const handleClearForm = () => {
    if (confirm('Are you sure you want to clear all form data?')) {
      clearLocalStorage();
      form.reset();
      setSelectedExercise(undefined);
    }
  };

  return (
    <div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
        className="space-y-6"
      >
        {/* MARK: date */}
        <form.Field name="date">
          {(field) => {
            return (
              <div className="grid w-full max-w-sm items-center gap-3">
                <Label>Date</Label>
                <DatePicker
                  value={field.state.value}
                  onChange={(date) => {
                    if (date) {
                      field.handleChange(date);
                    }
                  }}
                />
              </div>
            );
          }}
        </form.Field>
        {/* MARK: notes */}
        <form.Field name="notes">
          {(field) => {
            return (
              <div className="w-full max-w-sm space-y-3">
                <Label htmlFor={field.name}>Notes</Label>
                <Textarea
                  id={field.name}
                  name={field.name}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              </div>
            );
          }}
        </form.Field>
        <form.Field name="exercises" mode="array">
          {(field) => {
            return (
              <div className="space-y-4">
                {(field.state.value as Exercise[]).map((_, exerciseIndex) => {
                  return (
                    <div
                      key={`exercises[${exerciseIndex}]`}
                      className="p-4 border space-y-2 relative"
                    >
                      {/* MARK: exercise name */}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-2 h-6 w-6 rounded-full text-red-500 hover:text-red-700 hover:bg-red-100"
                        onClick={() => field.removeValue(exerciseIndex)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                      <form.Field name={`exercises[${exerciseIndex}].name`}>
                        {(subField) => {
                          return (
                            <div
                              className="grid w-full max-w-sm items-center gap-3"
                              key={`exercises[${exerciseIndex}].name`}
                            >
                              <Label htmlFor="email">Exercise Name</Label>
                              <Input
                                id={subField.name}
                                name={subField.name}
                                value={subField.state.value}
                                onBlur={subField.handleBlur}
                                onChange={(e) =>
                                  subField.handleChange(e.target.value)
                                }
                              />
                            </div>
                          );
                        }}
                      </form.Field>
                      <form.Field
                        name={`exercises[${exerciseIndex}].sets`}
                        mode="array"
                      >
                        {(field) => {
                          return (
                            <>
                              {(field.state.value as any[]).map(
                                (_, setIndex) => {
                                  return (
                                    <div
                                      key={`exercises[${exerciseIndex}].sets[${setIndex}]`}
                                      className="flex gap-3 items-end"
                                    >
                                      {/* MARK: set weight */}
                                      <form.Field
                                        name={`exercises[${exerciseIndex}].sets[${setIndex}].weight`}
                                      >
                                        {(subField) => {
                                          return (
                                            <div className="grid w-full max-w-sm items-center gap-3">
                                              <Label
                                                className={cn(
                                                  setIndex !== 0 && 'sr-only'
                                                )}
                                              >
                                                Weight
                                              </Label>
                                              <Input
                                                type="number"
                                                value={subField.state.value}
                                                onChange={(e) =>
                                                  subField.handleChange(
                                                    Number(e.target.value)
                                                  )
                                                }
                                                placeholder="Weight"
                                              />
                                            </div>
                                          );
                                        }}
                                      </form.Field>
                                      {/* MARK: set reps */}
                                      <form.Field
                                        name={`exercises[${exerciseIndex}].sets[${setIndex}].reps`}
                                      >
                                        {(subField) => {
                                          return (
                                            <div className="grid w-full max-w-sm items-center gap-3">
                                              <Label
                                                className={cn(
                                                  setIndex !== 0 && 'sr-only'
                                                )}
                                              >
                                                Reps
                                              </Label>
                                              <Input
                                                type="number"
                                                value={subField.state.value}
                                                onChange={(e) =>
                                                  subField.handleChange(
                                                    Number(e.target.value)
                                                  )
                                                }
                                                placeholder="Reps"
                                              />
                                            </div>
                                          );
                                        }}
                                      </form.Field>
                                      {/* MARK: set type */}
                                      <form.Field
                                        name={`exercises[${exerciseIndex}].sets[${setIndex}].setType`}
                                      >
                                        {(subField) => {
                                          return (
                                            <div className="grid w-full max-w-sm items-center gap-3">
                                              <Label
                                                className={cn(
                                                  setIndex !== 0 && 'sr-only'
                                                )}
                                              >
                                                Set Type
                                              </Label>
                                              <SetTypeSelect
                                                value={subField.state.value}
                                                onChange={subField.handleChange}
                                              />
                                            </div>
                                          );
                                        }}
                                      </form.Field>
                                      {/* MARK: remove set */}
                                      <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 rounded-full text-red-500 hover:text-red-700 hover:bg-red-100"
                                        onClick={() =>
                                          field.removeValue(setIndex)
                                        }
                                      >
                                        <Trash2 className="h-3 w-3" />
                                      </Button>
                                    </div>
                                  );
                                }
                              )}
                              {/* MARK: add set */}
                              <Button
                                onClick={() => {
                                  field.pushValue({
                                    weight: undefined,
                                    reps: undefined,
                                    setType: '',
                                  });
                                }}
                                type="button"
                              >
                                Add set
                              </Button>
                            </>
                          );
                        }}
                      </form.Field>
                    </div>
                  );
                })}
              </div>
            );
          }}
        </form.Field>
        {/* MARK: add exercise */}
        <div className="flex flex-col gap-4 md:flex-row md:items-end">
          <div className="space-y-2">
            <Label>Exercise</Label>
            <ExerciseCombobox
              options={exercises}
              selected={selectedExercise?.name ?? ''}
              onChange={handleSelect}
              onCreate={handleAppendGroup}
            />
          </div>
          <form.Field name="exercises">
            {(field) => (
              <Button
                onClick={() =>
                  field.pushValue({
                    name: selectedExercise?.name ?? '',
                    sets: [],
                  })
                }
                type="button"
                className="w-fit"
              >
                Add exercise
              </Button>
            )}
          </form.Field>
        </div>

        {/* Action buttons */}
        <div className="flex gap-4">
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
            children={([canSubmit, isSubmitting]) => (
              <Button type="submit" disabled={!canSubmit}>
                {isSubmitting ? '...' : 'Submit'}
              </Button>
            )}
          />
          <Button type="button" variant="outline" onClick={handleClearForm}>
            Clear Form
          </Button>
        </div>
      </form>
    </div>
  );
}
