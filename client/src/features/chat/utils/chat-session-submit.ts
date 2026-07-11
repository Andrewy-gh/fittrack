import {
  createAIChatConversation,
  streamAIChatMessage,
  type AIChatMessage,
} from "@/features/chat/api/ai-chat";
import {
  classifyRecoveryOutcome,
  classifyStreamInterruption,
  terminalStreamStage,
} from "@/features/chat/utils/ai-chat-observability";
import { getErrorMessage, showErrorToast } from "@/lib/errors";
import {
  clearResumeCursor,
  saveResumeCursor,
  updateStreamingMessageWithDelta,
  updateStreamingMessageWithDone,
  updateStreamingMessageWithError,
} from "./chat-resume";
import type {
  ChatSessionRefs,
  ChatSessionSetters,
  RecordChatTelemetry,
} from "./chat-session-types";
import type {
  LoadConversation,
  RecoverConversation,
} from "./chat-session-recovery";

type SubmitPromptOptions = {
  conversationId: number | null;
  prompt: string;
  isSubmitting: boolean;
  onConversationCreated: (conversationId: number) => Promise<void>;
  onPromptStarted: (conversationId: number) => void;
  onNewConversationCreated: (conversationId: number) => void;
  loadConversation: LoadConversation;
  recoverConversation: RecoverConversation;
  recordTelemetry: RecordChatTelemetry;
  refs: ChatSessionRefs;
  setters: ChatSessionSetters;
};

export async function submitPrompt({
  conversationId,
  prompt,
  isSubmitting,
  onConversationCreated,
  onPromptStarted,
  onNewConversationCreated,
  loadConversation,
  recoverConversation,
  recordTelemetry,
  refs,
  setters,
}: SubmitPromptOptions) {
  const nextPrompt = prompt.trim();
  if (!nextPrompt || isSubmitting) {
    return;
  }

  setters.setIsSubmitting(true);
  setters.setPrompt("");

  let activeConversationId = conversationId;

  try {
    if (!activeConversationId) {
      const createdConversation = await createAIChatConversation();
      activeConversationId = createdConversation.id;
      onNewConversationCreated(activeConversationId);
      setters.setConversation(createdConversation);
      await onConversationCreated(activeConversationId);
    }
  } catch (error) {
    setters.setPrompt(nextPrompt);
    setters.setIsSubmitting(false);
    showErrorToast(error, "Failed to create chat conversation");
    return;
  }

  if (!activeConversationId) {
    setters.setPrompt(nextPrompt);
    setters.setIsSubmitting(false);
    return;
  }

  const baseTimestamp = new Date().toISOString();
  const tempUserId = -Date.now();
  const tempAssistantId = tempUserId - 1;
  let streamStarted = false;
  let shouldRefreshConversation = false;
  const streamController = new AbortController();
  refs.pendingAssistantIdRef.current = tempAssistantId;
  refs.streamAbortRef.current = streamController;
  clearResumeCursor(activeConversationId);

  setters.setMessages((current) => [
    ...current,
    {
      id: tempUserId,
      conversation_id: activeConversationId,
      role: "user",
      content: nextPrompt,
      status: "completed",
      created_at: baseTimestamp,
      updated_at: baseTimestamp,
      completed_at: baseTimestamp,
    },
    {
      id: tempAssistantId,
      conversation_id: activeConversationId,
      role: "assistant",
      content: "",
      status: "streaming",
      created_at: baseTimestamp,
      updated_at: baseTimestamp,
    },
  ]);

  try {
    const streamResult = await streamAIChatMessage(
      activeConversationId,
      nextPrompt,
      {
        onStart: (event) => {
          streamStarted = true;
          onPromptStarted(activeConversationId);
          const assistantMessageId = event.message_id ?? tempAssistantId;
          refs.pendingAssistantIdRef.current = assistantMessageId;
          setters.setMessages((current) =>
            current.map((message) =>
              message.id === tempAssistantId
                ? { ...message, id: assistantMessageId }
                : message,
            ),
          );
          saveResumeCursor(activeConversationId, {
            runId: event.run_id ?? 0,
            sequence: event.sequence ?? 0,
            assistantMessageId,
          });
        },
        onDelta: (event) => {
          const targetId =
            refs.pendingAssistantIdRef.current ?? tempAssistantId;
          setters.setMessages((current) =>
            updateStreamingMessageWithDelta(
              current,
              targetId,
              event.delta ?? "",
            ),
          );
          if (event.sequence !== undefined) {
            saveResumeCursor(activeConversationId, {
              runId: event.run_id ?? 0,
              sequence: event.sequence,
              assistantMessageId: targetId,
            });
          }
        },
        onDone: (event) => {
          const targetId =
            refs.pendingAssistantIdRef.current ?? tempAssistantId;
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
          clearResumeCursor(activeConversationId);
        },
        onErrorEvent: (event) => {
          const targetId =
            refs.pendingAssistantIdRef.current ?? tempAssistantId;
          setters.setMessages((current) =>
            updateStreamingMessageWithError(current, targetId, event),
          );
          clearResumeCursor(activeConversationId);
          showErrorToast(
            { message: event.message ?? "AI chat streaming failed" },
            "AI chat streaming failed",
          );
        },
        signal: streamController.signal,
      },
    );

    recordTelemetry({
      category: "stream",
      outcome: streamResult.endedWithError ? "server_error" : "completed",
      stage: terminalStreamStage(),
    });
    if (streamResult.endedWithError) {
      recordTelemetry({
        category: "ux",
        outcome: "failure_toast_shown",
      });
    }
    shouldRefreshConversation = true;
  } catch (error) {
    if (!streamStarted && isPreflightAPIError(error)) {
      removeOptimisticMessages(setters, tempUserId, tempAssistantId);
      setters.setPrompt(nextPrompt);
      recordTelemetry({
        category: "stream",
        outcome: "server_error",
        stage: "pre_start",
      });
      recordTelemetry({
        category: "ux",
        outcome: "failure_toast_shown",
      });
      showErrorToast(error, "Failed to stream AI chat response");
      return;
    }

    const streamTelemetry = classifyStreamInterruption(error, streamStarted);
    recordTelemetry({
      category: "stream",
      outcome: streamTelemetry.outcome,
      stage: streamTelemetry.stage,
    });

    if (streamTelemetry.outcome === "client_aborted") {
      return;
    }

    await handleStreamFailureRecovery({
      activeConversationId,
      nextPrompt,
      streamStarted,
      tempAssistantId,
      error,
      recoverConversation,
      recordTelemetry,
      refs,
      setters,
    });
  } finally {
    if (refs.streamAbortRef.current === streamController) {
      refs.streamAbortRef.current = null;
    }
    refs.pendingAssistantIdRef.current = null;
    setters.setIsSubmitting(false);
  }

  if (shouldRefreshConversation) {
    await loadConversation(activeConversationId, { silent: true });
  }
}

async function handleStreamFailureRecovery({
  activeConversationId,
  nextPrompt,
  streamStarted,
  tempAssistantId,
  error,
  recoverConversation,
  recordTelemetry,
  refs,
  setters,
}: {
  activeConversationId: number;
  nextPrompt: string;
  streamStarted: boolean;
  tempAssistantId: number;
  error: unknown;
  recoverConversation: RecoverConversation;
  recordTelemetry: RecordChatTelemetry;
  refs: ChatSessionRefs;
  setters: ChatSessionSetters;
}) {
  const {
    detail: recoveredDetail,
    aborted: recoveryAborted,
    error: recoveryError,
  } = await recoverConversation(activeConversationId, { silent: true });
  const submitFailure = recoveryError ?? error;
  const recoveryOutcome = classifyRecoveryOutcome({
    messages: recoveredDetail?.messages,
    prompt: nextPrompt,
    aborted: recoveryAborted,
    error: recoveryError,
  });
  recordTelemetry({
    category: "recovery",
    outcome: recoveryOutcome,
  });
  if (recoveryAborted) {
    return;
  }

  const recoveredPromptStatus = findRecoveredPromptStatus(
    recoveredDetail?.messages ?? [],
    nextPrompt,
  );

  if (!recoveredDetail) {
    const targetId = streamStarted
      ? (refs.pendingAssistantIdRef.current ?? tempAssistantId)
      : tempAssistantId;
    markAssistantMessageFailed(setters, targetId, submitFailure);
  }

  if (recoveredPromptStatus !== "completed") {
    recordTelemetry({
      category: "ux",
      outcome: "failure_toast_shown",
    });
    showErrorToast(submitFailure, "Failed to stream AI chat response");
    return;
  }

  recordTelemetry({
    category: "ux",
    outcome: "failure_toast_suppressed_due_to_successful_recovery",
  });
}

function removeOptimisticMessages(
  setters: ChatSessionSetters,
  tempUserId: number,
  tempAssistantId: number,
) {
  setters.setMessages((current) =>
    current.filter(
      (message) => message.id !== tempUserId && message.id !== tempAssistantId,
    ),
  );
}

function markAssistantMessageFailed(
  setters: ChatSessionSetters,
  messageId: number,
  error: unknown,
) {
  setters.setMessages((current) =>
    current.map((message) =>
      message.id === messageId
        ? {
            ...message,
            status: "failed",
            error_message: getErrorMessage(error),
          }
        : message,
    ),
  );
}

function isPreflightAPIError(error: unknown): error is { message: string } {
  return (
    !(error instanceof Error) &&
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof (error as { message?: unknown }).message === "string"
  );
}

function findRecoveredPromptStatus(
  messages: AIChatMessage[],
  prompt: string,
): AIChatMessage["status"] | null {
  const normalizedPrompt = prompt.trim();
  if (!normalizedPrompt) {
    return null;
  }

  for (let index = messages.length - 1; index >= 0; index -= 1) {
    const message = messages[index];
    if (
      message.role !== "user" ||
      message.content.trim() !== normalizedPrompt
    ) {
      continue;
    }

    const assistant = messages[index + 1];
    if (assistant?.role === "assistant") {
      return assistant.status;
    }

    return null;
  }

  return null;
}
