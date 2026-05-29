import type { Dispatch, SetStateAction } from "react";
import type {
  AIChatConversation,
  AIChatConversationDetail,
  AIChatMessage,
  AIChatTelemetryEvent,
} from "@/features/chat/api/ai-chat";

export type ConversationRequestResult = {
  detail: AIChatConversationDetail | null;
  aborted: boolean;
  error?: unknown;
};

export type ConversationRequestOptions = {
  silent?: boolean;
};

export type ChatSessionRefs = {
  pendingAssistantIdRef: { current: number | null };
  loadAbortRef: { current: AbortController | null };
  recoveryAbortRef: { current: AbortController | null };
  resumeAbortRef: { current: AbortController | null };
  streamAbortRef: { current: AbortController | null };
};

export type ChatSessionSetters = {
  setConversation: Dispatch<SetStateAction<AIChatConversation | null>>;
  setMessages: Dispatch<SetStateAction<AIChatMessage[]>>;
  setPrompt: Dispatch<SetStateAction<string>>;
  setIsLoadingConversation: Dispatch<SetStateAction<boolean>>;
  setIsSubmitting: Dispatch<SetStateAction<boolean>>;
  setLoadError: Dispatch<SetStateAction<string | null>>;
  setIsSavingWorkoutDraft: Dispatch<SetStateAction<boolean>>;
  setLatestWorkoutDraftMessageId: Dispatch<SetStateAction<number | null>>;
};

export type RecordChatTelemetry = (event: AIChatTelemetryEvent) => void;
