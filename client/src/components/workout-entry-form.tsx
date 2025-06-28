import { useForm } from '@tanstack/react-form';
import { useState } from 'react';
import {
  clearLocalStorage,
  loadFromLocalStorage,
  saveToLocalStorage,
} from '@/lib/local-storage';
import type { Exercise, ExerciseOption } from '@/lib/types';
import { ExerciseCombobox } from '@/components/exercise-combobox';
import { Button } from '@/components/ui/button';
import { DatePicker } from '@/components/date-picker';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { SetTypeSelect } from '@/components/set-type-select';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Trash2,
  Plus,
  Target,
  Calendar,
  FileText,
  Dumbbell,
  RotateCcw,
  Save,
  X,
  AlertTriangle,
} from 'lucide-react';
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
    <div className="p-6 space-y-6 bg-black min-h-screen">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white tracking-wider mb-2">
          WORKOUT ENTRY PROTOCOL
        </h1>
        <p className="text-sm text-neutral-400">
          Training session data input and management
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
        className="space-y-6"
      >
        {/* Session Information */}
        <Card className="bg-neutral-900 border-neutral-700">
          <CardHeader>
            <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider flex items-center gap-2">
              <FileText className="w-4 h-4" />
              SESSION INFORMATION
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Date Field */}
            <form.Field name="date">
              {(field) => {
                return (
                  <div className="space-y-3">
                    <Label className="text-xs text-neutral-400 tracking-wider flex items-center gap-2">
                      <Calendar className="w-3 h-3" />
                      TRAINING DATE
                    </Label>
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

            {/* Notes Field */}
            <form.Field name="notes">
              {(field) => {
                return (
                  <div className="space-y-3">
                    <Label
                      htmlFor={field.name}
                      className="text-xs text-neutral-400 tracking-wider"
                    >
                      SESSION NOTES
                    </Label>
                    <Textarea
                      id={field.name}
                      name={field.name}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      className="bg-neutral-800 border-neutral-600 text-white placeholder-neutral-500 min-h-[80px]"
                      placeholder="Enter workout notes, focus areas, or observations..."
                    />
                  </div>
                );
              }}
            </form.Field>
          </CardContent>
        </Card>

        {/* Exercise Selection */}
        <Card className="bg-neutral-900 border-neutral-700">
          <CardHeader>
            <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider flex items-center gap-2">
              <Target className="w-4 h-4" />
              EXERCISE SELECTION
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4 md:flex-row md:items-end">
              <div className="space-y-3 flex-1">
                <Label className="text-xs text-neutral-400 tracking-wider">
                  EXERCISE DATABASE
                </Label>
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
                    className="bg-orange-500 hover:bg-orange-600 text-white"
                    disabled={!selectedExercise?.name}
                  >
                    <Plus className="w-4 h-4 mr-2" />
                    Deploy Exercise
                  </Button>
                )}
              </form.Field>
            </div>
          </CardContent>
        </Card>

        {/* Exercises List */}
        <form.Field name="exercises" mode="array">
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
                        className="bg-neutral-900 border-neutral-700 relative"
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
                          {/* Exercise Name Field */}
                          <form.Field name={`exercises[${exerciseIndex}].name`}>
                            {(subField) => {
                              return (
                                <div className="space-y-2">
                                  <Label className="text-xs text-neutral-400 tracking-wider">
                                    EXERCISE DESIGNATION
                                  </Label>
                                  <Input
                                    id={subField.name}
                                    name={subField.name}
                                    value={subField.state.value}
                                    onBlur={subField.handleBlur}
                                    onChange={(e) =>
                                      subField.handleChange(e.target.value)
                                    }
                                    className="bg-neutral-800 border-neutral-600 text-white placeholder-neutral-500"
                                    placeholder="Enter exercise name..."
                                  />
                                </div>
                              );
                            }}
                          </form.Field>

                          {/* Sets */}
                          <form.Field
                            name={`exercises[${exerciseIndex}].sets`}
                            mode="array"
                          >
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
                                  {(setsField.state.value as any[]).length >
                                    0 && (
                                    <div className="grid grid-cols-12 gap-3 text-xs text-neutral-400 tracking-wider px-2">
                                      <div className="col-span-1">SET</div>
                                      <div className="col-span-3">
                                        WEIGHT (LBS)
                                      </div>
                                      <div className="col-span-2">REPS</div>
                                      <div className="col-span-3">SET TYPE</div>
                                      <div className="col-span-2">VOLUME</div>
                                      <div className="col-span-1">ACTION</div>
                                    </div>
                                  )}

                                  {/* Sets List */}
                                  {(setsField.state.value as any[]).map(
                                    (set, setIndex) => {
                                      const volume =
                                        (set.weight || 0) * (set.reps || 0);

                                      return (
                                        <div
                                          key={`exercises[${exerciseIndex}].sets[${setIndex}]`}
                                          className="grid grid-cols-12 gap-3 items-end p-2 bg-neutral-800 rounded border border-neutral-700"
                                        >
                                          {/* Set Number */}
                                          <div className="col-span-1 text-center">
                                            <div className="text-white font-mono text-sm">
                                              #{setIndex + 1}
                                            </div>
                                          </div>

                                          {/* Weight */}
                                          <form.Field
                                            name={`exercises[${exerciseIndex}].sets[${setIndex}].weight`}
                                          >
                                            {(subField) => {
                                              return (
                                                <div className="col-span-3">
                                                  <Input
                                                    type="number"
                                                    value={
                                                      subField.state.value || ''
                                                    }
                                                    onChange={(e) =>
                                                      subField.handleChange(
                                                        Number(
                                                          e.target.value
                                                        ) || 0
                                                      )
                                                    }
                                                    placeholder="0"
                                                    className="bg-neutral-700 border-neutral-600 text-white text-center font-mono"
                                                  />
                                                </div>
                                              );
                                            }}
                                          </form.Field>

                                          {/* Reps */}
                                          <form.Field
                                            name={`exercises[${exerciseIndex}].sets[${setIndex}].reps`}
                                          >
                                            {(subField) => {
                                              return (
                                                <div className="col-span-2">
                                                  <Input
                                                    type="number"
                                                    value={
                                                      subField.state.value || ''
                                                    }
                                                    onChange={(e) =>
                                                      subField.handleChange(
                                                        Number(
                                                          e.target.value
                                                        ) || 0
                                                      )
                                                    }
                                                    placeholder="0"
                                                    className="bg-neutral-700 border-neutral-600 text-white text-center font-mono"
                                                  />
                                                </div>
                                              );
                                            }}
                                          </form.Field>

                                          {/* Set Type */}
                                          <form.Field
                                            name={`exercises[${exerciseIndex}].sets[${setIndex}].setType`}
                                          >
                                            {(subField) => {
                                              return (
                                                <div className="col-span-3">
                                                  <SetTypeSelect
                                                    value={subField.state.value}
                                                    onChange={
                                                      subField.handleChange
                                                    }
                                                  />
                                                </div>
                                              );
                                            }}
                                          </form.Field>

                                          {/* Volume */}
                                          <div className="col-span-2 text-center">
                                            <div className="text-orange-500 font-mono text-sm">
                                              {volume > 0
                                                ? volume.toLocaleString()
                                                : '-'}
                                            </div>
                                          </div>

                                          {/* Remove Set */}
                                          <div className="col-span-1 text-center">
                                            <Button
                                              type="button"
                                              variant="ghost"
                                              size="icon"
                                              className="h-8 w-8 text-red-500 hover:text-red-400 hover:bg-red-500/10"
                                              onClick={() =>
                                                setsField.removeValue(setIndex)
                                              }
                                            >
                                              <Trash2 className="h-3 w-3" />
                                            </Button>
                                          </div>
                                        </div>
                                      );
                                    }
                                  )}

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
                          </form.Field>
                        </CardContent>
                      </Card>
                    );
                  }
                )}
              </div>
            );
          }}
        </form.Field>

        {/* Action Buttons */}
        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="pt-6">
            <div className="flex gap-4 justify-between">
              <div className="flex gap-4">
                <form.Subscribe
                  selector={(state) => [state.canSubmit, state.isSubmitting]}
                  children={([canSubmit, isSubmitting]) => (
                    <Button
                      type="submit"
                      disabled={!canSubmit}
                      className="bg-orange-500 hover:bg-orange-600 text-white"
                    >
                      <Save className="w-4 h-4 mr-2" />
                      {isSubmitting ? 'TRANSMITTING...' : 'SUBMIT WORKOUT'}
                    </Button>
                  )}
                />

                <Button
                  type="button"
                  variant="outline"
                  onClick={handleClearForm}
                  className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
                >
                  <X className="w-4 h-4 mr-2" />
                  Clear Form
                </Button>
              </div>

              <div className="flex items-center gap-2 text-xs text-neutral-500">
                <AlertTriangle className="w-3 h-3" />
                <span>Auto-save enabled</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </form>
    </div>
  );
}
