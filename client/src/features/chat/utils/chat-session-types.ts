import type { Dispatch, SetStateAction } from "react";
import { getErrorMessage } from "@/lib/errors";
import type {
  AIChatConversation,
  AIChatConversationDetail,
  AIChatMessage,
  AIChatTelemetryEvent,
} from "@/features/chat/api/ai-chat";

/** Expected transport failures are values; their original cause stays available to UI/logging. */
export class ChatRequestFailure extends Error {
  /** Stable failure tag for callers handling request outcomes. */
  readonly kind = "ChatRequestFailure";

  /** Capture a boundary failure without claiming a persisted run outcome. */
  constructor(
    /** Request that failed. */
    readonly operation: "load" | "recover" | "resume" | "stop",
    /** Original API/transport error, including any request identifier. */
    readonly cause: unknown,
    /** A deadline is uncertainty about completion, not a server rejection. */
    readonly reason: "request" | "deadline" = "request",
  ) {
    super(getErrorMessage(cause));
  }
}

/** A detail request has exactly one outcome; cancellation cannot carry usable data. */
export type ConversationRequestResult =
  | { readonly kind: "loaded"; readonly detail: AIChatConversationDetail }
  | { readonly kind: "cancelled" }
  | { readonly kind: "failed"; readonly error: ChatRequestFailure };

/** Stream evidence and a subsequent reload can disagree or fail independently. */
export type ConversationResumeResult =
  | { readonly kind: "cancelled" }
  | { readonly kind: "failed"; readonly error: ChatRequestFailure }
  | { readonly kind: "not-needed"; readonly detail: AIChatConversationDetail }
  | {
      readonly kind: "finished";
      readonly reload: ConversationRequestResult;
      readonly stream:
        | {
            readonly kind: "terminal";
            readonly status: Exclude<AIChatMessage["status"], "streaming">;
          }
        | { readonly kind: "unconfirmed"; readonly error: ChatRequestFailure };
    };

/** Presentation and deadline options for a detail request. */
export type ConversationRequestOptions = {
  silent?: boolean;
  timeoutMs?: number;
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

/** Existing telemetry boundary shared by the controller's workflows. */
export type RecordChatTelemetry = (event: AIChatTelemetryEvent) => void;
