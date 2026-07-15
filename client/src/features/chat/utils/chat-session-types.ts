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

/** React state owned by useAIChatSession and updated by chat workflows. */
export type ChatSessionState = {
  setConversation: Dispatch<SetStateAction<AIChatConversation | null>>;
  setMessages: Dispatch<SetStateAction<AIChatMessage[]>>;
  setPrompt: Dispatch<SetStateAction<string>>;
  setIsLoadingConversation: Dispatch<SetStateAction<boolean>>;
  setLoadError: Dispatch<SetStateAction<string | null>>;
  setIsSavingWorkoutDraft: Dispatch<SetStateAction<boolean>>;
  setLatestWorkoutDraftMessageId: Dispatch<SetStateAction<number | null>>;
};

export type RecordChatTelemetry = (event: AIChatTelemetryEvent) => void;
