import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
} from "@/client";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import {
  type WorkoutDraftStorage,
  type StoredWorkoutDraft,
  workoutDraftStorage,
} from "@/lib/local-storage";

function cloneExercises(
  exercises: AIWorkoutDraft["exercises"],
): WorkoutExerciseInput[] {
  return exercises.map((exercise) => ({
    ...(exercise.catalog_id !== undefined && {
      catalog_id: exercise.catalog_id,
    }),
    name: exercise.name,
    sets: exercise.sets.map((set) => ({
      reps: set.reps,
      setType: set.setType,
      weight: set.weight,
    })),
  }));
}

export function toWorkoutFormDraft(
  draft: AIWorkoutDraft,
): StoredWorkoutDraft | null {
  if (draft.unsupportedRecommendations) return null;
  const baseline: WorkoutCreateWorkoutRequest = {
    date: draft.date,
    notes: draft.notes ?? "",
    workoutFocus: draft.workoutFocus ?? "",
    exercises: cloneExercises(draft.exercises),
  };
  return {
    ...baseline,
    ...(draft.recommendations !== undefined && {
      importedRecommendationContext: {
        baseline,
        recommendations: draft.recommendations,
      },
    }),
  };
}

export function saveAIWorkoutDraftToWorkoutForm(
  draft: AIWorkoutDraft,
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
): StoredWorkoutDraft | null {
  const nextDraft = toWorkoutFormDraft(draft);
  if (!nextDraft) return null;
  draftStorage.save(nextDraft, userId);
  return nextDraft;
}
