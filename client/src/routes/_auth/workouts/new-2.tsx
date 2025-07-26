import { createFileRoute } from '@tanstack/react-router';
import { Suspense, useState } from 'react';
import { Plus, Trash2, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { fetchExerciseOptions } from '@/lib/api/exercises';
import { clearLocalStorage, saveToLocalStorage } from '@/lib/local-storage';
import { useAppForm } from '@/hooks/form';
import type { ExerciseOption } from '@/lib/types';
import { MiniChart } from './-components/mini-chart';
import { formOpts, getInitialValues } from './-components/form-options';
import { ExerciseScreen2 } from './-components/exercise-screen';
import { AddExerciseScreen } from './-components/add-exercise';
import { Spinner } from '@/components/ui/spinner';

export const Route = createFileRoute('/_auth/workouts/new-2')({
  loader: async ({
    context,
  }): Promise<{
    accessToken: string;
    userId: string;
    exercises: ExerciseOption[];
  }> => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    if (!user.id || typeof user.id !== 'string') {
      throw new Error('User ID not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const exercises = await fetchExerciseOptions(accessToken);
    return { accessToken, userId: user.id, exercises };
  },
  component: WorkoutTracker,
});

export interface ExerciseNew {
  id: string;
  name: string;
  sets: number;
  volume: number;
  lastUpdated: string;
}

interface WorkoutSession {
  date: string;
  notes: string;
  exercises: ExerciseNew[];
}

// MARK: - WorkoutTracker
export default function WorkoutTracker() {
  const { accessToken, exercises, userId } = Route.useLoaderData();
  console.log('exercises', exercises);
  const [currentView, setCurrentView] = useState<
    'main' | 'exercise' | 'add-exercise'
  >('main');
  const [selectedExerciseIndex, setSelectedExerciseIndex] = useState<
    number | null
  >(null);

  // MARK: useForm
  const form = useAppForm({
    ...formOpts,
    defaultValues: getInitialValues(userId),
    listeners: {
      onChange: ({ formApi }) => {
        console.log('Saving form data to localStorage');
        saveToLocalStorage(formApi.state.values, userId);
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

        // localStorage clear after successful submission
        clearLocalStorage(userId);
        // Reset form to default values
        form.reset();
      } catch (error) {
        alert(error);
      }
    },
  });

  const handleAddExercise = (index: number) => {
    setSelectedExerciseIndex(index);
    setCurrentView('exercise');
  };

  const handleExerciseClick = (index: number) => {
    setSelectedExerciseIndex(index);
    setCurrentView('exercise');
  };

  const handleClearForm = () => {
    if (confirm('Are you sure you want to clear all form data?')) {
      clearLocalStorage(userId);
      form.reset();
      setSelectedExerciseIndex(null);
    }
  };

  // MARK: Screens
  if (currentView === 'add-exercise') {
    return (
      <Suspense
        fallback={
          <div className="fixed inset-0 flex items-center justify-center">
            <Spinner size="large" />
          </div>
        }
      >
        <AddExerciseScreen
          form={form}
          exercises={exercises}
          onAddExercise={handleAddExercise}
          onBack={() => setCurrentView('main')}
        />
      </Suspense>
    );
  }

  if (
    currentView === 'exercise' &&
    selectedExerciseIndex !== null &&
    form.state.values.exercises.length > 0
  ) {
    return (
      <Suspense
        fallback={
          <div className="fixed inset-0 flex items-center justify-center">
            <Spinner size="large" />
          </div>
        }
      >
        <ExerciseScreen2
          form={form}
          exerciseIndex={selectedExerciseIndex}
          onBack={() => setCurrentView('main')}
        />
      </Suspense>
    );
  }

  // MARK: Render
  return (
    <Suspense
      fallback={
        <div className="fixed inset-0 flex items-center justify-center">
          <Spinner size="large" />
        </div>
      }
    >
      <div>
        <div className="px-4 pb-8">
          <div className="max-w-md mx-auto space-y-6">
            {/* Header */}
            <div className="pt-6 pb-2">
              <h1 className="font-bold text-3xl tracking-tight text-foreground">
                Summary
              </h1>
            </div>

            {/* Quick Stats */}
            <form
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
              }}
            >
              <div className="grid grid-cols-2 gap-4 mb-4">
                {/* MARK: Date/Notes*/}
                <form.AppField
                  name="date"
                  children={(field) => <field.DatePicker2 />}
                />
                <form.AppField
                  name="notes"
                  children={(field) => <field.NotesTextarea2 />}
                />
              </div>

              {/* MARK: Exercise Cards */}
              <form.AppField
                name="exercises"
                mode="array"
                children={(field) => {
                  console.log('field', field.state.value);
                  return (
                    <div className="space-y-3">
                      {field.state.value.map((exercise, exerciseIndex) => (
                        <Card
                          key={`exercise-${exerciseIndex}`}
                          className="p-4 cursor-pointer hover:shadow-md transition-all duration-200"
                          onClick={() => handleExerciseClick(exerciseIndex)} // MARK: ! TODO index
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex-1">
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 mb-2">
                                  <div className="w-2 h-2 bg-primary rounded-full"></div>
                                  <span className="text-primary font-medium text-sm">
                                    {exercise.name}
                                  </span>
                                  {/* <span className="font-semibold text-sm tracking-tight uppercase text-muted-foreground ml-auto">
                                  {exercise.lastUpdated}
                                  </span> */}
                                </div>
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  className="h-8 w-8 text-primary hover:text-primary/80 hover:bg-primary/10"
                                  onClick={() =>
                                    field.removeValue(exerciseIndex)
                                  }
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
                                      {exercise.sets.reduce(
                                        (acc, set) =>
                                          acc +
                                          (set.reps || 0) * (set.weight || 0),
                                        0
                                      )}
                                    </div>
                                    <div className="font-semibold text-sm tracking-tight uppercase text-muted-foreground">
                                      volume
                                    </div>
                                  </div>
                                  <MiniChart
                                    data={[3, 5, 2, 4, 6, 3, 4]}
                                    activeIndex={6}
                                  />
                                </div>
                              </div>
                            </div>
                          </div>
                        </Card>
                      ))}
                    </div>
                  );
                }}
              />

              {/* Add Exercise Button */}
              <div className="pt-6">
                <Button
                  className="bg-primary text-primary-foreground hover:bg-primary/90 w-full py-4 text-base font-semibold rounded-lg"
                  onClick={() => setCurrentView('add-exercise')}
                >
                  <Plus className="w-5 h-5 mr-2" />
                  Add Exercise
                </Button>
              </div>
              {/* ! TODO: Save and Cancel buttons */}
              <div className="flex gap-2 mt-4">
                <Button>Save</Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleClearForm}
                >
                  <X className="w-3.5 h-3.5 mr-1.5" />
                  <span className="hidden sm:inline">Clear Form</span>
                  <span className="sm:hidden">Clear</span>
                </Button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Suspense>
  );
}
