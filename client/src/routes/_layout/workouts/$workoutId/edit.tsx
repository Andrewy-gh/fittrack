import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useSuspenseQuery, useMutation } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import {
  useUpdateWorkoutMutation,
  type WorkoutFocus,
} from '@/lib/api/workouts';
import { transformToWorkoutFormValues } from '@/lib/api/workouts';
import { putDemoWorkoutsByIdMutation } from '@/lib/demo-data/query-options';
import { initializeDemoData } from '@/lib/demo-data/storage';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import {
  getExercisesQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsFocusQueryOptions,
} from '@/lib/api/unified-query-options';
import { Suspense } from 'react';
import { useAppForm } from '@/hooks/form';
import { Button } from '@/components/ui/button';
import { MiniChart } from '../-components/mini-chart';
import { X } from 'lucide-react';
import { Spinner } from '@/components/ui/spinner';
import {
  ExerciseHeader,
  ExerciseScreen,
  ExerciseSets,
} from '../-components/exercise-screen';
import { AddExerciseScreen } from '../-components/add-exercise-screen';
import type {
  ExerciseExerciseResponse,
  WorkoutUpdateWorkoutRequest,
} from '@/client';
import {
  ErrorBoundary,
  FullScreenErrorFallback,
} from '@/components/error-boundary';
import { formatWeight } from '@/lib/utils';
import {
  WorkoutExerciseCards,
  WorkoutFormActions,
  WorkoutMetadataFields,
} from '../-components/workout-form-sections';

function EditWorkoutForm({
  user,
  exercises,
  workout,
  workoutId,
  workoutsFocus,
}: {
  user: CurrentUser | CurrentInternalUser | null;
  exercises: ExerciseExerciseResponse[];
  workout: WorkoutUpdateWorkoutRequest;
  workoutId: number;
  workoutsFocus: WorkoutFocus[];
}) {
  const { addExercise, exerciseIndex, newExercise } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });

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
            toast.success('Workout updated successfully');
            navigate({
              to: '/workouts/$workoutId',
              params: { workoutId },
            });
          },
        }
      );
    },
  });

  const handleClearForm = () => {
    if (confirm('Are you sure you want to clear all form data?')) {
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
              navigate({ search: { exerciseIndex: index, newExercise: isNewExercise } })
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
            <WorkoutMetadataFields form={form} workoutsFocus={workoutsFocus} />

            {/* MARK: Exercise Cards */}
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => {
                return (
                  <WorkoutExerciseCards
                    exercises={field.state.value}
                    dataTestId="edit-workout-exercise-card"
                    onRemoveExercise={field.removeValue}
                    formatVolume={formatWeight}
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
      </Suspense>
    </ErrorBoundary>
  );
}

const workoutSearchSchema = z.object({
  addExercise: z.boolean().optional(),
  exerciseIndex: z.coerce.number().int().optional(),
  newExercise: z.boolean().optional(),
});

export const Route = createFileRoute('/_layout/workouts/$workoutId/edit')({
  validateSearch: workoutSearchSchema,
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({
    context,
    params,
  }): Promise<{
    workoutId: number;
  }> => {
    const workoutId = params.workoutId;
    if (!context.user) initializeDemoData();

    context.queryClient.ensureQueryData(
      getWorkoutByIdQueryOptions(context.user, workoutId)
    );
    context.queryClient.ensureQueryData(getExercisesQueryOptions(context.user));
    context.queryClient.ensureQueryData(
      getWorkoutsFocusQueryOptions(context.user)
    );

    return { workoutId };
  },
  component: RouteComponent,
});

function RouteComponent() {
  const { workoutId } = Route.useLoaderData();
  const { user } = Route.useRouteContext();

  const { data: exercises } = useSuspenseQuery(getExercisesQueryOptions(user));
  const { data: workout } = useSuspenseQuery(
    getWorkoutByIdQueryOptions(user, workoutId)
  );
  const { data: workoutsFocusValues } = useSuspenseQuery(
    getWorkoutsFocusQueryOptions(user)
  );

  const workoutFormValues: WorkoutUpdateWorkoutRequest =
    transformToWorkoutFormValues(workout);

  const workoutsFocus: WorkoutFocus[] = workoutsFocusValues.map((wf) => ({
    name: wf,
  }));

  return (
    <EditWorkoutForm
      user={user}
      exercises={exercises}
      workout={workoutFormValues}
      workoutId={workoutId}
      workoutsFocus={workoutsFocus}
    />
  );
}
