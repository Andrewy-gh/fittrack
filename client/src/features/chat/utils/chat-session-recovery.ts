import {
  getAIChatConversation,
  pollAIChatConversationUntilSettled,
  requestAIChatMessageRecovery,
  resumeAIChatMessageStream,
  type AIChatConversationDetail,
} from "@/features/chat/api/ai-chat";
import {
  classifyRecoveryOutcome,
  isAbortError,
} from "@/features/chat/utils/ai-chat-observability";
import { getErrorMessage, showErrorToast } from "@/lib/errors";
import {
  clearResumeCursor,
  getResumeAfterSequence,
  saveResumeCursor,
} from "./chat-resume";
import type {
  ChatOperation,
  ChatOperationAttempt,
} from "./chat-operation-state";
import type { ChatOperationController } from "./chat-operation-controller";
import type {
  ChatSessionState,
  ConversationRequestOptions,
  ConversationRequestResult,
  ConversationResumeResult,
  RecordChatTelemetry,
} from "./chat-session-types";
import { ChatRequestFailure } from "./chat-session-types";
import { createChatStreamUpdates } from "./chat-stream-updates";

type RequestContext = {
  state: ChatSessionState;
  signal: AbortSignal;
  isCurrent: () => boolean;
};

type ResumeContext = {
  controller: ChatOperationController;
  state: ChatSessionState;
};

type RecoveryRequestFailure = {
  readonly cause: unknown;
};

/** Reload a conversation through the controller-owned request boundary. */
export type LoadConversation = (
  id: number,
  opts?: ConversationRequestOptions,
) => Promise<ConversationRequestResult>;

/** Recover and poll persisted state through the operation-owned request boundary. */
export type RecoverConversation = (
  id: number,
  opts?: ConversationRequestOptions,
) => Promise<ConversationRequestResult>;

/** Load detail only while this route/request still owns its callbacks. */
export async function loadConversation(
  id: number,
  opts: ConversationRequestOptions | undefined,
  { state, signal, isCurrent }: RequestContext,
): Promise<ConversationRequestResult> {
  try {
    const detail = await getAIChatConversation(id, { signal });
    if (signal.aborted || !isCurrent()) {
      return { kind: "cancelled" };
    }
    state.setConversation(detail.conversation);
    state.setMessages(detail.messages);
    if (!opts?.silent) {
      state.setLatestWorkoutDraftMessageId(null);
    }
    if (!detail.active_run) {
      clearResumeCursor(id);
    }
    state.setLoadError(null);
    return { kind: "loaded", detail };
  } catch (error) {
    if (signal.aborted || !isCurrent() || isAbortError(error)) {
      return { kind: "cancelled" };
    }
    state.setLoadError(getErrorMessage(error));
    return { kind: "failed", error: new ChatRequestFailure("load", error) };
  }
}

/** Request server recovery, then observe persisted state until settled or cancelled. */
export async function recoverConversation(
  id: number,
  opts: ConversationRequestOptions | undefined,
  { state, signal, isCurrent }: RequestContext,
): Promise<ConversationRequestResult> {
  let recoveryRequestError: RecoveryRequestFailure | null = null;
  let shouldRetryRecovery = false;

  const requestRecovery = async () => {
    try {
      const response = await requestAIChatMessageRecovery(id, { signal });
      shouldRetryRecovery = response.status === "not_needed";
    } catch (error) {
      if (signal.aborted || isAbortError(error)) throw error;
      recoveryRequestError = recoveryRequestError ?? { cause: error };
      shouldRetryRecovery = false;
    }
  };

  try {
    await requestRecovery();
    const detail = await pollAIChatConversationUntilSettled(id, {
      signal,
      onStreaming: async () => {
        if (shouldRetryRecovery) await requestRecovery();
      },
    });
    if (signal.aborted || !isCurrent()) {
      return { kind: "cancelled" };
    }
    state.setConversation(detail.conversation);
    state.setMessages(detail.messages);
    state.setLatestWorkoutDraftMessageId(null);
    state.setLoadError(null);
    return { kind: "loaded", detail };
  } catch (error) {
    if (signal.aborted || !isCurrent() || isAbortError(error)) {
      return { kind: "cancelled" };
    }
    const recoveryFailure: RecoveryRequestFailure = recoveryRequestError ?? {
      cause: undefined,
    };
    if (!opts?.silent) {
      state.setLoadError(getErrorMessage(recoveryFailure.cause ?? error));
    }
    return {
      kind: "failed",
      error: new ChatRequestFailure("recover", recoveryFailure.cause ?? error),
    };
  }
}

/** Resume the loaded active run without taking ownership of its lifecycle. */
export async function resumeConversation(
  detail: AIChatConversationDetail,
  loadConversation: LoadConversation,
  { controller, state }: ResumeContext,
  operation: ChatOperation,
  attempt: ChatOperationAttempt,
): Promise<ConversationResumeResult> {
  const activeRun = detail.active_run;
  if (!activeRun) return { kind: "not-needed", detail };

  const afterSequence = getResumeAfterSequence(detail);
  const ownsAttempt = () => controller.ownsAttempt(operation, attempt);

  try {
    const streamResult = await resumeAIChatMessageStream(
      detail.conversation.id,
      activeRun.id,
      afterSequence,
      {
        onStart: (event) => {
          if (!ownsAttempt()) return;
          const assistantMessageId =
            event.message_id ?? activeRun.assistant_message_id;
          controller.markStreaming(
            operation,
            attempt,
            activeRun.id,
            assistantMessageId,
          );
          if (event.sequence !== undefined) {
            saveResumeCursor(detail.conversation.id, {
              runId: activeRun.id,
              sequence: event.sequence,
              assistantMessageId,
            });
          }
        },
        ...createChatStreamUpdates({
          conversationId: detail.conversation.id,
          state,
          ownsAttempt,
          getMessageId: () =>
            controller.getSnapshot().assistantMessageId ??
            activeRun.assistant_message_id,
          getRunId: () => activeRun.id,
        }),
        // Replay/read errors do not establish persisted failure. Keep the
        // partial message and cursor until a reload or terminal event resolves it.
        onErrorEvent: (event) => {
          if (ownsAttempt()) state.setLoadError(event.message);
        },
        signal: attempt.controller.signal,
      },
    );
    if (attempt.controller.signal.aborted || !ownsAttempt()) {
      return { kind: "cancelled" };
    }

    const refreshed = await loadConversation(detail.conversation.id, {
      silent: true,
    });
    if (refreshed.kind === "cancelled" || !ownsAttempt())
      return { kind: "cancelled" };
    // An SSE error can mean a replay/read failure. Only a done event carries
    // terminal evidence; a failed reload cannot turn transport failure into it.
    const event = streamResult.doneEvent;
    return {
      kind: "finished",
      reload: refreshed,
      stream:
        event?.type === "done"
          ? { kind: "terminal", status: event.status ?? "completed" }
          : {
              kind: "unconfirmed",
              error: new ChatRequestFailure(
                "resume",
                new Error(
                  event?.type === "error"
                    ? event.message
                    : "AI chat resume failed",
                ),
              ),
            },
    };
  } catch (error) {
    if (
      attempt.controller.signal.aborted ||
      !ownsAttempt() ||
      isAbortError(error)
    ) {
      return { kind: "cancelled" };
    }
    return { kind: "failed", error: new ChatRequestFailure("resume", error) };
  }
}

/** Report recovery outcomes without transferring lifecycle ownership. */
export async function recoverLoadedConversation({
  id,
  recoverConversation,
  recordTelemetry,
  setLoadError,
}: {
  id: number;
  recoverConversation: RecoverConversation;
  recordTelemetry: RecordChatTelemetry;
  setLoadError: ChatSessionState["setLoadError"];
}) {
  const recoveryResult = await recoverConversation(id, { silent: true });
  const recoveryOutcome = classifyRecoveryOutcome({
    messages:
      recoveryResult.kind === "loaded"
        ? recoveryResult.detail.messages
        : undefined,
    aborted: recoveryResult.kind === "cancelled",
    error: recoveryResult.kind === "failed" ? recoveryResult.error : undefined,
  });
  recordTelemetry({ category: "recovery", outcome: recoveryOutcome });

  if (
    recoveryOutcome !== "recovered_completed" &&
    recoveryOutcome !== "recovery_aborted"
  ) {
    const recoveryError =
      recoveryResult.kind === "failed"
        ? recoveryResult.error.cause
        : new Error("Failed to recover AI chat conversation");
    recordTelemetry({ category: "ux", outcome: "failure_toast_shown" });
    if (recoveryResult.kind !== "loaded") {
      setLoadError(
        getErrorMessage(
          recoveryError,
          "Failed to recover AI chat conversation",
        ),
      );
    }
    showErrorToast(recoveryError, "Failed to recover AI chat conversation");
  }

  return recoveryResult;
}
