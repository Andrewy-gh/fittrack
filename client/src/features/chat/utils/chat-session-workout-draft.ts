import {
  saveAIChatLatestWorkoutDraft,
  type AIChatConversation,
} from "@/features/chat/api/ai-chat";
import { showErrorToast } from "@/lib/errors";
import { toast } from "sonner";
import type { ChatSessionSetters } from "./chat-session-types";

export async function saveLatestWorkoutDraft({
  conversation,
  setters,
}: {
  conversation: AIChatConversation | null;
  setters: ChatSessionSetters;
}) {
  if (!conversation) {
    return;
  }

  try {
    setters.setIsSavingWorkoutDraft(true);
    const updatedConversation = await saveAIChatLatestWorkoutDraft(
      conversation.id,
    );
    setters.setConversation(updatedConversation);
    toast.success("Workout saved successfully");
  } catch (error) {
    showErrorToast(error, "Failed to save workout");
  } finally {
    setters.setIsSavingWorkoutDraft(false);
  }
}
