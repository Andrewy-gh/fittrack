import { Suspense, type ReactNode } from "react";
import { type WorkoutFocus } from "@/features/workouts/api/workouts";
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
import { toast } from "sonner";
import { formatWeight } from "@/lib/utils";
import {
  ErrorBoundary,
  FullScreenErrorFallback,
} from "@/components/error-boundary";
import { ExerciseContextPanel } from "@/features/workouts/components/form/exercise-context-panel";
import { LastWorkoutNoteSection } from "@/features/workouts/components/last-workout-note-section";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutNewWorkoutContextResponse,
} from "@/client";
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
import { shouldShowRecentFocusAreaCard } from "@/features/workouts/components/form/workout-form-helpers";
import {
  WorkoutExerciseCards,
  type WorkoutExerciseArrayField,
  type WorkoutExerciseCard,
  WorkoutFormActions,
  type WorkoutFormSectionApi,
  WorkoutMetadataFields,
} from "@/features/workouts/components/form/workout-form-sections";
import { useExerciseReorder } from "@/features/workouts/components/form/use-exercise-reorder";
import { useNewWorkoutFormWorkflow } from "@/features/workouts/hooks/use-new-workout-form-workflow";

function WorkoutExerciseSection({
  field,
  form,
  formatVolume,
  renderExerciseGoalSummary,
}: {
  field: WorkoutExerciseArrayField;
  form: WorkoutFormSectionApi<WorkoutCreateWorkoutRequest>;
  formatVolume: (value: number) => string;
  renderExerciseGoalSummary: (exercise: { name: string }) => ReactNode;
}) {
  const exerciseReorder = useExerciseReorder<WorkoutExerciseCard>(
    field.state.value,
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
  newWorkoutContext,
  workoutsFocus,
  search,
  draftStorage = workoutDraftStorage,
}: {
  user: CurrentUser | CurrentInternalUser | null; // need user for localStorage
  exercises: DbExercise[];
  newWorkoutContext: WorkoutNewWorkoutContextResponse;
  workoutsFocus: WorkoutFocus[];
  search: WorkoutFormSearch;
  draftStorage?: WorkoutDraftStorage;
}) {
  const { addExercise, exerciseIndex, newExercise } = search;
  const workoutForm = useNewWorkoutFormWorkflow({
    user,
    exercises,
    newWorkoutContext,
    draftStorage,
  });
  const { form, focusAreaTemplates, latestWorkoutNote } = workoutForm;

  const renderExerciseGoalSummary = (exercise: { name: string }) => {
    const exerciseGoalSummary = workoutForm.getExerciseGoalSummary(
      exercise.name,
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
            onAction={workoutForm.clearSearch}
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
            onAddExercise={workoutForm.openExercise}
            onBack={workoutForm.clearSearch}
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
      workoutForm.clearSearch();
      return null;
    }

    const exerciseName = exercises[exerciseIndex]?.name;
    const exerciseId = workoutForm.getExerciseId(exerciseName);

    return (
      <ErrorBoundary
        fallback={
          <FullScreenErrorFallback
            message="Failed to load exercise details"
            onAction={workoutForm.clearSearch}
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
                onBack={() =>
                  workoutForm.closeExercise(exerciseIndex, newExercise)
                }
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
                onDiscardNewExercise={() =>
                  workoutForm.discardNewExercise(exerciseIndex, newExercise)
                }
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
                onClick={() => workoutForm.setIsClearDialogOpen(true)}
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
                          workoutForm.repeatWorkout(Number(value))
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
          open={workoutForm.isClearDialogOpen}
          onOpenChange={workoutForm.setIsClearDialogOpen}
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
              <AlertDialogAction onClick={workoutForm.clearForm}>
                Clear draft
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <AlertDialog
          open={workoutForm.pendingTemplateWorkoutId !== null}
          onOpenChange={(open) => {
            if (!open) {
              workoutForm.setPendingTemplateWorkoutId(null);
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
                onClick={workoutForm.replaceDraftWithPendingTemplate}
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
