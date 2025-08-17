import { createFileRoute } from '@tanstack/react-router';
import { Suspense, useState } from 'react';
import { useAppForm } from '@/hooks/form';
import { useSaveWorkoutMutation } from '@/lib/api/workouts';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { MiniChart } from './-components/mini-chart';
import { Plus, Save, Trash2, X } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import { clearLocalStorage, saveToLocalStorage } from '@/lib/local-storage';
import type { exercise_ExerciseResponse } from '@/generated';
import { exercisesQueryOptionsWithUser } from '@/lib/api/exercises';
import { getInitialValues } from './-components/form-options';
import { ExerciseScreen2 } from './-components/exercise-screen';
import { AddExerciseScreen } from './-components/add-exercise-screen';
import type { User } from '@/lib/api/auth';

function WorkoutTracker({
  user,
  exercises,
}: {
  user: Exclude<User, null>; // Use Exclude to remove null from the union type
  exercises: exercise_ExerciseResponse[];
}) {
  const [currentView, setCurrentView] = useState<
    'main' | 'exercise' | 'add-exercise'
  >('main');
  const [selectedExerciseIndex, setSelectedExerciseIndex] = useState<
    number | null
  >(null);

  const saveWorkout = useSaveWorkoutMutation(user);
  const form = useAppForm({
    defaultValues: getInitialValues(user.id),
    listeners: {
      onChange: ({ formApi }) => {
        console.log('Saving form data to localStorage');
        saveToLocalStorage(formApi.state.values, user.id);
      },
      onChangeDebounceMs: 500,
    },
    onSubmit: async ({ value }) => {
      await saveWorkout.mutateAsync(value, {
        onSuccess: () => {
          clearLocalStorage(user.id);
          form.reset();
          setSelectedExerciseIndex(null);
        },
        onError: (error) => {
          alert(error);
        },
      });
   
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
      clearLocalStorage(user.id);
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
      <div className="max-w-md mx-auto space-y-6 px-4 pb-8">
        <div className="flex items-center justify-between pt-6 pb-2">
          <div>
            <h1 className="font-bold text-2xl tracking-tight text-foreground">
              Today's Training
            </h1>
          </div>
          <div>
            <Button
              type="button"
              variant="ghost"
              onClick={handleClearForm}
              size="sm"
            >
              <X className="w-3.5 h-3.5 mr-1.5" />
              <span>Clear</span>
            </Button>
          </div>
        </div>
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
              return (
                <div className="space-y-3">
                  {field.state.value.map((exercise, exerciseIndex) => (
                    <Card
                      key={`exercise-${exerciseIndex}`}
                      className="p-4 cursor-pointer hover:shadow-md transition-all duration-200"
                      onClick={() => handleExerciseClick(exerciseIndex)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2 mb-2">
                              <div className="w-2 h-2 bg-primary rounded-full"></div>
                              <span className="text-primary font-medium text-sm">
                                {exercise.name}
                              </span>
                            </div>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-primary hover:text-primary/80 hover:bg-primary/10"
                              onClick={(e) => {
                                e.stopPropagation();
                                field.removeValue(exerciseIndex);
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
                                  {exercise.sets.reduce(
                                    (acc, set) =>
                                      acc + (set.reps || 0) * (set.weight || 0),
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

          {/* MARK: Buttons */}
          <div className="py-6">
            <Button
              type="button"
              variant="outline"
              className="w-full text-base font-semibold rounded-lg"
              onClick={() => setCurrentView('add-exercise')}
            >
              <Plus className="w-5 h-5 mr-2" />
              Add Exercise
            </Button>
          </div>
          <div className="mt-8">
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
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
        </form>
      </div>
    </Suspense>
  );
}

export const Route = createFileRoute('/_auth/workouts/new-2')({
  loader: async ({
    context,
  }): Promise<{
    user: Exclude<User, null>;
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
    context.queryClient.ensureQueryData(exercisesQueryOptionsWithUser(user));
    return { user };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useLoaderData();
  const { data: exercises } = useSuspenseQuery(
    exercisesQueryOptionsWithUser(user)
  );
  return <WorkoutTracker user={user} exercises={exercises} />;
}
