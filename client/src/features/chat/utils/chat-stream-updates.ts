import type {
  AIChatStreamDeltaEvent,
  AIChatStreamDoneEvent,
} from "@/features/chat/api/ai-chat";
import {
  clearResumeCursor,
  saveResumeCursor,
  updateStreamingMessageWithDelta,
  updateStreamingMessageWithDone,
} from "./chat-resume";
import type { ChatSessionState } from "./chat-session-types";

type StreamUpdateContext = {
  conversationId: number;
  state: ChatSessionState;
  ownsAttempt: () => boolean;
  getMessageId: () => number;
  getRunId: (event: AIChatStreamDeltaEvent) => number;
};

/**
 * Apply incoming stream updates only while the calling workflow owns its attempt.
 * Start events and operation lifecycle remain with submit/resume; resumed deltas
 * use the loaded run identity while new streams use the incoming event identity.
 */
export function createChatStreamUpdates({
  conversationId,
  state,
  ownsAttempt,
  getMessageId,
  getRunId,
}: StreamUpdateContext) {
  return {
    onDelta(event: AIChatStreamDeltaEvent) {
      if (!ownsAttempt()) return;
      const targetId = getMessageId();
      state.setMessages((current) =>
        updateStreamingMessageWithDelta(current, targetId, event.delta ?? ""),
      );
      if (event.sequence !== undefined) {
        saveResumeCursor(conversationId, {
          runId: getRunId(event),
          sequence: event.sequence,
          assistantMessageId: targetId,
        });
      }
    },
    onDone(event: AIChatStreamDoneEvent) {
      if (!ownsAttempt()) return;
      const targetId = getMessageId();
      state.setMessages((current) =>
        updateStreamingMessageWithDone(current, targetId, event),
      );
      if (event.workout_draft) {
        state.setConversation((current) =>
          current
            ? {
                ...current,
                latest_workout_draft: event.workout_draft,
                latest_workout_draft_status: { is_saved: false },
              }
            : current,
        );
        state.setLatestWorkoutDraftMessageId(event.message_id ?? targetId);
      }
      clearResumeCursor(conversationId);
    },
  };
}
