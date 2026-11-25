import type { WorkoutCreateWorkoutRequest } from "@/client";
import type { AIWorkoutDraft } from "@/features/chat/api/ai-chat";
import {
  type WorkoutDraftStorage,
  type StoredWorkoutDraft,
  workoutDraftStorage,
} from "@/lib/local-storage";

export function toWorkoutFormDraft(
  draft: AIWorkoutDraft,
): StoredWorkoutDraft | null {
  if (draft.unsupportedRecommendations) return null;
  const baseline: WorkoutCreateWorkoutRequest = {
    date: draft.date,
    notes: draft.notes ?? "",
    workoutFocus: draft.workoutFocus ?? "",
    exercises: draft.exercises.map((exercise) => ({
      ...exercise,
      sets: exercise.sets.map((set) => ({ ...set })),
    })),
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
