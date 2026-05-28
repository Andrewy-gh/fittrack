import { useNavigate } from "@tanstack/react-router";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  useUpdateWorkoutMutation,
  type WorkoutFocus,
} from "@/features/workouts/api/workouts";
import { putDemoWorkoutsByIdMutation } from "@/lib/demo-data/query-options";
import type { CurrentUser, CurrentInternalUser } from "@stackframe/react";
import { Suspense } from "react";
import { useAppForm } from "@/hooks/form";
import { Button } from "@/components/ui/button";
import { MiniChart } from "@/features/workouts/components/form/mini-chart";
import { X } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import {
  ExerciseHeader,
  ExerciseScreen,
  ExerciseSets,
} from "@/features/workouts/components/form/exercise-screen";
import { AddExerciseScreen } from "@/features/workouts/components/form/add-exercise-screen";
import type {
  ExerciseExerciseResponse,
  WorkoutUpdateWorkoutRequest,
} from "@/client";
import {
  ErrorBoundary,
  FullScreenErrorFallback,
} from "@/components/error-boundary";
import { formatWeight } from "@/lib/utils";
import {
  WorkoutExerciseCards,
  type WorkoutExerciseCard,
  WorkoutFormActions,
  WorkoutMetadataFields,
} from "@/features/workouts/components/form/workout-form-sections";
import { useExerciseReorder } from "@/features/workouts/components/form/use-exercise-reorder";

function WorkoutExerciseSection({ field, form }: { field: any; form: any }) {
  const exerciseReorder = useExerciseReorder<WorkoutExerciseCard>(
    field.state.value as WorkoutExerciseCard[],
  );

  return (
    <>
      <WorkoutExerciseCards
        exercises={exerciseReorder.displayEntries}
        dataTestId="edit-workout-exercise-card"
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
        formatVolume={formatWeight}
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

export function EditWorkoutPage({
  user,
  exercises,
  workout,
  workoutId,
  workoutsFocus,
  search,
}: {
  user: CurrentUser | CurrentInternalUser | null;
  exercises: ExerciseExerciseResponse[];
  workout: WorkoutUpdateWorkoutRequest;
  workoutId: number;
  workoutsFocus: WorkoutFocus[];
  search: {
    addExercise?: boolean;
    exerciseIndex?: number;
    newExercise?: boolean;
  };
}) {
  const { addExercise, exerciseIndex, newExercise } = search;
  const navigate = useNavigate({ from: "/workouts/$workoutId/edit" });

  const updateWorkoutApi = useUpdateWorkoutMutation();
  const updateWorkoutDemo = useMutation(putDemoWorkoutsByIdMutation());
  const updateWorkoutMutation = user ? updateWorkoutApi : updateWorkoutDemo;

  const form = useAppForm({
    defaultValues: workout,
    onSubmit: async ({ value }) => {
      const trimmedValue = {
        ...value,
        notes: value.notes?.trim() || undefined,
        workoutFocus: value.workoutFocus?.trim() || undefined,
      };
      await updateWorkoutMutation.mutateAsync(
        {
          path: { id: workoutId },
          body: trimmedValue,
        },
        {
          onSuccess: () => {
            toast.success("Workout updated successfully");
            navigate({
              to: "/workouts/$workoutId",
              params: { workoutId },
            });
          },
        },
      );
    },
  });

  const handleClearForm = () => {
    if (confirm("Are you sure you want to clear all form data?")) {
      form.reset();
      navigate({ search: {} });
    }
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
                Edit Training
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
                  />
                );
              }}
            />

            {/* MARK: Buttons */}
          </form>
        </div>
      </Suspense>
    </ErrorBoundary>
  );
}
