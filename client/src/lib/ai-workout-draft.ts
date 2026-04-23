import type { WorkoutCreateWorkoutRequest, WorkoutExerciseInput } from "@/client";
import type { AIWorkoutDraft } from "@/lib/api/ai-chat";
import {
  type WorkoutDraftStorage,
  workoutDraftStorage,
} from "@/lib/local-storage";

function cloneExercises(
  exercises: AIWorkoutDraft["exercises"],
): WorkoutExerciseInput[] {
  return exercises.map((exercise) => ({
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
): WorkoutCreateWorkoutRequest {
  return {
    date: draft.date,
    notes: draft.notes ?? "",
    workoutFocus: draft.workoutFocus ?? "",
    exercises: cloneExercises(draft.exercises),
  };
}

export function saveAIWorkoutDraftToWorkoutForm(
  draft: AIWorkoutDraft,
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage,
): WorkoutCreateWorkoutRequest {
  const nextDraft = toWorkoutFormDraft(draft);
  draftStorage.save(nextDraft, userId);
  return nextDraft;
}
