import type {
  AIChatConversationDetail,
  AIChatMessage,
  AIChatStreamDoneEvent,
  AIChatStreamErrorEvent,
} from "@/features/chat/api/ai-chat";

const storageKeyPrefix = "fittrack.ai-chat.resume";

export type StoredResumeCursor = {
  runId: number;
  sequence: number;
  assistantMessageId: number;
};

export function loadResumeCursor(
  conversationId: number,
): StoredResumeCursor | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(storageKey(conversationId));
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<StoredResumeCursor>;
    if (
      typeof parsed.runId !== "number" ||
      typeof parsed.sequence !== "number" ||
      typeof parsed.assistantMessageId !== "number"
    ) {
      return null;
    }

    return {
      runId: parsed.runId,
      sequence: parsed.sequence,
      assistantMessageId: parsed.assistantMessageId,
    };
  } catch {
    return null;
  }
}

export function saveResumeCursor(
  conversationId: number,
  cursor: StoredResumeCursor,
): void {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.setItem(
    storageKey(conversationId),
    JSON.stringify(cursor),
  );
}

export function clearResumeCursor(conversationId: number): void {
  if (typeof window === "undefined") {
    return;
  }

  window.sessionStorage.removeItem(storageKey(conversationId));
}

export function getResumeAfterSequence(
  detail: AIChatConversationDetail,
): number {
  const activeRun = detail.active_run;
  if (!activeRun) {
    return 0;
  }

  const stored = loadResumeCursor(detail.conversation.id);
  if (
    !stored ||
    stored.runId !== activeRun.id ||
    stored.assistantMessageId !== activeRun.assistant_message_id
  ) {
    return activeRun.latest_sequence;
  }

  return Math.max(activeRun.latest_sequence, stored.sequence);
}

export function updateStreamingMessageWithDelta(
  messages: AIChatMessage[],
  messageId: number,
  delta: string,
): AIChatMessage[] {
  return messages.map((message) =>
    message.id === messageId
      ? {
          ...message,
          content: `${message.content}${delta}`,
        }
      : message,
  );
}

export function updateStreamingMessageWithDone(
  messages: AIChatMessage[],
  messageId: number,
  event: AIChatStreamDoneEvent,
): AIChatMessage[] {
  return messages.map((message) =>
    message.id === messageId
      ? {
          ...message,
          id: event.message_id ?? message.id,
          status: event.status === "stopped" ? "stopped" : "completed",
          content: event.text,
          completed_at: new Date().toISOString(),
        }
      : message,
  );
}

export function updateStreamingMessageWithError(
  messages: AIChatMessage[],
  messageId: number,
  event: AIChatStreamErrorEvent,
): AIChatMessage[] {
  return messages.map((message) =>
    message.id === messageId
      ? {
          ...message,
          id: event.message_id ?? message.id,
          status: "failed",
          error_message: event.message,
        }
      : message,
  );
}

function storageKey(conversationId: number): string {
  return `${storageKeyPrefix}:${conversationId}`;
}
