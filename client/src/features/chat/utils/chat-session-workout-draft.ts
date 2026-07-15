import {
  saveAIChatLatestWorkoutDraft,
  type AIChatConversation,
} from "@/features/chat/api/ai-chat";
import { showErrorToast } from "@/lib/errors";
import { toast } from "sonner";
import type { ChatSessionState } from "./chat-session-types";

export async function saveLatestWorkoutDraft({
  conversation,
  state,
}: {
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
    toast.success("Workout saved successfully");
  } catch (error) {
    showErrorToast(error, "Failed to save workout");
  } finally {
    state.setIsSavingWorkoutDraft(false);
  }
}
