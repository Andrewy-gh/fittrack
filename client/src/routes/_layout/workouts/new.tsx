import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { Suspense, useMemo, useState } from 'react';
import { z } from 'zod';
import { useAppForm } from '@/hooks/form';
import { useSaveWorkoutMutation, type WorkoutFocus } from '@/lib/api/workouts';
import { useSuspenseQuery, useMutation } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { MiniChart } from './-components/mini-chart';
import { X } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  clearLocalStorage,
  loadFromLocalStorage,
  saveToLocalStorage,
} from '@/lib/local-storage';
import { type DbExercise } from '@/lib/api/exercises';
import { getInitialValues, MOCK_VALUES } from './-components/form-options';
import { postDemoWorkoutsMutation } from '@/lib/demo-data/query-options';
import { initializeDemoData } from '@/lib/demo-data/storage';
import { toast } from 'sonner';
import {
  getExercisesQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsQueryOptions,
  getWorkoutsFocusQueryOptions,
} from '@/lib/api/unified-query-options';
import { formatWeight } from '@/lib/utils';
import {
  ErrorBoundary,
  FullScreenErrorFallback,
} from '@/components/error-boundary';
import { queryClient } from '@/lib/api/api';
import { ExerciseContextPanel } from './-components/exercise-context-panel';
import { LastWorkoutNoteSection } from '@/components/workouts/last-workout-note-section';
import {
  buildWorkoutDraftFromHistory,
  getLatestWorkoutNote,
} from '@/lib/workout-insights';
import { formatExerciseGoalSummary, getExerciseGoal } from '@/lib/exercise-goals';
import type { WorkoutWorkoutResponse } from '@/client';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

import { AddExerciseScreen } from './-components/add-exercise-screen';
import {
  ExerciseHeader,
  ExerciseScreen,
  ExerciseSets,
} from './-components/exercise-screen';
import { RecentSets } from './-components/recent-sets-display';
import { hasWorkoutDraftContent } from './-components/workout-form-helpers';
import {
  WorkoutExerciseCards,
  WorkoutFormActions,
  WorkoutMetadataFields,
} from './-components/workout-form-sections';

function WorkoutTracker({
  user,
  exercises,
  workouts,
  workoutsFocus,
}: {
  user: CurrentUser | CurrentInternalUser | null; // need user for localStorage
  exercises: DbExercise[];
  workouts: WorkoutWorkoutResponse[];
  workoutsFocus: WorkoutFocus[];
}) {
  const { addExercise, exerciseIndex, newExercise } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const [isClearDialogOpen, setIsClearDialogOpen] = useState(false);
  const [pendingTemplateWorkoutId, setPendingTemplateWorkoutId] = useState<
    number | null
  >(null);

  const saveWorkoutApi = useSaveWorkoutMutation();
  const saveWorkoutDemo = useMutation(postDemoWorkoutsMutation());
  const saveWorkout = user ? saveWorkoutApi : saveWorkoutDemo;
  const form = useAppForm({
    defaultValues: getInitialValues(user?.id),
    listeners: {
      onChange: ({ formApi }) => {
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
            toast.success('Workout saved successfully');
            clearLocalStorage(user?.id);
            form.reset(MOCK_VALUES);
            navigate({ search: {} });
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

  const latestWorkoutNote = getLatestWorkoutNote(workouts);
  const focusAreaTemplates = useMemo(() => {
    const focusMap = new Map<
      string,
      { focus: string; workoutId: number }
    >();

    for (const workout of workouts) {
      const focus = workout.workout_focus?.trim();
      if (!focus || focusMap.has(focus.toLowerCase())) {
        continue;
      }

      focusMap.set(focus.toLowerCase(), {
        focus,
        workoutId: workout.id,
      });
    }

    return Array.from(focusMap.values());
  }, [workouts]);

  const loadWorkoutTemplate = async (workoutId: number) => {
    const workoutToRepeat = await queryClient.fetchQuery(
      getWorkoutByIdQueryOptions(user, workoutId)
    );
    const nextDraft = buildWorkoutDraftFromHistory(workoutToRepeat);

    form.reset(nextDraft);
    saveToLocalStorage(nextDraft, user?.id);
    navigate({ search: {} });
    toast.success('Loaded workout structure');
  };

  const handleRepeatWorkout = async (workoutId: number) => {
    const hasDraft =
      hasWorkoutDraftContent(form.state.values) ||
      loadFromLocalStorage(user?.id) !== null;
    if (hasDraft) {
      setPendingTemplateWorkoutId(workoutId);
      return;
    }

    await loadWorkoutTemplate(workoutId);
  };

  const handleClearForm = () => {
    clearLocalStorage(user?.id);
    form.reset(MOCK_VALUES);
    navigate({ search: {} });
    setIsClearDialogOpen(false);
  };

  const renderExerciseGoalSummary = (exercise: { name: string }) => {
    const exerciseGoalSummary = formatExerciseGoalSummary(
      getExerciseGoal({
        exerciseId: getExerciseId(exercise.name),
        exerciseName: exercise.name,
      })
    );

    if (!exerciseGoalSummary) {
      return null;
    }

    return (
      <p className="text-xs text-muted-foreground">
        Goal: {exerciseGoalSummary}
      </p>
    );
  };

  // MARK: Screens
  if (addExercise) {
    return (
      <ErrorBoundary
        fallback={
          <FullScreenErrorFallback
            message="Failed to load exercise selection"
            onAction={() => navigate({ search: {} })}
            actionLabel="Go Back"
          />
        }
      >
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
            onBack={() => navigate({ search: {} })}
          />
        </Suspense>
      </ErrorBoundary>
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

    const handleExerciseBack = () => {
      const currentExercise = form.state.values.exercises[exerciseIndex];
      if (newExercise && currentExercise && currentExercise.sets.length === 0) {
        form.removeFieldValue('exercises', exerciseIndex);
      }
      navigate({ search: {} });
    };

    const handleDiscardNewExercise = () => {
      if (newExercise) {
        form.removeFieldValue('exercises', exerciseIndex);
      }
      navigate({ search: {} });
    };

    return (
      <ErrorBoundary
        fallback={
          <FullScreenErrorFallback
            message="Failed to load exercise details"
            onAction={() => navigate({ search: {} })}
            actionLabel="Go Back"
          />
        }
      >
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
                onBack={handleExerciseBack}
              />
            }
            contextPanel={
              <ExerciseContextPanel
                exerciseId={exerciseId}
                exerciseName={exerciseName}
              />
            }
            recentSets={<RecentSets exerciseId={exerciseId} user={user} />}
            sets={
              <ExerciseSets
                form={form}
                exerciseIndex={exerciseIndex}
                isNewExercise={newExercise}
                onDiscardNewExercise={handleDiscardNewExercise}
              />
            }
          />
        </Suspense>
      </ErrorBoundary>
    );
  }

  // MARK: Render
  return (
    <ErrorBoundary
      fallback={
        <FullScreenErrorFallback message="Failed to load workout form" />
      }
    >
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
                onClick={() => setIsClearDialogOpen(true)}
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
            {focusAreaTemplates.length > 0 && (
              <Card className="mb-4 p-4">
                <div className="space-y-3">
                  <div className="space-y-2">
                    <p className="text-sm font-semibold text-card-foreground">
                      Start from a recent focus area
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Load the most recent workout for the focus area you want
                      to train today.
                    </p>
                  </div>
                  <Select
                    onValueChange={(value) => handleRepeatWorkout(Number(value))}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Choose a focus area" />
                    </SelectTrigger>
                    <SelectContent>
                      {focusAreaTemplates.map((template) => (
                        <SelectItem
                          key={template.focus}
                          value={String(template.workoutId)}
                        >
                          {template.focus}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </Card>
            )}

            {latestWorkoutNote && (
              <div className="mb-4">
                <LastWorkoutNoteSection
                  title="Last Workout Note"
                  note={latestWorkoutNote.note}
                  dateLabel={latestWorkoutNote.date}
                />
              </div>
            )}

            <WorkoutMetadataFields form={form} workoutsFocus={workoutsFocus} />

            {/* MARK: Exercise Cards */}
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => {
                return (
                  <WorkoutExerciseCards
                    exercises={field.state.value}
                    dataTestId="new-workout-exercise-card"
                    onRemoveExercise={field.removeValue}
                    formatVolume={formatWeight}
                    renderNameSupplement={renderExerciseGoalSummary}
                    renderMetrics={() => (
                      <MiniChart
                        data={[3, 5, 2, 4, 6, 3, 4]}
                        activeIndex={6}
                      />
                    )}
                  />
                );
              }}
            />

            {/* MARK: Buttons */}
            <WorkoutFormActions form={form} />
          </form>
        </div>
        <AlertDialog open={isClearDialogOpen} onOpenChange={setIsClearDialogOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Clear workout draft?</AlertDialogTitle>
              <AlertDialogDescription>
                This removes the current workout draft from the form and local
                storage.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleClearForm}>
                Clear draft
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <AlertDialog
          open={pendingTemplateWorkoutId !== null}
          onOpenChange={(open) => {
            if (!open) {
              setPendingTemplateWorkoutId(null);
            }
          }}
        >
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Replace current draft?</AlertDialogTitle>
              <AlertDialogDescription>
                Loading this focus-area template will replace the workout draft
                already in progress.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Keep current draft</AlertDialogCancel>
              <AlertDialogAction
                onClick={async () => {
                  if (pendingTemplateWorkoutId == null) {
                    return;
                  }

                  const workoutId = pendingTemplateWorkoutId;
                  setPendingTemplateWorkoutId(null);
                  await loadWorkoutTemplate(workoutId);
                }}
              >
                Replace draft
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </Suspense>
    </ErrorBoundary>
  );
}

const workoutSearchSchema = z.object({
  addExercise: z.boolean().optional(),
  exerciseIndex: z.coerce.number().int().optional(),
  newExercise: z.boolean().optional(),
});

export const Route = createFileRoute('/_layout/workouts/new')({
  validateSearch: workoutSearchSchema,
  loader: ({ context }) => {
    if (!context.user) initializeDemoData();
    context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
    context.queryClient.ensureQueryData(getWorkoutsQueryOptions(context.user));
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
  const { data: workouts } = useSuspenseQuery(getWorkoutsQueryOptions(user));
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
      workouts={workouts}
      workoutsFocus={workoutsFocus}
    />
  );
}
