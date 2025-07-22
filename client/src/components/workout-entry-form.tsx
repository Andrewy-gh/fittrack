import { useAppForm, withForm } from '@/hooks/form';
import { formOptions } from '@tanstack/react-form';
import { useState } from 'react';
import {
  clearLocalStorage,
  loadFromLocalStorage,
  saveToLocalStorage,
} from '@/lib/local-storage';
import type { Exercise, ExerciseOption } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogOverlay,
} from '@/components/ui/dialog';
import {
  Trash2,
  Plus,
  Target,
  FileText,
  Dumbbell,
  RotateCcw,
  Save,
  X,
  AlertTriangle,
} from 'lucide-react';
import type { WorkoutFormValues } from '@/lib/types';

// MARK: Init values
const MOCK_VALUES: WorkoutFormValues = {
  date: new Date(),
  notes: '',
  exercises: [] as Exercise[],
}

const getInitialValues = (userId: string): WorkoutFormValues => {
  const saved = loadFromLocalStorage(userId);
  return (
    saved || MOCK_VALUES
  );
};

const getMockValues = (): WorkoutFormValues => {
  return MOCK_VALUES;
};

// MARK: Form opts
const formOpts = formOptions({
  // defaultValues: getInitialValues
  listeners: {
    onChange: ({ formApi }) => {
      console.log('Saving form data to localStorage');
      saveToLocalStorage(formApi.state.values);
    },
    onChangeDebounceMs: 500,
  },
});

const formOptsMock = formOptions({
  defaultValues: getMockValues(),
  listeners: {
    onChange: ({ formApi }) => {
      console.log('Saving form data to localStorage');
      saveToLocalStorage(formApi.state.values);
    },
    onChangeDebounceMs: 500,
  },
});

// MARK: SUB Set Input Field
const SetInputField = withForm({
  ...formOptsMock,
  props: {} as { exerciseIndex: number },
  render: function Render({ form, exerciseIndex }) {
    return (
      <form.AppField name={`exercises[${exerciseIndex}].sets`} mode="array">
        {(setsField) => {
          return (
            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <RotateCcw className="w-4 h-4 text-neutral-400" />
                <Label className="text-xs text-neutral-400 tracking-wider">
                  SET CONFIGURATION
                </Label>
              </div>

              {/* Sets Header */}
              <div className="hidden sm:grid grid-cols-12 gap-2 text-[10px] text-neutral-400 tracking-wider px-1">
                <div className="col-span-2 text-center">SET</div>
                <div className="col-span-3 text-center">WEIGHT</div>
                <div className="col-span-2 text-center">REPS</div>
                <div className="col-span-3 text-center">TYPE</div>
                <div className="col-span-1 text-center">VOL</div>
                <div className="col-span-1"></div>
              </div>

              {/* Sets List */}
              <div className="space-y-3">
                {(setsField.state.value as any[]).map((set, setIndex) => {
                  const volume = (set.weight || 0) * (set.reps || 0);

                  return (
                    <div key={`exercises[${exerciseIndex}].sets[${setIndex}]`}>
                      {/* Mobile Layout (default) + Desktop Layout (sm:) */}
                      <div
                        className="space-y-3 rounded-lg border border-neutral-700 bg-neutral-800 p-3 
                  sm:grid sm:grid-cols-12 sm:items-end sm:gap-2 sm:space-y-0 
                  sm:rounded-none sm:border-0 sm:bg-transparent sm:p-0"
                      >
                        {/* Mobile: Header with Set # and Delete Button */}
                        <div className="flex justify-between items-center sm:hidden">
                          <div className="font-mono text-sm text-white">
                            Set #{setIndex + 1}
                          </div>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-red-500 hover:text-red-400 hover:bg-red-500/10"
                            onClick={() => setsField.removeValue(setIndex)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>

                        {/* Desktop: Set Number Column */}
                        <div className="hidden sm:block sm:col-span-2 sm:text-center">
                          <div className="text-white font-mono text-sm h-9 flex items-center justify-center">
                            #{setIndex + 1}
                          </div>
                        </div>

                        {/* Form Fields Container */}
                        <div className="grid grid-cols-2 gap-3 sm:contents">
                          {/* Weight Field */}
                          <div className="sm:col-span-3">
                            <form.AppField
                              name={`exercises[${exerciseIndex}].sets[${setIndex}].weight`}
                              children={(field) => (
                                <field.InputField
                                  label="Weight"
                                  type="number"
                                  className="sm:bg-neutral-700 sm:border-neutral-600 sm:text-white sm:text-center sm:font-mono sm:h-9"
                                />
                              )}
                            />
                          </div>

                          {/* Reps Field */}
                          <div className="sm:col-span-2">
                            <form.AppField
                              name={`exercises[${exerciseIndex}].sets[${setIndex}].reps`}
                              children={(field) => (
                                <field.InputField
                                  label="Reps"
                                  type="number"
                                  className="sm:bg-neutral-700 sm:border-neutral-600 sm:text-white sm:text-center sm:font-mono sm:h-9"
                                />
                              )}
                            />
                          </div>
                        </div>

                        {/* Set Type Field */}
                        <div className="sm:col-span-3">
                          <form.AppField
                            name={`exercises[${exerciseIndex}].sets[${setIndex}].setType`}
                            children={(field) => <field.SetTypeSelect />}
                          />
                        </div>

                        {/* Volume Display */}
                        <div
                          className="text-center text-sm text-neutral-400 pt-1 
                    sm:col-span-1 sm:text-center sm:pt-0"
                        >
                          <div className="sm:text-orange-500 sm:font-mono sm:text-sm sm:h-9 sm:flex sm:items-center sm:justify-center">
                            <span className="sm:hidden">Volume: </span>
                            <span className="font-mono text-orange-500 sm:text-orange-500">
                              {volume > 0 ? volume.toLocaleString() : '-'}
                            </span>
                          </div>
                        </div>

                        {/* Desktop: Delete Button Column */}
                        <div className="hidden sm:block sm:col-span-1 sm:text-center">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 text-red-500 hover:text-red-400 hover:bg-red-500/10"
                            onClick={() => setsField.removeValue(setIndex)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Add Set Button */}
              <Button
                onClick={() => {
                  setsField.pushValue({
                    weight: 0,
                    reps: 0,
                    setType: 'working',
                  });
                }}
                type="button"
                variant="outline"
                className="w-full border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
              >
                <Plus className="w-4 h-4 mr-2" />
                Add Set
              </Button>
            </div>
          );
        }}
      </form.AppField>
    );
  },
});

// MARK: SUB Exercise Input
const ExerciseInputField = withForm({
  ...formOptsMock,
  render: function Render({ form }) {
    return (
      <form.AppField name="exercises" mode="array">
        {(field) => {
          return (
            <div className="space-y-4">
              {(field.state.value as Exercise[]).map(
                (exercise, exerciseIndex) => {
                  const exerciseSets = exercise.sets || [];
                  const totalVolume = exerciseSets.reduce(
                    (sum, set) => sum + (set.weight || 0) * (set.reps || 0),
                    0
                  );

                  return (
                    <Card
                      key={`exercises[${exerciseIndex}]`}
                      className="relative"
                    >
                      <CardHeader className="pb-3">
                        <div className="flex items-start justify-between">
                          <div className="flex items-center gap-3">
                            <Dumbbell className="w-5 h-5 text-orange-500" />
                            <div>
                              <CardTitle className="text-sm font-bold text-white tracking-wider">
                                {exercise.name?.toUpperCase() ||
                                  'UNNAMED EXERCISE'}
                              </CardTitle>
                              <p className="text-xs text-neutral-400">
                                EX-
                                {(exerciseIndex + 1)
                                  .toString()
                                  .padStart(3, '0')}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <Badge className="bg-orange-500/20 text-orange-500">
                              {exerciseSets.length} SETS
                            </Badge>
                            {totalVolume > 0 && (
                              <Badge className="bg-white/20 text-white">
                                {totalVolume.toLocaleString()} VOL
                              </Badge>
                            )}
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-red-500 hover:text-red-400 hover:bg-red-500/10"
                              onClick={() => field.removeValue(exerciseIndex)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </div>
                        </div>
                      </CardHeader>

                      <CardContent className="space-y-4">
                        {/* MARK: Exercise Name */}
                        <form.AppField
                          name={`exercises[${exerciseIndex}].name`}
                          children={(field) => (
                            <field.InputField
                              label="EXERCISE DESIGNATION"
                              placeholder="Enter exercise name..."
                              type="text"
                            />
                          )}
                        />
                        {/* MARK: Sets */}
                        <SetInputField
                          form={form}
                          exerciseIndex={exerciseIndex}
                        />
                      </CardContent>
                    </Card>
                  );
                }
              )}
            </div>
          );
        }}
      </form.AppField>
    );
  },
});

// MARK: Main Workout Entry Form
export function WorkoutEntryForm({
  exercises,
  accessToken,
  userId,
}: {
  exercises: ExerciseOption[];
  accessToken: string;
  userId: string;
}) {
  const form = useAppForm({
    ...formOpts,    
    defaultValues: getInitialValues(userId),
    listeners: {
      onChange: ({ formApi }) => {
        console.log('Saving form data to localStorage');
        saveToLocalStorage(formApi.state.values , userId);
      },
      onChangeDebounceMs: 500,
    },
    onSubmit: async ({ value }) => {
      console.log('value', value);

      try {
        const response = await fetch('/api/workouts', {
          method: 'POST',
          headers: {
            'x-stack-access-token': accessToken,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(value),
        });

        if (!response.ok) {
          // ! TODO: Convert from text to json
          const errorText = await response.text();
          throw new Error(errorText ?? 'Failed to submit workout');
        }

        const result = await response.json();
        console.log(
          'Workout submitted! Server says: ' + JSON.stringify(result)
        );

        // MARK: localStorage clear after successful submission
        clearLocalStorage(userId);
        // Reset form to default values
        form.reset();
      } catch (error) {
        alert(error);
      }
    },
  });

  const [isModalOpen, setIsModalOpen] = useState(false);

  // Add a function to manually clear the form and localStorage
  const handleClearForm = () => {
    if (confirm('Are you sure you want to clear all form data?')) {
      clearLocalStorage(userId);
      form.reset();
      // ! TODO: Reset selected exercise
      // setSelectedExercise(undefined);
    }
  };

  return (
    <div className="min-h-screen space-y-4 bg-black p-2 lg:p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-xl md:text-2xl font-bold text-white tracking-wider mb-1">
          WORKOUT ENTRY
        </h1>
        <p className="text-xs md:text-sm text-neutral-400">
          Training session data input and management
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
        className="space-y-4"
      >
        {/* Session Information */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium tracking-wider flex items-center gap-2">
              <FileText className="w-4 h-4" />
              SESSION INFORMATION
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* MARK: Date Field */}
            <form.AppField
              name="date"
              children={(field) => <field.DatePicker />}
            />
            {/* MARK: Notes Field */}
            <form.AppField
              name="notes"
              children={(field) => <field.NotesTextarea />}
            />
          </CardContent>
        </Card>

        {/*MARK: Exercises List */}
        <ExerciseInputField form={form} />

        {/* Action Buttons */}
        <Card>
          <CardContent className="p-2 sm:p-3">
            <div className="flex flex-col gap-3">
              {/* MARK: Add Exercise Button */}
              <div>
                <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                  <DialogTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      className="w-full border-orange-500 text-orange-500 hover:bg-orange-500/10 hover:text-orange-400 bg-transparent"
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      Add Exercise
                    </Button>
                  </DialogTrigger>
                  <DialogOverlay className="bg-black/80 backdrop-blur-sm" />
                  <DialogContent className="bg-neutral-900 border-neutral-700 text-white">
                    <DialogHeader>
                      <DialogTitle className="text-sm font-medium text-neutral-300 tracking-wider flex items-center gap-2">
                        <Target className="w-4 h-4" />
                        EXERCISE SELECTION
                      </DialogTitle>
                    </DialogHeader>
                    <div className="mt-4">
                      <form.AppField
                        name="exercises"
                        mode="array"
                        children={(field) => (
                          <field.AddExerciseField
                            exercises={exercises}
                            onExerciseAdded={() => setIsModalOpen(false)}
                          />
                        )}
                      />
                    </div>
                  </DialogContent>
                </Dialog>
              </div>

              {/* Action Buttons Row */}
              <div className="flex flex-col sm:flex-row gap-2 sm:gap-3 justify-between items-stretch">
                <div className="flex flex-col sm:flex-row gap-2 flex-1">
                  <form.Subscribe
                    selector={(state) => [state.canSubmit, state.isSubmitting]}
                    children={([canSubmit, isSubmitting]) => (
                      <Button
                        type="submit"
                        disabled={!canSubmit}
                        className="flex-1 bg-orange-500 hover:bg-orange-600 text-white py-1.5 h-auto text-sm"
                      >
                        <Save className="w-3.5 h-3.5 mr-1.5" />
                        {isSubmitting ? 'SAVING...' : 'SAVE'}
                      </Button>
                    )}
                  />

                  <Button
                    type="button"
                    variant="outline"
                    onClick={handleClearForm}
                    className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent py-1.5 h-auto text-sm"
                  >
                    <X className="w-3.5 h-3.5 mr-1.5" />
                    <span className="hidden sm:inline">Clear Form</span>
                    <span className="sm:hidden">Clear</span>
                  </Button>
                </div>

                <div className="hidden sm:flex items-center justify-end gap-2 text-[10px] sm:text-xs text-neutral-500 px-2">
                  <AlertTriangle className="w-3 h-3 flex-shrink-0" />
                  <span>Auto-save</span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </form>
    </div>
  );
}
