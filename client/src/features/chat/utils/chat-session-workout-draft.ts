import {
  saveAIChatLatestWorkoutDraft,
  type AIChatConversation,
} from "@/features/chat/api/ai-chat";
import { showErrorToast } from "@/lib/errors";
import { invalidateTrainingData } from "@/features/workouts/api/training-data-cache";
import { toast } from "sonner";
import type { ChatSessionState } from "./chat-session-types";

export async function saveLatestWorkoutDraft({
  conversation,
  state,
  userId,
}: {
  userId: string;
  conversation: AIChatConversation | null;
  state: ChatSessionState;
}) {
  if (!conversation) {
    return;
  }

  try {
    state.setIsSavingWorkoutDraft(true);
    const updatedConversation = await saveAIChatLatestWorkoutDraft(
      conversation.id,
    );
    state.setConversation(updatedConversation);
    await invalidateTrainingData(userId);
    toast.success("Workout saved successfully");
  } catch (error) {
    showErrorToast(error, "Failed to save workout");
  } finally {
    state.setIsSavingWorkoutDraft(false);
  }
}
