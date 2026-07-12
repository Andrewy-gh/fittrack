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
  ChatSessionRefs,
  ChatSessionSetters,
  ChatSessionOperation,
  ConversationRequestOptions,
  ConversationRequestResult,
  RecordChatTelemetry,
} from "./chat-session-types";

type SessionStateContext = {
  refs: ChatSessionRefs;
  setters: ChatSessionSetters;
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
  { refs, setters }: SessionStateContext,
): Promise<ConversationRequestResult> {
  refs.loadAbortRef.current?.abort();
  const controller = new AbortController();
  refs.loadAbortRef.current = controller;

  if (!opts?.silent) {
    setters.setIsLoadingConversation(true);
  }

  try {
    const detail = await getAIChatConversation(id, {
      signal: controller.signal,
    });
    if (controller.signal.aborted) {
      return { detail: null, aborted: true };
    }
    setters.setConversation(detail.conversation);
    setters.setMessages(detail.messages);
    if (!opts?.silent) {
      setters.setLatestWorkoutDraftMessageId(null);
    }
    if (!detail.active_run) {
      clearResumeCursor(id);
    }
    setters.setLoadError(null);
    return { detail, aborted: false, error: undefined };
  } catch (error) {
    if (controller.signal.aborted || isAbortError(error)) {
      return { detail: null, aborted: true, error };
    }
    setters.setLoadError(getErrorMessage(error));
    return { detail: null, aborted: false, error };
  } finally {
    if (refs.loadAbortRef.current === controller) {
      refs.loadAbortRef.current = null;
      if (!opts?.silent) {
        setters.setIsLoadingConversation(false);
      }
    }
  }
}

export async function recoverConversation(
  id: number,
  opts: ConversationRequestOptions | undefined,
  { refs, setters }: SessionStateContext,
): Promise<ConversationRequestResult> {
  refs.recoveryAbortRef.current?.abort();
  const controller = new AbortController();
  refs.recoveryAbortRef.current = controller;
  let recoveryRequestError: unknown = null;
  let shouldRetryRecovery = false;

  const requestRecovery = async () => {
    try {
      const response = await requestAIChatMessageRecovery(id, {
        signal: controller.signal,
      });
      shouldRetryRecovery = response.status === "not_needed";
    } catch (error) {
      if (controller.signal.aborted || isAbortError(error)) {
        throw error;
      }
      recoveryRequestError = recoveryRequestError ?? error;
      shouldRetryRecovery = false;
    }
  };

  const retryRecoveryIfNeeded = async () => {
    if (!shouldRetryRecovery) {
      return;
    }
    await requestRecovery();
  };

  try {
    await requestRecovery();

    const detail = await pollAIChatConversationUntilSettled(id, {
      signal: controller.signal,
      onStreaming: retryRecoveryIfNeeded,
    });
    if (controller.signal.aborted) {
      return { detail: null, aborted: true };
    }
    setters.setConversation(detail.conversation);
    setters.setMessages(detail.messages);
    setters.setLatestWorkoutDraftMessageId(null);
    setters.setLoadError(null);
    return { detail, aborted: false, error: undefined };
  } catch (error) {
    if (controller.signal.aborted || isAbortError(error)) {
      return { detail: null, aborted: true, error };
    }
    if (!opts?.silent) {
      setters.setLoadError(getErrorMessage(recoveryRequestError ?? error));
    }
    return {
      detail: null,
      aborted: false,
      error: recoveryRequestError ?? error,
    };
  } finally {
    if (refs.recoveryAbortRef.current === controller) {
      refs.recoveryAbortRef.current = null;
    }
  }
}

export async function resumeConversation(
  detail: AIChatConversationDetail,
  loadConversation: LoadConversation,
  { refs, setters }: SessionStateContext,
  operation: ChatSessionOperation,
): Promise<ConversationRequestResult> {
  const activeRun = detail.active_run;
  if (!activeRun) {
    return { detail, aborted: false };
  }

  refs.resumeAbortRef.current?.abort();
  const controller = new AbortController();
  refs.resumeAbortRef.current = controller;
  refs.pendingAssistantIdRef.current = activeRun.assistant_message_id;
  const afterSequence = getResumeAfterSequence(detail);

  try {
    const streamResult = await resumeAIChatMessageStream(
      detail.conversation.id,
      activeRun.id,
      afterSequence,
      {
        onStart: (event) => {
          const assistantMessageId =
            event.message_id ?? activeRun.assistant_message_id;
          refs.pendingAssistantIdRef.current = assistantMessageId;
          if (event.sequence !== undefined) {
            saveResumeCursor(detail.conversation.id, {
              runId: activeRun.id,
              sequence: event.sequence,
              assistantMessageId,
            });
          }
        },
        onDelta: (event) => {
          const targetId =
            refs.pendingAssistantIdRef.current ??
            activeRun.assistant_message_id;
          setters.setMessages((current) =>
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
          const targetId =
            refs.pendingAssistantIdRef.current ??
            activeRun.assistant_message_id;
          setters.setMessages((current) =>
            updateStreamingMessageWithDone(current, targetId, event),
          );
          if (event.workout_draft) {
            setters.setConversation((current) =>
              current
                ? {
                    ...current,
                    latest_workout_draft: event.workout_draft,
                    latest_workout_draft_status: { is_saved: false },
                  }
                : current,
            );
            setters.setLatestWorkoutDraftMessageId(
              event.message_id ?? targetId,
            );
          }
          clearResumeCursor(detail.conversation.id);
        },
        onErrorEvent: (event) => {
          const targetId =
            refs.pendingAssistantIdRef.current ??
            activeRun.assistant_message_id;
          setters.setMessages((current) =>
            updateStreamingMessageWithError(current, targetId, event),
          );
          clearResumeCursor(detail.conversation.id);
        },
        signal: controller.signal,
      },
    );
    if (controller.signal.aborted) {
      return { detail: null, aborted: true };
    }

    const refreshed = await loadConversation(detail.conversation.id, {
      silent: true,
    });
    if (!streamResult.endedWithError) {
      return refreshed;
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
    };
  } catch (error) {
    if (controller.signal.aborted || isAbortError(error)) {
      return { detail: null, aborted: true, error };
    }
    return { detail: null, aborted: false, error };
  } finally {
    if (refs.resumeAbortRef.current === controller) {
      refs.resumeAbortRef.current = null;
    }
    if (refs.activeOperationRef.current === operation) {
      refs.pendingAssistantIdRef.current = null;
    }
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
  setLoadError: ChatSessionSetters["setLoadError"];
}) {
  const recoveryResult = await recoverConversation(id, { silent: true });
  const recoveryOutcome = classifyRecoveryOutcome({
    messages: recoveryResult.detail?.messages,
    aborted: recoveryResult.aborted,
    error: recoveryResult.error,
  });
  recordTelemetry({
    category: "recovery",
    outcome: recoveryOutcome,
  });

  if (
    recoveryOutcome === "recovered_completed" ||
    recoveryOutcome === "recovery_aborted"
  ) {
    return;
  }

  const recoveryError =
    recoveryResult.error ?? new Error("Failed to recover AI chat conversation");
  recordTelemetry({
    category: "ux",
    outcome: "failure_toast_shown",
  });
  if (!recoveryResult.detail) {
    setLoadError(
      getErrorMessage(recoveryError, "Failed to recover AI chat conversation"),
    );
  }
  showErrorToast(recoveryError, "Failed to recover AI chat conversation");
}
