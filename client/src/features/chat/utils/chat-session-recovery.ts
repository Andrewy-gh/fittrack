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
  updateStreamingMessageWithDelta,
  updateStreamingMessageWithDone,
  updateStreamingMessageWithError,
} from "./chat-resume";
import type {
  ChatOperation,
  ChatOperationAttempt,
  ChatOperationController,
} from "./chat-operation-controller";
import type {
  ChatSessionState,
  ConversationRequestOptions,
  ConversationRequestResult,
  RecordChatTelemetry,
} from "./chat-session-types";

type RequestContext = {
  state: ChatSessionState;
  signal: AbortSignal;
  isCurrent: () => boolean;
};

type ResumeContext = {
  controller: ChatOperationController;
  state: ChatSessionState;
};

export type LoadConversation = (
  id: number,
  opts?: ConversationRequestOptions,
) => Promise<ConversationRequestResult>;

export type RecoverConversation = (
  id: number,
  opts?: ConversationRequestOptions,
) => Promise<ConversationRequestResult>;

export async function loadConversation(
  id: number,
  opts: ConversationRequestOptions | undefined,
  { state, signal, isCurrent }: RequestContext,
): Promise<ConversationRequestResult> {
  try {
    const detail = await getAIChatConversation(id, { signal });
    if (signal.aborted || !isCurrent()) {
      return { detail: null, aborted: true };
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
    return { detail, aborted: false, error: undefined };
  } catch (error) {
    if (signal.aborted || !isCurrent() || isAbortError(error)) {
      return { detail: null, aborted: true, error };
    }
    state.setLoadError(getErrorMessage(error));
    return { detail: null, aborted: false, error };
  }
}

export async function recoverConversation(
  id: number,
  opts: ConversationRequestOptions | undefined,
  { state, signal, isCurrent }: RequestContext,
): Promise<ConversationRequestResult> {
  let recoveryRequestError: unknown = null;
  let shouldRetryRecovery = false;

  const requestRecovery = async () => {
    try {
      const response = await requestAIChatMessageRecovery(id, { signal });
      shouldRetryRecovery = response.status === "not_needed";
    } catch (error) {
      if (signal.aborted || isAbortError(error)) throw error;
      recoveryRequestError = recoveryRequestError ?? error;
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
      return { detail: null, aborted: true };
    }
    state.setConversation(detail.conversation);
    state.setMessages(detail.messages);
    state.setLatestWorkoutDraftMessageId(null);
    state.setLoadError(null);
    return { detail, aborted: false, error: undefined };
  } catch (error) {
    if (signal.aborted || !isCurrent() || isAbortError(error)) {
      return { detail: null, aborted: true, error };
    }
    if (!opts?.silent) {
      state.setLoadError(getErrorMessage(recoveryRequestError ?? error));
    }
    return {
      detail: null,
      aborted: false,
      error: recoveryRequestError ?? error,
    };
  }
}

export async function resumeConversation(
  detail: AIChatConversationDetail,
  loadConversation: LoadConversation,
  { controller, state }: ResumeContext,
  operation: ChatOperation,
  attempt: ChatOperationAttempt,
): Promise<ConversationRequestResult> {
  const activeRun = detail.active_run;
  if (!activeRun) return { detail, aborted: false };

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
        onDelta: (event) => {
          if (!ownsAttempt()) return;
          const targetId =
            operation.assistantMessageId ?? activeRun.assistant_message_id;
          state.setMessages((current) =>
            updateStreamingMessageWithDelta(
              current,
              targetId,
              event.delta ?? "",
            ),
          );
          if (event.sequence !== undefined) {
            saveResumeCursor(detail.conversation.id, {
              runId: activeRun.id,
              sequence: event.sequence,
              assistantMessageId: targetId,
            });
          }
        },
        onDone: (event) => {
          if (!ownsAttempt()) return;
          const targetId =
            operation.assistantMessageId ?? activeRun.assistant_message_id;
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
          clearResumeCursor(detail.conversation.id);
        },
        onErrorEvent: (event) => {
          if (!ownsAttempt()) return;
          const targetId =
            operation.assistantMessageId ?? activeRun.assistant_message_id;
          state.setMessages((current) =>
            updateStreamingMessageWithError(current, targetId, event),
          );
          clearResumeCursor(detail.conversation.id);
        },
        signal: attempt.controller.signal,
      },
    );
    if (attempt.controller.signal.aborted || !ownsAttempt()) {
      return { detail: null, aborted: true };
    }

    const refreshed = await loadConversation(detail.conversation.id, {
      silent: true,
    });
    const terminalStatus =
      streamResult.doneEvent?.type === "done"
        ? (streamResult.doneEvent.status ?? "completed")
        : streamResult.doneEvent?.type === "error"
          ? "failed"
          : undefined;
    if (!streamResult.endedWithError) {
      return { ...refreshed, terminalStatus };
    }

    return {
      detail: refreshed.detail,
      aborted: refreshed.aborted,
      error:
        refreshed.error ??
        new Error(
          streamResult.doneEvent?.type === "error"
            ? streamResult.doneEvent.message
            : "AI chat resume failed",
        ),
      terminalStatus,
    };
  } catch (error) {
    if (
      attempt.controller.signal.aborted ||
      !ownsAttempt() ||
      isAbortError(error)
    ) {
      return { detail: null, aborted: true, error };
    }
    return { detail: null, aborted: false, error };
  }
}

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
    messages: recoveryResult.detail?.messages,
    aborted: recoveryResult.aborted,
    error: recoveryResult.error,
  });
  recordTelemetry({ category: "recovery", outcome: recoveryOutcome });

  if (
    recoveryOutcome !== "recovered_completed" &&
    recoveryOutcome !== "recovery_aborted"
  ) {
    const recoveryError =
      recoveryResult.error ??
      new Error("Failed to recover AI chat conversation");
    recordTelemetry({ category: "ux", outcome: "failure_toast_shown" });
    if (!recoveryResult.detail) {
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
