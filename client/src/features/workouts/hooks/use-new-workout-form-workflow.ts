import { useNavigate } from "@tanstack/react-router";
import { useMutation } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { toast } from "sonner";
import type { WorkoutNewWorkoutContextResponse } from "@/client";
import type { DbExercise } from "@/features/exercises/api/exercises";
import {
  formatExerciseGoalSummary,
  getExerciseGoal,
} from "@/features/exercises/utils/exercise-goals";
import { useSaveWorkoutMutation } from "@/features/workouts/api/workouts";
import {
  getInitialValues,
  MOCK_VALUES,
} from "@/features/workouts/components/form/form-options";
import { hasWorkoutDraftContent } from "@/features/workouts/components/form/workout-form-helpers";
import { buildWorkoutDraftFromHistory } from "@/features/workouts/utils/workout-insights";
import { useAppForm } from "@/hooks/form";
import { queryClient } from "@/lib/api/api";
import { getWorkoutByIdQueryOptions } from "@/features/workouts/api/workout-query-options";
import { postDemoWorkoutsMutation } from "@/lib/demo-data/query-options";
import {
  type WorkoutDraftStorage,
  workoutDraftStorage,
} from "@/lib/local-storage";

export function useNewWorkoutFormWorkflow({
  user,
  exercises,
  newWorkoutContext,
  draftStorage = workoutDraftStorage,
}: {
  user: CurrentUser | CurrentInternalUser | null;
  exercises: DbExercise[];
  newWorkoutContext: WorkoutNewWorkoutContextResponse;
  draftStorage?: WorkoutDraftStorage;
}) {
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
            clearSearch();
          },
        },
      );
    },
  });

  const exerciseIdsByName = useMemo(() => {
    return new Map(exercises.map((exercise) => [exercise.name, exercise.id]));
  }, [exercises]);
  const latestWorkoutNote = newWorkoutContext.latestWorkoutNote;
  const focusAreaTemplates = useMemo(
    () => newWorkoutContext.focusTemplates ?? [],
    [newWorkoutContext.focusTemplates],
  );

  function getExerciseId(exerciseName: string): number | null {
    return exerciseIdsByName.get(exerciseName) ?? null;
  }

  function getExerciseGoalSummary(exerciseName: string): string | null {
    return formatExerciseGoalSummary(
      getExerciseGoal({
        exerciseId: getExerciseId(exerciseName),
        exerciseName,
      }),
    );
  }

  function clearSearch() {
    navigate({ search: {} });
  }

  const openExercise = (index: number, isNewExercise?: boolean) => {
    navigate({
      search: { exerciseIndex: index, newExercise: isNewExercise },
    });
  };

  const closeExercise = (exerciseIndex: number, newExercise?: boolean) => {
    const currentExercise = form.state.values.exercises[exerciseIndex];
    if (newExercise && currentExercise && currentExercise.sets.length === 0) {
      form.removeFieldValue("exercises", exerciseIndex);
    }
    clearSearch();
  };

  const discardNewExercise = (exerciseIndex: number, newExercise?: boolean) => {
    if (newExercise) {
      form.removeFieldValue("exercises", exerciseIndex);
    }
    clearSearch();
  };

  const loadWorkoutTemplate = async (workoutId: number) => {
    const workoutToRepeat = await queryClient.fetchQuery(
      getWorkoutByIdQueryOptions(user, workoutId),
    );
    const nextDraft = buildWorkoutDraftFromHistory(workoutToRepeat);

    form.reset(nextDraft);
    draftStorage.save(nextDraft, user?.id);
    clearSearch();
    toast.success("Loaded workout structure");
  };

  const repeatWorkout = async (workoutId: number) => {
    const hasDraft =
      hasWorkoutDraftContent(form.state.values) ||
      draftStorage.load(user?.id) !== null;
    if (hasDraft) {
      setPendingTemplateWorkoutId(workoutId);
      return;
    }

    await loadWorkoutTemplate(workoutId);
  };

  const replaceDraftWithPendingTemplate = async () => {
    if (pendingTemplateWorkoutId == null) {
      return;
    }

    const workoutId = pendingTemplateWorkoutId;
    setPendingTemplateWorkoutId(null);
    await loadWorkoutTemplate(workoutId);
  };

  const clearForm = () => {
    draftStorage.clear(user?.id);
    form.reset(MOCK_VALUES);
    clearSearch();
    setIsClearDialogOpen(false);
  };

  return {
    form,
    latestWorkoutNote,
    focusAreaTemplates,
    isClearDialogOpen,
    pendingTemplateWorkoutId,
    setIsClearDialogOpen,
    setPendingTemplateWorkoutId,
    clearForm,
    clearSearch,
    openExercise,
    closeExercise,
    discardNewExercise,
    repeatWorkout,
    replaceDraftWithPendingTemplate,
    getExerciseId,
    getExerciseGoalSummary,
  };
}
