import { useNavigate } from "@tanstack/react-router";
import { Suspense, useMemo, useState, type ReactNode } from "react";
import { useAppForm } from "@/hooks/form";
import {
  useSaveWorkoutMutation,
  type WorkoutFocus,
} from "@/features/workouts/api/workouts";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { MiniChart } from "@/features/workouts/components/form/mini-chart";
import { X } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import type { CurrentUser, CurrentInternalUser } from "@stackframe/react";
import {
  type WorkoutDraftStorage,
  workoutDraftStorage,
} from "@/lib/local-storage";
import { type DbExercise } from "@/features/exercises/api/exercises";
import {
  getInitialValues,
  MOCK_VALUES,
} from "@/features/workouts/components/form/form-options";
import { postDemoWorkoutsMutation } from "@/lib/demo-data/query-options";
import { toast } from "sonner";
import { getWorkoutByIdQueryOptions } from "@/lib/api/unified-query-options";
import { formatWeight } from "@/lib/utils";
import {
  ErrorBoundary,
  FullScreenErrorFallback,
} from "@/components/error-boundary";
import { queryClient } from "@/lib/api/api";
import { ExerciseContextPanel } from "@/features/workouts/components/form/exercise-context-panel";
import { LastWorkoutNoteSection } from "@/features/workouts/components/last-workout-note-section";
import {
  buildWorkoutDraftFromHistory,
  getLatestWorkoutNote,
} from "@/features/workouts/utils/workout-insights";
import {
  formatExerciseGoalSummary,
  getExerciseGoal,
} from "@/features/exercises/utils/exercise-goals";
import type { WorkoutWorkoutResponse } from "@/client";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import { AddExerciseScreen } from "@/features/workouts/components/form/add-exercise-screen";
import {
  ExerciseHeader,
  ExerciseScreen,
  ExerciseSets,
} from "@/features/workouts/components/form/exercise-screen";
import { RecentSets } from "@/features/workouts/components/form/recent-sets-display";
import {
  hasWorkoutDraftContent,
  shouldShowRecentFocusAreaCard,
} from "@/features/workouts/components/form/workout-form-helpers";
import {
  WorkoutExerciseCards,
  type WorkoutExerciseCard,
  WorkoutFormActions,
  WorkoutMetadataFields,
} from "@/features/workouts/components/form/workout-form-sections";
import { useExerciseReorder } from "@/features/workouts/components/form/use-exercise-reorder";

function WorkoutExerciseSection({
  field,
  form,
  formatVolume,
  renderExerciseGoalSummary,
}: {
  field: any;
  form: any;
  formatVolume: (value: number) => string;
  renderExerciseGoalSummary: (exercise: { name: string }) => ReactNode;
}) {
  const exerciseReorder = useExerciseReorder<WorkoutExerciseCard>(
    field.state.value as WorkoutExerciseCard[],
  );

  return (
    <>
      <WorkoutExerciseCards
        exercises={exerciseReorder.displayEntries}
        dataTestId="new-workout-exercise-card"
        canEditOrder={exerciseReorder.canReorder}
        hasPendingOrderChanges={exerciseReorder.hasPendingOrderChanges}
        isReorderMode={exerciseReorder.isReorderMode}
        onCancelOrder={exerciseReorder.cancelReorder}
        onEditOrder={exerciseReorder.startReorder}
        onRemoveExercise={field.removeValue}
        onReorderExercises={exerciseReorder.moveExercise}
        onSaveOrder={() => {
          field.handleChange(exerciseReorder.commitReorder());
          toast.success("Exercise order saved");
        }}
        formatVolume={formatVolume}
        renderNameSupplement={renderExerciseGoalSummary}
        renderMetrics={() => (
          <MiniChart
            data={[3, 5, 2, 4, 6, 3, 4]}
            activeIndex={6}
          />
        )}
      />

      <WorkoutFormActions
        form={form}
        isReorderMode={exerciseReorder.isReorderMode}
      />
    </>
  );
}

export type WorkoutFormSearch = {
  addExercise?: boolean;
  exerciseIndex?: number;
  newExercise?: boolean;
};

export function NewWorkoutPage({
  user,
  exercises,
  workouts,
  workoutsFocus,
  search,
  draftStorage = workoutDraftStorage,
}: {
  user: CurrentUser | CurrentInternalUser | null; // need user for localStorage
  exercises: DbExercise[];
  workouts: WorkoutWorkoutResponse[];
  workoutsFocus: WorkoutFocus[];
  search: WorkoutFormSearch;
  draftStorage?: WorkoutDraftStorage;
}) {
  const { addExercise, exerciseIndex, newExercise } = search;
  const navigate = useNavigate({ from: "/workouts/new" });
  const [isClearDialogOpen, setIsClearDialogOpen] = useState(false);
  const [pendingTemplateWorkoutId, setPendingTemplateWorkoutId] = useState<
    number | null
  >(null);

  const saveWorkoutApi = useSaveWorkoutMutation();
  const saveWorkoutDemo = useMutation(postDemoWorkoutsMutation());
  const saveWorkout = user ? saveWorkoutApi : saveWorkoutDemo;
  const form = useAppForm({
    defaultValues: getInitialValues(user?.id, draftStorage),
    listeners: {
      onChange: ({ formApi }) => {
        draftStorage.save(formApi.state.values, user?.id);
      },
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
            toast.success("Workout saved successfully");
            draftStorage.clear(user?.id);
            form.reset(MOCK_VALUES);
            navigate({ search: {} });
          },
        },
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
    const focusMap = new Map<string, { focus: string; workoutId: number }>();

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
      getWorkoutByIdQueryOptions(user, workoutId),
    );
    const nextDraft = buildWorkoutDraftFromHistory(workoutToRepeat);

    form.reset(nextDraft);
    draftStorage.save(nextDraft, user?.id);
    navigate({ search: {} });
    toast.success("Loaded workout structure");
  };

  const handleRepeatWorkout = async (workoutId: number) => {
    const hasDraft =
      hasWorkoutDraftContent(form.state.values) ||
      draftStorage.load(user?.id) !== null;
    if (hasDraft) {
      setPendingTemplateWorkoutId(workoutId);
      return;
    }

    await loadWorkoutTemplate(workoutId);
  };

  const handleClearForm = () => {
    draftStorage.clear(user?.id);
    form.reset(MOCK_VALUES);
    navigate({ search: {} });
    setIsClearDialogOpen(false);
  };

  const renderExerciseGoalSummary = (exercise: { name: string }) => {
    const exerciseGoalSummary = formatExerciseGoalSummary(
      getExerciseGoal({
        exerciseId: getExerciseId(exercise.name),
        exerciseName: exercise.name,
      }),
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
            onAddExercise={(index, isNewExercise) =>
              navigate({
                search: { exerciseIndex: index, newExercise: isNewExercise },
              })
            }
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
        form.removeFieldValue("exercises", exerciseIndex);
      }
      navigate({ search: {} });
    };

    const handleDiscardNewExercise = () => {
      if (newExercise) {
        form.removeFieldValue("exercises", exerciseIndex);
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
            recentSets={
              <RecentSets
                exerciseId={exerciseId}
                user={user}
              />
            }
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
            <form.Subscribe
              selector={(state) =>
                shouldShowRecentFocusAreaCard({
                  focusAreaTemplateCount: focusAreaTemplates.length,
                  isDirty: state.isDirty,
                  value: state.values,
                })
              }
            >
              {(showRecentFocusAreaCard) =>
                showRecentFocusAreaCard ? (
                  <Card className="mb-4 p-4">
                    <div className="space-y-3">
                      <div className="space-y-2">
                        <p className="text-sm font-semibold text-card-foreground">
                          Start from a recent focus area
                        </p>
                        <p className="text-sm text-muted-foreground">
                          Load the most recent workout for the focus area you
                          want to train today.
                        </p>
                      </div>
                      <Select
                        onValueChange={(value) =>
                          handleRepeatWorkout(Number(value))
                        }
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
                ) : null
              }
            </form.Subscribe>

            {latestWorkoutNote && (
              <div className="mb-4">
                <LastWorkoutNoteSection
                  title="Last Workout Note"
                  note={latestWorkoutNote.note}
                  dateLabel={latestWorkoutNote.date}
                />
              </div>
            )}

            <WorkoutMetadataFields
              form={form}
              workoutsFocus={workoutsFocus}
            />

            {/* MARK: Exercise Cards */}
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => {
                return (
                  <WorkoutExerciseSection
                    field={field}
                    form={form}
                    formatVolume={formatWeight}
                    renderExerciseGoalSummary={renderExerciseGoalSummary}
                  />
                );
              }}
            />

            {/* MARK: Buttons */}
          </form>
        </div>
        <AlertDialog
          open={isClearDialogOpen}
          onOpenChange={setIsClearDialogOpen}
        >
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
