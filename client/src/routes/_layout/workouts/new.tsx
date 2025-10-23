import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { Suspense } from 'react';
import { z } from 'zod';
import { useAppForm } from '@/hooks/form';
import { useSaveWorkoutMutation, type WorkoutFocus } from '@/lib/api/workouts';
import { useSuspenseQuery, useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { MiniChart } from './-components/mini-chart';
import { Plus, Save, Trash2, X } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import { clearLocalStorage, saveToLocalStorage } from '@/lib/local-storage';
import { type DbExercise } from '@/lib/api/exercises';
import { getInitialValues } from './-components/form-options';
import { postDemoWorkoutsMutation } from '@/lib/demo-data/query-options';
import { initializeDemoData } from '@/lib/demo-data/storage';
import {
  getExercisesQueryOptions,
  getWorkoutsFocusQueryOptions,
} from '@/lib/api/unified-query-options';

import { AddExerciseScreen } from './-components/add-exercise-screen';
import {
  ExerciseHeader,
  ExerciseScreen,
  ExerciseSets,
} from './-components/exercise-screen';
import { RecentSets } from './-components/recent-sets-display';

function WorkoutTracker({
  user,
  exercises,
  workoutsFocus,
}: {
  user: CurrentUser | CurrentInternalUser | null; // need user for localStorage
  exercises: DbExercise[];
  workoutsFocus: WorkoutFocus[];
}) {
  const { addExercise, exerciseIndex } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

  const saveWorkoutApi = useSaveWorkoutMutation();
  const saveWorkoutDemo = useMutation(postDemoWorkoutsMutation());
  const saveWorkout = user ? saveWorkoutApi : saveWorkoutDemo;
  const form = useAppForm({
    // MARK: TODO: Handle user.id when chat is implemented
    defaultValues: getInitialValues(user?.id),
    listeners: {
      onChange: ({ formApi }) => {
        console.log('Saving form data to localStorage');
        saveToLocalStorage(formApi.state.values, user?.id);
      },
      onChangeDebounceMs: 500,
    },
    onSubmit: async ({ value }) => {
      const trimmedValue = {
        ...value,
        notes: value.notes?.trim() || undefined,
        workoutFocus: value.workoutFocus?.trim() || undefined,
      };

      await saveWorkout.mutateAsync(
        { body: trimmedValue },
        {
          onSuccess: () => {
            clearLocalStorage(user?.id);
            form.reset();
            navigate({ search: {} });
          },
          onError: (error) => {
            alert(error);
          },
        }
      );
    },
  });

  // Helper to get exercise ID from exercises list
  const getExerciseId = (exerciseName: string): number | null => {
    const exercise = exercises.find((ex) => ex.name === exerciseName);
    return exercise?.id || null;
  };

  const handleAddExercise = (index: number) => {
    navigate({ search: { exerciseIndex: index } });
  };

  const handleExerciseClick = (index: number) => {
    navigate({ search: { exerciseIndex: index } });
  };

  const handleClearForm = () => {
    if (confirm('Are you sure you want to clear all form data?')) {
      clearLocalStorage(user?.id);
      form.reset();
      navigate({ search: {} });
    }
  };

  // MARK: Screens
  if (addExercise) {
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
          onBack={() => navigate({ search: {} })}
        />
      </Suspense>
    );
  }

  if (exerciseIndex !== undefined) {
    const exercises = form.state.values.exercises;

    // Validate exercise index
    if (exerciseIndex < 0 || exerciseIndex >= exercises.length) {
      // Silently redirect to main
      navigate({ search: {} });
      return null;
    }

    const exerciseName = exercises[exerciseIndex]?.name;
    const exerciseId = getExerciseId(exerciseName);

    return (
      <Suspense
        fallback={
          <div className="fixed inset-0 flex items-center justify-center">
            <Spinner size="large" />
          </div>
        }
      >
        <ExerciseScreen
          header={
            <ExerciseHeader
              form={form}
              exerciseIndex={exerciseIndex}
              onBack={() => navigate({ search: {} })}
            />
          }
          recentSets={
            <RecentSets exerciseId={exerciseId} user={user} />
          }
          sets={
            <ExerciseSets form={form} exerciseIndex={exerciseIndex} />
          }
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
            {/* MARK: Date/Notes/Focus */}
            <form.AppField
              name="date"
              children={(field) => <field.DatePicker2 />}
            />
            <form.AppField
              name="workoutFocus"
              children={(field) => (
                <field.WorkoutFocusCombobox workoutsFocus={workoutsFocus} />
              )}
            />
            <div className="col-span-2">
              <form.AppField
                name="notes"
                children={(field) => <field.NotesTextarea2 />}
              />
            </div>
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
                      data-testid="new-workout-exercise-card"
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
              onClick={() => navigate({ search: { addExercise: true } })}
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

const workoutSearchSchema = z.object({
  addExercise: z.boolean().default(false),
  exerciseIndex: z.coerce.number().int().optional(),
});

export const Route = createFileRoute('/_layout/workouts/new')({
  validateSearch: workoutSearchSchema,
  loader: ({ context }) => {
    if (!context.user) initializeDemoData();
    context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
    context.queryClient.ensureQueryData(
      getWorkoutsFocusQueryOptions(context.user)
    );
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  const { data: exercisesResponse } = useSuspenseQuery(
    getExercisesQueryOptions(user)
  );
  const { data: workoutsFocusValues } = useSuspenseQuery(
    getWorkoutsFocusQueryOptions(user)
  );

  // Convert API response to our cleaner DbExercise type
  const exercises: DbExercise[] = exercisesResponse.map((ex) => ({
    id: ex.id,
    name: ex.name,
  }));

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <WorkoutTracker
      user={user}
      exercises={exercises}
      workoutsFocus={workoutsFocus}
    />
  );
}
